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

### Update 2: struktur, auth, UI, dan hardening (2026-11-23)

- **Struktur**: `algohub-docs` menjadi workspace member `apps/docs`; git Vercel terhubung ke repo dengan Root Directory `apps/docs` (auto-deploy aktif). Folder lama dihapus.
- **R2 ditolak**: pemakaian storage aktual ~33 MB dari free tier 10 GB — migrasi tidak menghemat biaya. Tetap Supabase Storage.
- **Versioning**: dipilih disiplin expand-contract (migration additive, deploy backend→frontend), tanpa infra `/api/v1`-`/api/v2`.
- **Auth**: anti-spoof `ClientIP`, cron secret timing-safe, logout praktikum mencabut token via tabel `revoked_tokens` (migrasi dijalankan ke Supabase).
- **UI praktikum**: skeleton loading, kontras badge ≥12:1, card-stack riwayat di ponsel, nav admin dikelompokkan 5 grup, focus-visible ring, error form dekat field, format angka `Intl.NumberFormat`.
- **Mobile baseline 4 app**: viewport-fit=cover, theme-color light/dark, input 16px (anti iOS zoom), tap-highlight transparan, overscroll none, 100vh→100dvh.
- **Keamanan**: XSS `{@html}` disanitasi DOMPurify; crash dashboard karena `nama` null ditutup; RLS + GRANT `revoked_tokens` & `schema_migrations` diperketat; arsip `legacy_preserve` (42 MB) dibackup lalu dihapus.

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
