---
title: Umum & CORS
description: Health check dan kebijakan CORS API AlgoHub.
---

## Health check

| Method | Path | Auth | Response |
|---|---|---|---|
| GET | `/api/health` | — | `200 {"service":"algohub-api","status":"ok"}` |

## CORS

- **Origin diizinkan** berasal dari daftar eksak `CORS_ORIGINS` dan regex
  `CORS_ORIGIN_PATTERNS`. Tanpa env di non-production, `http://localhost:5173`
  sampai `5176` otomatis diizinkan.
- Origin yang diizinkan menerima `Access-Control-Allow-Origin: <origin>`,
  `Access-Control-Allow-Credentials: true`,
  `Access-Control-Expose-Headers: Content-Disposition, Retry-After`, dan
  `Vary: Origin`. Header yang sama ada pada response error.
- **Preflight** (`OPTIONS` dengan `Access-Control-Request-Method`) dijawab `204`
  sebelum handler lain berjalan, dengan:
  - `Access-Control-Allow-Methods: GET, POST, PUT, PATCH, DELETE, OPTIONS`
  - `Access-Control-Allow-Headers: Authorization, Content-Type, X-Requested-With, X-Cron-Secret`
  - `Access-Control-Max-Age: 86400`
- Origin asing: request tetap diproses tanpa header CORS, lalu browser yang
  memblokir.
