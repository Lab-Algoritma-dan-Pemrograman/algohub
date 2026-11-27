# Update: migrasi aset publik ke Cloudflare R2 + perbaikan data mahasiswa terhapus

Tanggal: 2026-11-27
Status: sebagian besar selesai, deploy live. Dua hal dikerjakan bertahap.

## Latar belakang

Dua laporan dari mahasiswa dan admin:

1. **Web terasa lambat saat mengakses, terutama download modul.**
2. **Data yang sudah diisi admin tiba-tiba hilang sebagian atau seluruhnya** — nama, kelas, shift, asisten, sesi.

## Penyebab data hilang (akar masalah)

### Jalur #1 — login sinkronisasi profil (paling kerasa)

`refresh()` di `backend/internal/usecase/elearning/profile_usecase.go` mengirim `updates["kelas"] = ""` setiap login ketika token tidak membawa kelas. Trigger DB `users_sync_kelas_jurusan` (BEFORE UPDATE) lalu masuk cabang `NEW.kelas IS DISTINCT FROM OLD.kelas`, menimpa `class_code` dengan string kosong, dan men-NULL-kan `kelas_id`. Trigger `users_sync_group_members` (AFTER UPDATE) menghapus baris `group_members` yang tidak cocok dengan schedule — asisten ikut lepas tanpa kembali otomatis.

Bukti: dari 28 baris yang kelasnya kosong, **27** memiliki `|updated_at - last_login_at| < 10 ms`. Artinya: dua tulisan dalam satu alur login (last_login_at oleh praktikum, updated_at oleh sync profil), berjarak ~3 ms.

### Jalur #2 — admin edit parsial diperlakukan sebagai PUT

`UpdateMahasiswa` memakai `Save(u)` full-struct. `dto.UserRequest` punya `KelasID/Shift/Gelombang/Kelompok` sebagai pointer opsional, tetapi `Save()` menulis SEMUA kolom termasuk zero-value. Form admin yang mengirim payload parsial (mis. hanya ubah nama) mengosongkan kolom lainnya.

### Jalur #3 — impor CSV dengan kolom kosong

`BulkUpsert` memakai `DoUpdates: AssignmentColumns([nama, kelas_id, shift, gelombang, kelompok])`. Kolom yang kosong di CSV menimpa nilai lama menjadi NULL.

### Yang TIDAK terbukti hilang

`kelompok` (398/399 NULL), `gelombang` (231 NULL), `nomor_hp` (240 NULL), `email` (335 NULL). Backup lama (`users_backup_prestige`, `_bak_users_20260929_shiftfix`) juga sudah NULL untuk kolom ini, jadi itu baseline yang memang tak pernah diisi, bukan data yang hilang. Tidak ada trigger yang men-NULL-kan kolom-kolom tersebut.

## Perbaikan yang diterapkan

### Backend (Go)

| File | Perubahan |
|---|---|
| `usecase/elearning/profile_usecase.go` | Guard `strings.TrimSpace(p.Kelas) != ""` — kelas kosong tidak pernah dikirim ke DB. |
| `usecase/praktikum/user_usecase.go` | `UpdateMahasiswa` memakai `BulkUpdateAttrs` (map atribut); hanya kolom yang pointer-nya non-nil yang ditimpa. |
| `repository/praktikum/user_repository.go` | `BulkUpsert` memakai `COALESCE(EXCLUDED.kolom, users.kolom)` untuk `kelas_id/shift/gelombang/kelompok`. |

Negative check: dengan kode lama, test baru `TestUpdateMahasiswa_PartialRequest_DoesNotClearOptionalFields` FAIL — mock membuktikan `Update(*entity.User)` menerima `KelasID/Shift/Gelombang/Kelompok = nil`.

### Migrasi aset ke Cloudflare R2

| Langkah | Bukti |
|---|---|
| Bucket `public-assets` (APAC) + `laporan-besar` | Dibuat; domain publik `assets.algohub.web.id` aktif dengan SSL |
| 45 gambar soal + ujian | MD5 lokal = ETag R2 semua |
| 3 PDF modul + 8 pedoman + 5 foto asisten | MD5 lokal = ETag R2 semua |
| Download modul live | 200, `application/pdf`, 1.79 MB |
| CSP `img-src` | + `https://assets.algohub.web.id` |
| DB: 74 `soal.gambar_url` | 74 ke R2, 0 Supabase tersisa |
| DB: `konfigurasi`, `pedoman_laporan`, `users.foto_url` | 0 referensi `supabase.co` |

Sumber byte: `E:\tmp\legacy-storage-backup-20261115\public-assets\` (92 file). Backup `_bak_konfigurasi_20261127`, `_bak_pedoman_laporan_20261127`, `soal_gambar_url_backup_20261127`.

### Re-backfill data kelas

Excel `Penilaian Praktikum (2).xlsx` (303 baris valid). Empat re-backfill (v1–v4) karena bug login terus mengikis sebelum backend fix live.

Setelah v4: seluruh 303 NIM Excel punya `kelas_id` + `kelas` yang benar. Sisa 9 baris `kelas_id IS NULL` hanya mahasiswa TE AD/AE (di luar Excel) + akun uji. Backup: `_bak_users_20261127_kelasfix4` (377 baris).

## Yang belum dikerjakan (perlu keputusan)

### Backend upload handler masih menulis ke Supabase Storage

`pkg/supabase/supabase.go` `Upload` + `PublicURL` membentuk URL Supabase. Admin mengunggah modul/pedoman/foto berikutnya akan menyimpan URL Supabase lagi. Cutover writer ke R2 belum dilakukan.

### Cold start backend 3.4s

Vercel Function `sin1`. Ping keep-warm cron tiap menit memakai ~720% kuota function Hobby — **tidak diterapkan**. Alternatif migrasi (Rust/CF Workers, Fly.io, Railway) tertunda.

### Kasus C: NIM 202515015

RISVA TAUFIQ ZAKI. Excel TE E, DB TE A (`kelas_id=2`). Sengaja tidak diubah, menunggu keputusan.

### `users.kelompok` kosong

376/377 NULL. Tidak ada sumber di production. Perlu input manual.