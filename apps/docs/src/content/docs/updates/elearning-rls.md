---
title: E-Learning RLS
description: Dari RLS lama ke aturan Go.
---

Backend Go e-learning terhubung ke DB Supabase e-learning sebagai role
`postgres` (pooler `postgres.<ref>`). Dengan role itu **RLS tidak berlaku**:
`postgres` adalah pemilik tabel dan tidak ada `FORCE ROW LEVEL SECURITY`.
Fungsi `public.auth_role()` juga selalu mengembalikan `'admin'` lewat cabang
`current_user IN ('postgres','supabase_admin')`. Akibatnya seluruh policy dan
trigger yang bergantung `auth_role()` tidak lagi melindungi apa pun —
perlindungan dipindah ke aturan di `internal/usecase/elearning`.

## Notasi peran

Dari claim `user_role` hasil normalisasi server:

- **A** = admin
- **K** = kordas
- **S** = asisten
- **P** = praktikan

"Milik sendiri" berarti kolom `nim` baris sama dengan `nim` token
(`auth_nim()`).

## Sumber kebijakan

Dokumen sumber menginventarisasi 20 file migrasi SQL di
`elearning/docs/*.sql`, diurutkan berdasarkan waktu tulis. Dua file
(`fix-supabase-permissions.sql` dan `setup-new-database.sql`) **dikecualikan**
karena dianggap tidak pernah dijalankan di produksi — buktinya ada di
`migration-fix-users-column-grants.sql` yang ditulis sesudahnya.

## Aturan yang ikut menentukan hasil

- Semua policy di repo bersifat **PERMISSIVE** (default). Postgres menggabungkan
  policy permissive untuk satu perintah dengan **OR**, jadi policy lama yang
  tidak di-`DROP` tetap aktif dan bisa membuka akses lebih luas dari niat awal.
- Policy `FOR ALL` tanpa `WITH CHECK` memakai ekspresi `USING` sebagai
  `WITH CHECK`.
- `UPDATE`/`DELETE` dengan `WHERE` juga menyaring baris lewat policy SELECT.
- `INSERT ... ON CONFLICT DO UPDATE` memeriksa baris baru terhadap policy SELECT
  dan policy INSERT `WITH CHECK`; bila gagal, statement ditolak (bukan disaring).
- Tabel ber-RLS tanpa policy untuk suatu perintah = default deny.

Rincian per policy dan status tidak pasti ada di `docs/updates/elearning-rls.md`
di repo.
