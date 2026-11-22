---
title: Aplikasi E-Learning
description: Platform gamifikasi pembelajaran pemrograman, kurikulum berjenjang, runner kode WebAssembly & Pyodide, streak belajar, dan SSO SIAKAD.
---

# Aplikasi E-Learning (`apps/elearning`)

**E-Learning AlgoHub** (`elearning.algohub.web.id`) adalah platform pembelajaran mandiri interaktif yang dirancang untuk memperkuat kemampuan logika algoritma dan pemrograman mahasiswa di luar jam praktikum reguler melalui pendekatan gamifikasi.

---

## 1. Struktur Kurikulum & Format Pembelajaran

Materi pembelajaran di E-Learning disusun dalam hierarki tiga tingkat:

$$\text{Level} \longrightarrow \text{Modul} \longrightarrow \text{Lesson}$$

1. **Level**: Tingkatan kemahiran (misalnya: *Dasar-Dasar C*, *Percabangan & Perulangan*, *Array & String*, *Fungsi & Pointer*, *Struktur Data*).
2. **Modul**: Pengelompokan topik tematik di dalam satu level.
3. **Lesson**: Unit pelajaran individual yang dapat diselesaikan oleh mahasiswa.

### Tipe Konten Lesson
Setiap lesson dapat berisi satu atau kombinasi dari metode interaktif:
- **Materi Rich Text / Markdown**: Penjelasan konsep dengan sintaks terformat, diagram visual, dan tabel.
- **Kuis Pilihan Ganda (Multiple Choice Quiz)**: Soal evaluasi pemahaman konsep dasar dengan kunci jawaban terenkripsi di server.
- **Flowchart Puzzle**: Latihan menyusun alur logika bagan alir algoritma secara visual sebelum menuliskan implementasi kode.
- **Latihan Pemrograman (Code Exercise)**: Editor kode Monaco terintegrasi dengan instruksi spesifik dan pengujian *test case* otomatis.

---

## 2. In-Browser Code Runner (Clang WASM & Pyodide)

E-Learning mengeksekusi kode mahasiswa langsung di dalam peramban peramban (*client-side execution*) menggunakan package `@algohub/runner`:

```
  +-----------------------------------------------------------+
  | Halaman Lesson / Playground (Main UI Thread)              |
  | Monaco Code Editor -> Tombol "Jalankan Kode"              |
  +-----------------------------+-----------------------------+
                                | postMessage()
                                v
  +-----------------------------------------------------------+
  | Web Worker Terisolasi (@algohub/runner)                   |
  | - Bahasa C      : Clang WebAssembly + memfs virtual POSIX  |
  | - Bahasa Python : Pyodide (CPython WebAssembly Runtime)   |
  +-----------------------------+-----------------------------+
                                | Input/Output via SharedArrayBuffer
                                v
  +-----------------------------------------------------------+
  | xterm Terminal Display / Matplotlib Plot Canvas           |
  +-----------------------------------------------------------+
```

- **Bahasa C**: Dijalankan oleh biner Clang & LLD WebAssembly yang diunduh dari `/compiler/v1/`. Input interaktif (`scanf`) didukung melalui sinkronisasi thread `SharedArrayBuffer` dan `Atomics.wait()`.
- **Bahasa Python**: Menggunakan runtime Pyodide WebAssembly. Mendukung eksekusi library komputasi ilmiah dan visualisasi grafik **Matplotlib** yang otomatis diekstraksi ke elemen gambar web (`images[]`).
- **Isolasi Penuh**: Eksekusi program yang mengalami infinite loop (misal `while(1)`) dapat dihentikan (*force terminate*) kapan saja tanpa membuat tab peramban macet.

---

## 3. Sistem Gamifikasi & Retensi Belajar

E-Learning menerapkan mekanisme retensi psikologis untuk mendorong konsistensi belajar:

### Streak Harian (*Daily Streak*)
- Sistem mencatat timestamp aktivitas belajar terakhir mahasiswa (`last_active_date`).
- Jika mahasiswa menyelesaikan minimal satu lesson atau beraktivitas setiap hari berturut-turut, angka *streak* akan bertambah.
- Mahasiswa yang absen belajar satu hari akan kehilangan streak aktifnya.

### Penghitungan Detik Belajar (*Study Seconds*)
- Peramban melacak waktu interaksi aktif mahasiswa pada halaman pelajaran.
- Setiap interval 30 detik, akumulasi detik belajar dikirimkan ke server melalui endpoint gabungan `PUT /api/elearning/sessions/me` guna menghemat alokasi invocation serverless.

### Peringkat & Leaderboard Mingguan
- **XP (Experience Points)**: Diberikan setiap kali mahasiswa menyelesaikan lesson, kuis, atau tantangan kode.
- **Level Akun**: Dihitung secara otomatis di basis data menggunakan formula fungsi `calculate_level(xp)`.
- **Leaderboard Global & Kelas**: Menyajikan peringkat mahasiswa berdasarkan akumulasi XP total atau filter per kelas/jurusan.
- **Reset Mingguan**: Leaderboard mingguan di-reset setiap hari **Senin pukul 00:00 WIB**, memberikan kesempatan yang adil bagi seluruh mahasiswa untuk memperebutkan peringkat teratas setiap pekannya.

### Sistem Prestige Reset
Bagi mahasiswa yang telah mencapai batas level tertinggi kurikulum, sistem menyediakan fitur **Prestige Reset** sukarela melalui fungsi RPC `do_prestige_reset`:
- Mengatur ulang total XP menjadi 0.
- Memberikan lencana kehormatan eksklusif (*Prestige Badge*) pada profil mahasiswa.
- Membuka level tantangan tingkat mahir.

---

## 4. Game Logika: Bug Hunt

**Bug Hunt** adalah permainan logika mini di mana mahasiswa disajikan potongan kode pemrograman C atau Python yang memiliki satu baris kesalahan (*bug*):
- Mahasiswa harus mengidentifikasi baris yang keliru sebelum waktu habis.
- Memiliki batas kuota bermain mingguan untuk menjaga keseimbangan perolehan XP.
- Status kuota dan perizinan bermain diverifikasi secara realtime via `GET /api/elearning/game/plays/me`.

---

## 5. Single Sign-On (SSO) dari SIAKAD

Mahasiswa tidak perlu membuat akun terpisah untuk mengakses E-Learning. Sistem menggunakan alur otentikasi SSO terenkripsi:

```
  +--------------+            +-------------------+            +-----------------+
  | SIAKAD       |            | Backend Go API    |            | E-Learning      |
  | (siakad.*)   |            | (api.*)           |            | (elearning.*)   |
  +-------+------+            +---------+---------+            +--------+--------+
          |                             |                               |
          | 1. Form POST /api/sso       |                               |
          +---------------------------->|                               |
          |                             | 2. Generate Code AES-256-GCM  |
          |                             |    (TTL 60 detik)             |
          | 3. HTTP 303 Redirect        |                               |
          |<----------------------------+                               |
          |                                                             |
          | 4. Buka URL: /sso?code=<one_time_code>                      |
          +------------------------------------------------------------>|
          |                                                             |
          |                             | 5. POST /receive-token {code} |
          |                             |<------------------------------+
          |                             | 6. Sesi JWT Elearning         |
          |                             +------------------------------>|
          |                                                             |
          |                                                             | 7. Masuk Dashboard
```

1. SIAKAD mengirimkan data handshake melalui form POST ke `/api/elearning/sso`.
2. Backend Go menerbitkan kode otentikasi sekali pakai yang dienkripsi dengan algoritma AES-256-GCM (masa berlaku $\le$ 60 detik).
3. Pengguna dialihkan (HTTP 303) ke `https://elearning.algohub.web.id/sso?code=...`. Token sesi panjang tidak pernah diekspos di URL peramban.
4. Halaman `/sso` menukarkan kode tersebut melalui `POST /api/elearning/receive-token` untuk mendapatkan token sesi terotentikasi.
