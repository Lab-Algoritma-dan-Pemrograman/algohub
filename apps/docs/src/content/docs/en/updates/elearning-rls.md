---
title: E-Learning RLS
description: From the old RLS to Go rules.
---

The Go e-learning backend connects to the e-learning Supabase DB as the
`postgres` role (pooler `postgres.<ref>`). With that role **RLS does not
apply**: `postgres` owns the tables and there is no `FORCE ROW LEVEL SECURITY`.
The `public.auth_role()` function also always returns `'admin'` through the
`current_user IN ('postgres','supabase_admin')` branch. As a result, every
policy and trigger depending on `auth_role()` no longer protects anything —
protection was moved to rules in `internal/usecase/elearning`.

## Role notation

From the `user_role` claim after server normalisation:

- **A** = admin
- **K** = kordas
- **S** = asisten
- **P** = praktikan

"Own row" means the row's `nim` column equals the token's `nim` (`auth_nim()`).

## Policy sources

The source document inventories 20 migration SQL files in
`elearning/docs/*.sql`, ordered by write time. Two files
(`fix-supabase-permissions.sql` and `setup-new-database.sql`) are **excluded**
because they are believed never to have run in production — the evidence is in
`migration-fix-users-column-grants.sql`, written afterwards.

## Rules that shape the outcome

- Every policy in the repo is **PERMISSIVE** (the default). Postgres combines
  permissive policies for one command with **OR**, so an old policy that was not
  `DROP`ped stays active and can open more access than intended.
- A `FOR ALL` policy without `WITH CHECK` uses its `USING` expression as the
  `WITH CHECK`.
- `UPDATE`/`DELETE` with a `WHERE` also filters rows through the SELECT policy.
- `INSERT ... ON CONFLICT DO UPDATE` checks the new row against the SELECT
  policy and the INSERT `WITH CHECK`; on failure the statement is rejected (not
  filtered).
- An RLS table without a policy for a command = default deny.

Per-policy detail and uncertain status are in
`docs/updates/elearning-rls.md` in the repo.
