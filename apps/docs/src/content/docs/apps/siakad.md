---
title: Aplikasi SIAKAD
description: Modul operasional akademik laboratorium, presensi dynamic QR, penjadwalan asisten, notifikasi push/WhatsApp, dan arsitektur RPC Gateway.
---

# Aplikasi SIAKAD (`apps/siakad`)

**SIAKAD AlgoHub** (`siakad.algohub.web.id`) adalah Sistem Informasi Akademik yang mengelola seluruh administrasi laboratorium, mulai dari absensi harian mahasiswa, penjadwalan giliran jaga asisten, manajemen inventaris perangkat lab, hingga alur pertukaran jadwal (swap shift).

---

## 1. Modul Fitur Utama

SIAKAD terdiri dari 21 halaman dengan kontrol hak akses berbasis peran (*Role-Based Access Control*):

| Modul | Rute | Izin Peran | Deskripsi |
|---|---|---|---|
| **Beranda** | `/beranda` | Semua Login | Ringkasan pengumuman lab, jadwal terdekat, dan status absensi |
| **Absensi Mahasiswa** | `/absensi` | Semua Login | Pemindai Dynamic QR Code untuk presensi masuk dan keluar praktikum |
| **Validasi Absensi** | `/validasi-absensi` | Asisten, Koordinator | Verifikasi dan persetujuan kehadiran mahasiswa secara manual/massal |
| **Buat QR Code** | `/buat-qr` | Asisten, Koordinator | Generator Dynamic QR Code presensi yang berganti token secara periodik |
| **Jadwal Jaga** | `/jadwal-jaga` | Asisten, Koordinator | Kalender shift jaga lab dan antarmuka pengajuan tukar jadwal (Swap Shift) |
| **Jadwal Saya** | `/jadwal-saya` | Semua Login | Jadwal praktikum mahasiswa atau jadwal shift asistensi spesifik |
| **Inventaris Lab** | `/inventaris` | Asisten, Koordinator | Katalog perangkat keras laboratorium dan status ketersediaannya |
| **Sewa Barang** | `/sewa-barang`, `/manajemen-sewa` | Semua Login / Staf | Peminjaman inventaris lab oleh mahasiswa dan persetujuan staf |
| **Laporan Keuangan** | `/laporan-keuangan` | Asisten, Koordinator | Pencatatan kas laboratorium dan ekspor laporan neraca ke Excel (`.xlsx`) |
| **Audit Log** | `/audit-log` | Asisten, Koordinator | Riwayat jejak audit aktivitas sensitif pengguna di laboratorium |
| **Pengaturan** | `/pengaturan` | Koordinator | Konfigurasi global sistem, template pesan WA, dan parameter lab |

---

## 2. Presensi Dynamic QR Code

Sistem presensi AlgoHub dirancang anti-kecurangan menggunakan token QR Code dinamis:
1. **Generator Token (`/buat-qr`)**: Asisten mengaktifkan sesi presensi di layar lab. Sistem menghasilkan QR code berisi payload terenkripsi yang disegarkan otomatis setiap beberapa detik (mencegah foto QR dibagikan ke mahasiswa di luar ruangan).
2. **Scanner Kamera (`/absensi`)**: Mahasiswa memindai QR code menggunakan peramban gawai (didukung pustaka `html5-qrcode`).
3. **Validasi Server**: Gateway RPC memverifikasi keabsahan waktu token dan ID mahasiswa melalui fungsi basis data `record_attendance_secure`.

---

## 3. Jadwal Jaga Asisten & Swap Shift

Untuk memfasilitasi asisten yang berhalangan hadir pada jadwal piket praktikum:
1. **Pengajuan Swap**: Asisten pemohon memilih jadwal jaga yang ingin ditukar dan menunjuk asisten pengganti melalui antarmuka `/jadwal-jaga`.
2. **Notifikasi Otomatis**: Sistem mengirimkan notifikasi peramban (Web Push) dan pesan WhatsApp ke asisten yang bersangkutan.
3. **Persetujuan (Handshake)**: Asisten pengganti menerima permohonan, dan jadwal di database diperbarui secara otomatis tanpa intervensi manual koordinator.

---

## 4. Sistem Notifikasi Multi-Channel

SIAKAD mengintegrasikan dua saluran notifikasi otomatis:

### Web Push Notifications (VAPID)
- Berjalan melalui Service Worker di latar belakang (`/service-worker.js`).
- Diimplementasikan di backend Go menggunakan pustaka `SherClockHolmes/webpush-go`.
- Mahasiswa dan asisten menerima pemberitahuan langsung di layar desktop atau ponsel cerdas saat jadwal shift tiba atau status laporan berubah.

### WhatsApp Notification Gateway
- Mengirim pesan konfirmasi kehadiran, pengingat jadwal jaga, dan permohonan swap shift ke nomor WhatsApp pengguna.
- Token gateway dan URL penyedia dikelola secara dinamis di tabel `system_settings`.
- Menggunakan sanitasi nomor telepon internasional (`+62` / `08xx`).

---

## 5. Arsitektur Postgres RPC Gateway

Frontend SIAKAD tidak melakukan query SQL atau manipulasi tabel secara langsung ke Supabase. Seluruh operasi dilewatkan melalui **RPC Gateway** pada backend Go (`POST /api/siakad/rpc?fn=<nama_fungsi>`).

```
  +-----------------------------------------------------------+
  | Frontend SIAKAD (@algohub/api-client/siakad)              |
  | api.rpc('get_attendance_records_secure', { session_id })  |
  +-----------------------------+-----------------------------+
                                | HTTP POST (Cookie / Bearer)
                                v
  +-----------------------------------------------------------+
  | Backend Go Gin: /api/siakad/rpc?fn=...                    |
  | 1. Normalisasi Role (ROLE_ALIASES)                        |
  | 2. Pengecekan Hak Akses Whitelist (ROLE_PERMISSIONS)       |
  | 3. Injeksi Identitas Caller (p_caller_id dari JWT)        |
  | 4. Sanitasi Keamanan Output (stripCredentials)            |
  +-----------------------------+-----------------------------+
                                | SQL Parameterized Query
                                v
  +-----------------------------------------------------------+
  | PostgreSQL Database:                                      |
  | SELECT * FROM public.get_attendance_records_secure(...)   |
  +-----------------------------------------------------------+
```

### Keunggulan Desain Gateway:
1. **Whitelist Fungsi Ketat**: Hanya fungsi Postgres terdaftar yang diizinkan untuk dipanggil. Permintaan fungsi liar langsung ditolak dengan `403 Forbidden`.
2. **Injeksi Parameter Otomatis**: Parameter krusial seperti `p_caller_id`, `p_student_id`, atau `p_user_id` tidak diambil dari input pengguna, melainkan diekstrak langsung dari token JWT yang terverifikasi untuk mencegah *identity spoofing*.
3. **`stripCredentials`**: Respons dari basis data disaring untuk memastikan kolom rahasia seperti `password`, `password_hash`, atau garam (*salt*) tidak pernah bocor ke peramban klien.
