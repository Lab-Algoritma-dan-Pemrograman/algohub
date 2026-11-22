---
title: Ringkasan API
description: Base URL, autentikasi, CORS, dan rate limit API AlgoHub.
---

Sumber lengkap ada di `docs/api.md` di dalam repo.

## Base URL

```
https://api.algohub.web.id
```

Semua route berada di bawah `/api`.

## Autentikasi

Kebanyakan endpoint memakai header:

```
Authorization: Bearer <token>
```

Jenis token mengikuti `AUTH_MODE` (`legacy`, `dual`, `supabase`). Endpoint cron
tidak memakai Bearer melainkan header `X-Cron-Secret`.

## CORS

- Origin diizinkan berasal dari daftar eksak `CORS_ORIGINS` plus regex
  `CORS_ORIGIN_PATTERNS`. Tanpa env di non-production, `http://localhost:5173`
  sampai `5176` diizinkan.
- Preflight `OPTIONS` dijawab `204` sebelum handler lain berjalan, dengan
  `Access-Control-Max-Age: 86400`.
- Header `Access-Control-Allow-Origin`, `Access-Control-Allow-Credentials`, dan
  `Vary: Origin` juga hadir pada response error.

## Rate limit

Rate limiter dipakai bersama oleh path baru dan alias, dengan kunci IP klien
dari `X-Forwarded-For` (entri pertama) atau `X-Real-IP`. Limiter `run` memakai
ID user. Beberapa endpoint punya batas eksplisit, misalnya login 10/menit/IP dan
`POST /api/praktikum/run` 30/menit/user.

## Alias route lama

Route praktikum lama `/api/X` pindah ke `/api/praktikum/X`. Path lama tetap
dilayani sebagai alias dengan handler dan middleware yang sama sampai cutover.
