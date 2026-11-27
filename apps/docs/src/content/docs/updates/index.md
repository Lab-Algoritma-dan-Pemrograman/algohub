---
title: Index Pembaruan
description: Ringkasan dokumen pembaruan di docs/updates/.
---

Folder `docs/updates/` di monorepo memuat catatan keputusan teknis yang
memengaruhi porting ke backend Go. Dokumen utama:

- [`elearning-rls`](/updates/elearning-rls) — penerjemahan kebijakan RLS lama
  Supabase menjadi aturan di service Go e-learning.
- [`temuan`](/updates/temuan) — daftar bug dan kejanggalan kode lama yang
  sengaja **tidak** diperbaiki demi menjaga paritas perilaku dengan produksi.
- [`r2-migrasi-dan-data-mahasiswa`](/updates/r2-migrasi-dan-data-mahasiswa) —
  migrasi aset publik ke Cloudflare R2, perbaikan tiga jalur tulis destruktif
  penyebab data mahasiswa terhapus, re-backfill data kelas, dan status yang
  belum dikerjakan (writer backend masih Supabase, cold start 3.4s).

Dokumen pembaruan lain akan ditambahkan seiring berjalannya porting.
