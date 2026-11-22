---
title: Pengenalan AlgoHub
description: Platform terpadu laboratorium algoritma dan pemrograman untuk pembelajaran, praktikum, dan administrasi akademik.
---

# Pengenalan AlgoHub

**AlgoHub** adalah platform ekosistem digital terintegrasi yang dirancang khusus untuk mendukung seluruh aktivitas operasional pada **Laboratorium Algoritma & Pemrograman**. Platform ini menggabungkan kegiatan perkuliahan praktikum interaktif, pembelajaran mandiri tergamifikasi, presensi digital, hingga tata kelola inventaris dan jadwal akademik ke dalam satu kesatuan sistem modern.

---

## Visi & Misi

- **Visi**: Menjadi pusat keunggulan laboratorium komputasi yang menyediakan lingkungan praktikum interaktif, otomatis, dan berdaya guna tinggi dalam membentuk kompetensi pemrograman mahasiswa.
- **Misi**:
  1. Menyediakan sarana praktikum pemrograman bahasa C dan Python yang efisien tanpa hambatan konfigurasi lingkungan lokal mahasiswa.
  2. Meningkatkan motivasi dan pemahaman logika algoritma melalui pembelajaran bertingkat (gamifikasi kurikulum dan tantangan Bug Hunt).
  3. Memodernisasi administrasi laboratorium melalui presensi dinamis, penjadwalan presisi, integrasi notifikasi instan, dan evaluasi berbasis data.
  4. Menerapkan arsitektur perangkat lunak yang andal, aman, hemat sumber daya, dan mudah dirawat.

---

## Ekosistem Arsitektur AlgoHub

AlgoHub dibangun dengan arsitektur **Monorepo** yang memisahkan aplikasi ke dalam 4 modul antarmuka web (Frontend SPA SvelteKit) dan 1 layanan pusat API (Backend Go Gin), didukung basis data terpusat Supabase PostgreSQL.

```
                    +---------------------------+
                    |       Cloudflare DNS      |
                    | (DNS-Only / Gray Cloud)   |
                    +-------------+-------------+
                                  |
                +-----------------+-----------------+
                |                                   |
                v                                   v
   +-------------------------+         +-------------------------+
   |   4 Frontend SPA        |         |   1 Central Backend     |
   |   (SvelteKit Static)    |         |   (Go Gin Serverless)   |
   |   - Portal              |         |   - api.algohub.web.id  |
   |   - SIAKAD              |         +------------+------------+
   |   - E-Learning          |                      |
   |   - Praktikum           |                      | Transaction Pooler
   +------------+------------+                      | (Port 6543)
                |                                   v
                | CORS Fetch           +-------------------------+
                +--------------------->|   Supabase PostgreSQL   |
                                       |   (Database Bersama)    |
                                       +-------------------------+
```

### 1. Portal (`algohub.web.id`)
Landing page publik dan gerbang utama AlgoHub. Selain memperkenalkan profil laboratorium dan tautan ke seluruh layanan, Portal berfungsi sebagai server distribusi aset biner compiler C berbasis WebAssembly (`/compiler/v1/`) serta menangani pengalihan rute permanen (301 redirect) dari pranala praktikum lama.

### 2. SIAKAD (`siakad.algohub.web.id`)
Sistem Informasi Akademik Laboratorium yang mengelola seluruh administrasi laboratorium:
- Presensi mahasiswa dan asisten berbasis Dynamic QR Code yang berubah secara berkala.
- Penjadwalan jaga laboratorium dan mekanisme tukar giliran (swap shift) antar asisten.
- Pengelolaan peminjaman inventaris dan aset peralatan laboratorium.
- Notifikasi push peramban (Web Push VAPID) dan pesan WhatsApp gateway.
- Integrasi ke database melalui arsitektur RPC Gateway Postgres yang terlindungi.

### 3. E-Learning (`elearning.algohub.web.id`)
Platform pembelajaran mandiri interaktif dengan pendekatan gamifikasi:
- Kurikulum terstruktur berjenjang: **Level** &rarr; **Modul** &rarr; **Lesson**.
- Runner kode terisolasi di peramban (Clang WebAssembly & Pyodide Python).
- Sistem pemeliharaan streak harian dan penghitungan detik belajar aktif (*study seconds*).
- Gamifikasi XP, leaderboard mingguan, tingkatan level, dan mekanisme *Prestige Reset*.
- Permainan interaktif logika kode *Bug Hunt*.
- Otentikasi Seamless Single Sign-On (SSO) dari akun SIAKAD.

### 4. Praktikum (`praktikum.algohub.web.id`)
Sistem inti pelaksanaan sesi praktikum resmi di ruang laboratorium:
- Manajemen alur ujian berbatas waktu: Pretest, Posttest, Keterampilan, dan Ujian Utama.
- Fitur autosave draft jawaban dan kode ke backend dan penyimpanan lokal IndexedDB.
- Lingkungan eksekusi kode C di sisi peramban menggunakan Web Worker dan SharedArrayBuffer tanpa membebani server.
- Pengunggahan dan pengelolaan Laporan Besar terintegrasi Supabase Storage.
- Penilaian terstruktur manual oleh asisten dosen dan bantuan evaluasi rubrik AI (AI-assisted grading).

### 5. Backend Go API (`api.algohub.web.id`)
Satu-satunya backend layanan RESTful yang melayani seluruh frontend:
- Dibangun dengan Go 1.26 dan Gin framework, berjalan di Vercel Fluid Compute.
- Menangani autentikasi dual-token JWT, rate limiting per IP & User, dan penegakan logika RLS (Row Level Security) aplikasi.

---

## Peta Subdomain Resmi

Seluruh layanan AlgoHub berjalan di bawah domain utama `algohub.web.id`:

| Subdomain | Aplikasi / Layanan | Teknologi Utama | Akses Publik / Hak Akses |
|---|---|---|---|
| `algohub.web.id` | Portal & Landing Page | SvelteKit (Prerendered HTML) | Publik Bebas |
| `siakad.algohub.web.id` | SIAKAD Laboratorium | SvelteKit SPA + RPC Gateway | Mahasiswa, Asisten, Koordinator |
| `elearning.algohub.web.id` | E-Learning Gamifikasi | SvelteKit SPA + PWA | Mahasiswa, Asisten, Koordinator |
| `praktikum.algohub.web.id` | Sistem Praktikum Lab | SvelteKit SPA + Web Worker | Mahasiswa, Asisten, Koordinator |
| `api.algohub.web.id` | Central Backend API | Go Gin (Serverless Function) | Terproteksi Token / HMAC Cron |
| `docs.algohub.web.id` | Dokumentasi Resmi | Astro Starlight | Publik Bebas |

---

## Peran Pengguna (Role System)

AlgoHub menerapkan sistem hak akses berjenjang:
1. **Praktikan / Mahasiswa**: Menjalankan sesi praktikum, belajar di e-learning, melakukan absensi QR, dan mengunggah laporan.
2. **Penyewa**: Pengguna luar atau mahasiswa yang memiliki izin peminjaman alat lab di SIAKAD.
3. **Asisten Laboratorium**: Mengaktivasi sesi praktikum, memonitor status mahasiswa secara realtime, menilai tugas/laporan, mengatur jadwal jaga lab, dan memvalidasi presensi.
4. **Koordinator Asisten (Kordas)**: Memiliki hak manajerial penuh atas kurikulum, konfigurasi penilaian, manajemen pengguna lab, dan persetujuan jadwal.
5. **Superadmin**: Mengelola perizinan sistem mendalam dan parameter global platform.
