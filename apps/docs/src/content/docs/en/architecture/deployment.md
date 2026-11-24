---
title: Deployment
description: Procedures for deploying the Go backend, SvelteKit frontends, and documentation to Vercel, including post-deploy verification and rollback.
---

# Deployment

All production services run on Vercel from a single monorepo. Each project uses a different Root Directory, so deployment commands must be run from the repository root.

## 1. Project Mapping

| Vercel Project | Root Directory | Production alias |
|---|---|---|
| `algohub-api` | `backend` | `api.algohub.web.id` |
| `algohub-praktikum` | `apps/praktikum` | `praktikum.algohub.web.id` |
| `algohub-siakad` | `apps/siakad` | `siakad.algohub.web.id` |
| `algohub-elearning` | `apps/elearning` | `elearning.algohub.web.id` |
| `algohub-portal` | `apps/portal` | `algohub.web.id` |
| `algohub-docs` | `apps/docs` | `docs.algohub.web.id` |

CLI scope: `lab-algoritma-pemrograman`.

## 2. Deploy the Backend

Run from the repository root, **not** from `backend/`:

```bash
vercel deploy --prod --project algohub-api --scope lab-algoritma-pemrograman --yes --logs
```

Project `algohub-api` already sets Root Directory to `backend`, so running the command inside `backend/` doubles the path and fails.

### Required production environment
- `DATABASE_URL_PRAKTIKUM`, `DATABASE_URL_SIAKAD`, `DATABASE_URL_ELEARNING`
- `JWT_SECRET`, `SIAKAD_JWT_SECRET`, `AUTH_MODE=dual`
- `SUPABASE_URL`, `SUPABASE_SERVICE_KEY`, `SUPABASE_BUCKET`, `SUPABASE_JWKS_URL`
- `PUBLIC_BLOB_READ_WRITE_TOKEN`, `REPORT_BLOB_READ_WRITE_TOKEN`
- `CRON_SECRET`, `CORS_ORIGINS`

Blob tokens must stay in the server environment; never expose them to a frontend bundle.

## 3. Deploy Frontends & Docs

```bash
vercel deploy --prod --project algohub-praktikum --scope lab-algoritma-pemrograman --yes --logs
vercel deploy --prod --project algohub-siakad --scope lab-algoritma-pemrograman --yes --logs
vercel deploy --prod --project algohub-elearning --scope lab-algoritma-pemrograman --yes --logs
vercel deploy --prod --project algohub-portal --scope lab-algoritma-pemrograman --yes --logs
vercel deploy --prod --project algohub-docs --scope lab-algoritma-pemrograman --yes --logs
```

`algohub-docs` is also connected to Git (auto-deploy on `main`). Manual deployment is only needed to publish changes before pushing.

## 4. Post-Deploy Verification

```bash
curl -s https://api.algohub.web.id/api/health
curl -s -o /dev/null -w "%{http_code}\n" https://praktikum.algohub.web.id/
curl -s -o /dev/null -w "%{http_code}\n" https://siakad.algohub.web.id/
```

Check the following:
1. The health endpoint returns `{"service":"algohub-api","status":"ok"}`.
2. Each app's home page returns `200`.
3. The CSP header includes `*.public.blob.vercel-storage.com` (praktikum/siakad).
4. A small file upload succeeds and the resulting URL can be opened.
5. For changes touching exam data, run the `BEGIN/UPDATE/ROLLBACK` write probe (see the Database & Auth page).

## 5. Rollback

Vercel deployments are immutable. Open the project dashboard, select the last healthy deployment, and choose **Promote to Production**. CLI alternative: `vercel rollback` from the project root. Re-run the verification steps above after rolling back.

## 6. Release Order

- Database schema changes must be **additive** and deployed before the code that reads them.
- Deploy the backend first, then the frontends.
- Never run `AutoMigrate` in production.
