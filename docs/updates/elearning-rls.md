# E-Learning: RLS lama → aturan Go

Backend Go elearning terhubung ke DB Supabase elearning sebagai role `postgres` (pooler `postgres.<ref>`). Dengan role itu RLS tidak berlaku, karena `postgres` pemilik tabel dan tidak ada `FORCE ROW LEVEL SECURITY`. `public.auth_role()` juga mengembalikan `'admin'` lewat cabang `current_user IN ('postgres','supabase_admin')` dan cabang "tanpa `request.jwt.claims`". Akibatnya semua policy dan trigger yang bergantung `auth_role()` tidak lagi melindungi apa pun. Dokumen ini mencatat kebijakan efektif terakhir dari 20 file `E:/rekapProject/project_ap/elearning/docs/*.sql` dan menerjemahkannya ke aturan di `internal/usecase/elearning`. Semantik penolakan mengikuti keputusan 1 di plan Fase 7.

Notasi peran token (claim `user_role` hasil normalisasi server): **A** = admin, **K** = kordas, **S** = asisten, **P** = praktikan. "Milik sendiri" artinya kolom `nim` baris = `nim` token (`auth_nim()`).

## Sumber & urutan

Urutan eksekusi diasumsikan `LastWriteTime` naik, lalu nama file (`Get-ChildItem E:/rekapProject/project_ap/elearning/docs/*.sql | Sort LastWriteTime,Name`):

| # | Waktu tulis | File | Isi yang relevan |
|---|---|---|---|
| 1 | 2026-09-12 21:45:07 | `migration-assessment-settings.sql` | tabel + 2 policy `assessment_grading_rules` |
| 2 | 2026-09-12 21:45:07 | `migration-fix-leaderboard-rls.sql` | 4 policy `users`, 1 policy `activity_logs` |
| 3 | 2026-09-12 21:45:07 | `migration-fix-question-types.sql` | constraint `assessment_questions`, tanpa policy |
| 4 | 2026-09-12 21:45:07 | `migration-fix-rls.sql` | Pilihan A: `DISABLE ROW LEVEL SECURITY` levels/modules/lessons/game_questions/game_settings. Pilihan B (10 `CREATE POLICY`) ada di dalam blok komentar `/* ... */` baris 27-72, jadi tidak pernah jalan |
| 5 | 2026-09-18 00:22:53 | `migration-elearning-updates.sql` | kolom `levels.sort_order`, `users.study_time` |
| 6 | 2026-09-18 00:22:53 | `migration-final-security.sql` | RLS diaktifkan lagi untuk tabel kurikulum, policy storage, game_history, achievements, student_progress, trigger `update_user_last_active_from_session` |
| 7 | 2026-09-18 00:22:53 | `migration-harden-rls.sql` | set policy utama untuk 12 tabel, trigger `check_progress_velocity` |
| 8 | 2026-09-18 00:22:53 | `migration-xp-security.sql` | `lessons.xp_reward`, view `student_lessons` v1, policy lessons khusus staf |
| 9 | 2026-09-25 17:50:43 | `fix-role-update-trigger.sql` | `auth_role()` v1, `check_user_updates`, `check_user_xp_level` |
| 10 | 2026-09-25 17:50:43 | `migration-canonical-role.sql` | `users.elearning_role`, trigger `canonicalize_user_role` |
| 11 | 2026-09-25 17:50:43 | `migration-fix-role-claim.sql` | `auth_role()` v1 (teks fungsi identik dengan #9) |
| 12 | 2026-09-25 17:50:43 | `migration-users-rls.sql` | RLS `users` diaktifkan, 6 policy `users`, `check_user_updates` |
| 13 | 2026-09-28 19:06:57 | `fix-supabase-permissions.sql` | **dikecualikan** |
| 14 | 2026-09-28 19:06:57 | `migration-flowchart-exercise.sql` | kolom `exercise_type`, `flowchart_distractors`, view `student_lessons` v2 |
| 15 | 2026-09-28 19:06:57 | `migration-flowchart-solution-visible.sql` | view `student_lessons` v3 (efektif) |
| 16 | 2026-09-28 19:06:57 | `setup-new-database.sql` | **dikecualikan** |
| 17 | 2026-10-01 21:37:35 | `migration-fix-users-column-grants.sql` | grant kolom `users`, `check_user_updates` (efektif) |
| 18 | 2026-10-25 20:26:25 | `migration-prestige.sql` | kolom prestige, 4 policy, `guard_prestige_columns`, `do_prestige_reset` |
| 19 | 2026-10-25 20:26:25 | `migration-weekly-standings.sql` | tabel + 3 policy `weekly_standings` |
| 20 | 2026-11-16 10:50:09 | `migration-xp-fix-20261026.sql` | `calculate_level`, `check_user_xp_level` (efektif) |

**Dua file dikecualikan** (`fix-supabase-permissions.sql`, `setup-new-database.sql`): keduanya dianggap tidak dijalankan di prod. Buktinya ada di `migration-fix-users-column-grants.sql`, file yang ditulis sesudahnya:

- § JEBAKAN (baris 96-101) melarang menjalankan kedua file itu karena `GRANT ALL ON ALL TABLES ... TO anon` membuka lagi skema yang "sudah dikunci ketat".
- § 3 (baris 85-92) mencatat hasil uji di prod: "praktikan PATCH row lain → 200 []" (RLS menyaring). Hasil itu mustahil bila `"Allow update users" FOR UPDATE USING (true)` dari kedua file aktif.
- Komentar § SEBAB (baris 9) menyebut `authenticated` hanya punya GRANT UPDATE pada 9 kolom. Itu juga bertentangan dengan `GRANT ALL ON ALL TABLES` di kedua file.

Policy dari kedua file didaftar terpisah di § Kebijakan efektif → "Policy di file yang dikecualikan" dan di § Status tidak pasti.

**Aturan OR permissive.** Semua policy di repo bersifat PERMISSIVE (default). PostgreSQL menggabungkan policy permissive yang berlaku untuk satu perintah dengan OR. Policy lama yang tidak di-DROP oleh file berikutnya tetap aktif. Contoh: `"Semua user terautentikasi bisa membaca data seluruh user"` (`migration-fix-leaderboard-rls.sql:18`, `USING (true)`) tetap aktif setelah `migration-users-rls.sql:23` menambah policy SELECT yang lebih sempit, sehingga SELECT `users` efektif terbuka untuk semua token.

**Aturan Postgres yang ikut menentukan hasil** (dokumentasi `CREATE POLICY`, tabel "Policies Applied by Command Type"):

- Policy `FOR ALL` tanpa `WITH CHECK` memakai ekspresi `USING` sebagai `WITH CHECK`.
- `UPDATE`/`DELETE` dengan `WHERE` juga menyaring baris lewat policy SELECT. Baris yang tidak terlihat lewat SELECT tidak bisa di-UPDATE/DELETE (0 baris, tanpa error).
- `INSERT ... ON CONFLICT DO UPDATE` (upsert PostgREST) memeriksa baris baru terhadap policy SELECT dan policy INSERT `WITH CHECK`. Kalau gagal, statement ditolak dengan error (bukan disaring).
- `WITH CHECK` dievaluasi sesudah trigger BEFORE, jadi trigger yang mengubah `NEW` ikut menentukan hasil cek.
- Tabel ber-RLS tanpa policy untuk suatu perintah = default deny (0 baris untuk UPDATE/DELETE, error untuk INSERT).

## Fungsi bantu

### `auth_nim()`

Satu-satunya definisi di repo ada di `setup-new-database.sql:11` (file dikecualikan), padahal dipakai oleh puluhan policy aktif. Isinya: `COALESCE(claims->>'nim', claims->>'sub', claims->>'username')`, dan `NULL` bila terjadi exception. Token re-sign elearning selalu membawa `nim` dan `sub = nim`, jadi `auth_nim()` = claim `nim`. Di Go: `Principal.NIM`.

### `auth_role()` versi 1 (efektif)

`fix-role-update-trigger.sql:9`, teks fungsinya sama persis dengan `migration-fix-role-claim.sql:11`:

1. `current_user IN ('postgres','supabase_admin')` → `'admin'`.
2. `request.jwt.claims` NULL atau `''` → `'admin'`.
3. `claims->>'role' = 'service_role'` → `'admin'`.
4. `user_role = 'koordinator'` → `'kordas'`; `user_role = 'user'` → `'praktikan'`; nilai lain dikembalikan mentah; `user_role` tidak ada → `'praktikan'`.

### `auth_role()` versi 2 (file dikecualikan)

`setup-new-database.sql:24`: langkah 1-3 sama, lalu `lower(trim(coalesce(user_role, role, 'praktikan')))` dipetakan: `admin|superadmin|super_admin|administrator|super admin` → admin, `kordas|koordinator|korda|coordinator` → kordas, `asisten|assistant|laboran|ast` → asisten, lainnya → praktikan. Exception → praktikan.

### Keduanya identik untuk token re-sign

Token re-sign Go selalu berisi `role:'authenticated'` dan `user_role ∈ {admin, kordas, asisten, praktikan}` (keputusan 12). Untuk keempat nilai itu versi 1 mengembalikan nilai mentah, dan versi 2 memetakan setiap nilai ke dirinya sendiri. Bila `user_role` tidak ada, versi 1 memberi `praktikan` dan versi 2 jatuh ke `role` = `authenticated` → `praktikan`. Hasilnya sama. Perbedaan hanya muncul untuk nilai non-kanonik (`superadmin`, `laboran`, huruf besar), yang tidak pernah ditandatangani server. Pengecualian: tanpa secret Supabase, token asli dikembalikan apa adanya (keputusan 7), dan claim `user_role`-nya bisa non-kanonik. Peran token di Go tetap dinormalisasi dengan `normalizeRole` sebelum aturan di bawah dipakai.

### Koneksi Go

Koneksi Go memakai role `postgres`, jadi `auth_role()` = `'admin'` di semua trigger dan view, dan RLS dilewati sebagai pemilik tabel. Karena itu setiap aturan di § Aturan usecase dicek di Go memakai peran token. Trigger yang tidak memakai `auth_role()` tetap berjalan di DB (lihat § Trigger & view).

## Kebijakan efektif

Kolom "TO" kosong berarti `PUBLIC` (termasuk `anon`). Semua endpoint Go elearning yang menyentuh tabel ini butuh token, jadi perbedaan `PUBLIC` vs `authenticated` tidak muncul di Go.

### users (RLS aktif: `migration-users-rls.sql:8`)

| Operasi | Policy aktif | File:baris | TO | USING | WITH CHECK |
|---|---|---|---|---|---|
| SELECT | Semua user terautentikasi bisa membaca data seluruh user | `migration-fix-leaderboard-rls.sql:18` | authenticated | `true` | — |
| SELECT | Semua user terautentikasi bisa membaca profil sendiri dan praktikan lain | `migration-users-rls.sql:23` | authenticated | `nim = auth_nim() OR role = 'praktikan' OR auth_role() IN ('admin','kordas','asisten')` | — |
| INSERT | Praktikan hanya bisa membuat data miliknya sendiri | `migration-fix-leaderboard-rls.sql:25` | authenticated | — | `nim = auth_nim()` |
| INSERT | User terautentikasi bisa mendaftarkan profil sendiri | `migration-users-rls.sql:34` | authenticated | — | `nim = auth_nim() AND role = auth_role()` |
| INSERT | Staf bisa mendaftarkan profil praktikan | `migration-users-rls.sql:41` | authenticated | — | `auth_role() IN ('admin','kordas')` |
| UPDATE | Praktikan hanya bisa mengubah data miliknya sendiri | `migration-fix-leaderboard-rls.sql:32` | authenticated | `nim = auth_nim()` | `nim = auth_nim()` |
| UPDATE | User terautentikasi bisa memperbarui profil sendiri | `migration-users-rls.sql:51` | authenticated | `nim = auth_nim()` | `nim = auth_nim()` |
| UPDATE | Staf bisa memperbarui profil praktikan | `migration-users-rls.sql:56` | authenticated | `auth_role() IN ('admin','kordas')` | `true` |
| DELETE | Staf bisa menghapus profil praktikan | `migration-users-rls.sql:63` | authenticated | `auth_role() IN ('admin','kordas')` | — |
| ALL | Admin & Kordas bisa mengelola seluruh user | `migration-fix-leaderboard-rls.sql:41` | authenticated | `auth_role() IN ('admin','kordas')` | `auth_role() IN ('admin','kordas')` |

Efektif: SELECT semua baris. INSERT `nim = auth_nim() OR A/K`. UPDATE baris milik sendiri `OR A/K`; `WITH CHECK` gabungan selalu lolos karena ada `WITH CHECK (true)`. DELETE A/K. Ditambah column grant (§ Column grant users) dan trigger (§ Trigger & view).

### levels (RLS: dimatikan `migration-fix-rls.sql:12`, diaktifkan lagi `migration-final-security.sql:8`)

| Operasi | Policy aktif | File:baris | TO | USING | WITH CHECK |
|---|---|---|---|---|---|
| SELECT | Semua user terautentikasi bisa melihat level | `migration-harden-rls.sql:35` | authenticated | `true` | — |
| ALL | Hanya Admin & Kordas yang bisa mengelola level | `migration-harden-rls.sql:37` | authenticated | `auth_role() IN ('admin','kordas')` | (= USING) |

### modules (RLS: dimatikan `migration-fix-rls.sql:13`, diaktifkan lagi `migration-final-security.sql:9`)

| Operasi | Policy aktif | File:baris | TO | USING | WITH CHECK |
|---|---|---|---|---|---|
| SELECT | Semua user terautentikasi bisa melihat modul | `migration-harden-rls.sql:48` | authenticated | `true` | — |
| ALL | Hanya Admin & Kordas yang bisa mengelola modul | `migration-harden-rls.sql:50` | authenticated | `auth_role() IN ('admin','kordas')` | (= USING) |

### lessons (RLS: dimatikan `migration-fix-rls.sql:14`, diaktifkan lagi `migration-final-security.sql:10`)

| Operasi | Policy aktif | File:baris | TO | USING | WITH CHECK |
|---|---|---|---|---|---|
| SELECT | Staf bisa melihat pelajaran secara penuh | `migration-xp-security.sql:82` | authenticated | `auth_role() IN ('admin','kordas','asisten')` | — |
| ALL | Hanya Admin & Kordas yang bisa mengelola pelajaran | `migration-harden-rls.sql:63` | authenticated | `auth_role() IN ('admin','kordas')` | (= USING) |

`"Semua user terautentikasi bisa melihat pelajaran"` (`migration-harden-rls.sql:61`) di-DROP oleh `migration-xp-security.sql:80`. P tidak bisa membaca tabel `lessons` sama sekali, termasuk lewat embed PostgREST (`lessons:lesson_id(xp_reward)` → `null`).

### student_lessons (view)

Definisi efektif `migration-flowchart-solution-visible.sql:13`. `GRANT SELECT ... TO authenticated, anon` (`:36`). View biasa (bukan `security_invoker`), jadi membaca `lessons` dengan hak pemilik view dan tidak terkena RLS `lessons`. Isinya semua baris `lessons` dengan kolom `id, module_id, title, explanation, code_example, initial_code, hint, quiz::jsonb - 'correctAnswer' AS quiz, test_cases, validation_rules, sort_order, xp_reward, exercise_type, flowchart_distractors, solution`. `solution` diisi bila `exercise_type = 'flowchart'` atau `auth_role() IN ('admin','kordas','asisten')`, selain itu `NULL`. View tidak punya policy.

### student_progress (RLS aktif: `migration-harden-rls.sql:22`)

| Operasi | Policy aktif | File:baris | TO | USING | WITH CHECK |
|---|---|---|---|---|---|
| SELECT | Mahasiswa hanya bisa melihat progresnya sendiri | `migration-harden-rls.sql:207` | authenticated | `nim = auth_nim()` | — |
| SELECT | Staf bisa melihat seluruh progres mahasiswa | `migration-harden-rls.sql:211` | authenticated | `auth_role() IN ('admin','kordas','asisten')` | — |
| INSERT | Mahasiswa hanya bisa menambah progresnya sendiri | `migration-harden-rls.sql:216` | authenticated | — | `nim = auth_nim()` |
| ALL | Staf bisa mengelola progres belajar seluruh mahasiswa | `migration-harden-rls.sql:221` | authenticated | `auth_role() IN ('admin','kordas','asisten')` | (= USING) |

Efektif: SELECT milik sendiri `OR A/K/S`. INSERT milik sendiri `OR A/K/S`. UPDATE/DELETE hanya A/K/S (P tidak bisa mengubah atau menghapus progresnya sendiri).

### game_questions (RLS: dimatikan `migration-fix-rls.sql:15`, diaktifkan lagi `migration-final-security.sql:11`)

| Operasi | Policy aktif | File:baris | TO | USING | WITH CHECK |
|---|---|---|---|---|---|
| SELECT | Semua user terautentikasi bisa melihat game questions | `migration-harden-rls.sql:74` | authenticated | `true` | — |
| ALL | Hanya Admin & Kordas yang bisa mengelola game questions | `migration-harden-rls.sql:76` | authenticated | `auth_role() IN ('admin','kordas')` | (= USING) |

### game_settings (RLS: dimatikan `migration-fix-rls.sql:16`, diaktifkan lagi `migration-final-security.sql:12`)

| Operasi | Policy aktif | File:baris | TO | USING | WITH CHECK |
|---|---|---|---|---|---|
| SELECT | Semua user terautentikasi bisa melihat game settings | `migration-harden-rls.sql:87` | authenticated | `true` | — |
| ALL | Hanya Admin & Kordas yang bisa mengelola game settings | `migration-harden-rls.sql:89` | authenticated | `auth_role() IN ('admin','kordas')` | (= USING) |

### game_history (RLS aktif: `migration-final-security.sql:108`)

| Operasi | Policy aktif | File:baris | TO | USING | WITH CHECK |
|---|---|---|---|---|---|
| SELECT | Mahasiswa hanya bisa melihat riwayat game sendiri | `migration-harden-rls.sql:144` | authenticated | `nim = auth_nim()` | — |
| SELECT | Staf bisa melihat riwayat game seluruh praktikan | `migration-harden-rls.sql:146` | authenticated | `auth_role() IN ('admin','kordas','asisten')` | — |
| INSERT | Mahasiswa hanya bisa mencatat riwayat game sendiri | `migration-harden-rls.sql:148` | authenticated | — | `nim = auth_nim()` |

UPDATE/DELETE tanpa policy: default deny untuk semua peran.

### achievements (RLS aktif: `migration-final-security.sql:125`)

| Operasi | Policy aktif | File:baris | TO | USING | WITH CHECK |
|---|---|---|---|---|---|
| SELECT | Semua user terautentikasi bisa melihat pencapaian | `migration-harden-rls.sql:113` | authenticated | `true` | — |
| ALL | Hanya Admin & Kordas yang bisa mengelola pencapaian | `migration-harden-rls.sql:115` | authenticated | `auth_role() IN ('admin','kordas')` | (= USING) |

### unlocked_achievements (RLS aktif: `migration-final-security.sql:126`)

| Operasi | Policy aktif | File:baris | TO | USING | WITH CHECK |
|---|---|---|---|---|---|
| SELECT | Mahasiswa hanya bisa melihat pencapaian miliknya sendiri | `migration-harden-rls.sql:127` | authenticated | `nim = auth_nim()` | — |
| SELECT | Staf bisa melihat pencapaian seluruh praktikan | `migration-harden-rls.sql:129` | authenticated | `auth_role() IN ('admin','kordas','asisten')` | — |
| INSERT | Mahasiswa hanya bisa mencatat pencapaian miliknya sendiri | `migration-harden-rls.sql:131` | authenticated | — | `nim = auth_nim()` |
| DELETE | Staf bisa menghapus pencapaian praktikan | `migration-harden-rls.sql:133` | authenticated | `auth_role() IN ('admin','kordas','asisten')` | — |

UPDATE tanpa policy: default deny.

### playground_examples (RLS aktif: `migration-harden-rls.sql:19`)

| Operasi | Policy aktif | File:baris | TO | USING | WITH CHECK |
|---|---|---|---|---|---|
| SELECT | Semua user terautentikasi bisa melihat contoh playground | `migration-harden-rls.sql:101` | authenticated | `true` | — |
| ALL | Hanya Admin & Kordas yang bisa mengelola contoh playground | `migration-harden-rls.sql:103` | authenticated | `auth_role() IN ('admin','kordas')` | (= USING) |

`"Anyone can read examples"` dan `"Admin can manage examples"` di-DROP oleh `migration-harden-rls.sql:96`/`:97`. Keduanya tidak dibuat oleh file mana pun di repo.

### active_sessions (RLS aktif: `migration-harden-rls.sql:24`)

| Operasi | Policy aktif | File:baris | TO | USING | WITH CHECK |
|---|---|---|---|---|---|
| SELECT | Staf bisa memonitor seluruh sesi aktif | `migration-harden-rls.sql:163` | authenticated | `auth_role() IN ('admin','kordas','asisten')` | — |
| INSERT | Mahasiswa bisa mendaftarkan sesi miliknya | `migration-harden-rls.sql:165` | authenticated | — | `nim = auth_nim()` |
| UPDATE | Mahasiswa bisa memperbarui heartbeat miliknya | `migration-harden-rls.sql:167` | authenticated | `nim = auth_nim()` | `nim = auth_nim()` |
| DELETE | Mahasiswa bisa menghapus sesi miliknya | `migration-harden-rls.sql:169` | authenticated | `nim = auth_nim()` | — |
| DELETE | Staf bisa menghapus sesi aktif apa saja | `migration-harden-rls.sql:171` | authenticated | `auth_role() IN ('admin','kordas','asisten')` | — |

Akibat aturan Postgres di § Sumber & urutan: P tidak lolos policy SELECT, sehingga upsert heartbeat P (`Layout.tsx:198`, `:237`, `ON CONFLICT (nim) DO UPDATE`) ditolak dengan error, dan `DELETE ... WHERE nim = <sendiri>` (`Layout.tsx:70`) menghapus 0 baris. Kode lama menelan kedua hasil itu. Lihat § Status tidak pasti.

### activity_logs (RLS aktif: `migration-harden-rls.sql:25`)

| Operasi | Policy aktif | File:baris | TO | USING | WITH CHECK |
|---|---|---|---|---|---|
| SELECT | Staf bisa membaca seluruh log audit | `migration-harden-rls.sql:186` | authenticated | `auth_role() IN ('admin','kordas','asisten')` | — |
| INSERT | Mahasiswa hanya bisa mencatat log aktivitas miliknya | `migration-harden-rls.sql:188` | authenticated | — | `nim = auth_nim()` |
| DELETE | Admin & Kordas bisa menghapus log audit | `migration-harden-rls.sql:190` | authenticated | `auth_role() IN ('admin','kordas')` | — |

UPDATE tanpa policy: default deny.

### prestige_history (RLS aktif: `migration-prestige.sql:47`)

| Operasi | Policy aktif | File:baris | TO | USING | WITH CHECK |
|---|---|---|---|---|---|
| SELECT | Mahasiswa baca prestige sendiri, staff baca semua | `migration-prestige.sql:61` | authenticated | `nim = auth_nim() OR auth_role() IN ('admin','kordas','asisten')` | — |
| ALL | Staff kelola prestige_history | `migration-prestige.sql:66` | authenticated | `auth_role() IN ('admin','kordas')` | `auth_role() IN ('admin','kordas')` |

### progress_archive (RLS aktif: `migration-prestige.sql:46`)

| Operasi | Policy aktif | File:baris | TO | USING | WITH CHECK |
|---|---|---|---|---|---|
| SELECT | Mahasiswa baca arsip sendiri, staff baca semua | `migration-prestige.sql:50` | authenticated | `nim = auth_nim() OR auth_role() IN ('admin','kordas','asisten')` | — |
| ALL | Staff kelola arsip progres | `migration-prestige.sql:55` | authenticated | `auth_role() IN ('admin','kordas')` | `auth_role() IN ('admin','kordas')` |

Ditulis hanya oleh `do_prestige_reset` (SECURITY DEFINER, melewati RLS).

### weekly_standings (RLS aktif: `migration-weekly-standings.sql:34`)

| Operasi | Policy aktif | File:baris | TO | USING | WITH CHECK |
|---|---|---|---|---|---|
| SELECT | Semua user bisa baca weekly_standings | `migration-weekly-standings.sql:37` | authenticated | `true` | — |
| INSERT | Staff bisa arsip weekly_standings | `migration-weekly-standings.sql:41` | authenticated | — | `auth_role() IN ('admin','kordas','asisten')` |
| UPDATE | Staff bisa update weekly_standings | `migration-weekly-standings.sql:46` | authenticated | `auth_role() IN ('admin','kordas','asisten')` | `auth_role() IN ('admin','kordas','asisten')` |

Semua pemakainya kode mati (keputusan 15), jadi tidak ada endpoint.

### assessment_grading_rules (RLS aktif: `migration-assessment-settings.sql:147`)

| Operasi | Policy aktif | File:baris | TO | USING | WITH CHECK |
|---|---|---|---|---|---|
| SELECT | Allow authenticated read grading rules | `migration-assessment-settings.sql:150` | authenticated | `true` | — |
| ALL | Allow admin update grading rules | `migration-assessment-settings.sql:154` | authenticated | `true` | `true` |

Nama policy kedua menyebut admin, tetapi ekspresinya `true`, jadi semua token boleh menulis. Go hanya membacanya (`POST /grade`).

### storage.objects, bucket `lesson-images`

| Operasi | Policy aktif | File:baris | TO | USING | WITH CHECK |
|---|---|---|---|---|---|
| SELECT | Izinkan publik membaca gambar materi | `migration-final-security.sql:72` | | `bucket_id = 'lesson-images'` | — |
| INSERT | Izinkan staf mengunggah gambar materi | `migration-final-security.sql:77` | | — | `bucket_id = 'lesson-images' AND auth_role() IN ('admin','kordas','asisten')` |
| UPDATE | Izinkan staf memperbarui gambar materi | `migration-final-security.sql:85` | | `bucket_id = 'lesson-images' AND auth_role() IN ('admin','kordas','asisten')` | (= USING) |
| DELETE | Izinkan staf menghapus gambar materi | `migration-final-security.sql:93` | | `bucket_id = 'lesson-images' AND auth_role() IN ('admin','kordas','asisten')` | — |

Go mengunggah lewat `pkg/supabase.Client` dengan service role key, yang melewati policy ini, jadi cek staf dilakukan di usecase.

### Tabel tanpa policy di repo

`assessment_attempts`, `assessment_questions` (dipakai `api/grade.ts` dengan token pengguna) dan `elearning_progress` (`api/report.ts`, service role key). Lihat § Status tidak pasti.

### Policy yang di-DROP oleh file berikutnya (25)

| CREATE POLICY | Tabel | Di-DROP oleh |
|---|---|---|
| `migration-fix-leaderboard-rls.sql:50` Kordas bisa menghapus seluruh log audit | activity_logs | `migration-harden-rls.sql:180` |
| `migration-final-security.sql:18` Semua user bisa melihat level | levels | `migration-harden-rls.sql:31` |
| `migration-final-security.sql:20` Hanya Admin & Kordas yang bisa mengelola level | levels | `migration-harden-rls.sql:33` (dibuat ulang `:37`) |
| `migration-final-security.sql:27` Semua user bisa melihat modul | modules | `migration-harden-rls.sql:44` |
| `migration-final-security.sql:29` Hanya Admin & Kordas yang bisa mengelola modul | modules | `migration-harden-rls.sql:46` (dibuat ulang `:50`) |
| `migration-final-security.sql:36` Semua user bisa melihat pelajaran | lessons | `migration-harden-rls.sql:57` |
| `migration-final-security.sql:38` Hanya Admin & Kordas yang bisa mengelola pelajaran | lessons | `migration-harden-rls.sql:59` (dibuat ulang `:63`) |
| `migration-final-security.sql:45` Semua user bisa melihat game questions | game_questions | `migration-harden-rls.sql:70` |
| `migration-final-security.sql:47` Hanya Admin & Kordas yang bisa mengelola game questions | game_questions | `migration-harden-rls.sql:72` (dibuat ulang `:76`) |
| `migration-final-security.sql:54` Semua user bisa melihat game settings | game_settings | `migration-harden-rls.sql:83` |
| `migration-final-security.sql:56` Hanya Admin & Kordas yang bisa mengelola game settings | game_settings | `migration-harden-rls.sql:85` (dibuat ulang `:89`) |
| `migration-final-security.sql:63` Hanya Admin yang bisa menghapus log audit | activity_logs | `migration-harden-rls.sql:179` |
| `migration-final-security.sql:103` Staf bisa menghapus sesi aktif apa saja | active_sessions | `migration-harden-rls.sql:157` (dibuat ulang `:171`) |
| `migration-final-security.sql:111` Mahasiswa hanya bisa melihat riwayat game sendiri | game_history | `migration-harden-rls.sql:140` (dibuat ulang `:144`) |
| `migration-final-security.sql:115` Staf bisa melihat riwayat game seluruh praktikan | game_history | `migration-harden-rls.sql:141` (dibuat ulang `:146`) |
| `migration-final-security.sql:119` Mahasiswa hanya bisa mencatat riwayat game sendiri | game_history | `migration-harden-rls.sql:142` (dibuat ulang `:148`) |
| `migration-final-security.sql:129` Semua user terautentikasi bisa melihat pencapaian | achievements | `migration-harden-rls.sql:110` (dibuat ulang `:113`) |
| `migration-final-security.sql:133` Hanya Admin & Kordas yang bisa mengelola pencapaian | achievements | `migration-harden-rls.sql:111` (dibuat ulang `:115`) |
| `migration-final-security.sql:137` Mahasiswa hanya bisa melihat pencapaian miliknya sendiri | unlocked_achievements | `migration-harden-rls.sql:122` (dibuat ulang `:127`) |
| `migration-final-security.sql:141` Staf bisa melihat pencapaian seluruh praktikan | unlocked_achievements | `migration-harden-rls.sql:123` (dibuat ulang `:129`) |
| `migration-final-security.sql:145` Mahasiswa hanya bisa mencatat pencapaian miliknya sendiri | unlocked_achievements | `migration-harden-rls.sql:124` (dibuat ulang `:131`) |
| `migration-final-security.sql:149` Staf bisa menghapus pencapaian praktikan | unlocked_achievements | `migration-harden-rls.sql:125` (dibuat ulang `:133`) |
| `migration-final-security.sql:178` Mahasiswa hanya bisa modifikasi progress-nya sendiri | student_progress | `migration-harden-rls.sql:199` |
| `migration-final-security.sql:181` Staf bisa mengelola progres belajar praktikan | student_progress | `migration-harden-rls.sql:200` |
| `migration-harden-rls.sql:61` Semua user terautentikasi bisa melihat pelajaran | lessons | `migration-xp-security.sql:80` |

### Policy yang tidak pernah jalan: di dalam blok komentar (10)

`migration-fix-rls.sql:37, :39, :45, :47, :53, :55, :61, :63, :69, :71` (Pilihan B, di dalam `/* ... */` baris 27-72): "Semua user bisa melihat level", "Hanya Admin & Kordas yang bisa mengelola level", dan pasangan yang sama untuk modules, lessons, game questions, game settings.

### Policy di file yang dikecualikan (46)

`fix-supabase-permissions.sql` (23): `:57` Allow select users, `:60` Allow insert users, `:63` Allow update users, `:66` Allow delete users, `:74` Allow all for active_sessions, `:80` Public select levels, `:82` Admin manage levels, `:85` Public select modules, `:87` Admin manage modules, `:90` Public select lessons, `:92` Admin manage lessons, `:95` Public select game_questions, `:97` Admin manage game_questions, `:100` Public select game_settings, `:102` Admin manage game_settings, `:105` Public select playground_examples, `:107` Admin manage playground_examples, `:110` Public select achievements, `:112` Admin manage achievements, `:117` Manage student_progress, `:120` Manage unlocked_achievements, `:123` Manage game_history, `:126` Manage activity_logs.

`setup-new-database.sql` (23): `:280` Allow select users, `:281` Allow insert users, `:282` Allow update users, `:283` Allow delete users, `:287` Allow all for active_sessions, `:291` Public select levels, `:293` Admin manage levels, `:296` Public select modules, `:298` Admin manage modules, `:301` Public select lessons, `:303` Admin manage lessons, `:307` Public select game_questions, `:309` Admin manage game_questions, `:312` Public select game_settings, `:314` Admin manage game_settings, `:317` Public select achievements, `:319` Admin manage achievements, `:322` Public select playground_examples, `:324` Admin manage playground_examples, `:328` Manage student_progress, `:331` Manage unlocked_achievements, `:334` Manage game_history, `:337` Manage activity_logs.

Kalau salah satu file ini ternyata dijalankan di prod, `users`, `active_sessions`, `student_progress`, `unlocked_achievements`, `game_history`, dan `activity_logs` praktis terbuka (`USING (true)`), dan `fix-supabase-permissions.sql:51`-`:55` men-DROP lima policy `users` dari `migration-users-rls.sql`.

### Rekap hitungan

`Select-String -Path E:/rekapProject/project_ap/elearning/docs/*.sql -Pattern 'CREATE POLICY'` = 137 baris:

| Status | Jumlah |
|---|---|
| Aktif (tabel di atas: users 10, levels 2, modules 2, lessons 2, student_progress 4, game_questions 2, game_settings 2, game_history 3, achievements 2, unlocked_achievements 4, playground_examples 2, active_sessions 5, activity_logs 3, prestige_history 2, progress_archive 2, weekly_standings 3, assessment_grading_rules 2, storage.objects 4) | 56 |
| Di-DROP oleh file berikutnya | 25 |
| Di dalam blok komentar `migration-fix-rls.sql` | 10 |
| Di file yang dikecualikan | 46 |
| **Total** | **137** |

## Column grant users

Sumber: komentar `migration-fix-users-column-grants.sql:9`-`:30` (tidak pasti, grant aslinya tidak ada di repo).

- UPDATE untuk `authenticated` hanya pada kolom `role, nama, xp, level, streak, study_time, last_active, kelas, jurusan` (komentar baris 9-10), ditambah `elearning_role, level_access_overrides` (`:22`).
- SELECT `elearning_role` (`:28`). Komentar baris 25 menyebut "15 kolom lain yang grant-nya sudah ada", tanpa daftar.
- DELETE pada tabel `users` (`:30`).
- `migration-prestige.sql:111` hanya menyarankan `REVOKE UPDATE (prestige_mult, prestige_tier, prestige_count, last_prestige_at)`. Kolom itu memang tidak ada di daftar UPDATE di atas.

PostgREST menolak seluruh statement (403, `42501 permission denied for table users`) bila payload UPDATE memuat satu saja kolom di luar daftar. Grant berlaku untuk role DB `authenticated`, jadi berlaku untuk semua peran token termasuk A/K. Di Go hanya grant UPDATE yang ditiru (keputusan 4, aturan `R-users-UPDATE-2`). Grant SELECT/INSERT tidak ditiru karena daftarnya tidak diketahui.

## Trigger & view yang ditiru di Go

Urutan trigger BEFORE pada `users` (Postgres menjalankan trigger BEFORE berurutan alfabet):

1. `trigger_canonicalize_user_role` (`migration-canonical-role.sql:30`, `BEFORE INSERT OR UPDATE OF role`): `admin|kordas|korda` → `koordinator`, `praktikan|mahasiswa` → `mahasiswa`, `asisten`, `koordinator` tetap; nilai lain dibiarkan. Input di-`lower(trim())`. **Tetap berjalan di DB**, tidak ditiru, tetapi Go harus menerapkan pemetaan yang sama pada nilai `role` baru sebelum mengevaluasi `check_user_updates`, karena di DB trigger ini jalan lebih dulu.
2. `trigger_check_user_updates` (`BEFORE UPDATE`, fungsi efektif `migration-fix-users-column-grants.sql:40`, trigger `:78`): **ditiru**.
3. `trigger_check_user_xp_level` (`BEFORE INSERT OR UPDATE`, trigger `migration-prestige.sql:176`, fungsi efektif `migration-xp-fix-20261026.sql:54`): **ditiru**.
4. `trigger_guard_prestige_columns` (`BEFORE UPDATE`, `migration-prestige.sql:115`, trigger `:130`): **ditiru**.

### check_user_updates (efektif `migration-fix-users-column-grants.sql:40`)

Dilewati bila `current_user` postgres/supabase_admin, tanpa claims, atau `auth_role() IN ('admin','kordas','service_role')`. Untuk peran lain, dicek berurutan dan berhenti di exception pertama (SQLSTATE `P0001`):

1. `NEW.role IS DISTINCT FROM OLD.role` dan `NEW.role IS DISTINCT FROM auth_role()` → `Akses Ditolak: Anda tidak diperbolehkan mengubah peran (role) Anda sendiri!`. `NEW.role` di sini sudah dikanonikkan trigger 1, sedangkan `auth_role()` bernilai `praktikan`/`asisten`. Artinya P tidak pernah bisa mengubah `role` (nilai kanonik tidak pernah `praktikan`), sedangkan S hanya boleh mengubahnya menjadi `asisten`.
2. `level_access_overrides` berubah → `Akses Ditolak: Anda tidak memiliki wewenang untuk mengubah level_access_overrides!`
3. `assessment_access` berubah → `Akses Ditolak: Anda tidak memiliki wewenang untuk mengubah assessment_access!`
4. `elearning_role` berubah → `Akses Ditolak: Anda tidak memiliki wewenang untuk mengubah elearning_role!`

Versi sebelumnya (`migration-users-rls.sql:73`, `fix-role-update-trigger.sql:42`) tidak punya cek 4.

### check_user_xp_level (efektif `migration-xp-fix-20261026.sql:54`)

Untuk `auth_role() NOT IN ('admin','kordas')` (S dan P), pada INSERT maupun UPDATE:

- `lesson_base = COALESCE(SUM(student_lessons.xp_reward), 0)` dari `student_progress sp JOIN student_lessons l ON sp.lesson_id = l.id WHERE sp.nim = NEW.nim AND sp.completed`.
- `game_sum = COALESCE(SUM(game_history.xp_earned), 0) WHERE nim = NEW.nim`.
- `mult = clamp(COALESCE(NEW.prestige_mult, 1.00), 1.00, 1.50)`, lalu `NEW.prestige_mult := mult`.
- `NEW.xp := FLOOR(lesson_base * mult)::int + game_sum` (nilai dari klien selalu ditimpa; di atas cap hanya `RAISE NOTICE`, tanpa error).
- `NEW.level := calculate_level(NEW.xp)` = `1` bila `xp` NULL atau ≤ 0, selain itu `FLOOR(SQRT(xp::numeric / 50.0))::int + 1` (`migration-xp-fix-20261026.sql:40`).

A/K: `xp` dan `level` disimpan apa adanya, `level` tidak dihitung ulang. Di Go penimpaan dilakukan dalam satu statement UPDATE dengan subquery (keputusan 3). Fungsi trigger sengaja bukan SECURITY DEFINER (komentar `:135`-`:141`).

### guard_prestige_columns (`migration-prestige.sql:115`)

Untuk `auth_role() NOT IN ('admin','kordas')`: bila `prestige_mult`, `prestige_tier`, atau `prestige_count` berubah → `Akses Ditolak: kolom prestige hanya bisa diubah lewat Prestige resmi (tamat 6/6).` (`P0001`). `last_prestige_at` tidak dicek, walaupun komentar `:112`-`:113` menyebutnya. Trigger ini jalan sesudah `check_user_xp_level`, jadi yang dibandingkan adalah `prestige_mult` yang sudah di-clamp. Bila `OLD.prestige_mult` di luar 1.00-1.50, setiap UPDATE oleh S/P akan ditolak.

### View student_lessons (efektif `migration-flowchart-solution-visible.sql:13`)

- `quiz` = `quiz::jsonb - 'correctAnswer'` untuk semua pembaca (hanya kunci tingkat atas).
- `solution` = isi bila `exercise_type = 'flowchart'`; bila tidak, hanya untuk `auth_role() IN ('admin','kordas','asisten')`, selain itu `NULL`.
- Dari koneksi Go `auth_role()` = admin, jadi `solution` selalu terisi. Go menerapkan aturan `solution` sendiri berdasarkan peran token.

### Trigger dan fungsi yang tetap berjalan di DB (tidak ditiru)

- `canonicalize_user_role` (di atas).
- `check_progress_velocity` (`migration-harden-rls.sql:232`, trigger `BEFORE INSERT ON student_progress` `:290`): jeda minimal 1 menit antar penyelesaian dan maksimal 1 level baru per 24 jam. Pesan `Aktivitas Mencurigakan: Anda menyelesaikan materi terlalu cepat! Harap baca materi dengan seksama.` dan `Batas Tingkat Harian Terlampaui: Anda hanya diperbolehkan mempelajari maksimal 1 level dalam 24 jam. Harap kuasai materi hari ini sebelum melaju ke tingkat berikutnya.` Tidak memakai `auth_role()`, jadi tetap berlaku untuk insert dari Go (semua peran). Go memetakan `P0001` → 400 dengan pesan apa adanya.
- `update_user_last_active_from_session` (`migration-final-security.sql:154`, SECURITY DEFINER, `AFTER INSERT OR UPDATE ON active_sessions` `:165`): menyalin `last_heartbeat` ke `users.last_active`. Karena SECURITY DEFINER, trigger `users` melihat `auth_role()` = admin dan tidak menimpa XP.
- `do_prestige_reset(p_nim)` (`migration-prestige.sql:187`, SECURITY DEFINER): tidak mengecek siapa pemanggilnya, jadi token mana pun bisa mereset NIM mana pun yang memenuhi syarat. Exception-nya (`NIM % tidak ditemukan.`, `Kurikulum belum 6 level ...`, `Syarat belum terpenuhi: ...`, `Cooldown: ...`, `Kuota semester habis (2/2). ...`) diteruskan sebagai 400. Guard pemanggil ada di service lama `prestigeService.doVoluntaryPrestigeReset`: bukan diri sendiri dan bukan staf (peran efektif UI A/K/S) → `Akses ditolak.`.
- `calculate_level`, `jakarta_semester`, `prestige_count_this_semester`, `jakarta_week_start`.

## Aturan usecase

Semantik hasil (keputusan 1):

- SELECT disaring → baris disaring atau hasil kosong, 200.
- UPDATE/DELETE gagal `USING` → 0 baris terpengaruh, sukses 200.
- INSERT/UPSERT gagal `WITH CHECK` → 403 `new row violates row-level security policy for table "<tabel>"`.
- Exception trigger (`P0001`) → 400 dengan pesan exception.
- Guard service lama → 403 dengan pesan lama, dicek sebelum aturan RLS (memakai peran efektif UI, keputusan 2).

Peran pada aturan di bawah = peran token. Nama test ada di `internal/usecase/elearning/rls_test.go`. ID memakai tanda `-`, sedangkan nama fungsi Go tidak boleh, jadi `-` diganti `_`: `R-users-UPDATE-1` → `TestRLS_R_users_UPDATE_1_Diizinkan` / `TestRLS_R_users_UPDATE_1_Ditolak`. Untuk aturan yang terbuka bagi semua token, kasus "Ditolak" adalah principal kosong (tanpa token) → 401, karena policy-nya `TO authenticated`. Aturan tanpa endpoint (kode mati atau tidak ada pemanggil) tetap dicatat. Test-nya hanya wajib dibuat bila endpointnya ada.

Endpoint `POST /ai`, `POST /report`, `POST /validate-quiz` memakai service role key di kode lama (`api/ai.ts:6`, `api/report.ts:6`, `api/validate-quiz.ts:5`), jadi RLS tidak berlaku dan tidak ada aturan R-* untuk query di dalamnya.

### users

| ID | Diizinkan | Ditolak → hasil | Endpoint | Test |
|---|---|---|---|---|
| R-users-SELECT-1 | semua peran, semua baris | tanpa token → 401 | `POST /profile/sync`, `GET /admin/users`, `GET /monitoring/students`, `GET /leaderboard`, `GET /leaderboard/rank`, `GET /leaderboard/filters`, `GET /leaderboard/weekly`, `POST /grade` | `TestRLS_R_users_SELECT_1_Diizinkan`, `TestRLS_R_users_SELECT_1_Ditolak` |
| R-users-INSERT-1 | baris milik sendiri (semua peran); A/K baris mana pun | S/P untuk NIM lain → 403 `new row violates row-level security policy for table "users"` | `POST /profile/sync` | `TestRLS_R_users_INSERT_1_Diizinkan`, `TestRLS_R_users_INSERT_1_Ditolak` |
| R-users-UPDATE-1 | baris milik sendiri (semua peran); A/K baris mana pun | S/P untuk baris lain → 0 baris, 200 | `POST /profile/sync`, `POST /progress/study-time`, `POST /streak`, `PUT /sessions/me` (touchStreak), `POST /progress/lessons/:lessonId/complete`, `POST /progress/xp`, `POST /game/history`, `PATCH /admin/users/:nim/level-access`, `PATCH /admin/users/:nim/role`, `POST /admin/users/:nim/reset`, `POST /admin/users/:nim/reset-level`, `POST /admin/users/:nim/reset-lesson`, `PUT /admin/users/:nim/xp` | `TestRLS_R_users_UPDATE_1_Diizinkan`, `TestRLS_R_users_UPDATE_1_Ditolak` |
| R-users-UPDATE-2 | semua peran, bila semua kolom ⊆ {role, nama, xp, level, streak, study_time, last_active, kelas, jurusan, elearning_role, level_access_overrides} | ada kolom lain (mis. `shift`) → 403 `permission denied for table users`, seluruh statement batal (tidak pasti) | semua endpoint R-users-UPDATE-1 | `TestRLS_R_users_UPDATE_2_Diizinkan`, `TestRLS_R_users_UPDATE_2_Ditolak` |
| R-users-DELETE-1 | A/K | S/P → 0 baris, 200 | `DELETE /admin/users/:nim` | `TestRLS_R_users_DELETE_1_Diizinkan`, `TestRLS_R_users_DELETE_1_Ditolak` |

### levels, modules, game_questions, game_settings, playground_examples, achievements

Pola sama untuk keenam tabel (`<t>` = nama tabel). Pesan 403 memakai nama tabel masing-masing.

| ID | Diizinkan | Ditolak → hasil | Endpoint | Test |
|---|---|---|---|---|
| R-levels-SELECT-1 | semua peran | tanpa token → 401 | `GET /curriculum/levels` | `TestRLS_R_levels_SELECT_1_Diizinkan`, `TestRLS_R_levels_SELECT_1_Ditolak` |
| R-levels-INSERT-1 | A/K (termasuk upsert) | S/P → 403 `... for table "levels"` | `PUT /curriculum` | `TestRLS_R_levels_INSERT_1_Diizinkan`, `TestRLS_R_levels_INSERT_1_Ditolak` |
| R-levels-UPDATE-1 | A/K | S/P → 0 baris, 200 | — (`updateLevel` kode mati; upsert ikut R-levels-INSERT-1) | `TestRLS_R_levels_UPDATE_1_Diizinkan`, `TestRLS_R_levels_UPDATE_1_Ditolak` |
| R-levels-DELETE-1 | A/K | S/P → 0 baris, 200 | `PUT /curriculum`, `DELETE /curriculum` | `TestRLS_R_levels_DELETE_1_Diizinkan`, `TestRLS_R_levels_DELETE_1_Ditolak` |
| R-modules-SELECT-1 | semua peran | tanpa token → 401 | `GET /curriculum/modules` | `TestRLS_R_modules_SELECT_1_Diizinkan`, `TestRLS_R_modules_SELECT_1_Ditolak` |
| R-modules-INSERT-1 | A/K (termasuk upsert) | S/P → 403 `... for table "modules"` | `PUT /curriculum` | `TestRLS_R_modules_INSERT_1_Diizinkan`, `TestRLS_R_modules_INSERT_1_Ditolak` |
| R-modules-UPDATE-1 | A/K | S/P → 0 baris, 200 | — (upsert ikut R-modules-INSERT-1) | `TestRLS_R_modules_UPDATE_1_Diizinkan`, `TestRLS_R_modules_UPDATE_1_Ditolak` |
| R-modules-DELETE-1 | A/K | S/P → 0 baris, 200 | `PUT /curriculum` | `TestRLS_R_modules_DELETE_1_Diizinkan`, `TestRLS_R_modules_DELETE_1_Ditolak` |
| R-game_questions-SELECT-1 | semua peran | tanpa token → 401 | `GET /game/questions`, `POST /game/questions/seed` (cek kosong) | `TestRLS_R_game_questions_SELECT_1_Diizinkan`, `TestRLS_R_game_questions_SELECT_1_Ditolak` |
| R-game_questions-INSERT-1 | A/K | S/P → 403 `... for table "game_questions"` | `POST /game/questions`, `POST /game/questions/seed`, `PUT /game/questions` | `TestRLS_R_game_questions_INSERT_1_Diizinkan`, `TestRLS_R_game_questions_INSERT_1_Ditolak` |
| R-game_questions-UPDATE-1 | A/K | S/P → 0 baris, 200 | `PATCH /game/questions/:id` | `TestRLS_R_game_questions_UPDATE_1_Diizinkan`, `TestRLS_R_game_questions_UPDATE_1_Ditolak` |
| R-game_questions-DELETE-1 | A/K | S/P → 0 baris, 200 | `DELETE /game/questions/:id`, `PUT /game/questions` | `TestRLS_R_game_questions_DELETE_1_Diizinkan`, `TestRLS_R_game_questions_DELETE_1_Ditolak` |
| R-game_settings-SELECT-1 | semua peran | tanpa token → 401 | `GET /game/settings`, `GET /game/plays/me` | `TestRLS_R_game_settings_SELECT_1_Diizinkan`, `TestRLS_R_game_settings_SELECT_1_Ditolak` |
| R-game_settings-INSERT-1 | A/K (upsert) | S/P → 403 `... for table "game_settings"` | `PUT /game/settings` | `TestRLS_R_game_settings_INSERT_1_Diizinkan`, `TestRLS_R_game_settings_INSERT_1_Ditolak` |
| R-game_settings-UPDATE-1 | A/K | S/P → 0 baris, 200 | — (upsert ikut R-game_settings-INSERT-1) | `TestRLS_R_game_settings_UPDATE_1_Diizinkan`, `TestRLS_R_game_settings_UPDATE_1_Ditolak` |
| R-game_settings-DELETE-1 | A/K | S/P → 0 baris, 200 | — | `TestRLS_R_game_settings_DELETE_1_Diizinkan`, `TestRLS_R_game_settings_DELETE_1_Ditolak` |
| R-playground_examples-SELECT-1 | semua peran | tanpa token → 401 | `GET /playground/examples` | `TestRLS_R_playground_examples_SELECT_1_Diizinkan`, `TestRLS_R_playground_examples_SELECT_1_Ditolak` |
| R-playground_examples-INSERT-1 | A/K | S/P → 403 `... for table "playground_examples"` | `POST /playground/examples` | `TestRLS_R_playground_examples_INSERT_1_Diizinkan`, `TestRLS_R_playground_examples_INSERT_1_Ditolak` |
| R-playground_examples-UPDATE-1 | A/K | S/P → 0 baris, 200 | `PATCH /playground/examples/:id` | `TestRLS_R_playground_examples_UPDATE_1_Diizinkan`, `TestRLS_R_playground_examples_UPDATE_1_Ditolak` |
| R-playground_examples-DELETE-1 | A/K | S/P → 0 baris, 200 | `DELETE /playground/examples/:id` | `TestRLS_R_playground_examples_DELETE_1_Diizinkan`, `TestRLS_R_playground_examples_DELETE_1_Ditolak` |
| R-achievements-SELECT-1 | semua peran | tanpa token → 401 | `GET /achievements`, `POST /achievements/check`, `POST /achievements/check-xp` | `TestRLS_R_achievements_SELECT_1_Diizinkan`, `TestRLS_R_achievements_SELECT_1_Ditolak` |
| R-achievements-INSERT-1 | A/K | S/P → 403 `... for table "achievements"` | — | `TestRLS_R_achievements_INSERT_1_Diizinkan`, `TestRLS_R_achievements_INSERT_1_Ditolak` |
| R-achievements-UPDATE-1 | A/K | S/P → 0 baris, 200 | — | `TestRLS_R_achievements_UPDATE_1_Diizinkan`, `TestRLS_R_achievements_UPDATE_1_Ditolak` |
| R-achievements-DELETE-1 | A/K | S/P → 0 baris, 200 | — | `TestRLS_R_achievements_DELETE_1_Diizinkan`, `TestRLS_R_achievements_DELETE_1_Ditolak` |

### lessons dan student_lessons

| ID | Diizinkan | Ditolak → hasil | Endpoint | Test |
|---|---|---|---|---|
| R-lessons-SELECT-1 | A/K/S: baris `lessons` penuh | P → kosong (P dilayani lewat R-student_lessons-*) | `GET /curriculum/lessons`, `PUT /curriculum` (baca id lama) | `TestRLS_R_lessons_SELECT_1_Diizinkan`, `TestRLS_R_lessons_SELECT_1_Ditolak` |
| R-lessons-SELECT-2 | A/K/S: embed `xp_reward` terisi | P → embed `null`, fallback kode lama `?? 60` | `GET /leaderboard/weekly` | `TestRLS_R_lessons_SELECT_2_Diizinkan`, `TestRLS_R_lessons_SELECT_2_Ditolak` |
| R-lessons-INSERT-1 | A/K (upsert) | S/P → 403 `... for table "lessons"` | `PUT /curriculum` | `TestRLS_R_lessons_INSERT_1_Diizinkan`, `TestRLS_R_lessons_INSERT_1_Ditolak` |
| R-lessons-UPDATE-1 | A/K | S/P → 0 baris, 200 | — (upsert ikut R-lessons-INSERT-1) | `TestRLS_R_lessons_UPDATE_1_Diizinkan`, `TestRLS_R_lessons_UPDATE_1_Ditolak` |
| R-lessons-DELETE-1 | A/K | S/P → 0 baris, 200 | `PUT /curriculum`, `DELETE /curriculum` (cascade dari levels) | `TestRLS_R_lessons_DELETE_1_Diizinkan`, `TestRLS_R_lessons_DELETE_1_Ditolak` |
| R-student_lessons-SELECT-1 | semua peran: semua baris, `quiz` tanpa kunci `correctAnswer` | tanpa token → 401 | `GET /curriculum/lessons` (P) | `TestRLS_R_student_lessons_SELECT_1_Diizinkan`, `TestRLS_R_student_lessons_SELECT_1_Ditolak` |
| R-student_lessons-SELECT-2 | `solution` terisi bila `exercise_type = 'flowchart'` atau peran A/K/S | P untuk lesson non-flowchart → `solution = null` | `GET /curriculum/lessons` (P) | `TestRLS_R_student_lessons_SELECT_2_Diizinkan`, `TestRLS_R_student_lessons_SELECT_2_Ditolak` |

### student_progress

| ID | Diizinkan | Ditolak → hasil | Endpoint | Test |
|---|---|---|---|---|
| R-student_progress-SELECT-1 | baris milik sendiri (semua peran); A/K/S semua baris | P baris orang lain → disaring | `GET /progress/me/lessons`, `GET /progress/me/recent`, `GET /admin/users/:nim/progress`, `GET /monitoring/progress`, `GET /prestige/eligibility`, `GET /leaderboard/weekly`, `POST /progress/lessons/:lessonId/complete` (cek duplikat) | `TestRLS_R_student_progress_SELECT_1_Diizinkan`, `TestRLS_R_student_progress_SELECT_1_Ditolak` |
| R-student_progress-INSERT-1 | baris milik sendiri (semua peran); A/K/S NIM mana pun | P untuk NIM lain → 403 `... for table "student_progress"` | `POST /progress/lessons/:lessonId/complete` | `TestRLS_R_student_progress_INSERT_1_Diizinkan`, `TestRLS_R_student_progress_INSERT_1_Ditolak` |
| R-student_progress-UPDATE-1 | A/K/S | P (termasuk baris sendiri) → 0 baris, 200 | — | `TestRLS_R_student_progress_UPDATE_1_Diizinkan`, `TestRLS_R_student_progress_UPDATE_1_Ditolak` |
| R-student_progress-DELETE-1 | A/K/S | P (termasuk baris sendiri) → 0 baris, 200 | `POST /admin/users/:nim/reset`, `POST /admin/users/:nim/reset-level`, `POST /admin/users/:nim/reset-lesson` | `TestRLS_R_student_progress_DELETE_1_Diizinkan`, `TestRLS_R_student_progress_DELETE_1_Ditolak` |

### game_history

| ID | Diizinkan | Ditolak → hasil | Endpoint | Test |
|---|---|---|---|---|
| R-game_history-SELECT-1 | baris milik sendiri (semua peran); A/K/S semua baris | P baris orang lain → disaring | `GET /game/plays/me`, `GET /leaderboard/weekly` | `TestRLS_R_game_history_SELECT_1_Diizinkan`, `TestRLS_R_game_history_SELECT_1_Ditolak` |
| R-game_history-INSERT-1 | baris milik sendiri (semua peran, termasuk A/K) | NIM lain (semua peran) → 403 `... for table "game_history"` | `POST /game/history` | `TestRLS_R_game_history_INSERT_1_Diizinkan`, `TestRLS_R_game_history_INSERT_1_Ditolak` |

### unlocked_achievements

| ID | Diizinkan | Ditolak → hasil | Endpoint | Test |
|---|---|---|---|---|
| R-unlocked_achievements-SELECT-1 | baris milik sendiri (semua peran); A/K/S semua baris | P baris orang lain → disaring | `GET /achievements/me`, `POST /achievements/check`, `POST /achievements/check-xp` | `TestRLS_R_unlocked_achievements_SELECT_1_Diizinkan`, `TestRLS_R_unlocked_achievements_SELECT_1_Ditolak` |
| R-unlocked_achievements-INSERT-1 | baris milik sendiri (semua peran) | NIM lain (semua peran) → 403 `... for table "unlocked_achievements"` | `POST /achievements/check`, `POST /achievements/check-xp` | `TestRLS_R_unlocked_achievements_INSERT_1_Diizinkan`, `TestRLS_R_unlocked_achievements_INSERT_1_Ditolak` |
| R-unlocked_achievements-DELETE-1 | A/K/S | P → 0 baris, 200 | `POST /admin/users/:nim/reset` | `TestRLS_R_unlocked_achievements_DELETE_1_Diizinkan`, `TestRLS_R_unlocked_achievements_DELETE_1_Ditolak` |

### active_sessions

| ID | Diizinkan | Ditolak → hasil | Endpoint | Test |
|---|---|---|---|---|
| R-active_sessions-SELECT-1 | A/K/S semua baris | P → kosong | `GET /monitoring/sessions` | `TestRLS_R_active_sessions_SELECT_1_Diizinkan`, `TestRLS_R_active_sessions_SELECT_1_Ditolak` |
| R-active_sessions-INSERT-1 | A/K/S untuk NIM sendiri (upsert) | NIM lain → 403 `... for table "active_sessions"`; P untuk NIM sendiri juga 403 karena baris baru tidak lolos policy SELECT pada upsert (tidak pasti). `Layout.tsx:198`/`:237` menelan error ini dan tetap lanjut ke update `users`; `PUT /sessions/me` meniru dengan `200` + `sessionSaved:false`, bagian `touchStreak`/`studySeconds` tetap jalan | `PUT /sessions/me` | `TestRLS_R_active_sessions_INSERT_1_Diizinkan`, `TestRLS_R_active_sessions_INSERT_1_Ditolak` |
| R-active_sessions-UPDATE-1 | A/K/S untuk baris sendiri | baris lain → 0 baris; P baris sendiri → 0 baris (disaring policy SELECT) | — (upsert ikut R-active_sessions-INSERT-1) | `TestRLS_R_active_sessions_UPDATE_1_Diizinkan`, `TestRLS_R_active_sessions_UPDATE_1_Ditolak` |
| R-active_sessions-DELETE-1 | A/K/S baris mana pun | P (termasuk baris sendiri, disaring policy SELECT) → 0 baris, 200 | `DELETE /sessions/me`, `DELETE /monitoring/sessions/stale` | `TestRLS_R_active_sessions_DELETE_1_Diizinkan`, `TestRLS_R_active_sessions_DELETE_1_Ditolak` |

### activity_logs

| ID | Diizinkan | Ditolak → hasil | Endpoint | Test |
|---|---|---|---|---|
| R-activity_logs-SELECT-1 | A/K/S | P → kosong | `GET /activity-logs` | `TestRLS_R_activity_logs_SELECT_1_Diizinkan`, `TestRLS_R_activity_logs_SELECT_1_Ditolak` |
| R-activity_logs-INSERT-1 | `nim` log = NIM token (semua peran) | NIM lain → 403 `... for table "activity_logs"`. `monitoringService.addAuditLog` tidak memeriksa error, jadi penulisan log internal oleh endpoint admin/prestige gagal diam-diam | `POST /activity-logs`, penulisan internal di endpoint `/admin/users/*`, `PATCH /admin/users/:nim/role`, `POST /prestige/reset` | `TestRLS_R_activity_logs_INSERT_1_Diizinkan`, `TestRLS_R_activity_logs_INSERT_1_Ditolak` |
| R-activity_logs-DELETE-1 | A/K | S/P → 0 baris, 200 | `DELETE /activity-logs/:id` | `TestRLS_R_activity_logs_DELETE_1_Diizinkan`, `TestRLS_R_activity_logs_DELETE_1_Ditolak` |

### prestige_history, progress_archive

| ID | Diizinkan | Ditolak → hasil | Endpoint | Test |
|---|---|---|---|---|
| R-prestige_history-SELECT-1 | baris milik sendiri; A/K/S semua | P baris orang lain → disaring | `GET /prestige/eligibility` | `TestRLS_R_prestige_history_SELECT_1_Diizinkan`, `TestRLS_R_prestige_history_SELECT_1_Ditolak` |
| R-prestige_history-INSERT-1 | A/K (UPDATE/DELETE sama, policy ALL) | S/P → 403 `... for table "prestige_history"` (UPDATE/DELETE: 0 baris) | — (hanya ditulis `do_prestige_reset`) | `TestRLS_R_prestige_history_INSERT_1_Diizinkan`, `TestRLS_R_prestige_history_INSERT_1_Ditolak` |
| R-progress_archive-SELECT-1 | baris milik sendiri; A/K/S semua | P baris orang lain → disaring | — | `TestRLS_R_progress_archive_SELECT_1_Diizinkan`, `TestRLS_R_progress_archive_SELECT_1_Ditolak` |
| R-progress_archive-INSERT-1 | A/K (UPDATE/DELETE sama, policy ALL) | S/P → 403 `... for table "progress_archive"` (UPDATE/DELETE: 0 baris) | — (hanya ditulis `do_prestige_reset`) | `TestRLS_R_progress_archive_INSERT_1_Diizinkan`, `TestRLS_R_progress_archive_INSERT_1_Ditolak` |

### weekly_standings, assessment_grading_rules, storage.objects

| ID | Diizinkan | Ditolak → hasil | Endpoint | Test |
|---|---|---|---|---|
| R-weekly_standings-SELECT-1 | semua peran | tanpa token → 401 | — (kode mati) | `TestRLS_R_weekly_standings_SELECT_1_Diizinkan`, `TestRLS_R_weekly_standings_SELECT_1_Ditolak` |
| R-weekly_standings-INSERT-1 | A/K/S | P → 403 `... for table "weekly_standings"` | — (kode mati) | `TestRLS_R_weekly_standings_INSERT_1_Diizinkan`, `TestRLS_R_weekly_standings_INSERT_1_Ditolak` |
| R-weekly_standings-UPDATE-1 | A/K/S | P → 0 baris, 200 | — (kode mati) | `TestRLS_R_weekly_standings_UPDATE_1_Diizinkan`, `TestRLS_R_weekly_standings_UPDATE_1_Ditolak` |
| R-assessment_grading_rules-SELECT-1 | semua peran | tanpa token → 401 | `POST /grade` | `TestRLS_R_assessment_grading_rules_SELECT_1_Diizinkan`, `TestRLS_R_assessment_grading_rules_SELECT_1_Ditolak` |
| R-assessment_grading_rules-UPDATE-1 | semua peran (policy ALL `true`, berlaku juga untuk INSERT/DELETE) | tanpa token → 401 | — | `TestRLS_R_assessment_grading_rules_UPDATE_1_Diizinkan`, `TestRLS_R_assessment_grading_rules_UPDATE_1_Ditolak` |
| R-storage_objects-SELECT-1 | publik (URL publik bucket) | — | — (dilayani Supabase Storage langsung) | `TestRLS_R_storage_objects_SELECT_1_Diizinkan`, `TestRLS_R_storage_objects_SELECT_1_Ditolak` (tanpa token → 401 di endpoint Go) |
| R-storage_objects-INSERT-1 | A/K/S | P → 403 `new row violates row-level security policy` (teks Storage API tanpa nama tabel, tidak pasti). Dicek sesudah validasi tipe file | `POST /storage/lesson-images` | `TestRLS_R_storage_objects_INSERT_1_Diizinkan`, `TestRLS_R_storage_objects_INSERT_1_Ditolak` |
| R-storage_objects-UPDATE-1 | A/K/S (DELETE sama) | P → 403 dari Storage API | — (`upsert: false`, tidak ada hapus) | `TestRLS_R_storage_objects_UPDATE_1_Diizinkan`, `TestRLS_R_storage_objects_UPDATE_1_Ditolak` |

### Aturan trigger dan RPC

| ID | Diizinkan | Ditolak → hasil | Endpoint | Test |
|---|---|---|---|---|
| T-check_user_updates-1 | A/K ubah `role` apa pun; S ubah `role` ke nilai yang dikanonikkan jadi `asisten` | S/P selain itu → 400 `Akses Ditolak: Anda tidak diperbolehkan mengubah peran (role) Anda sendiri!` | `POST /profile/sync`, `PATCH /admin/users/:nim/role` | `TestRLS_T_check_user_updates_1_Diizinkan`, `TestRLS_T_check_user_updates_1_Ditolak` |
| T-check_user_updates-2 | A/K ubah `level_access_overrides` | S/P → 400 `Akses Ditolak: Anda tidak memiliki wewenang untuk mengubah level_access_overrides!` | `PATCH /admin/users/:nim/level-access` | `TestRLS_T_check_user_updates_2_Diizinkan`, `TestRLS_T_check_user_updates_2_Ditolak` |
| T-check_user_updates-3 | A/K ubah `assessment_access` | S/P → 400 `Akses Ditolak: Anda tidak memiliki wewenang untuk mengubah assessment_access!` | — (tidak ada pemanggil yang menulis kolom ini; tetap dicek di guard) | `TestRLS_T_check_user_updates_3_Diizinkan`, `TestRLS_T_check_user_updates_3_Ditolak` |
| T-check_user_updates-4 | A/K ubah `elearning_role` | S/P → 400 `Akses Ditolak: Anda tidak memiliki wewenang untuk mengubah elearning_role!` | `PATCH /admin/users/:nim/role` | `TestRLS_T_check_user_updates_4_Diizinkan`, `TestRLS_T_check_user_updates_4_Ditolak` |
| T-check_user_xp_level-1 | A/K: `xp`/`level` disimpan apa adanya | S/P (UPDATE): `xp` ditimpa rumus sumber, `level = calculate_level(xp)`, `prestige_mult` di-clamp; tidak ada error | semua endpoint R-users-UPDATE-1 | `TestRLS_T_check_user_xp_level_1_Diizinkan`, `TestRLS_T_check_user_xp_level_1_Ditolak` |
| T-check_user_xp_level-2 | A/K: INSERT profil dengan `xp`/`level` apa adanya | S/P (INSERT): sama seperti T-check_user_xp_level-1 | `POST /profile/sync` | `TestRLS_T_check_user_xp_level_2_Diizinkan`, `TestRLS_T_check_user_xp_level_2_Ditolak` |
| T-guard_prestige_columns-1 | A/K ubah `prestige_mult`/`prestige_tier`/`prestige_count` | S/P → 400 `Akses Ditolak: kolom prestige hanya bisa diubah lewat Prestige resmi (tamat 6/6).` (dibandingkan sesudah clamp T-check_user_xp_level-1) | semua endpoint R-users-UPDATE-1 | `TestRLS_T_guard_prestige_columns_1_Diizinkan`, `TestRLS_T_guard_prestige_columns_1_Ditolak` |
| T-check_progress_velocity-1 | insert lolos trigger DB | DB mengembalikan `P0001` → 400 pesan apa adanya (test memakai `*pgconn.PgError` palsu dari mock) | `POST /progress/lessons/:lessonId/complete` | `TestRLS_T_check_progress_velocity_1_Diizinkan`, `TestRLS_T_check_progress_velocity_1_Ditolak` |
| T-do_prestige_reset-1 | diri sendiri, atau peran efektif UI A/K/S, lalu RPC dipanggil apa adanya | guard service: bukan diri sendiri dan bukan staf → 403 `Akses ditolak.`; exception RPC (`P0001`) → 400 pesan apa adanya | `POST /prestige/reset` | `TestRLS_T_do_prestige_reset_1_Diizinkan`, `TestRLS_T_do_prestige_reset_1_Ditolak` |

Urutan evaluasi untuk satu UPDATE `users` atas nama token di Go: R-users-UPDATE-2 (grant, 403) → R-users-UPDATE-1 (USING, 0 baris) → kanonikkan `role` → T-check_user_updates-1..4 (400) → T-check_user_xp_level-1 (timpa) → T-guard_prestige_columns-1 (400). Grant dicek PostgreSQL saat parse sebelum baris dibaca. USING menyaring baris sebelum trigger BEFORE jalan. Trigger berjalan alfabet.

### Operasi tanpa policy (default deny, tidak dipakai kode)

`game_history` UPDATE/DELETE, `unlocked_achievements` UPDATE, `activity_logs` UPDATE, `weekly_standings` DELETE. Tidak ada endpoint, jadi tidak ada aturan atau test.

## Status tidak pasti

1. `assessment_attempts` dan `assessment_questions` tidak punya `CREATE POLICY` atau `ENABLE ROW LEVEL SECURITY` di repo, padahal `api/grade.ts:53` memakai klien dengan token pengguna. Go menganggap kedua tabel terbaca/tertulis oleh pemanggil yang sudah lolos cek peran `grade.ts`. Bila prod punya RLS yang menyaring, perilaku lama bisa berbeda (attempt tidak ditemukan). `elearning_progress` (`api/report.ts:95`) juga tanpa policy di repo, tetapi ditulis dengan service role key.
2. `fix-supabase-permissions.sql` dan `setup-new-database.sql` dianggap tidak dijalankan (bukti di § Sumber & urutan). Bila ternyata dijalankan, sebagian besar tabel terbuka penuh dan aturan di atas lebih ketat daripada prod.
3. Empat file berbagi timestamp `2026-09-25 17:50:43` (`fix-role-update-trigger`, `migration-canonical-role`, `migration-fix-role-claim`, `migration-users-rls`), empat file `2026-09-28 19:06:57`, empat file `2026-09-18 00:22:53`, dan empat file `2026-09-12 21:45:07`. Urutan nama dipakai sebagai tie-breaker. Untuk `auth_role()` urutan `fix-role-update-trigger` vs `migration-fix-role-claim` tidak berpengaruh karena teks fungsinya identik. `check_user_updates` dan `check_user_xp_level` dari kelompok itu tertimpa file yang lebih baru.
4. `auth_nim()` hanya didefinisikan di `setup-new-database.sql:11` (file yang dikecualikan). Fungsi itu pasti ada di prod karena dipakai policy aktif, tetapi definisinya tidak bisa dipastikan dari repo. Diasumsikan `nim → sub → username`.
5. Belum pasti apakah `migration-xp-fix-20261026.sql` sudah dijalankan. Bila belum, `check_user_xp_level` efektif adalah `migration-prestige.sql:138`: menolak dengan `Akses Ditolak: Nilai XP yang dikirim (%) melebihi batas XP riil belajar Anda (%)!` bila XP di atas cap, dan membaca tabel `lessons` yang tersaring RLS bagi P (cap = `game_sum`). Keputusan 3 memakai versi xp-fix.
6. Column grant `users` hanya diketahui dari komentar `migration-fix-users-column-grants.sql`. Daftar SELECT ("15 kolom lain") dan INSERT tidak diketahui. Contoh dampak: sinkron login `SupabaseProvider.tsx:212` yang ikut menulis `shift` gagal seluruhnya, dan kode lama menelan error itu.
7. Leaderboard mingguan untuk P: `student_progress` dan `game_history` orang lain tersaring (R-student_progress-SELECT-1, R-game_history-SELECT-1), dan embed `lessons:lesson_id(xp_reward)` bernilai `null` sehingga reward = 60 (R-lessons-SELECT-2). Akibatnya P hanya melihat XP mingguan dirinya sendiri. Perilaku ini ditiru, bukan diperbaiki.
8. `api/report.ts:54` membaca `levels(*, modules(*, lessons(*)))` tanpa `order`, jadi urutan level/modul/lesson mengikuti urutan fisik Postgres. Go memakai `sort_order` (plan langkah 3), dan perbedaan ini perlu dicatat di temuan.
9. Upsert `active_sessions` oleh P: menurut dokumentasi PostgreSQL, `INSERT ... ON CONFLICT DO UPDATE` memeriksa baris baru terhadap policy SELECT. Satu-satunya policy SELECT `active_sessions` hanya untuk staf, jadi heartbeat P seharusnya selalu ditolak (`42501`), dan `DELETE` sesi milik sendiri saat logout menghapus 0 baris. Kode lama tidak memeriksa error keduanya. Kalau prod ternyata berisi sesi mahasiswa, berarti policy di prod berbeda dari repo. Pemilik perlu menjalankan query verifikasi 6 di bawah.
10. Pesan 403 Storage API untuk upload non-staf diasumsikan `new row violates row-level security policy` (tanpa nama tabel), sesuai teks error Supabase Storage. Belum diverifikasi.
11. `users.role` sudah kanonik (`mahasiswa/asisten/koordinator`) sejak `migration-canonical-role.sql`, sehingga cabang `role = 'praktikan'` di policy `migration-users-rls.sql:23` dan pembanding `NEW.role = auth_role()` di `check_user_updates` hampir tidak pernah cocok untuk P. Efeknya dijelaskan di T-check_user_updates-1.

## Query verifikasi untuk pemilik

Semua query di bawah read-only dan **TIDAK dijalankan agent** (agent dilarang konek ke DB mana pun). Jalankan oleh pemilik di Supabase SQL Editor project elearning. Kalau hasilnya berbeda dari dokumen ini, dokumen dan aturan Go yang mengikuti DB.

1. Semua policy public:
   ```sql
   select * from pg_policies where schemaname='public';
   ```
2. Column grant `users` untuk `authenticated`:
   ```sql
   select grantee, privilege_type, column_name from information_schema.column_privileges where table_schema='public' and table_name='users' and grantee='authenticated';
   ```
3. Trigger pada `users`:
   ```sql
   select tgname, pg_get_triggerdef(oid) from pg_trigger where tgrelid='public.users'::regclass and not tgisinternal;
   ```
4. Definisi view `student_lessons`:
   ```sql
   select pg_get_viewdef('public.student_lessons'::regclass, true);
   ```
5. Fungsi dan status SECURITY DEFINER:
   ```sql
   select proname, prosecdef from pg_proc where proname in ('auth_role','auth_nim','check_user_xp_level','check_user_updates','guard_prestige_columns','do_prestige_reset');
   ```
6. Tambahan untuk § Status tidak pasti (juga read-only):
   ```sql
   select * from pg_policies where schemaname='storage' and tablename='objects';
   select relname, relrowsecurity, relforcerowsecurity from pg_class where relnamespace='public'::regnamespace and relkind in ('r','v') order by relname;
   select pg_get_functiondef('public.auth_nim'::regproc);
   select pg_get_functiondef('public.check_user_xp_level'::regproc);
   select count(*) filter (where u.role='mahasiswa') as sesi_mahasiswa, count(*) as total from public.active_sessions a left join public.users u on u.nim=a.nim;
   ```

## Peta query

Semua path di bawah `/api/elearning`. Kolom "Status" diisi FEAT-002..005: `belum` atau `selesai (<nama test route>)`. Baris "kode mati" tidak dibuat endpoint (keputusan 15).

| Sumber lama | Tabel/RPC | Endpoint baru | Status |
|---|---|---|---|
| `components/Layout.tsx:70` | active_sessions delete own | `DELETE /sessions/me` | selesai (`TestElearningLearningRoutesRequireToken`) |
| `Layout.tsx:143`, `:152`, raw `PATCH /rest/v1/users` `:181` | users study_time | `POST /progress/study-time {seconds}` (increment dalam satu UPDATE) | selesai (`TestElearningLearningRoutesRequireToken`) |
| `Layout.tsx:198`, `:237` | active_sessions upsert | `PUT /sessions/me {currentActivity,nama?,kelas?,attemptId?,touchStreak?,studySeconds?}` | selesai (`TestElearningLearningRoutesRequireToken`) |
| `Layout.tsx:218` | users last_active/streak | `POST /streak` (atau `touchStreak` di `PUT /sessions/me`) | selesai (`TestElearningLearningRoutesRequireToken`) |
| `components/SupabaseProvider.tsx:91`, `:126`, `:212` | users select/insert/update | `POST /profile/sync` | selesai (`TestElearningLearningRoutesRequireToken`) |
| `SupabaseProvider.tsx:298`, `:321` | student_progress own completed | `GET /progress/me/lessons` | selesai (`TestElearningLearningRoutesRequireToken`) |
| `SupabaseProvider.tsx:370`, `:379` | levels | `GET /curriculum/levels` | selesai (`TestElearningLearningRoutesRequireToken`) |
| `SupabaseProvider.tsx:397` | modules | `GET /curriculum/modules` | selesai (`TestElearningLearningRoutesRequireToken`) |
| `SupabaseProvider.tsx:405` | lessons / student_lessons | `GET /curriculum/lessons` (staf: lessons penuh; lainnya: aturan view) | selesai (`TestElearningLearningRoutesRequireToken`) |
| `pages/AdminDashboard.tsx:1196` | users level_access_overrides | `PATCH /admin/users/:nim/level-access` | selesai (`TestElearningLearningRoutesRequireToken`) |
| `AdminDashboard.tsx:1219` | users list | `GET /admin/users?page&limit&q` | selesai (`TestElearningLearningRoutesRequireToken`) |
| `AdminDashboard.tsx:1266` | student_progress by nim | `GET /admin/users/:nim/progress` | selesai (`TestElearningLearningRoutesRequireToken`) |
| `AdminDashboard.tsx:1328` | users role/elearning_role | `PATCH /admin/users/:nim/role` | selesai (`TestElearningLearningRoutesRequireToken`) |
| `pages/Dashboard.tsx:81` | student_progress recent | `GET /progress/me/recent?limit=5` | selesai (`TestElearningLearningRoutesRequireToken`) |
| `pages/Profile.tsx:27` | unlocked_achievements own | `GET /achievements/me` | selesai (`TestElearningLearningRoutesRequireToken`) |
| `pages/StudentMonitoring.tsx:62`, `:156` | active_sessions + users(role) | `GET /monitoring/sessions` | selesai (`TestElearningLearningRoutesRequireToken`) |
| `StudentMonitoring.tsx:81` | users mahasiswa | `GET /monitoring/students?page&limit` | selesai (`TestElearningLearningRoutesRequireToken`) |
| `StudentMonitoring.tsx:111` | student_progress semua (paging 1000) | `GET /monitoring/progress?page&limit` (dikelompokkan per nim, `array_agg`) | selesai (`TestElearningLearningRoutesRequireToken`) |
| `services/achievementService.ts:32`, `:55`, `:91` | achievements, unlocked select/insert | `POST /achievements/check` | selesai (`TestElearningLearningRoutesRequireToken`) |
| `achievementService.ts:116` | achievements | `GET /achievements?requirementType=` | selesai (`TestElearningLearningRoutesRequireToken`) |
| `achievementService.ts:145`, `:162`, `:176` | achievements xp, unlocked | `POST /achievements/check-xp {xp}` | selesai (`TestElearningLearningRoutesRequireToken`) |
| `services/curriculumService.ts:12`, `:22` / `:38` / `:43` | levels / modules / lessons | `GET /curriculum/levels` / `modules` / `lessons` | selesai (`TestElearningLearningRoutesRequireToken`) |
| `curriculumService.ts:160`, `:162`, `:164`, `:179`, `:187`, `:195`, `:251`, `:255`, `:259` | saveFullCurriculum | `PUT /curriculum` (satu transaksi) | selesai (`TestElearningLearningRoutesRequireToken`) |
| `curriculumService.ts:271` | levels update | kode mati (`updateLevel` tanpa pemanggil) | kode mati |
| `curriculumService.ts:290`, `:293` | clearCurriculum | `DELETE /curriculum` | selesai (`TestElearningLearningRoutesRequireToken`) |
| `services/gameService.ts:27` | game_questions by language | `GET /game/questions?language&limit` (acak, maks 100) | selesai (`TestElearningLearningRoutesRequireToken`) |
| `gameService.ts:83`, `:105` | game_history insert + users | `POST /game/history` | selesai (`TestElearningLearningRoutesRequireToken`) |
| `gameService.ts:143`, `:160` | seed bila kosong | `POST /game/questions/seed` | selesai (`TestElearningLearningRoutesRequireToken`) |
| `gameService.ts:176`, `:179`, `:193` | force reset | `PUT /game/questions` (satu transaksi) | selesai (`TestElearningLearningRoutesRequireToken`) |
| `gameService.ts:206` / `:232` / `:241` | add / update / delete | `POST /game/questions` / `PATCH /game/questions/:id` / `DELETE /game/questions/:id` | selesai (`TestElearningLearningRoutesRequireToken`) |
| `gameService.ts:251` / `:294` | game_settings get / upsert | `GET /game/settings` / `PUT /game/settings` | selesai (`TestElearningLearningRoutesRequireToken`) |
| `gameService.ts:314` | game_history count minggu ini | `GET /game/plays/me` (= `canPlayBugHunt`) | selesai (`TestElearningLearningRoutesRequireToken`) |
| `services/leaderboardService.ts:82`, `:102`, `:109`, `:169`, `:176` | weekly | `GET /leaderboard/weekly?kelas&jurusan&shift&limit` | selesai (`TestElearningLearningRoutesRequireToken`) |
| `leaderboardService.ts:266`, `:280`, `:288` | weekly_standings | kode mati | kode mati |
| `leaderboardService.ts:306` / `:356` / `:384` | users | `GET /leaderboard` / `GET /leaderboard/rank?xp` / `GET /leaderboard/filters` | selesai (`TestElearningLearningRoutesRequireToken`) |
| `services/monitoringService.ts:50`, `:65` | active_sessions | kode mati (operasi sama dengan `PUT /sessions/me` dan `GET /monitoring/sessions`) | kode mati |
| `monitoringService.ts:99` | active_sessions delete stale | `DELETE /monitoring/sessions/stale?thresholdMinutes=2` | selesai (`TestElearningLearningRoutesRequireToken`) |
| `monitoringService.ts:132` | activity_logs insert | `POST /activity-logs` (juga dipakai internal oleh endpoint admin dan prestige) | selesai (`TestElearningLearningRoutesRequireToken`) |
| `monitoringService.ts:143`, `:158` | activity_logs list | `GET /activity-logs?page&limit` | selesai (`TestElearningLearningRoutesRequireToken`) |
| `monitoringService.ts:190` | activity_logs delete | `DELETE /activity-logs/:id` | selesai (`TestElearningLearningRoutesRequireToken`) |
| `services/playgroundService.ts:15` / `:43` / `:67` / `:76` | playground_examples | `GET` / `POST /playground/examples`, `PATCH` / `DELETE /playground/examples/:id` | selesai (`TestElearningLearningRoutesRequireToken`) |
| `services/prestigeService.ts:55`, `:57`, `:58`, `:118`, `:132`, `:139` | eligibility | `GET /prestige/eligibility?nim=` | selesai (`TestElearningLearningRoutesRequireToken`) |
| `prestigeService.ts:207` | RPC `do_prestige_reset` | `POST /prestige/reset {nim}` (RPC dipanggil apa adanya) | selesai (`TestElearningLearningRoutesRequireToken`) |
| `services/progressService.ts:57`, `:84`, `:102` | completeLesson | `POST /progress/lessons/:lessonId/complete {xpReward}` | selesai (`TestElearningLearningRoutesRequireToken`) |
| `progressService.ts:162` | grantXp | `POST /progress/xp {amount}` | selesai (`TestElearningLearningRoutesRequireToken`) |
| `progressService.ts:193` | syncProgress | kode mati (= `GET /progress/me/lessons`) | kode mati |
| `progressService.ts:236`, `:244`, `:254` | resetUserProgress | `POST /admin/users/:nim/reset` | selesai (`TestElearningLearningRoutesRequireToken`) |
| `progressService.ts:311`, `:322`, `:331`, `:340` | resetLevelProgress | `POST /admin/users/:nim/reset-level {levelId}` | selesai (`TestElearningLearningRoutesRequireToken`) |
| `progressService.ts:395`, `:404`, `:413` | resetLessonProgress | `POST /admin/users/:nim/reset-lesson {lessonId,lessonTitle}` | selesai (`TestElearningLearningRoutesRequireToken`) |
| `progressService.ts:465`, `:473` | adjustUserXp | `PUT /admin/users/:nim/xp {xp}` | selesai (`TestElearningLearningRoutesRequireToken`) |
| `progressService.ts:538` | deleteUser | `DELETE /admin/users/:nim` | selesai (`TestElearningLearningRoutesRequireToken`) |
| `services/storageService.ts:81`, `:98` | storage `lesson-images` | `POST /storage/lesson-images` (multipart `file`) | selesai (`TestElearningLearningRoutesRequireToken`) |
| `services/streakService.ts:88` | users streak | kode mati (= `POST /streak`) | kode mati |
| `api/validate-quiz.ts` | lessons.quiz | `POST /validate-quiz` | selesai (`TestElearningLegacyAPIRoutes`) |
| `api/grade.ts` | users ×3, assessment_attempts ×2, assessment_grading_rules, assessment_questions | `POST /grade` | selesai (`TestElearningLegacyAPIRoutes`) |
| `api/ai.ts` | users.role | `POST /ai` | selesai (`TestElearningLegacyAPIRoutes`) |
| `api/report.ts` | lessons, student_progress ×2, levels(modules(lessons)), elearning_progress | `POST /report` | selesai (`TestElearningLegacyAPIRoutes`) |
| `api/cron/ping.ts:30` | lessons (keep-alive) | — | di luar Fase 7: keep-alive milik Fase 5 (`GET /api/cron/keep-alive`), sengaja tidak dibuat |
| `api/index.ts` / `receive-token.ts` / `login.ts` / `verify.ts` / `auth.ts` | — | `POST /sso` / `POST /receive-token` / `POST /login` / `POST /verify` / `POST /auth/supabase-token` | selesai (`TestElearningSSOFormToCodeExchange`, `TestElearningSSORedirects`, `TestElearningAuthEndpoints`, `TestElearningLoginRateLimit`) |

Baris 06-elearning "`assessment_*` → `/api/elearning/assessment/*`": di `src/` tidak ada query langsung ke tabel assessment. Baris ini dipenuhi oleh `POST /grade` dan `POST /validate-quiz`, dan dicatat sebagai penyimpangan.
