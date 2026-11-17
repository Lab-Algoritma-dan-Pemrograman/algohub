# Deploy

Rencana lengkap: `docs/updates/` (salinan `_update/2026-11-17-monorepo-sveltekit-go/08-deploy.md`).

## Hasil Fase 0 (2026-11-17)

### Project Vercel yang melayani domain produksi sekarang
Semua ada di team `lab-algoritma-pemrograman` (Hobby), akun `lordbgsdts-afk`. Sumber: `vercel alias ls`.

| Domain | Project Vercel | Deployment |
|---|---|---|
| `algohub.web.id`, `www.algohub.web.id` | `labap-frontend` | `labap-frontend-859rlu8ta-…` |
| `api.algohub.web.id` | `labap-backend` | `labap-backend-jusv7yhfo-…` |
| `siakad.algohub.web.id` | `labap-siakad` | `labap-siakad-f39nbz45m-…` |
| `elearning.algohub.web.id` | `labap-elearning` | `labap-elearning-gknsdm2fm-…` |
| `praktikum.algohub.web.id` | — (belum dipakai, 404) | — |

Domain dibeli di IDCloudHost, nameserver di Cloudflare. Record `CNAME cname.vercel-dns.com` dengan mode DNS only: IP hasil resolve (`76.76.21.x`, `66.33.60.x`) adalah IP Vercel, bukan IP Cloudflare. `vercel domains ls` menampilkan 0 domain, karena domain dikelola di luar Vercel. Itu sebabnya `vercel domains inspect` menolak, padahal alias-nya ada di team ini.

### Project Supabase per app
Hanya project ref yang dicatat. Sumber: bundle JS produksi yang publik, response publik `/api/info/*`, dan file `.env` lokal (hanya ref yang dibaca).

| App | Supabase project ref | Sumber |
|---|---|---|
| siakad | `huddecanbapwjairmaef` | bundle JS `siakad.algohub.web.id` (prod), `siakad/.env` |
| elearning | `oidwvfcxlxfyojkyurmm` | bundle JS `elearning.algohub.web.id` (prod) |
| praktikum | `oidwvfcxlxfyojkyurmm` | URL storage di `api.algohub.web.id/api/info/*` (prod), `praktikum/backend/.env` |

Kesimpulan: **2 project Supabase**, bukan 1. Siakad memakai project sendiri, sedangkan elearning dan praktikum berbagi satu project. Ref lain yang muncul di repo lama sudah tidak dipakai prod: `smebowpudhqbemttiqbo` di `siakad/supabase/config.toml`, dan `tvsawtkevzfqobsfkiag` di `elearning/.env` lokal.

Dampak ke backend (sudah sesuai desain `02-backend.md`):
- `DATABASE_URL_SIAKAD` → project `huddecanbapwjairmaef`
- `DATABASE_URL_PRAKTIKUM` = `DATABASE_URL_ELEARNING` → project `oidwvfcxlxfyojkyurmm`
- keep-alive cron harus mem-ping kedua project

Belum terverifikasi: nilai prod `ELEARNING_SUPABASE_URL` milik siakad (`sync-elearning`), karena nilai env tidak dibaca (rahasia). Cek di dashboard `labap-siakad` → Settings → Environment Variables, cukup bagian project ref.

### Scheduler eksternal
Menurut `praktikum/updateAndPRDERD/deploy.md` dan `update5.md`: **cron-job.org** mengirim `POST /api/cron/auto-submit` setiap 1 menit dengan header `X-Cron-Secret`. Status aktifnya belum terverifikasi karena butuh login cron-job.org.

### Pemakaian Vercel 30 hari terakhir
Tidak bisa dibaca lewat CLI. `vercel usage` di team Hobby mengembalikan "Billing cost data is unavailable". Pemilik perlu mencatatnya dari dashboard (Usage) team `lab-algoritma-pemrograman`.

| Metrik | 30 hari terakhir |
|---|---|
| Function Invocations | ? |
| Fluid Active CPU | ? |
| Fluid Provisioned Memory | ? |
| Fast Origin Transfer | ? |
| Fast Data Transfer | ? |
| CDN Requests | ? |

## Project baru

### `algohub-api` (Fase 1, 2026-11-17)
| Setting | Nilai |
|---|---|
| Project ID | `prj_nI7yzo8IPyQFqKQ4ITu1uyOlXUwj` |
| Team | `lab-algoritma-pemrograman` (`team_xEm9gJWxtM1sb09eyCeIN4js`) |
| Root Directory | `backend` |
| Framework | Go (Go 1.26.11 dari `toolchain` di `go.mod`) |
| Git integration | **belum disambungkan** (deploy lewat `vercel deploy` dari laptop) |
| Domain | belum ada. `api.algohub.web.id` masih di `labap-backend`. |
| Env (preview) | `APP_ENV=production`, `CORS_ORIGINS` (5 domain prod), `CORS_ORIGIN_PATTERNS=https://algohub-[a-z0-9-]+-lab-algoritma-pemrograman\.vercel\.app`. Sejak Fase 2 juga `DATABASE_URL_PRAKTIKUM` (encrypted, disusun dari `DB_*` `labap-backend`), `JWT_SECRET` (encrypted, sama dengan prod), `JWT_EXPIRE_HOURS`, `AUTH_MODE`, `SUPABASE_JWKS_URL`, `SUPABASE_JWT_ISSUER`, `SUPABASE_JWT_AUD`. `FB_SCRYPT_*` sengaja tidak dipasang supaya login uji tidak memigrasi hash password. |
| Env (production) | belum diisi. **Jangan deploy production `algohub-api` sebelum cutover**: `vercel.json` berisi cron `check-shifts` (07.00) dan `keep-alive`, sehingga deployment production akan mengirim notifikasi push shift **ganda** bersama siakad lama. |

Deployment:
- `algohub-r1g6oxqvq-…vercel.app` (preview): lulus uji `/api/health` dan CORS (lihat di bawah).
- `algohub-et9mmhajw-…vercel.app` (production, alias `algohub-api.vercel.app`): deploy pertama sebuah project selalu dijadikan production oleh Vercel. Deployment ini jalan tanpa env production, sehingga `APP_ENV` = `development` dan CORS hanya mengizinkan localhost. Isinya hanya `/api/health`. Tidak terhubung ke domain prod mana pun.

Catatan:
- Deployment Protection standar aktif. URL deployment preview mengembalikan 302 ke login Vercel. Alias production `algohub-api.vercel.app` publik.
- `vercel curl` membuat **Protection Bypass for Automation** secret di project ini (otomatis, untuk uji). Secret bisa dicabut di Settings → Deployment Protection.

Uji di preview (`vercel curl /api/health --deployment <url>`):

| Kasus | Hasil |
|---|---|
| GET, origin `https://siakad.algohub.web.id` | 200, `Allow-Origin` = origin, `Allow-Credentials: true`, `Expose-Headers`, `Vary: Origin` |
| Preflight POST dari origin prod | 204, `Allow-Methods`, `Allow-Headers`, `Max-Age: 86400` |
| GET, origin preview `algohub-portal-…-lab-algoritma-pemrograman.vercel.app` | 200 dengan header CORS |
| GET, origin asing | 200 tanpa header CORS |
| GET, origin tiruan `…vercel.app.evil.com` | 200 tanpa header CORS |
| Preflight dari origin asing | 204 tanpa header CORS |

### Smoke test Fase 2 dengan DB prod (2026-11-17)
Deployment preview `algohub-ds05mrkbf-…vercel.app` (commit `28708e4`), DB praktikum `oidwvfcxlxfyojkyurmm` lewat pooler 6543. Disetujui pemilik, dengan satu akun uji mahasiswa.

| Uji | Hasil |
|---|---|
| `GET /api/health` | 200 |
| `GET /api/praktikum/info/{asisten,laporan,modul,jadwal,jadwal/config,announcements}` | 200, JSON identik dengan `api.algohub.web.id/api/info/*` (prod) |
| `GET /api/info/asisten` (alias lama) | 200, identik dengan prod |
| `POST /api/praktikum/auth/login` (akun uji) | 200, token HS256, `role=mahasiswa`, tanpa field kredensial di response |
| `GET /api/praktikum/profile` dengan token | 200, NIM sama |
| `GET /api/praktikum/profile` tanpa token | 401 |
| `POST elearning.algohub.web.id/api/verify` (prod lama) dengan token baru | 401, lihat temuan di bawah |
| `POST elearning.algohub.web.id/api/verify` token rusak | 401 |

Data prod yang tertulis oleh uji ini, sama dengan login biasa: `users.last_login_at` akun uji dan satu baris `audit_logs` (`LOGIN`). Tidak ada data yang dihapus, dan tidak ada data user lain yang berubah.

**Temuan prod (belum diverifikasi lebih lanjut, tidak diubah):** API Vercel mengembalikan **0 env var** untuk project `labap-elearning` dan `labap-siakad`, di level project maupun shared env team. `labap-backend` punya 29. Akibatnya, `elearning/api/verify` prod tidak punya `JWT_SECRET`/`SUPABASE_JWT_SECRET` untuk memverifikasi token apa pun, jadi 401 di atas bukan regresi backend baru. Kode JWT lama dan baru identik. Pemilik perlu mengecek dashboard kedua project (Settings → Environment Variables) untuk memastikan apakah env memang kosong dan fitur apa yang terdampak di prod.