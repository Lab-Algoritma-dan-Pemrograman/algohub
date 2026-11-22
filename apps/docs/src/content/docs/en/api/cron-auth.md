---
title: Cron & Auth
description: Cron endpoints, public auth, public info, and login for all roles.
---

## Cron

Triggered by an external scheduler, authenticated through the `X-Cron-Secret`
header (configured via `CRON_SECRET`).

| Method | New path | Old path (alias) |
|---|---|---|
| POST | `/api/cron/praktikum/auto-submit` | `/api/cron/auto-submit` |
| POST | `/api/cron/praktikum/rumpang` | `/api/cron/rumpang` |

## Auth (public)

| Method | New path | Rate limit |
|---|---|---|
| POST | `/api/praktikum/auth/cek-nim` | 20/minute/IP |
| POST | `/api/praktikum/auth/login` | 10/minute/IP + 20/minute/IP |
| POST | `/api/praktikum/auth/register` | 20/minute/IP |
| POST | `/api/praktikum/auth/forgot-password` | 20/minute/IP |
| POST | `/api/praktikum/auth/reset-password` | 20/minute/IP |

## Info (public)

`/api/praktikum/info/*` serves public data without login:

- `asisten`, `laporan`, `modul`, `jadwal`, `jadwal/config`,
  `kelas/:id/mahasiswa`, `announcements`.

## Login for all roles

After login, the Bearer token opens the following endpoints:

- **Session & profile** — `logout`, `me`, `GET/PUT /profile`.
- **Learning** — `level`, `modul`, `materi`, `materi/:id`, `progres`,
  `materi/:id/selesai`, `pencapaian`, `sesi`.
- **Game** — `konfigurasi`, `status`, `mulai`, `selesai`, `peringkat`, `detak`.

All live under `/api/praktikum/`. The old routes remain available as aliases
until cutover.
