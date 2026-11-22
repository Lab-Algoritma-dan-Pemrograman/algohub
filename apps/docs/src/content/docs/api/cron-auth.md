---
title: Cron & Auth
description: Endpoint cron, autentikasi publik, info publik, dan login semua role.
---

## Cron

Dipicu scheduler eksternal, diautentikasi lewat header `X-Cron-Secret`
(dikonfigurasi lewat `CRON_SECRET`).

| Method | Path baru | Path lama (alias) |
|---|---|---|
| POST | `/api/cron/praktikum/auto-submit` | `/api/cron/auto-submit` |
| POST | `/api/cron/praktikum/rumpang` | `/api/cron/rumpang` |

## Auth (publik)

| Method | Path baru | Rate limit |
|---|---|---|
| POST | `/api/praktikum/auth/cek-nim` | 20/menit/IP |
| POST | `/api/praktikum/auth/login` | 10/menit/IP + 20/menit/IP |
| POST | `/api/praktikum/auth/register` | 20/menit/IP |
| POST | `/api/praktikum/auth/forgot-password` | 20/menit/IP |
| POST | `/api/praktikum/auth/reset-password` | 20/menit/IP |

## Info (publik)

`/api/praktikum/info/*` menyediakan data publik tanpa login:

- `asisten`, `laporan`, `modul`, `jadwal`, `jadwal/config`,
  `kelas/:id/mahasiswa`, `announcements`.

## Login semua role

Setelah login, token Bearer membuka endpoint berikut:

- **Sesi & profil** — `logout`, `me`, `GET/PUT /profile`.
- **Belajar** — `level`, `modul`, `materi`, `materi/:id`, `progres`,
  `materi/:id/selesai`, `pencapaian`, `sesi`.
- **Game** — `konfigurasi`, `status`, `mulai`, `selesai`, `peringkat`, `detak`.

Semua berada di bawah `/api/praktikum/`. Route lama tetap tersedia sebagai
alias sampai cutover.
