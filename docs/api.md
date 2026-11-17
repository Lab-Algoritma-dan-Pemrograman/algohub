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

## E-Learning

Semua route di bawah `/api/elearning`. Pengganti `api/*.ts` lama dan query Supabase langsung dari browser `elearning`.

- Response memakai envelope `pkg/response`: `{success, message, data, error}`. Pesan error lama masuk `message` persis, status code sama, field tambahan lama (`receivedKeys`, `code`) masuk `error`, dan body sukses lama masuk `data`.
- Tanpa `DATABASE_URL_ELEARNING`, semua path `/api/elearning/*` dijawab `503 "Backend elearning belum dikonfigurasi (DATABASE_URL_ELEARNING)."`. Route praktikum tidak terpengaruh.
- `POST /api/elearning/login` dibatasi 10 request/menit per IP klien (`X-Forwarded-For` entri pertama atau `X-Real-IP`).
- Auth "Bearer" = header `Authorization: Bearer <token>`. Token yang diterima: token Supabase elearning (HS256/384/512, `ELEARNING_SUPABASE_JWT_SECRET`), token Go praktikum (`JWT_SECRET`), dan token siakad (`SIAKAD_JWT_SECRET`). Tanpa token atau token rusak → `401`. Semua route Bearer mengirim `Cache-Control: private, no-store`, termasuk respons `401`.
- Peran A/K/S/P = peran token ternormalisasi (`admin`, `kordas`, `asisten`, `praktikan`); aturan RLS lama ditiru di Go dengan peran ini (`docs/updates/elearning-rls.md`). "Peran UI" = `resolveElearningRole(users.role, users.elearning_role)` dari DB, dipakai untuk guard service lama.
- Penolakan mengikuti semantik RLS: baca yang tersaring → `200` kosong, UPDATE/DELETE yang tidak lolos → `200` tanpa efek, INSERT yang tidak lolos → `403 new row violates row-level security policy for table "<tabel>"`, exception trigger/RPC (`P0001`) → `400` dengan pesan apa adanya.
- Pagination `?page=&limit=`: default limit 100, maksimum 100; `data` berbentuk `{items, total, page, limit}`.

| Method | Path | Auth/peran | Sumber lama | Pagination | Catatan |
|---|---|---|---|---|---|
| POST | `/api/elearning/sso` | publik (`token` form/JSON) | `api/index.ts` | — | `303` ke `<ELEARNING_APP_URL>/sso?code=…` (code AES-GCM, sekali pakai, 60 detik); error `302` ke `/?error=invalid_token\|missing_fields\|server_error` |
| POST | `/api/elearning/receive-token` | publik | `api/receive-token.ts` | — | `{token}` atau `{code}` dari `/sso` |
| POST | `/api/elearning/login` | publik, 10/menit/IP | `api/login.ts` | — | lewat usecase login praktikum + audit `LOGIN`/`LOGIN_FAILED`; gagal → `401 "NIM/email atau password salah."` |
| POST | `/api/elearning/verify` | publik | `api/verify.ts` | — | |
| POST | `/api/elearning/ai` | Bearer atau `token` body; `users.role` admin/kordas | `api/ai.ts` | — | Gemini (rotasi `GEMINI_API_KEY`) atau Ollama `ELEARNING_OLLAMA_*`; role mentah (T-14) |
| POST | `/api/elearning/grade` | Bearer atau `token` body; `users.role` admin/kordas/asisten | `api/grade.ts` | — | maks 3 attempt dinilai bersamaan |
| POST | `/api/elearning/report` | `token` + `nim` body (atau Bearer) | `api/report.ts` | — | upsert `elearning_progress` |
| POST | `/api/elearning/validate-quiz` | Bearer | `api/validate-quiz.ts` | — | token rusak → `500` seperti lama (T-12) |
| POST | `/api/elearning/auth/supabase-token` | Bearer | `api/auth.ts` | — | token re-sign Supabase |
| GET | `/api/elearning/curriculum/levels` | Bearer | `curriculumService.ts:12`, `SupabaseProvider.tsx:370` | ya | |
| GET | `/api/elearning/curriculum/modules` | Bearer | `curriculumService.ts:38`, `SupabaseProvider.tsx:397` | ya | |
| GET | `/api/elearning/curriculum/lessons` | Bearer | `curriculumService.ts:43`, `SupabaseProvider.tsx:405` | ya | A/K/S: tabel `lessons`; P: view `student_lessons` |
| PUT | `/api/elearning/curriculum` | A/K | `curriculumService.saveFullCurriculum` | — | `{levels:[...]}`, satu transaksi; S/P → `403` (`levels`) |
| DELETE | `/api/elearning/curriculum` | A/K | `curriculumService.clearCurriculum` | — | S/P → `200` tanpa efek |
| GET | `/api/elearning/progress/me/lessons` | Bearer | `SupabaseProvider.tsx:298` | ya | |
| GET | `/api/elearning/progress/me/recent` | Bearer | `Dashboard.tsx:81` | `?limit` (default 5, maks 100) | |
| POST | `/api/elearning/progress/lessons/:lessonId/complete` | Bearer | `progressService.completeLesson` | — | `{xpReward}` |
| POST | `/api/elearning/progress/xp` | Bearer | `progressService.grantXp` | — | `{amount}` (T-15) |
| POST | `/api/elearning/progress/study-time` | Bearer | `Layout.tsx:143` | — | `{seconds}`, increment dalam satu UPDATE |
| POST | `/api/elearning/streak` | Bearer | `Layout.tsx:218` | — | |
| POST | `/api/elearning/profile/sync` | Bearer | `SupabaseProvider.tsx:91` | — | |
| POST | `/api/elearning/activity-logs` | Bearer | `monitoringService.addAuditLog` | — | `{eventType, details, nama?}`, `nim` dari token |
| GET | `/api/elearning/activity-logs` | Bearer; A/K/S (lainnya kosong) | `monitoringService.getAuditLogs` | ya | urut `timestamp` desc |
| DELETE | `/api/elearning/activity-logs/:id` | A/K (lainnya `200` tanpa efek) | `monitoringService.deleteAuditLog` | — | |
| GET | `/api/elearning/achievements` | Bearer | `achievementService.ts:116` | tidak (katalog) | `?requirementType=` |
| GET | `/api/elearning/achievements/me` | Bearer | `Profile.tsx:27` | tidak | |
| POST | `/api/elearning/achievements/check` | Bearer | `achievementService.checkAndUnlockAchievements` | — | |
| POST | `/api/elearning/achievements/check-xp` | Bearer | `achievementService.ts:145` | — | `{xp}` |
| GET | `/api/elearning/admin/users` | Bearer | `AdminDashboard.tsx:1219` | ya | `?q=` (nama/email) |
| GET | `/api/elearning/admin/users/:nim/progress` | Bearer; progres orang lain hanya A/K/S | `AdminDashboard.tsx:1266` | ya | |
| PATCH | `/api/elearning/admin/users/:nim/level-access` | peran UI A/K | `AdminDashboard.tsx:1196` | — | `{levelAccessOverrides}` |
| PATCH | `/api/elearning/admin/users/:nim/role` | guard `handleToggleRole` | `AdminDashboard.tsx:1328` | — | `{role}` |
| POST | `/api/elearning/admin/users/:nim/reset` | peran UI A/K/S | `progressService.resetUserProgress` | — | audit `progress_reset` |
| POST | `/api/elearning/admin/users/:nim/reset-level` | peran UI A/K/S | `progressService.resetLevelProgress` | — | `{levelId}` |
| POST | `/api/elearning/admin/users/:nim/reset-lesson` | peran UI A/K/S | `progressService.resetLessonProgress` | — | `{lessonId, lessonTitle?}` |
| PUT | `/api/elearning/admin/users/:nim/xp` | peran UI A/K | `progressService.adjustUserXp` | — | `{xp}` |
| DELETE | `/api/elearning/admin/users/:nim` | peran UI A/K | `progressService.deleteUser` | — | |
| GET | `/api/elearning/game/questions` | Bearer | `gameService.getGameQuestions` | `?limit` (default 5, maks 100) | `?language=`; diacak lalu dipotong |
| POST | `/api/elearning/game/questions` | A/K (lainnya `403`) | `gameService.addGameQuestion` | — | `{language, difficulty, title, code, bugLine, explanation}` → `{id}` |
| PUT | `/api/elearning/game/questions` | A/K (lainnya `403`) | `gameService.forceResetGameQuestions` | — | `{questions:[...]}`, hapus semua + insert dalam satu transaksi |
| POST | `/api/elearning/game/questions/seed` | Bearer; insert hanya A/K | `gameService.seedInitialQuestions` | — | `{questions:[...]}`, hanya bila bank soal kosong → `{seeded}` |
| PATCH | `/api/elearning/game/questions/:id` | A/K (lainnya `200` tanpa efek) | `gameService.updateGameQuestion` | — | field kosong dilewati, `bugLine` ditulis bila dikirim |
| DELETE | `/api/elearning/game/questions/:id` | A/K (lainnya `200` tanpa efek) | `gameService.deleteGameQuestion` | — | |
| GET | `/api/elearning/game/settings` | Bearer | `gameService.getGameSettings` | — | baris `default` tidak ada/error → `{true,true,true,3,5}` |
| PUT | `/api/elearning/game/settings` | A/K (lainnya `403`) | `gameService.updateGameSettings` | — | upsert `id='default'`, field yang dikirim saja |
| GET | `/api/elearning/game/plays/me` | Bearer | `gameService.canPlayBugHunt` | — | `{allowed, playsUsed, limit}`, minggu mulai Senin 00:00 WIB |
| POST | `/api/elearning/game/history` | Bearer | `gameService.saveGameHistory` | — | `{gameType, xpEarned}`; kuota habis → `400 WEEKLY_LIMIT_REACHED` |
| GET | `/api/elearning/leaderboard` | Bearer | `leaderboardService.getLeaderboard` | ya (default 10, `limit=0` → 100) | `?kelas&jurusan&shift`, role `mahasiswa`, urut xp desc |
| GET | `/api/elearning/leaderboard/rank` | Bearer | `leaderboardService.getUserRank` | — | `?xp&kelas&jurusan&shift`; xp 0 → `rank: null` |
| GET | `/api/elearning/leaderboard/filters` | Bearer | `leaderboardService.getLeaderboardFilters` | — | |
| GET | `/api/elearning/leaderboard/weekly` | Bearer | `leaderboardService.getWeeklyLeaderboard` | ya (default 10, `limit=0` → 100) | P hanya melihat XP mingguan miliknya, reward lesson 60 (T-19) |
| GET | `/api/elearning/prestige/eligibility` | Bearer | `prestigeService.checkPrestigeEligibility` | — | `?nim=` (default diri sendiri) |
| POST | `/api/elearning/prestige/reset` | diri sendiri atau peran UI A/K/S | `prestigeService.doVoluntaryPrestigeReset` | — | `{nim?}`; RPC `do_prestige_reset` apa adanya; lainnya `403 "Akses ditolak."` |
| GET | `/api/elearning/playground/examples` | Bearer | `playgroundService.getPlaygroundExamples` | ya | `?language=`, urut `sort_order` |
| POST | `/api/elearning/playground/examples` | A/K (lainnya `403`) | `playgroundService.addPlaygroundExample` | — | → `{id}` |
| PATCH | `/api/elearning/playground/examples/:id` | A/K (lainnya `200` tanpa efek) | `playgroundService.updatePlaygroundExample` | — | |
| DELETE | `/api/elearning/playground/examples/:id` | A/K (lainnya `200` tanpa efek) | `playgroundService.deletePlaygroundExample` | — | |
| PUT | `/api/elearning/sessions/me` | Bearer | `Layout.tsx:198`, `:237` | — | `{currentActivity, nama?, kelas?, attemptId?, touchStreak?, studySeconds?}`; P → `sessionSaved:false` (RLS), streak/waktu belajar tetap jalan |
| DELETE | `/api/elearning/sessions/me` | Bearer | `Layout.tsx:70` | — | P → `200` tanpa efek |
| GET | `/api/elearning/monitoring/sessions` | A/K/S (lainnya kosong) | `StudentMonitoring.tsx:62` | ya | `{nim, last_heartbeat, users:{role}}` |
| DELETE | `/api/elearning/monitoring/sessions/stale` | A/K/S (lainnya `200` tanpa efek) | `monitoringService.cleanupStaleSessions` | — | `?thresholdMinutes=2` |
| GET | `/api/elearning/monitoring/students` | Bearer | `StudentMonitoring.tsx:81` | ya | role `praktikan`/`mahasiswa` |
| GET | `/api/elearning/monitoring/progress` | Bearer; P hanya miliknya | `StudentMonitoring.tsx:111` | ya (per NIM) | `{nim, lessonIds}` |
| POST | `/api/elearning/storage/lesson-images` | A/K/S (lainnya `403`) | `storageService.uploadImage` | — | multipart `file` (JPG/PNG/GIF/WEBP) → `{publicUrl}`, bucket `lesson-images` |
