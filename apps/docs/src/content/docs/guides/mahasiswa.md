---
title: Panduan Mahasiswa
description: Panduan langkah demi langkah bagi mahasiswa dalam mengikuti praktikum, belajar di e-learning, absensi presensi, dan submit laporan.
---

# Panduan Mahasiswa

Panduan ini ditujukan bagi mahasiswa peserta praktikum (praktikan) Laboratorium Algoritma dan Pemrograman untuk mengarahkan seluruh kegiatan perkuliahan praktikum, pembelajaran mandiri, hingga administrasi laboratorium di AlgoHub.

---

## 1. Akses Akun & Single Sign-On (SSO)

Mahasiswa memiliki satu identitas terpadu yang terhubung ke seluruh sistem AlgoHub menggunakan **Nomor Induk Mahasiswa (NIM)**.

### Cara Masuk ke Layanan:
1. **SIAKAD (`siakad.algohub.web.id`)**:
   - Masukkan NIM dan password akun laboratorium Anda.
   - Jika baru pertama kali mengakses, gunakan kredensial default yang diberikan oleh koordinator asisten, lalu segera perbarui kata sandi di menu **Profil**.
2. **SSO ke E-Learning**:
   - Dari menu samping SIAKAD, klik menu **E-Learning**.
   - Sistem akan melakukan otentikasi otomatis (SSO) dan mengarahkan Anda langsung ke dasbor pembelajaran tanpa perlu memasukkan kata sandi ulang.
3. **Praktikum (`praktikum.algohub.web.id`)**:
   - Masuk melalui halaman `/praktikum/login` menggunakan NIM dan kata sandi Anda.

---

## 2. Pelaksanaan Sesi Praktikum Laboratorium

### Memeriksa Jadwal
Sebelum jadwal sesi dimulai, periksa jadwal shift kelas Anda di menu `https://praktikum.algohub.web.id/info/jadwal`. Pastikan Anda hadir di laboratorium 10 menit sebelum sesi dibuka.

### Masuk Ruang Praktikum (`/praktikum/ruang`)
1. Hubungkan perangkat Anda ke jaringan laboratorium yang ditentukan.
2. Buka menu **Ruang Praktikum**. Jika sesi telah diaktivasi oleh asisten jaga, status ruangan akan berubah menjadi **Aktif**.
3. Jika sesi memerlukan token keamanan, masukkan token 6 digit yang diumumkan oleh asisten di depan kelas.
4. Klik tombol **Masuk Sesi**. Waktu pengerjaan (*countdown timer*) akan segera mulai berjalan di sudut kanan atas layar.

### Mengerjakan Soal & Menguji Kode
Di dalam ruang praktikum, Anda akan disajikan daftar soal (Pretest, Posttest, Keterampilan, atau Ujian):
- **Monaco Code Editor**: Tuliskan kode solusi Anda pada editor. Anda dapat mengatur tema editor dan ukuran font melalui panel setelan.
- **Uji Kode (Compile & Run)**: Klik tombol **Jalankan Kode**. Program C atau Python Anda akan dikompilasi secara instan di peramban.
- **Terminal Interaktif**: Jika program membutuhkan masukan (`scanf` pada C atau `input()` pada Python), ketikkan masukan langsung ke dalam terminal hitam di bawah editor dan tekan `Enter`.
- **Indikator Autosave**: Setiap karakter yang Anda ketik disimpan secara otomatis ke memori lokal peramban dan disinkronkan ke server. Pastikan indikator menunjukkan status **"Tersimpan otomatis"**.

### Pengumpulan Jawaban (Submit)
1. Periksa kembali seluruh jawaban sebelum waktu habis.
2. Klik tombol **Submit Jawaban** di bagian bawah halaman.
3. Konfirmasi pengumpulan akhir. Begitu konfirmasi dilakukan, jawaban Anda terkunci dan tidak dapat diubah kembali.
4. **Batas Waktu**: Jika waktu pada *countdown timer* habis (`00:00:00`), sistem secara otomatis mengunci lembar kerja dan men-submit seluruh draf jawaban terakhir Anda ke server.

---

## 3. Pengumpulan Laporan Besar

Pada akhir semester, mahasiswa wajib mengumpulkan berkas proyek pemrograman (Laporan Besar):
1. Buka menu **Laporan Besar** di `/praktikum/laporan-besar`.
2. Klik tombol **Unggah Berkas**.
3. Pilih berkas laporan proyek Anda (format `.pdf` atau arsip `.zip` sesuai ketentuan asisten pembimbing dengan ukuran maksimal yang ditetapkan).
4. Klik **Kirim Laporan**.
5. Pantau status pengumpulan:
   - **Terkumpul**: Berkas berhasil diterima sistem dan menunggu penilaian asisten.
   - **Perlu Revisi**: Asisten memberikan catatan perbaikan. Anda dapat mengunggah berkas revisi sebelum batas tanggal penutupan.
   - **Diterima (Marked)**: Laporan telah diverifikasi dan nilai telah diinput ke sistem.

---

## 4. Pembelajaran Gamifikasi di E-Learning

Manfaatkan platform E-Learning untuk memperdalam penguasaan materi di luar lab:

### Mengakses Kursus & Lesson
1. Buka `https://elearning.algohub.web.id/courses`.
2. Pilih **Level** dan **Modul** yang ingin dipelajari (materi tersusun dari dasar hingga konsep tingkat lanjut).
3. Buka **Lesson**: Pelajari materi bacaan, selesaikan flowchart puzzle, jawab kuis konsep, dan tuliskan kode solusi pada latihan pemrograman.

### Mempertahankan Streak Belajar Harian
- Buka E-Learning dan selesaikan minimal 1 aktivitas lesson atau kuis setiap hari.
- Ikon api (*streak*) pada profil Anda akan terus bertambah. Jangan lewatkan satu hari kalender pun agar streak tidak kembali ke nol.

### Mini-Game Bug Hunt
- Buka menu game di E-Learning setiap minggunya.
- Temukan baris kode yang mengandung error logika sebelum waktu habis untuk mengklaim bonus XP mingguan.

### Mengikuti Peringkat Leaderboard
- Pantau posisi Anda di papan peringkat mingguan dan semesteran pada menu **Leaderboard**.
- Bersainglah secara sehat dengan rekan sekelas untuk meraih skor dan level tertinggi.

---

## 5. Presensi Kehadiran di SIAKAD

Saat memasuki dan meninggalkan laboratorium praktikum:
1. Masuk ke `https://siakad.algohub.web.id/absensi`.
2. Izinkan peramban untuk mengakses kamera ponsel atau laptop Anda.
3. Arahkan kamera ke layar proyektor laboratorium yang menampilkan **Dynamic QR Code** yang dibangkitkan asisten jaga.
4. Tunggu bunyi notifikasi verifikasi berhasil.
5. Periksa riwayat kehadiran Anda pada tabel rekap absensi untuk memastikan status kehadiran tercatat valid.
