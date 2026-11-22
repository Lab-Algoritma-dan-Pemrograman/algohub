---
title: Aplikasi Praktikum
description: Sistem inti pelaksanaan praktikum pemrograman, ruang pengerjaan berbatas waktu, autosave draft, runner WebAssembly, dan tata kelola laporan besar.
---

# Aplikasi Praktikum (`apps/praktikum`)

**Praktikum AlgoHub** (`praktikum.algohub.web.id`) adalah sistem inti pelaksanaan kegiatan laboratorium resmi bagi seluruh mahasiswa yang menempuh mata kuliah Algoritma dan Pemrograman. Aplikasi ini mengelola ujian berbasis komputer (*Computer-Based Test*), penilaian kode otomatis, pengumpulan tugas akhir laporan besar, hingga rekapitulasi nilai akademik.

---

## 1. Alur Sesi Praktikum Laboratorium

Setiap pertemuan praktikum di laboratorium terbagi ke dalam empat jenis sesi terstandar:

```
  +------------------+     +-------------------+     +---------------------+
  |     PRETEST      |     |     POSTTEST      |     |    KETERAMPILAN     |
  |  (10-15 Menit)   |---->|   (15-20 Menit)   |---->|    (45-60 Menit)    |
  | Uji Konsep Dasar |     | Evaluasi Teori Lab|     | Coding Kasus Nyata  |
  +------------------+     +-------------------+     +---------------------+
                                                                |
                                                                v
                                                     +---------------------+
                                                     |    UJIAN UTAMA      |
                                                     |    (UTS / UAS)      |
                                                     | Ujian Praktik Penuh |
                                                     +---------------------+
```

### Prosedur Pelaksanaan Sesi:
1. **Pemeriksaan Jadwal**: Mahasiswa memeriksa jadwal sesi praktikum di `/info/jadwal`.
2. **Aktivasi Ruang oleh Asisten**: Asisten membuka sesi melalui `/admin/aktivasi` dan mengumumkan token sesi jika diperlukan.
3. **Pengerjaan Soal (`/praktikum/ruang`)**: Mahasiswa masuk ke ruang praktikum. Layar menampilkan daftar soal esai, pilihan ganda, atau kasus pemrograman dengan **countdown timer** yang berjalan sinkron dengan jam server.
4. **Uji Kode & Debugging**: Mahasiswa menulis solusi kode C atau Python di Monaco Editor dan mengujinya terhadap berbagai skenario input.
5. **Final Submission**: Mahasiswa melakukan submit final. Jika waktu habis sebelum tombol submit ditekan, sistem secara otomatis melakukan *auto-submit* seluruh jawaban terakhir.

---

## 2. Mekanisme Autosave & Proteksi Jawaban

Untuk mencegah hilangnya jawaban mahasiswa akibat gangguan listrik atau peramban tertutup:

### Dual-Layer Storage
- **Local Storage / IndexedDB (`algohub-praktikum-<id>`)**: Setiap ketikan kode atau perubahan jawaban disimpan secara instan ke database peramban lokal.
- **Remote Autosave (`POST /api/praktikum/autosave`)**: Perubahan disinkronkan ke server secara berkala menggunakan teknik *debounce* (hanya mengirim saat ada jeda ketikan atau perubahan nilai nyata) guna mencegah pembengkakan pemanggilan API.

### Graceful Auto-Submit
- Ketika timer di sisi klien mencapai `00:00:00`, sistem mengunci editor dan mengirimkan draf terakhir ke `/api/praktikum/submit`.
- Scheduler background di backend (`/api/cron/praktikum/auto-submit`) secara berkala memeriksa sesi yang telah melewati batas waktu toleransi dan secara otomatis membekukan jawaban mahasiswa yang belum tersubmit manual.

---

## 3. In-Browser Runner C WebAssembly

Berbeda dengan sistem judge konvensional yang mengeksekusi kode di server sandbox (yang boros CPU dan rentan eksploitasi keamanan), AlgoHub menjalankan kompilasi dan eksekusi kode C **100% di dalam peramban mahasiswa**:

```
  +-----------------------------------------------------------+
  | Halaman Ruang Praktikum (Origin: praktikum.algohub.web.id) |
  | Header HTTP: Cross-Origin-Opener-Policy: same-origin      |
  | Header HTTP: Cross-Origin-Embedder-Policy: require-corp   |
  +-----------------------------+-----------------------------+
                                | Aktifkan crossOriginIsolated
                                v
  +-----------------------------------------------------------+
  | Web Worker Kompiler C (@algohub/runner)                   |
  | 1. Unduh Clang WASM dari algohub.web.id/compiler/v1/      |
  | 2. Kompilasi source.c -> output.wasm                      |
  | 3. Jalankan output.wasm di Virtual POSIX Sandbox          |
  | 4. Tangani scanf() via SharedArrayBuffer & Atomics        |
  +-----------------------------+-----------------------------+
                                | Salurkan stdout / stderr
                                v
  +-----------------------------------------------------------+
  | Terminal Interaktif Mahasiswa (xterm.js)                  |
  +-----------------------------------------------------------+
```

### Keuntungan Arsitektur:
- **Nol Beban Server**: Kompilasi puluhan ribu baris kode oleh ratusan mahasiswa saat ujian berlangsung tidak membebani server backend sama sekali.
- **Keamanan Absolut**: Kode berbahaya (seperti upaya membaca filesystem atau serangan fork-bomb) terkurung di dalam sandbox WebAssembly peramban pengguna.
- **Responsif**: Umpan balik error kompilasi muncul seketika tanpa latensi jaringan.

---

## 4. Pengumpulan & Pengelolaan Laporan Besar

Pada akhir semester, mahasiswa wajib mengumpulkan berkas Laporan Besar tugas proyek:
- **Unggah Berkas**: Mahasiswa meminta presigned upload URL melalui `POST /api/praktikum/laporan-besar/upload-url` dan mengunggah berkas PDF/arsip proyek langsung ke Supabase Storage bucket terenkripsi.
- **Review Asisten**: Asisten meninjau berkas di `/admin/laporan-besar`, memberikan catatan revisi, dan menetapkan status *marked*.
- **Unduh Massal (*Batch Download*)**: Koordinator asisten dapat mengunduh seluruh berkas laporan satu angkatan dalam satu arsip ZIP terstruktur menggunakan `POST /api/praktikum/admin/laporan-besar/:id/download-batch`.

---

## 5. Sistem Penilaian & Rekapitulasi Data

- **Penilaian Manual & Rubrik**: Asisten mengoreksi jawaban mahasiswa melalui antarmuka rekapitulasi nilai dengan rubrik penilaian terstandar.
- **AI-Assisted Grading**: Pilihan evaluasi awal berbantuan model AI (Ollama atau Gemini) untuk mendeteksi kesesuaian logika dan pola plagiarisme.
- **Virtual List Scrollable**: Tabel rekapitulasi nilai di `/admin/rekap-nilai` memanfaatkan `@tanstack/svelte-virtual` untuk merender lebih dari 1.200 baris data mahasiswa secara mulus dengan konsumsi memori minimal.
- **Ekspor Nilai**: Nilai akhir dapat diekspor ke format CSV terstandar RFC 4180 untuk diserahkan ke dosen pengampu.
