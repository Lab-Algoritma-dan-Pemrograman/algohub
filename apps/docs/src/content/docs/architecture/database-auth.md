---
title: Basis Data & Keamanan Autentikasi
description: Arsitektur Supabase PostgreSQL pooler 6543, sistem autentikasi multi-token JWT, penegakan RLS di usecase Go, dan strategi rate limiting.
---

# Basis Data & Keamanan Autentikasi

AlgoHub mengoperasikan basis data terpusat **Supabase PostgreSQL** yang melayani empat frontend berbeda, dengan lapisan keamanan autentikasi multi-token JWT dan penegakan *Row Level Security* (RLS) pada tingkat kode usecase backend Go.

---

## 1. Arsitektur Basis Data Supabase PostgreSQL

Seluruh modul (Praktikum, SIAKAD, dan E-Learning) terhubung ke basis data terpadu (`oidwvfcxlxfyojkyurmm`).

```
                    +------------------------------------+
                    |        Backend Go Serverless       |
                    |          (api.algohub.web.id)      |
                    +-----------------+------------------+
                                      |
                                      | Transaction Pooler
                                      | Port 6543 (pgx driver)
                                      v
                    +------------------------------------+
                    |      Supabase Supavisor Pooler     |
                    +-----------------+------------------+
                                      |
                                      | Internal Unix Socket
                                      v
                    +------------------------------------+
                    |     PostgreSQL Core Database       |
                    |  (Schemas: public, auth, storage)  |
                    +------------------------------------+
```

### Konfigurasi Connection Pooler (Port 6543)
Karena backend berjalan di lingkungan *serverless* Vercel Fluid Compute yang dapat melakukan scaling horizontal secara dinamis, koneksi ke PostgreSQL tidak dibuka langsung ke port 5432 (yang memiliki batas koneksi ketat), melainkan melalui **Supavisor Transaction Pooler** pada port `6543`.

Parameter koneksi GORM diatur secara spesifik:
- **`PreferSimpleProtocol: true`**: **Wajib diaktifkan**. Pada transaction pooler, koneksi fisik berganti-ganti antar-query, sehingga *prepared statement* biner PostgreSQL bawaan akan memicu tabrakan identifikasi statement. Mode simple protocol mengirim query sebagai teks mentah berparameter.
- **Ukuran Pool Serverless**:
  - `MaxOpenConns`: 5 (membatasi penggunaan koneksi per instance function)
  - `MaxIdleConns`: 2
  - `ConnMaxIdleTime`: 1 menit
  - `ConnMaxLifetime`: 5 menit
- **Larangan AutoMigrate di Produksi**: `AutoMigrate` dan seed otomatis dinonaktifkan sepenuhnya di lingkungan produksi untuk mencegah *table lock* yang dapat mengganggu jalannya ujian mahasiswa.

---

### Dua Bidang Supabase: HTTP API vs PostgreSQL

Supabase mengekspos dua bidang yang berbeda, dan dampak gangguannya juga berbeda:

| Bidang | Dipakai untuk | Dampak bila 402 |
|---|---|---|
| HTTP API (Auth, REST, Storage) | OTP, reset password, berkas lama | Gagal, tetapi tidak memblokir ujian |
| PostgreSQL pooler port 6543 | Data inti: jawaban, nilai, absensi, user, konfigurasi | Tetap normal |

Sejak insiden 2026-11-24, backend tidak bergantung pada REST/Storage Supabase untuk data inti:

- Jawaban, nilai, dan absensi ditulis langsung ke PostgreSQL via GORM (`DATABASE_URL_*`).
- Upload berkas memakai `pkg/objectstore`: Vercel Blob (publik/privat) sebagai primary, Supabase Storage sebagai fallback.
- Register tetap berjalan saat Auth membalas 402 dengan bcrypt lokal; akun Supabase/OTP disinkronkan setelah kuota pulih.
- Foto lama yang masih menunjuk ke Supabase Storage tidak dihapus; UI memakai `ui-avatars` sampai berkas bisa diakses lagi.

Operator dapat memverifikasi jalur tulis tanpa mengubah isi tabel:

```sql
BEGIN;
UPDATE jawaban_mahasiswa SET updated_at = updated_at
WHERE id = (SELECT id FROM jawaban_mahasiswa ORDER BY id DESC LIMIT 1);
ROLLBACK;
```

---

## 2. Sistem Autentikasi Multi-Token JWT

Masing-masing subsistem AlgoHub memiliki karakteristik sesi yang disesuaikan dengan kebutuhan arsitekturnya:

| Modul | Mekanisme Sesi | Secret Key Terkait | Bentuk Klaim JWT Utama |
|---|---|---|---|
| **Praktikum** | Header `Authorization: Bearer <token>` di `localStorage` | `JWT_SECRET` (HS256) | `{ uid, nim, role, exp }` |
| **SIAKAD** | Cookie HTTP-Only `siakad_token` di `.algohub.web.id` + Bearer | `SIAKAD_JWT_SECRET` (HS256) | `{ username, full_name, role, exp }` |
| **E-Learning** | Header `Authorization: Bearer <token>` | `ELEARNING_SUPABASE_JWT_SECRET` (HS256) | `{ nim, sub, role: "authenticated", user_role }` |

### Mode Autentikasi Backend (`AUTH_MODE`)
Backend Go mendukung tiga mode validasi token fleksibel melalui variabel `AUTH_MODE`:
1. `legacy`: Hanya memvalidasi token JWT Go bawaan.
2. `supabase`: Memvalidasi token Supabase RS256/HS256 via JWKS.
3. `dual`: **Mode Produksi Aktif**. Backend secara cerdas mengenali format token yang masuk (apakah token Praktikum, cookie SIAKAD, atau token Supabase E-Learning) dan memvalidasinya menggunakan parser yang sesuai.

### Normalisasi Peran Terpusat (`pkg/jwt/roles.go`)
Untuk mengatasi inkonsistensi penamaan peran antar-sistem warisan, fungsi `NormalizeRole()` menyeragamkan nilai peran menjadi empat standar kanonik:
- `mahasiswa`, `user`, `student` $\longrightarrow$ `praktikan`
- `kordas`, `coordinator`, `super_admin` $\longrightarrow$ `koordinator`
- `admin`, `laboran`, `assistant` $\longrightarrow$ `asisten`
- `penyewa` $\longrightarrow$ `penyewa`

### Kompatibilitas Hash Kata Sandi (Bcrypt & Firebase Scrypt)
Mahasiswa lama yang terdaftar sejak era sistem berbasis Firebase diautentikasi secara transparan menggunakan verifikator `FB_SCRYPT_*` (`pkg/hash/scrypt.go`). Saat pengguna memperbarui kata sandi, sistem secara otomatis mengonversi hash ke algoritma standar modern **Bcrypt**.

---

## 3. Penegakan Row Level Security (RLS) di Usecase Go

Backend Go terhubung ke database Supabase menggunakan peran `postgres` (pemilik basis data). Pada arsitektur PostgreSQL, peran `postgres` secara otomatis **melewati (*bypass*) seluruh kebijakan RLS tabel**.

Oleh karena itu, **seluruh 56 kebijakan RLS aktif ditiru dan ditegakkan secara ketat pada layer aplikasi (`usecase/elearning`)**:

```
Client Request -> Router -> Middleware Auth (Ekstrak NIM & Role) -> Usecase Check:
  |
  +-- Apakah mahasiswa mengakses data orang lain?
  |     -> Ya: Kembalikan data tersaring / array kosong (Semantik RLS SELECT)
  |
  +-- Apakah praktikan mencoba INSERT ke tabel kurikulum/soal?
  |     -> Ya: Tolak langsung dengan HTTP 403 "violates row-level security"
  |
  +-- Apakah praktikan mengupdate row milik user lain?
  |     -> Ya: Update 0 baris, respons HTTP 200 (Semantik RLS UPDATE)
  |
  v Lolos Validasi
Panggil Repository -> Eksekusi SQL ke PostgreSQL
```

### Penanganan Khusus View `student_lessons`
Pada E-Learning, tabel `lessons` menyimpan kunci jawaban kuis dan solusi kode. Mahasiswa praktikan tidak pernah diizinkan membaca tabel `lessons` langsung; usecase secara otomatis mengarahkan query mahasiswa ke view `student_lessons` yang menyembunyikan kolom `correct_answer`, `solution`, dan `test_cases_hidden`.

---

## 4. Mekanisme Rate Limiting

Untuk melindungi server dari serangan brute-force login dan *denial of service*:
- **Pendeteksian IP Klien Asli**: Middleware `internal/middleware/client_ip.go` mengekstrak IP riil pengguna dari entri pertama header `X-Forwarded-For` atau `X-Real-IP` yang diverifikasi oleh Vercel Edge.
- **Batas Kuota**:
  - `POST /api/praktikum/auth/login`: Maksimal 10 percobaan/menit/IP.
  - `POST /api/praktikum/auth/cek-nim`: Maksimal 20 request/menit/IP.
  - `POST /api/praktikum/run` (Runner Glot fallback): Maksimal 30 request/menit/user ID.
  - `POST /api/siakad/auth/login`: Maksimal 5 percobaan/menit/IP.
- **Respons 429**: Permintaan yang melampaui batas dibalas dengan HTTP status `429 Too Many Requests` disertai header `Retry-After: <detik>`. Header CORS tetap disertakan pada respons 429 agar peramban dapat membaca pesan penolakan dengan jelas.
