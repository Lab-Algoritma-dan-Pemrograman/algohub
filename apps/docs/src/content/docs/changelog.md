---
title: Riwayat Perubahan
description: Format changelog dan kebiasaan pencatatan perubahan.
---

Semua perubahan penting dicatat di file `CHANGELOG.md` di root monorepo. Setiap
perubahan **wajib** punya entri di sana.

## Format

Mengikuti [Keep a Changelog](https://keepachangelog.com/id-ID/1.1.0/) dan
[Semantic Versioning](https://semver.org/lang/id/). Entri terbaru berada di
bagian `## [Belum dirilis]`, dikelompokkan per perubahan bertanggal, lalu
di dalamnya dipakai sub-judul seperti:

- `#### Ditambahkan`
- `#### Diubah`
- `#### Dihapus`
- `#### Diperbaiki`
- `#### Verifikasi`

## Contoh entri terbaru

### Redesign landing portal (2026-11-20)

Landing `algohub.web.id` dibangun ulang mengikuti export situs Framer, memakai
spesifikasi visual asli dari HTML/CSS export (Ground `#f4f4f4`, Ink/Primary
`#0a0a0a`). Hero kini dua kolom, tombol menjadi pill, kartu platform memakai
cover gambar, dan footer menjadi panel gelap dengan wordmark "AlgoHub"
beranimasi wave.

### Cutover (2026-11-19)

Semua domain produksi dipindahkan dari proyek Vercel lama ke proyek baru:
`api.algohub.web.id` → `algohub-api`, `algohub.web.id` + `www` → `algohub-portal`,
`siakad.algohub.web.id` → `algohub-siakad`, `elearning.algohub.web.id` →
`algohub-elearning`, `praktikum.algohub.web.id` → `algohub-praktikum`. DNS
Cloudflare tidak diubah (semua tetap `CNAME cname.vercel-dns.com`).

### Persiapan cutover (2026-11-18)

Dibuat `.vercelignore` di root serta project Vercel `algohub-portal`,
`algohub-praktikum`, `algohub-siakad`, `algohub-elearning` dengan Root Directory
`apps/*` dan alias production masing-masing. Env production `algohub-api`
lengkap dengan `CRON_SECRET`, `SUPABASE_*`, dan variabel lain, termasuk
`AUTH_MODE=dual`.

> Rujuk `CHANGELOG.md` di repo untuk daftar lengkap dan entri setelah tanggal
> di atas.
