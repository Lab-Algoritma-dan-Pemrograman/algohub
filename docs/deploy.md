# Deploy

Rencana lengkap: `docs/updates/` (salinan `_update/2026-11-17-monorepo-sveltekit-go/08-deploy.md`).

## Hasil Fase 0 (2026-11-17)

### Project Vercel yang melayani domain produksi sekarang
Dicek dengan `curl -I` ke setiap domain dan `vercel project ls` di akun `lordbgsdts-afk`.

| Domain | Status | Project Vercel | Team |
|---|---|---|---|
| `algohub.web.id`, `www.algohub.web.id` | 200 | `labap-frontend` | `lab-algoritma-pemrograman` |
| `api.algohub.web.id` | `/` 404, `/api/info/*` mengembalikan data | `labap-backend` | `lab-algoritma-pemrograman` |
| `siakad.algohub.web.id` | 200 | **tidak diketahui** | team lain, tidak bisa diakses akun ini |
| `elearning.algohub.web.id` | 200 | **tidak diketahui** | team lain, tidak bisa diakses akun ini |
| `praktikum.algohub.web.id` | 404 (belum dipakai) | — | — |

Project `labap-siakad` dan `labap-elearning` ada di team `lab-algoritma-pemrograman`, tapi hanya punya URL `*.vercel.app` dan **tidak** melayani `siakad.algohub.web.id` atau `elearning.algohub.web.id`. Domain kedua subdomain itu (`vercel domains inspect`) tidak bisa diakses dari team ini.

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

Belum terverifikasi: nilai prod `ELEARNING_SUPABASE_URL` milik siakad (`sync-elearning`), karena env project siakad ada di team yang tidak bisa diakses.

### Scheduler eksternal
Menurut `praktikum/updateAndPRDERD/deploy.md` dan `update5.md`: **cron-job.org** mengirim `POST /api/cron/auto-submit` setiap 1 menit dengan header `X-Cron-Secret`. Status aktifnya belum terverifikasi karena butuh login cron-job.org.

### Pemakaian Vercel 30 hari terakhir
Tidak bisa dibaca lewat CLI. `vercel usage` di team Hobby mengembalikan "Billing cost data is unavailable". Pemilik perlu mencatatnya dari dashboard (Usage), untuk team `lab-algoritma-pemrograman` dan team yang memegang siakad/elearning.

| Metrik | lab-algoritma-pemrograman | team siakad/elearning |
|---|---|---|
| Function Invocations | ? | ? |
| Fluid Active CPU | ? | ? |
| Fluid Provisioned Memory | ? | ? |
| Fast Origin Transfer | ? | ? |
| Fast Data Transfer | ? | ? |
| CDN Requests | ? | ? |
