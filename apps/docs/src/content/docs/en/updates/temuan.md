---
title: Findings
description: Bugs and oddities in the old code that were kept deliberately.
---

`docs/updates/temuan.md` records bugs and oddities in the old code that were
**deliberately not fixed** during the port, because the porting rule is that
behaviour must match production. Each finding gets a `T-n` number so it can be
referenced from `TODO(T-n)` in the code.

## Scope

The list covers the praktikum, siakad, and e-learning modules. Example
categories:

- **Light security** — `X-Cron-Secret` compared with `!=` instead of
  constant-time (T-3); the Gemini API key sent via query string (T-21); Bug Hunt
  `xp_earned` taken from the client without bounds (T-39).
- **Fail-open** — an `is_token_revoked` error treated as "not revoked" (T-10).
- **Ambiguous behaviour** — roles compared raw without aliasing, so kordas gets
  403 (T-15, T-22, T-27); a failed upsert still returns 200 (T-18).
- **Dead code** — env and functions with no callers (T-2, T-28, T-36).
- **Quality** — duplication, an unforwarded placeholder, a11y warnings.

## Conventions

- `T-n` numbers are permanent; do not reuse them.
- New fixes may land once the new frontend is ready and must be recorded in
  `CHANGELOG.md`.
- Some findings are marked "already replaced" because the new backend already
  handles them (for example T-11, T-20, T-5).

The full list from T-1 to T-54, with impact and proposed fixes, is in
`docs/updates/temuan.md` in the repo.
