# API

Base URL produksi: `https://api.algohub.web.id`. Semua route di bawah `/api`.

## Umum
| Method | Path | Auth | Response |
|---|---|---|---|
| GET | `/api/health` | — | `200 {"service":"algohub-api","status":"ok"}` |

## CORS
- Origin diizinkan: daftar eksak `CORS_ORIGINS` + regex `CORS_ORIGIN_PATTERNS`. Tanpa env di non-production: `http://localhost:5173`–`5176`.
- Origin diizinkan mendapat `Access-Control-Allow-Origin: <origin>`, `Access-Control-Allow-Credentials: true`, `Access-Control-Expose-Headers: Content-Disposition, Retry-After`, dan `Vary: Origin`. Header ini juga ada di response error.
- Preflight (`OPTIONS` dengan `Access-Control-Request-Method`) dijawab `204` sebelum handler lain berjalan, dengan `Access-Control-Allow-Methods: GET, POST, PUT, PATCH, DELETE, OPTIONS`, `Access-Control-Allow-Headers: Authorization, Content-Type, X-Requested-With, X-Cron-Secret`, dan `Access-Control-Max-Age: 86400`.
- Origin asing: request tetap diproses tanpa header CORS, dan browser yang memblokir.

## Pemetaan route lama → baru
Diisi mulai Fase 2.
