# API

Base URL produksi: `https://api.algohub.web.id`. Semua route di bawah `/api`.

## Umum
| Method | Path | Auth | Response |
|---|---|---|---|
| GET | `/api/health` | — | `200 {"service":"algohub-api","status":"ok"}` |

## CORS
- Origin diizinkan: daftar eksak `CORS_ORIGINS` + regex `CORS_ORIGIN_PATTERNS`. Tanpa env di non-production: `http://localhost:5173`–`5176`.
- Origin diizinkan mendapat `Access-Control-Allow-Origin: <origin>`, `Access-Control-Allow-Credentials: true`, `Access-Control-Expose-Headers: Content-Disposition, Retry-After`, dan `Vary: Origin`. Header ini juga ada di response error.
- Preflight (`OPTIONS` dengan `Access-Control-Request-Method`) dijawab `204` sebelum handler lain berjalan, dengan `Access-Control-Allow-Methods: GET, POST, PUT, PATCH, DELETE, OPTIONS`, `Access-Control-Allow-Headers: Authorization, Content-Type, X-Requested-With, X-Cron-Secret`, dan `Access-Control-Max-Age: 86400`.
- Origin asing: request tetap diproses tanpa header CORS, dan browser yang memblokir.

## Pemetaan route lama → baru
- Route praktikum lama `/api/X` pindah ke `/api/praktikum/X`. Path lama tetap dilayani sebagai alias (`ALIAS-LAMA` di `router.go`) dengan handler dan middleware yang sama sampai cutover.
- Route lama yang sudah berawalan `/api/praktikum/` (grup mahasiswa) tidak berubah dan tidak punya alias.
- Cron pindah ke `/api/cron/praktikum/*`. Path lama `/api/cron/auto-submit` dan `/api/cron/rumpang` tetap aktif sebagai alias.
- `POST /api/praktikum/compile-c` dihapus. Compile C hanya dilakukan di browser.
- `/swagger/*any` hanya dipasang saat `APP_ENV` bukan `production`.
- Auth memakai header `Authorization: Bearer <token>`. Jenis token mengikuti `AUTH_MODE` (`legacy`, `dual`, `supabase`). Cron memakai header `X-Cron-Secret`.
- Rate limiter dipakai bersama oleh path baru dan alias, dan kuncinya IP klien dari `X-Forwarded-For` (entri pertama) atau `X-Real-IP`. Limiter `run` memakai ID user.
- Kolom Izin berisi permission `RequirePermission` (`resource:write`) di atas role.
### Cron
| Method | Path baru | Path lama (alias) | Akses | Izin | Rate limit |
|---|---|---|---|---|---|
| POST | `/api/cron/praktikum/auto-submit` | `/api/cron/auto-submit` | X-Cron-Secret | — | — |
| POST | `/api/cron/praktikum/rumpang` | `/api/cron/rumpang` | X-Cron-Secret | — | — |

### Auth
| Method | Path baru | Path lama (alias) | Akses | Izin | Rate limit |
|---|---|---|---|---|---|
| POST | `/api/praktikum/auth/cek-nim` | `/api/auth/cek-nim` | publik | — | 20/menit/IP |
| POST | `/api/praktikum/auth/login` | `/api/auth/login` | publik | — | 10/menit/IP + 20/menit/IP |
| POST | `/api/praktikum/auth/register` | `/api/auth/register` | publik | — | 20/menit/IP |
| POST | `/api/praktikum/auth/forgot-password` | `/api/auth/forgot-password` | publik | — | 20/menit/IP |
| POST | `/api/praktikum/auth/reset-password` | `/api/auth/reset-password` | publik | — | 20/menit/IP |

### Info
| Method | Path baru | Path lama (alias) | Akses | Izin | Rate limit |
|---|---|---|---|---|---|
| GET | `/api/praktikum/info/asisten` | `/api/info/asisten` | publik | — | — |
| GET | `/api/praktikum/info/laporan` | `/api/info/laporan` | publik | — | — |
| GET | `/api/praktikum/info/modul` | `/api/info/modul` | publik | — | — |
| GET | `/api/praktikum/info/jadwal` | `/api/info/jadwal` | publik | — | — |
| GET | `/api/praktikum/info/jadwal/config` | `/api/info/jadwal/config` | publik | — | — |
| GET | `/api/praktikum/info/kelas/:id/mahasiswa` | `/api/info/kelas/:id/mahasiswa` | publik | — | — |
| GET | `/api/praktikum/info/announcements` | `/api/info/announcements` | publik | — | — |

### Login semua role
| Method | Path baru | Path lama (alias) | Akses | Izin | Rate limit |
|---|---|---|---|---|---|
| POST | `/api/praktikum/auth/logout` | `/api/auth/logout` | login | — | — |
| GET | `/api/praktikum/auth/me` | `/api/auth/me` | login | — | — |
| GET | `/api/praktikum/profile` | `/api/profile` | login | — | — |
| PUT | `/api/praktikum/profile` | `/api/profile` | login | — | — |
| GET | `/api/praktikum/belajar/level` | `/api/belajar/level` | login | — | — |
| GET | `/api/praktikum/belajar/modul` | `/api/belajar/modul` | login | — | — |
| GET | `/api/praktikum/belajar/materi` | `/api/belajar/materi` | login | — | — |
| GET | `/api/praktikum/belajar/materi/:id` | `/api/belajar/materi/:id` | login | — | — |
| GET | `/api/praktikum/belajar/progres` | `/api/belajar/progres` | login | — | — |
| POST | `/api/praktikum/belajar/materi/:id/selesai` | `/api/belajar/materi/:id/selesai` | login | — | — |
| GET | `/api/praktikum/belajar/pencapaian` | `/api/belajar/pencapaian` | login | — | — |
| GET | `/api/praktikum/belajar/sesi` | `/api/belajar/sesi` | login | — | — |
| GET | `/api/praktikum/game/konfigurasi` | `/api/game/konfigurasi` | login | — | — |
| GET | `/api/praktikum/game/status` | `/api/game/status` | login | — | — |
| POST | `/api/praktikum/game/mulai` | `/api/game/mulai` | login | — | — |
| POST | `/api/praktikum/game/selesai` | `/api/game/selesai` | login | — | — |
| GET | `/api/praktikum/game/peringkat` | `/api/game/peringkat` | login | — | — |
| POST | `/api/praktikum/game/detak` | `/api/game/detak` | login | — | — |

### Mahasiswa
| Method | Path baru | Path lama (alias) | Akses | Izin | Rate limit |
|---|---|---|---|---|---|
| GET | `/api/praktikum/dashboard` | — | mahasiswa | — | — |
| GET | `/api/praktikum/sesi` | — | mahasiswa | — | — |
| GET | `/api/praktikum/ruang` | — | mahasiswa | — | — |
| POST | `/api/praktikum/mulai` | — | mahasiswa | — | — |
| POST | `/api/praktikum/autosave` | — | mahasiswa | — | — |
| POST | `/api/praktikum/submit` | — | mahasiswa | — | — |
| POST | `/api/praktikum/run` | — | mahasiswa | — | 30/menit/user |
| GET | `/api/praktikum/laporan-besar` | — | mahasiswa | — | — |
| POST | `/api/praktikum/laporan-besar/upload-url` | — | mahasiswa | — | — |
| POST | `/api/praktikum/laporan-besar/kumpul` | — | mahasiswa | — | — |
| DELETE | `/api/praktikum/laporan-besar/kumpul` | — | mahasiswa | — | — |

### Admin
| Method | Path baru | Path lama (alias) | Akses | Izin | Rate limit |
|---|---|---|---|---|---|
| GET | `/api/praktikum/admin/dashboard` | `/api/admin/dashboard` | asisten, koordinator | — | — |
| GET | `/api/praktikum/admin/users` | `/api/admin/users` | asisten, koordinator | — | — |
| POST | `/api/praktikum/admin/users` | `/api/admin/users` | asisten, koordinator | `users:write` | — |
| POST | `/api/praktikum/admin/users/bulk` | `/api/admin/users/bulk` | asisten, koordinator | `users:write` | — |
| PUT | `/api/praktikum/admin/users/bulk-attrs` | `/api/admin/users/bulk-attrs` | asisten, koordinator | `users:write` | — |
| PUT | `/api/praktikum/admin/users/:id` | `/api/admin/users/:id` | asisten, koordinator | `users:write` | — |
| DELETE | `/api/praktikum/admin/users/:id` | `/api/admin/users/:id` | asisten, koordinator | `users:write` | — |
| POST | `/api/praktikum/admin/users/:id/reset-password` | `/api/admin/users/:id/reset-password` | asisten, koordinator | `users:write` | — |
| GET | `/api/praktikum/admin/materi` | `/api/admin/materi` | asisten, koordinator | — | — |
| GET | `/api/praktikum/admin/materi/:id` | `/api/admin/materi/:id` | asisten, koordinator | — | — |
| POST | `/api/praktikum/admin/materi` | `/api/admin/materi` | asisten, koordinator | `sesi:write` | — |
| PUT | `/api/praktikum/admin/materi/:id` | `/api/admin/materi/:id` | asisten, koordinator | `sesi:write` | — |
| DELETE | `/api/praktikum/admin/materi/:id` | `/api/admin/materi/:id` | asisten, koordinator | `sesi:write` | — |
| PUT | `/api/praktikum/admin/belajar/kurikulum` | `/api/admin/belajar/kurikulum` | asisten, koordinator | `sesi:write` | — |
| DELETE | `/api/praktikum/admin/belajar/kurikulum` | `/api/admin/belajar/kurikulum` | asisten, koordinator | `sesi:write` | — |
| GET | `/api/praktikum/admin/belajar/users` | `/api/admin/belajar/users` | asisten, koordinator | — | — |
| GET | `/api/praktikum/admin/belajar/users/:id/progres` | `/api/admin/belajar/users/:id/progres` | asisten, koordinator | — | — |
| PUT | `/api/praktikum/admin/belajar/users/:id/akses-level` | `/api/admin/belajar/users/:id/akses-level` | asisten, koordinator | `penilaian:write` | — |
| POST | `/api/praktikum/admin/belajar/users/:id/reset` | `/api/admin/belajar/users/:id/reset` | asisten, koordinator | `penilaian:write` | — |
| PUT | `/api/praktikum/admin/belajar/users/:id/xp` | `/api/admin/belajar/users/:id/xp` | asisten, koordinator | `penilaian:write` | — |
| GET | `/api/praktikum/admin/monitoring/sesi` | `/api/admin/monitoring/sesi` | asisten, koordinator | — | — |
| GET | `/api/praktikum/admin/game/soal` | `/api/admin/game/soal` | asisten, koordinator | — | — |
| POST | `/api/praktikum/admin/game/soal` | `/api/admin/game/soal` | asisten, koordinator | `sesi:write` | — |
| PUT | `/api/praktikum/admin/game/soal/:id` | `/api/admin/game/soal/:id` | asisten, koordinator | `sesi:write` | — |
| DELETE | `/api/praktikum/admin/game/soal/:id` | `/api/admin/game/soal/:id` | asisten, koordinator | `sesi:write` | — |
| PUT | `/api/praktikum/admin/game/konfigurasi` | `/api/admin/game/konfigurasi` | asisten, koordinator | `konfigurasi:write` | — |
| GET | `/api/praktikum/admin/kelas` | `/api/admin/kelas` | asisten, koordinator | — | — |
| POST | `/api/praktikum/admin/kelas` | `/api/admin/kelas` | asisten, koordinator | `kelas:write` | — |
| PUT | `/api/praktikum/admin/kelas/:id` | `/api/admin/kelas/:id` | asisten, koordinator | `kelas:write` | — |
| DELETE | `/api/praktikum/admin/kelas/:id` | `/api/admin/kelas/:id` | asisten, koordinator | `kelas:write` | — |
| GET | `/api/praktikum/admin/asisten` | `/api/admin/asisten` | asisten, koordinator | — | — |
| POST | `/api/praktikum/admin/asisten` | `/api/admin/asisten` | asisten, koordinator | `asisten:write` | — |
| PUT | `/api/praktikum/admin/asisten/:id` | `/api/admin/asisten/:id` | asisten, koordinator | `asisten:write` | — |
| GET | `/api/praktikum/admin/jadwal` | `/api/admin/jadwal` | asisten, koordinator | — | — |
| POST | `/api/praktikum/admin/jadwal` | `/api/admin/jadwal` | asisten, koordinator | `jadwal:write` | — |
| PUT | `/api/praktikum/admin/jadwal/:id` | `/api/admin/jadwal/:id` | asisten, koordinator | `jadwal:write` | — |
| DELETE | `/api/praktikum/admin/jadwal/:id` | `/api/admin/jadwal/:id` | asisten, koordinator | `jadwal:write` | — |
| GET | `/api/praktikum/admin/pedoman` | `/api/admin/pedoman` | asisten, koordinator | — | — |
| POST | `/api/praktikum/admin/pedoman` | `/api/admin/pedoman` | asisten, koordinator | `konfigurasi:write` | — |
| PUT | `/api/praktikum/admin/pedoman/:id` | `/api/admin/pedoman/:id` | asisten, koordinator | `konfigurasi:write` | — |
| DELETE | `/api/praktikum/admin/pedoman/:id` | `/api/admin/pedoman/:id` | asisten, koordinator | `konfigurasi:write` | — |
| GET | `/api/praktikum/admin/konfigurasi` | `/api/admin/konfigurasi` | asisten, koordinator | — | — |
| GET | `/api/praktikum/admin/konfigurasi/ai-model` | `/api/admin/konfigurasi/ai-model` | asisten, koordinator | — | — |
| POST | `/api/praktikum/admin/konfigurasi` | `/api/admin/konfigurasi` | asisten, koordinator | `konfigurasi:write` | — |
| GET | `/api/praktikum/admin/sesi` | `/api/admin/sesi` | asisten, koordinator | — | — |
| GET | `/api/praktikum/admin/sesi/:id` | `/api/admin/sesi/:id` | asisten, koordinator | — | — |
| POST | `/api/praktikum/admin/sesi` | `/api/admin/sesi` | asisten, koordinator | `sesi:write` | — |
| PUT | `/api/praktikum/admin/sesi/:id` | `/api/admin/sesi/:id` | asisten, koordinator | `sesi:write` | — |
| DELETE | `/api/praktikum/admin/sesi/:id` | `/api/admin/sesi/:id` | asisten, koordinator | `sesi:write` | — |
| GET | `/api/praktikum/admin/sesi/:id/course` | `/api/admin/sesi/:id/course` | asisten, koordinator | — | — |
| POST | `/api/praktikum/admin/sesi/:id/course` | `/api/admin/sesi/:id/course` | asisten, koordinator | `sesi:write` | — |
| PUT | `/api/praktikum/admin/course/:courseId` | `/api/admin/course/:courseId` | asisten, koordinator | `sesi:write` | — |
| DELETE | `/api/praktikum/admin/course/:courseId` | `/api/admin/course/:courseId` | asisten, koordinator | `sesi:write` | — |
| GET | `/api/praktikum/admin/soal` | `/api/admin/soal` | asisten, koordinator | — | — |
| POST | `/api/praktikum/admin/soal` | `/api/admin/soal` | asisten, koordinator | `sesi:write` | — |
| PUT | `/api/praktikum/admin/soal/:id` | `/api/admin/soal/:id` | asisten, koordinator | `sesi:write` | — |
| DELETE | `/api/praktikum/admin/soal/:id` | `/api/admin/soal/:id` | asisten, koordinator | `sesi:write` | — |
| GET | `/api/praktikum/admin/aktivasi` | `/api/admin/aktivasi` | asisten, koordinator | — | — |
| POST | `/api/praktikum/admin/aktivasi` | `/api/admin/aktivasi` | asisten, koordinator | `aktivasi:write` | — |
| GET | `/api/praktikum/admin/aktivasi/:id` | `/api/admin/aktivasi/:id` | asisten, koordinator | — | — |
| DELETE | `/api/praktikum/admin/aktivasi/:id` | `/api/admin/aktivasi/:id` | asisten, koordinator | `aktivasi:write` | — |
| POST | `/api/praktikum/admin/aktivasi/:id/token` | `/api/admin/aktivasi/:id/token` | asisten, koordinator | `aktivasi:write` | — |
| POST | `/api/praktikum/admin/aktivasi/:id/susulan` | `/api/admin/aktivasi/:id/susulan` | asisten, koordinator | `aktivasi:write` | — |
| GET | `/api/praktikum/admin/aktivasi/:id/susulan` | `/api/admin/aktivasi/:id/susulan` | asisten, koordinator | — | — |
| GET | `/api/praktikum/admin/aktivasi/:id/peserta` | `/api/admin/aktivasi/:id/peserta` | asisten, koordinator | — | — |
| DELETE | `/api/praktikum/admin/aktivasi/:id/susulan/:mahasiswaId` | `/api/admin/aktivasi/:id/susulan/:mahasiswaId` | asisten, koordinator | `aktivasi:write` | — |
| POST | `/api/praktikum/admin/aktivasi-course/buka-tutup` | `/api/admin/aktivasi-course/buka-tutup` | asisten, koordinator | `aktivasi:write` | — |
| POST | `/api/praktikum/admin/aktivasi-course/buka-tutup-peserta` | `/api/admin/aktivasi-course/buka-tutup-peserta` | asisten, koordinator | `aktivasi:write` | — |
| GET | `/api/praktikum/admin/penilaian/rekap` | `/api/admin/penilaian/rekap` | asisten, koordinator | — | — |
| POST | `/api/praktikum/admin/penilaian` | `/api/admin/penilaian` | asisten, koordinator | `penilaian:write` | — |
| POST | `/api/praktikum/admin/keaktifan` | `/api/admin/keaktifan` | asisten, koordinator | `penilaian:write` | — |
| GET | `/api/praktikum/admin/penilaian/ai-grade/targets` | `/api/admin/penilaian/ai-grade/targets` | asisten, koordinator | — | — |
| POST | `/api/praktikum/admin/penilaian/ai-grade/one` | `/api/admin/penilaian/ai-grade/one` | asisten, koordinator | `penilaian:write` | — |
| GET | `/api/praktikum/admin/rekap-jawaban` | `/api/admin/rekap-jawaban` | asisten, koordinator | — | — |
| POST | `/api/praktikum/admin/penilaian/bulk-action` | `/api/admin/penilaian/bulk-action` | asisten, koordinator | `penilaian:write` | — |
| POST | `/api/praktikum/admin/jawaban/inject` | `/api/admin/jawaban/inject` | asisten, koordinator | `penilaian:write` | — |
| GET | `/api/praktikum/admin/rekap/kelas/:id_kelas` | `/api/admin/rekap/kelas/:id_kelas` | asisten, koordinator | — | — |
| POST | `/api/praktikum/admin/upload` | `/api/admin/upload` | asisten, koordinator | `sesi:write` | — |
| GET | `/api/praktikum/admin/ampuan` | `/api/admin/ampuan` | asisten, koordinator | — | — |
| POST | `/api/praktikum/admin/ampuan` | `/api/admin/ampuan` | asisten, koordinator | `asisten:write` | — |
| DELETE | `/api/praktikum/admin/ampuan/:id` | `/api/admin/ampuan/:id` | asisten, koordinator | `asisten:write` | — |
| GET | `/api/praktikum/admin/audit-logs` | `/api/admin/audit-logs` | asisten, koordinator | — | — |
| GET | `/api/praktikum/admin/laporan-besar` | `/api/admin/laporan-besar` | asisten, koordinator | — | — |
| POST | `/api/praktikum/admin/laporan-besar` | `/api/admin/laporan-besar` | asisten, koordinator | `sesi:write` | — |
| GET | `/api/praktikum/admin/laporan-besar/:id` | `/api/admin/laporan-besar/:id` | asisten, koordinator | — | — |
| PUT | `/api/praktikum/admin/laporan-besar/:id` | `/api/admin/laporan-besar/:id` | asisten, koordinator | `sesi:write` | — |
| DELETE | `/api/praktikum/admin/laporan-besar/:id` | `/api/admin/laporan-besar/:id` | asisten, koordinator | `sesi:write` | — |
| PUT | `/api/praktikum/admin/laporan-besar/:id/target` | `/api/admin/laporan-besar/:id/target` | asisten, koordinator | `sesi:write` | — |
| GET | `/api/praktikum/admin/laporan-besar/:id/pengumpulan` | `/api/admin/laporan-besar/:id/pengumpulan` | asisten, koordinator | — | — |
| PUT | `/api/praktikum/admin/laporan-besar/:id/pengumpulan/:mhsId/mark` | `/api/admin/laporan-besar/:id/pengumpulan/:mhsId/mark` | asisten, koordinator | `sesi:write` | — |
| PUT | `/api/praktikum/admin/laporan-besar/:id/pengumpulan/:mhsId/override` | `/api/admin/laporan-besar/:id/pengumpulan/:mhsId/override` | asisten, koordinator | `sesi:write` | — |
| DELETE | `/api/praktikum/admin/laporan-besar/:id/pengumpulan/:mhsId/override` | `/api/admin/laporan-besar/:id/pengumpulan/:mhsId/override` | asisten, koordinator | `sesi:write` | — |
| GET | `/api/praktikum/admin/laporan-besar/:id/pengumpulan/:mhsId/preview` | `/api/admin/laporan-besar/:id/pengumpulan/:mhsId/preview` | asisten, koordinator | — | — |
| GET | `/api/praktikum/admin/laporan-besar/:id/pengumpulan/:mhsId/download` | `/api/admin/laporan-besar/:id/pengumpulan/:mhsId/download` | asisten, koordinator | — | — |
| POST | `/api/praktikum/admin/laporan-besar/:id/download-batch` | `/api/admin/laporan-besar/:id/download-batch` | asisten, koordinator | `sesi:write` | — |

### Superadmin
| Method | Path baru | Path lama (alias) | Akses | Izin | Rate limit |
|---|---|---|---|---|---|
| GET | `/api/praktikum/superadmin/permissions` | `/api/superadmin/permissions` | koordinator | — | — |
| PUT | `/api/praktikum/superadmin/permissions` | `/api/superadmin/permissions` | koordinator | — | — |

### Dihapus
| Method | Path | Keterangan |
|---|---|---|
| POST | `/api/praktikum/compile-c` | Compile C pindah ke browser |

## Siakad

- Endpoint siakad lama (Vercel function di `siakad.algohub.web.id/api/*`) pindah ke `api.algohub.web.id/api/siakad/*`, dan cron-nya ke `/api/cron/*`. Path lama tidak punya alias di backend ini.
- Body response sama dengan kode lama: JSON mentah (`{"error": "..."}` untuk error, data RPC apa adanya), bukan envelope praktikum.
- Token sesi dikirim lewat cookie `siakad_token` yang ditulis login: `Domain=.algohub.web.id; Path=/; Max-Age=3600; HttpOnly; Secure; SameSite=Lax`. Logout dan verify yang gagal menghapusnya (`Max-Age=0`). Kalau cookie tidak ada, header `Authorization: Bearer <token>` dipakai (token juga dikirim di body login). `analyze-schedule` dan `sync-elearning` hanya membaca header Bearer, seperti kode lama.
- Rate limit 5/menit/IP, kunci IP sama dengan praktikum: login, dan rpc khusus `fn=login_user`/`register_user`. Kuota keduanya terpisah. Ditolak `429 {"error": ...}`.
- Method di luar kolom Method dibalas `405 {"error":"Method not allowed"}`. `send-push` dan `send-wa` membalas `OPTIONS` non-preflight dengan 200 kosong.
- Cron memakai `Authorization: Bearer <CRON_SECRET>` (dikirim Vercel Cron) atau `X-Cron-Secret: <CRON_SECRET>`. Kalau `CRON_SECRET` kosong, semua request ditolak `401 {"error":"Unauthorized: Access restricted to Vercel Cron only"}`. Jadwal ada di `backend/vercel.json`: check-shifts `0 7 * * *`, keep-alive `0 0 * * *`.

| Method | Lama (`siakad.algohub.web.id`) | Baru (`api.algohub.web.id`) | Auth | Catatan |
|---|---|---|---|---|
| POST | `/api/auth/login` | `/api/siakad/auth/login` | publik | 5/menit/IP; set cookie `siakad_token` |
| POST | `/api/auth/logout` | `/api/siakad/auth/logout` | cookie/Bearer opsional | 204; mencabut token (`revoke_token`) dan menghapus cookie |
| POST | `/api/auth/verify` | `GET`/`POST /api/siakad/auth/verify` | cookie `siakad_token` atau Bearer | token dicabut atau invalid → 401 dan cookie dihapus |
| POST | `/api/rpc?fn=` | `/api/siakad/rpc?fn=` | cookie/Bearer, kecuali `PUBLIC_RPCS` | whitelist RPC per role; `login_user` → 403; `p_*_id` diisi dari token |
| POST | `/api/generate-jwt` | `/api/siakad/generate-jwt` | `handshake_code` di body | token SSO elearning berlaku 2 jam |
| POST | `/api/analyze-schedule` | `/api/siakad/analyze-schedule` | Bearer, role asisten/koordinator | Gemini (default) atau Bynara `stepfun-3.7-flash` (hanya gambar) |
| POST | `/api/send-push` | `/api/siakad/send-push` | `X-Cron-Secret` atau token staf | Web Push; subscription 404/410 dihapus |
| POST | `/api/send-wa` | `/api/siakad/send-wa` | `X-Cron-Secret` atau token staf | gateway dari `system_settings` |
| POST | `/api/sync-elearning` | `/api/siakad/sync-elearning` | Bearer | baca DB elearning, tulis `elearning_progress` di DB siakad |
| GET | `/api/cron-check-shifts` | `/api/cron/siakad/check-shifts` | `CRON_SECRET` | push pengingat shift jaga hari ini |
| GET | `/api/keep-alive` | `/api/cron/keep-alive` | `CRON_SECRET` | ping DB siakad dan elearning |
