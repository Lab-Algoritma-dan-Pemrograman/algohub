---
title: Panduan Asisten & Koordinator
description: Panduan operasional bagi asisten laboratorium dan koordinator dalam aktivasi sesi, monitoring ujian, penilaian AI, dan manajemen jadwal jaga.
---

# Panduan Asisten & Koordinator

Panduan ini disusun sebagai pedoman operasional bagi **Asisten Laboratorium** dan **Koordinator Asisten (Kordas)** dalam memandu jalannya sesi praktikum, mengawasi ujian, mengelola penilaian terstandar, serta mengatur administrasi presensi dan jadwal jaga di AlgoHub.

---

## 1. Aktivasi Sesi Praktikum & Token Ujian

Asisten yang bertugas pada suatu shift laboratorium bertanggung jawab membuka dan mengontrol akses ruang pengerjaan mahasiswa:

### Langkah Membuka Sesi Praktikum:
1. Buka antarmuka admin praktikum di `https://praktikum.algohub.web.id/praktikum/admin/aktivasi`.
2. Klik tombol **Aktivasi Sesi Baru**.
3. Pilih parameter sesi:
   - **Kelas & Modul**: Tentukan kelas yang sedang berjalan dan modul materi yang diujikan.
   - **Tipe Sesi**: Pilih jenis sesi (`Pretest`, `Posttest`, `Keterampilan`, atau `Ujian`).
   - **Durasi Waktu**: Tetapkan durasi pengerjaan dalam menit (misal: Pretest 15 menit, Keterampilan 60 menit).
4. **Generate Token Keamanan**:
   - Klik **Bangkitkan Token** untuk menghasilkan kode unik 6 karakter.
   - Tampilkan token ini di layar proyektor ruang laboratorium agar dapat dimasukkan oleh mahasiswa yang hadir secara fisik.
5. Klik **Buka Sesi**. Status sesi akan berubah menjadi aktif dan mahasiswa diizinkan masuk ke ruang praktikum.

### Pengelolaan Sesi Susulan (*Make-Up Sessions*)
Bagi mahasiswa yang berhalangan hadir dengan alasan sah (surat sakit/tugas kampus):
1. Buka detail sesi terkait di `/admin/aktivasi/:id`.
2. Masuk ke tab **Peserta Susulan**.
3. Cari mahasiswa berdasarkan NIM atau nama, lalu klik **Izinkan Susulan**.
4. Sistem akan mengizinkan mahasiswa tersebut mengakses ruang ujian dengan alokasi waktu penuh meskipun sesi kelas reguler telah ditutup.
5. Setelah pengerjaan susulan selesai, asisten dapat menutup izin susulan melalui tombol **Hapus Susulan**.

### Kontrol Buka/Tutup Fleksibel
- **Buka/Tutup Course**: Jika seluruh mahasiswa telah selesai lebih awal, asisten dapat langsung mengunci seluruh sesi melalui tombol **Tutup Sesi**. Aksi ini memicu fungsi backend untuk men-submit otomatis seluruh draft jawaban yang belum tersubmit manual.
- **Buka/Tutup Akses Individu**: Asisten dapat memperpanjang waktu khusus atau mengunci pengerjaan mahasiswa tertentu secara granular jika terjadi kendala teknis perangkat keras di lab.

---

## 2. Monitoring Mahasiswa Realtime

Selama sesi berlangsung, asisten bertugas mengawasi ketertiban ujian melalui dasbor monitoring:
1. Buka menu **Monitoring Sesi** di `/admin/monitoring/sesi`.
2. Dasbor menyajikan status setiap praktikan secara langsung (*realtime*):
   - **Status Koneksi**: Indikator hijau menandakan peramban mahasiswa aktif mengirimkan *heartbeat* rutin.
   - **Status Pengerjaan**: Menampilkan nomor soal yang sedang aktif dikerjakan oleh mahasiswa.
   - **Status Simpan**: Menampilkan timestamp terakhir jawaban mahasiswa berhasil disimpan ke server.
   - **Status Submit**: Berubah warna saat mahasiswa telah melakukan submit final.
3. **Deteksi Anomali**: Jika seorang mahasiswa terdeteksi offline lebih dari 2 menit saat ujian berlangsung, asisten dapat segera menghampiri meja praktikan untuk memeriksa kendala jaringan atau perangkat.

---

## 3. Penilaian & Evaluasi Mahasiswa

AlgoHub menyediakan sistem penilaian terpusat yang menggabungkan koreksi manusia dengan bantuan analisis cerdas model AI.

### Koreksi Jawaban Esai & Kode Pemrograman
1. Buka menu **Penilaian** di `/admin/penilaian/rekap`.
2. Pilih kelas dan sesi yang telah selesai.
3. Klik nama mahasiswa untuk membuka lembar jawaban komparatif:
   - Sisi kiri menampilkan soal dan rubrik acuan jawaban.
   - Sisi kanan menampilkan kode atau jawaban esai mahasiswa.
4. Inputkan skor nilai (rentang 0–100) dan berikan catatan umpan balik yang konstruktif bagi mahasiswa.

### AI-Assisted Grading (Penilaian Berbantuan AI)
Untuk mempercepat evaluasi awal pada ratusan baris kode pemrograman:
1. Klik tombol **AI-Grade** pada lembar jawaban mahasiswa.
2. Backend Go AlgoHub akan menganalisis kode mahasiswa terhadap rubrik soal menggunakan model bahasa (Ollama lokal atau Gemini API).
3. AI memberikan rekomendasi nilai numerik beserta ringkasan analisis kesalahan logika, kompleksitas waktu, dan sintaksis.
4. **Peran Asisten**: Asisten memverifikasi rekomendasi AI, dapat mengubah nilai rekomendasi sesuai pertimbangan objektif, lalu mengonfirmasi nilai akhir ke basis data.

### Input Nilai Keaktifan Lab
Asisten yang bertugas di ruang lab dapat mencatat keaktifan tanya-jawab atau kedisiplinan mahasiswa selama praktikum berlangsung melalui menu `/admin/keaktifan`.

### Verifikasi Laporan Besar
1. Buka menu **Laporan Besar** di `/admin/laporan-besar`.
2. Pilih kelas yang diampu.
3. Klik **Preview** untuk meninjau berkas laporan PDF mahasiswa secara langsung di peramban tanpa perlu mengunduh ke laptop lokal.
4. Tentukan status laporan:
   - **Mark**: Laporan disetujui dan nilai akhir dimasukkan.
   - **Override**: Memberikan izin kepada mahasiswa untuk mengunggah ulang dokumen revisi.
5. **Download Batch**: Untuk keperluan arsip jurusan atau dosen pengampu, klik **Unduh Semua (ZIP)** untuk mengunduh seluruh berkas laporan satu kelas dalam satu arsip terkompresi.

---

## 4. Manajemen Jadwal Jaga & Swap Shift di SIAKAD

### Mengecek Jadwal Jaga
1. Buka `https://siakad.algohub.web.id/jadwal-jaga`.
2. Kalender kerja menampilkan seluruh shift piket asistensi Anda selama satu semester.

### Prosedur Tukar Giliran Jaga (Swap Shift)
Jika Anda memiliki jadwal kuliah pengganti atau keperluan mendesak:
1. Pada kartu jadwal jaga yang ingin ditukar, klik tombol **Ajukan Swap**.
2. Pilih asisten rekanan yang telah disepakati untuk menggantikan Anda.
3. Tuliskan alasan singkat permohonan tukar jadwal.
4. Klik **Kirim Permohonan**. Sistem akan mengirimkan pesan notifikasi WhatsApp dan Web Push ke rekan asisten tersebut.
5. Begitu rekan Anda menekan tombol **Setuju** di akun SIAKAD-nya, database otomatis menukar nama penanggung jawab jadwal tersebut.

### Presensi Praktikum & Buat Dynamic QR
1. Sebelum praktikum dimulai, asisten jaga membuka `https://siakad.algohub.web.id/buat-qr` pada komputer proyektor lab.
2. Pilih kelas praktikum yang sedang berlangsung, lalu klik **Tampilkan QR**.
3. Layar proyektor akan menampilkan Dynamic QR Code yang berotasi otomatis setiap beberapa detik untuk dipindai oleh mahasiswa yang hadir.
4. Pantau daftar mahasiswa yang berhasil absen di menu **Validasi Absensi** (`/validasi-absensi`). Asisten dapat memvalidasi kehadiran secara manual jika mahasiswa mengalami kendala kamera perangkat.
