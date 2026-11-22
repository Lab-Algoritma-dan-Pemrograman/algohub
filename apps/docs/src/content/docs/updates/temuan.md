---
title: Temuan
description: Bug dan kejanggalan kode lama yang sengaja dipertahankan.
---

`docs/updates/temuan.md` mencatat bug dan kejanggalan dari kode lama yang
**sengaja tidak diperbaiki** selama porting, karena aturan porting adalah
perilaku harus sama dengan produksi. Setiap temuan diberi nomor `T-n` supaya
bisa dirujuk dari `TODO(T-n)` di kode.

## Cakupan

Daftar mencakup modul praktikum, siakad, dan e-learning. Contoh kategori
temuan:

- **Keamanan ringan** — `X-Cron-Secret` dibandingkan dengan `!=` alih-alih
  constant-time (T-3); API key Gemini dikirim lewat query string (T-21);
  `xp_earned` Bug Hunt diambil dari klien tanpa batas (T-39).
- **Fail-open** — error `is_token_revoked` dianggap "tidak dicabut" (T-10).
- **Perilaku rancu** — role dibandingkan mentah tanpa alias sehingga koordinator
  ditolak 403 (T-15, T-22, T-27); upsert gagal tapi response tetap 200 (T-18).
- **Kode mati** — env dan fungsi tanpa pemanggil (T-2, T-28, T-36).
- **Kualitas** — duplikasi, placeholder tidak diteruskan, peringatan a11y.

## Kebiasaan

- Nomor `T-n` bersifat permanen; jangan pakai ulang.
- Perbaikan baru boleh dilakukan setelah frontend baru siap, dan harus dicatat
  di `CHANGELOG.md`.
- Sebagian temuan punya status "sudah diganti" karena sudah ditangani di backend
  baru (misal T-11, T-20, T-5).

Daftar lengkap T-1 sampai T-54 beserta dampak dan usulan perbaikan ada di
`docs/updates/temuan.md` di repo.
