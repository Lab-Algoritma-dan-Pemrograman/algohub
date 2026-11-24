---
title: Changelog
description: Changelog format and the conventions for recording changes.
---

All significant changes are recorded in `CHANGELOG.md` at the monorepo root.
Every change **must** have an entry there.

## Format

Following [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and
[Semantic Versioning](https://semver.org/). The newest entries live under
`## [Unreleased]`, grouped per dated change, using sub-headings such as:

- `#### Added`
- `#### Changed`
- `#### Removed`
- `#### Fixed`
- `#### Verification`

## Recent entry examples

### Update 3: storage failover & answer protection during Supabase 402 (2026-11-24)

- **Cause**: Supabase Auth/REST/Storage returned HTTP 402 because the billing quota was exhausted. The PostgreSQL database stayed healthy and is used directly by the backend, so exam answers were unaffected.
- **Registration**: when Auth returns 402, the account is still created with local bcrypt and the roster is claimed; OTP is resynced once the quota is restored.
- **Storage**: praktikum uploads, E-Learning lesson images, and final reports now use a failover object store - Vercel Blob (public/private) first, Supabase Storage as fallback. New env vars: `PUBLIC_BLOB_READ_WRITE_TOKEN`, `REPORT_BLOB_READ_WRITE_TOKEN`.
- **Answer safety**: `localStorage` drafts per user+session+course (7 days), autosave every 15 seconds, submit is blocked while any answer has not reached the server, and a 401 no longer deletes drafts.
- **UI**: avatars fall back to `ui-avatars` when the Supabase photo fails; praktikum/siakad CSP now allows Vercel Blob hosts.
- **Verification**: `/api/health` 200, production upload 201 with the file readable from Blob, 4,792 answer rows in the DB (877 updated in the last 7 days), all Go/Vitest suites green.
- **Remaining**: old photos/PDFs in Supabase Storage and OTP only recover after Supabase billing is resolved.

See `CHANGELOG.md` at the monorepo root for full details.

### Update 2: structure, auth, UI, and hardening (2026-11-23)

- **Structure**: `algohub-docs` moved into the monorepo as `apps/docs`; the Vercel git connection now uses Root Directory `apps/docs` (auto-deploy active). The old folder was removed.
- **R2 rejected**: actual storage usage is ~33 MB of the 10 GB free tier — migration would save nothing. Supabase Storage stays.
- **Versioning**: chose the expand-contract discipline (additive migrations, deploy backend→frontend) instead of `/api/v1`-`/api/v2` infrastructure.
- **Auth**: `ClientIP` anti-spoofing, timing-safe cron secret, praktikum logout now revokes tokens via the `revoked_tokens` table (migration applied to Supabase).
- **Praktikum UI**: skeleton loading, ≥12:1 badge contrast, mobile card-stack for the grade history, admin nav grouped into 5 groups, focus-visible rings, form errors near fields, `Intl.NumberFormat`.
- **Mobile baseline (4 apps)**: viewport-fit=cover, light/dark theme-color, 16px inputs (no iOS zoom), transparent tap-highlight, overscroll none, 100vh→100dvh.
- **Security**: `{@html}` XSS now sanitized with DOMPurify; dashboard crash on null `nama` fixed; RLS + grants tightened on `revoked_tokens` & `schema_migrations`; the `legacy_preserve` archive (42 MB) was backed up then dropped.

### Landing portal redesign (2026-11-20)

The `algohub.web.id` landing was rebuilt following a Framer site export, using
the original visual spec from the HTML/CSS export (Ground `#f4f4f4`,
Ink/Primary `#0a0a0a`). The hero is now two columns, buttons became pills,
platform cards use cover images, and the footer became a dark panel with an
animated "AlgoHub" wordmark.

### Cutover (2026-11-19)

All production domains moved from the old Vercel project to new ones:
`api.algohub.web.id` → `algohub-api`, `algohub.web.id` + `www` →
`algohub-portal`, `siakad.algohub.web.id` → `algohub-siakad`,
`elearning.algohub.web.id` → `algohub-elearning`,
`praktikum.algohub.web.id` → `algohub-praktikum`. The Cloudflare DNS was not
changed (all remain `CNAME cname.vercel-dns.com`).

### Cutover preparation (2026-11-18)

Added `.vercelignore` at the root and created the Vercel projects
`algohub-portal`, `algohub-praktikum`, `algohub-siakad`, `algohub-elearning`
with Root Directory `apps/*` and their respective production aliases. The
`algohub-api` production env was completed with `CRON_SECRET`, `SUPABASE_*`,
and other variables, including `AUTH_MODE=dual`.

> Refer to `CHANGELOG.md` in the repo for the full list and entries after the
> dates above.
