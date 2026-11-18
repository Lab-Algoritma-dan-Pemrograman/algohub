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
| Env (production) | terisi sejak 2026-11-18 (Langkah A). Kecuali `DATABASE_URL_SIAKAD`, `ELEARNING_SUPABASE_JWT_SECRET`, `WA_GATEWAY_TOKEN`, dan `BYNARA_API_KEY` yang nilainya belum tersedia. Selama `DATABASE_URL_SIAKAD` kosong, route siakad dan cron `check-shifts` tidak dipasang, jadi tidak ada push ganda dengan siakad lama. |

Deployment:
- `algohub-r1g6oxqvq-…vercel.app` (preview): lulus uji `/api/health` dan CORS (lihat di bawah).
- `algohub-et9mmhajw-…vercel.app` (production, alias `algohub-api.vercel.app`): deploy pertama sebuah project selalu dijadikan production oleh Vercel. Deployment ini jalan tanpa env production, sehingga `APP_ENV` = `development` dan CORS hanya mengizinkan localhost. Isinya hanya `/api/health`. Tidak terhubung ke domain prod mana pun.

Catatan:
- Deployment Protection standar aktif. URL deployment preview mengembalikan 302 ke login Vercel. Alias production `algohub-api.vercel.app` publik.
- `vercel curl` membuat **Protection Bypass for Automation** secret di project ini (otomatis, untuk uji). Secret bisa dicabut di Settings → Deployment Protection.

### Project frontend baru (Langkah A, 2026-11-18)
| Project | ID | Root Directory | Alias production |
|---|---|---|---|
| `algohub-portal` | `prj_RLTSjcnv7qYQaMp941hlDWp0zWJ8` | `apps/portal` | `algohub-portal-gamma.vercel.app` |
| `algohub-praktikum` | `prj_uUPzNn1jct8cAivSf36ZVRqNNiYH` | `apps/praktikum` | `algohub-praktikum.vercel.app` |
| `algohub-siakad` | `prj_rg5hLev48lfWazZEvwsnmk73KEAF` | `apps/siakad` | `algohub-siakad-snowy.vercel.app` |
| `algohub-elearning` | `prj_Al0bkfl2oeptiah50Mm5AddZ0U4U` | `apps/elearning` | `algohub-elearning-kappa.vercel.app` |

- Keempat project: framework `null`, `sourceFilesOutsideRootDirectory=true`, env `PUBLIC_*` target production+preview. Belum ada domain kustom dan belum disambungkan ke Git.
- Nama `<project>.vercel.app` untuk portal, siakad, dan elearning sudah dipakai akun Vercel lain, jadi Vercel menambahkan akhiran (`-gamma`, `-snowy`, `-kappa`). Hanya `algohub-praktikum.vercel.app` yang bersih.
- Alias berakhiran `-lab-algoritma-pemrograman` (mis. `algohub-portal-lab-algoritma-pemrograman.vercel.app`) dilindungi Deployment Protection, jadi tidak dipakai sebagai `PUBLIC_COMPILER_BASE_URL`.
- `PUBLIC_COMPILER_BASE_URL` praktikum dan elearning: `https://algohub-portal-gamma.vercel.app/compiler/v1/` sampai cutover langkah 3. Origin yang sama ada di CSP `connect-src` kedua app.
- Hasil uji: portal `/` 200 dan redirect lama 301 ke `praktikum.algohub.web.id`; `/compiler/v1/clang` 200, `Cache-Control: immutable`, `Access-Control-Allow-Origin: *`, `Cross-Origin-Resource-Policy: cross-origin`, SHA-256 sama dengan GitHub Release. Praktikum, siakad, dan elearning `/` 200 dan bundle-nya memuat env yang benar.
- `/compiler/v1/SHA256SUMS` 404: file itu hanya dipakai `fetch-compiler` untuk verifikasi, tidak ikut ditulis ke `build/`.


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