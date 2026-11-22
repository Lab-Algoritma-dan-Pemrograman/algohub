---
title: Praktikum
description: Endpoint mahasiswa dan admin pada modul praktikum.
---

## Mahasiswa

Diperlukan role `mahasiswa`.

| Method | Path |
|---|---|
| GET | `/api/praktikum/dashboard` |
| GET | `/api/praktikum/sesi` |
| GET | `/api/praktikum/ruang` |
| POST | `/api/praktikum/mulai` |
| POST | `/api/praktikum/autosave` |
| POST | `/api/praktikum/submit` |
| POST | `/api/praktikum/run` (30/menit/user) |
| GET | `/api/praktikum/laporan-besar` |
| POST | `/api/praktikum/laporan-besar/upload-url` |
| POST | `/api/praktikum/laporan-besar/kumpul` |
| DELETE | `/api/praktikum/laporan-besar/kumpul` |

`POST /api/praktikum/compile-c` sudah dihapus — compile C kini dilakukan di
browser.

## Admin

Role `asisten` dan `koordinator`. Beberapa endpoint butuh permission
`RequirePermission` (`resource:write`) di atas role:

- `users:write` — kelola user (`users`, `users/bulk`, `users/bulk-attrs`,
  `users/:id`, `users/:id/reset-password`).
- `sesi:write` — kelola materi, kurikulum belajar, dan soal game.
- `penilaian:write` — akses level, reset progres, dan XP user.
- `konfigurasi:write` — ubah konfigurasi game.

Endpoint lain pada grup admin (dashboard, monitoring, kelas, soal, dan
sebagainya) hanya butuh role `asisten`/`koordinator`.

Lihat `docs/api.md` di repo untuk daftar lengkap path beserta alias lamanya.
