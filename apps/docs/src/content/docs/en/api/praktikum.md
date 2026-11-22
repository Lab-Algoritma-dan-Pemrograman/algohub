---
title: Praktikum
description: Student and admin endpoints in the praktikum module.
---

## Students

Requires the `mahasiswa` role.

| Method | Path |
|---|---|
| GET | `/api/praktikum/dashboard` |
| GET | `/api/praktikum/sesi` |
| GET | `/api/praktikum/ruang` |
| POST | `/api/praktikum/mulai` |
| POST | `/api/praktikum/autosave` |
| POST | `/api/praktikum/submit` |
| POST | `/api/praktikum/run` (30/minute/user) |
| GET | `/api/praktikum/laporan-besar` |
| POST | `/api/praktikum/laporan-besar/upload-url` |
| POST | `/api/praktikum/laporan-besar/kumpul` |
| DELETE | `/api/praktikum/laporan-besar/kumpul` |

`POST /api/praktikum/compile-c` has been removed — C compilation now happens in
the browser.

## Admin

Roles `asisten` and `koordinator`. Some endpoints require a
`RequirePermission` permission (`resource:write`) above the role:

- `users:write` — user management (`users`, `users/bulk`, `users/bulk-attrs`,
  `users/:id`, `users/:id/reset-password`).
- `sesi:write` — manage learning materials, curriculum, and game questions.
- `penilaian:write` — level access, progress reset, and user XP.
- `konfigurasi:write` — change game configuration.

Other endpoints in the admin group (dashboard, monitoring, classes, questions,
and so on) only require the `asisten`/`koordinator` role.

See `docs/api.md` in the repo for the full list of paths and their legacy
aliases.
