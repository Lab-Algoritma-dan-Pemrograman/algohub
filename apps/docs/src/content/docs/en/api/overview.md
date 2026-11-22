---
title: API Overview
description: Base URL, authentication, CORS, and rate limits of the AlgoHub API.
---

The full source lives in `docs/api.md` in the monorepo.

## Base URL

```
https://api.algohub.web.id
```

All routes live under `/api`.

## Authentication

Most endpoints use the header:

```
Authorization: Bearer <token>
```

The token type follows `AUTH_MODE` (`legacy`, `dual`, `supabase`). Cron
endpoints do not use Bearer; they use the `X-Cron-Secret` header instead.

## CORS

- Allowed origins come from the exact list `CORS_ORIGINS` plus the regex
  `CORS_ORIGIN_PATTERNS`. Without env in non-production, `http://localhost:5173`
  through `5176` are allowed.
- Preflight `OPTIONS` is answered `204` before any other handler runs, with
  `Access-Control-Max-Age: 86400`.
- `Access-Control-Allow-Origin`, `Access-Control-Allow-Credentials`, and
  `Vary: Origin` are also present on error responses.

## Rate limit

The rate limiter is shared by new paths and aliases, keyed on the client IP from
`X-Forwarded-For` (first entry) or `X-Real-IP`. The `run` limiter uses the user
ID. Some endpoints have explicit limits, for example login at 10/minute/IP and
`POST /api/praktikum/run` at 30/minute/user.

## Legacy route aliases

Old praktikum routes `/api/X` moved to `/api/praktikum/X`. The old paths are
still served as aliases with the same handler and middleware until cutover.
