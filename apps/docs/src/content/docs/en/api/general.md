---
title: General & CORS
description: Health check and the CORS policy of the AlgoHub API.
---

## Health check

| Method | Path | Auth | Response |
|---|---|---|---|
| GET | `/api/health` | — | `200 {"service":"algohub-api","status":"ok"}` |

## CORS

- **Allowed origins** come from the exact list `CORS_ORIGINS` and the regex
  `CORS_ORIGIN_PATTERNS`. Without env in non-production, `http://localhost:5173`
  through `5176` are allowed automatically.
- Allowed origins receive `Access-Control-Allow-Origin: <origin>`,
  `Access-Control-Allow-Credentials: true`,
  `Access-Control-Expose-Headers: Content-Disposition, Retry-After`, and
  `Vary: Origin`. The same headers appear on error responses.
- **Preflight** (`OPTIONS` with `Access-Control-Request-Method`) is answered
  `204` before any other handler runs, with:
  - `Access-Control-Allow-Methods: GET, POST, PUT, PATCH, DELETE, OPTIONS`
  - `Access-Control-Allow-Headers: Authorization, Content-Type, X-Requested-With, X-Cron-Secret`
  - `Access-Control-Max-Age: 86400`
- Foreign origins: the request is still processed without CORS headers, and the
  browser blocks it.
