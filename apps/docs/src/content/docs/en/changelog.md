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
