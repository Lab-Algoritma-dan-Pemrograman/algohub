# Changelog

Semua perubahan penting dicatat di file ini. Format mengikuti [Keep a Changelog](https://keepachangelog.com/id-ID/1.1.0/), dan versi mengikuti [Semantic Versioning](https://semver.org/lang/id/).

## [Belum dirilis]

### Persiapan cutover (2026-11-18)

#### Diubah
- Domain siakad di backend menjadi opsional. Kalau `DATABASE_URL_SIAKAD` kosong, route siakad dan cron siakad tidak dipasang, sementara praktikum dan elearning tetap jalan. Kalau DSN siakad diisi tanpa `DATABASE_URL_ELEARNING` atau `SIAKAD_JWT_SECRET`, server tetap gagal start. Sebelumnya env siakad yang belum lengkap ikut mematikan praktikum.
### Fase 9 — apps/portal (2026-11-18)

#### Ditambahkan
- `apps/portal`: landing `algohub.web.id` yang di-prerender penuh (tanpa function), dengan link ke siakad, elearning, dan praktikum, plus halaman `/404`. Aksesibilitas Lighthouse 100.
- `/compiler/v1/{clang,lld,memfs,sysroot.tar}` diunduh saat build oleh `fetch-compiler` dari GitHub Release `compiler-v1` dan dicek SHA-256. Build gagal kalau ada file yang tidak cocok. File compiler tidak disimpan di git.
- `apps/portal/vercel.json`: redirect 301 semua URL lama praktikum (`/info`, `/praktikum`, `/playground`, `/playground-c`, `/layanan`) ke `praktikum.algohub.web.id` dengan path dan query yang sama, header compiler (`immutable`, CORS `*`, CORP `cross-origin`), header keamanan, dan cache aset.
- Test Vitest untuk setiap redirect dan header, test `fetch-compiler` dengan release palsu (cocok dan tidak cocok), dan 4 e2e Playwright terhadap hasil build.

#### Diubah
- `scripts/fetch-compiler.ts` tidak lagi memakai streaming Bun, yang kadang macet di Windows.
- Test yang menjalankan proses bun diberi batas waktu 60 detik.

#### Diketahui
- Header HSTS `includeSubDomains; preload` memaksa HTTPS di semua `*.algohub.web.id`. Sebelum cutover langkah 3, pastikan tidak ada subdomain yang hanya melayani HTTP.
- Font KaTeX ±1,2 MB ikut masuk build portal lewat CSS `@algohub/ui` (T-73).

### Fase 8 — apps/elearning (2026-11-18)

#### Ditambahkan
- `apps/elearning`: SvelteKit 3 SPA statis dengan 13 route. Navigasi berbasis store lama diganti URL: `/`, `/login`, `/sso`, `/dashboard`, `/courses`, `/lesson/[id]`, `/playground`, `/leaderboard`, `/profile`, `/admin`, `/monitoring`, `/auditlog`, dan `/terminal-demo`. Guard role sama dengan `App.tsx` lama.
- Login lewat SSO (`/sso?code=`, kode ditukar ke `receive-token`) atau login langsung. Token Realtime diambil dari `auth/supabase-token` dan diperbarui sebelum kedaluwarsa. `supabase-js` hanya dipakai untuk `.channel()`.
- Compiler C dan Python lewat `@algohub/runner`. Grafik matplotlib dari `images[]`.
- PWA lewat `@vite-pwa/sveltekit`: manifest dan service worker. Service worker tidak meng-cache respons API (NetworkOnly).
- `@algohub/api-client/elearning` (client untuk 64 route + `elearningKeys`) dan `@algohub/ui` MarkdownRenderer serta RichTextRenderer.
- 243 unit test dan 30 e2e Playwright (mock API + mock Realtime, compiler asli). Script root `dev:elearning`.

#### Diubah
- Yang dipersist ke IndexedDB hanya levels, modules, dan leaderboard. Query lesson **tidak** dipersist, karena staf menerima kunci jawaban (`correctAnswer`, `solution`). Akibatnya daftar lesson juga tidak tersedia offline untuk mahasiswa.

#### Diperbaiki
- E2E lesson Python kadang gagal karena menunggu teks "Menyiapkan" hilang, padahal cek itu bisa lolos sebelum teks dirender. Sekarang e2e menunggu tombol Jalankan aktif. Diverifikasi 30/30 tiga kali berturut-turut.

#### Diketahui
- Belum diuji dengan backend nyata: SSO dari siakad, Realtime, instal PWA offline, upload gambar lesson, dan AI grading.
- Pola tunggu yang sama masih ada di `playground.spec.ts` elearning dan `compiler.spec.ts` praktikum. Keduanya stabil di semua run (T-72).
### Fase 6 — apps/siakad (2026-11-18)

#### Ditambahkan
- `apps/siakad`: SvelteKit 3 SPA statis dengan 21 halaman. Role dan menu key sama dengan `App.tsx` lama, dan menu key dicek lewat `get_division_access_secure` seperti `lib/auth.tsx` lama.
- Data lewat backend Go (`/api/siakad/*`) dengan TanStack Query, `credentials: 'include'`, dan Bearer. Yang dipersist ke IndexedDB hanya pengaturan publik, jadwal, tautan, dan katalog barang.
- Realtime Supabase hanya untuk `.channel()` (mati kalau `PUBLIC_SUPABASE_URL` kosong). Service worker `/service-worker.js` menangani push dan klik notifikasi.
- `@algohub/api-client/siakad`: `createSiakadApi` (`rpc` dengan `{ data, error }`, `.single()`, `.maybeSingle()`, endpoint auth/notify/AI/SSO) dan `siakadKeys`.
- `@algohub/ui`: input, label, select, tabs, badge, switch, checkbox, textarea, separator, sheet, tooltip, skeleton, sidebar, dan hook `is-mobile`.
- E2E Playwright `apps/siakad/e2e` (mock API + mock Realtime): login per role, guard role dan menu key, aksi tulis halaman inti, Realtime, dan logout. Hasilnya 21 lulus. Script root `dev:siakad`.

#### Diubah
- Sesi memakai `token`/`user` dari `@algohub/auth`, bukan `lab_jwt_token`. User perlu login ulang setelah cutover.
- `/service-worker.js` menggantikan `/sw.js`, jadi subscription push lama perlu subscribe ulang.
- Logout berurutan: unsubscribe push → revoke token → bersihkan cache Query, IndexedDB, dan token.

#### Diketahui
- Belum diuji dengan backend nyata, dan SQL siakad masih diuji dengan mock (lihat Fase 5).
- Aksi tulis kritik-saran, buat-qr, manajemen-jadwal, manajemen-kelas, ketersediaan, laporan-keuangan, manajemen-sewa, dan pengaturan belum punya e2e.
### Fase 4 — apps/praktikum (2026-11-18)

#### Ditambahkan
- `apps/praktikum`: SvelteKit 3 SPA (`adapter-static`, fallback `index.html`, `ssr = false`) dengan 39 route lama. Isinya: info, layanan, playground, auth, mahasiswa, ruang pengerjaan pretest/posttest/keterampilan/ujian, 15 halaman admin, dan superadmin/pengaturan. Guard token dan role ada di `+layout.ts`.
- Semua data lewat TanStack Query. Persist IndexedDB per user (`algohub-praktikum-<id>`) hanya untuk whitelist (`info.*`, `admin.kelas`, `profile`). Logout mengosongkan cache memori dan IndexedDB.
- Package `@algohub/api-client` (modul praktikum + query key factory), `@algohub/query` (QueryClient + persister IndexedDB), dan `@algohub/auth` (sesi, `requireRole`, logout). `@algohub/ui` mendapat DataTable, VirtualList, dan RichTextEditor. `@algohub/runner` mendapat `compilerDir()` dan `formatResult()`.
- `apps/praktikum/vercel.json`: header keamanan, CSP, COOP/COEP di semua path, cache `/runner` dan `/_app/immutable`, dan rewrite SPA.
- `scripts/fetch-compiler.ts`: unduh Release `compiler-v1` dan cocokkan dengan `SHA256SUMS`. Script root `dev:praktikum`, `fetch-compiler`, dan `test:e2e`.
- E2E Playwright `tests/e2e/praktikum` (mock API same-origin, compiler asli): auth per role, compiler C/Python interaktif, sesi sampai submit dan timer habis, rekap nilai 1.200 baris dengan Virtual, admin, dan persist. Hasilnya 30 lulus.

#### Diubah
- Halaman memanggil API di `/api/praktikum/*`. Token disimpan di origin baru, jadi pengguna perlu login ulang saat cutover.
- Worker compiler hanya dibuat di halaman yang punya editor, dan dashboard mem-prefetch file compiler.
- Cache Query: `staleTime` default 30 detik. `/info` dan `/info/jadwal` 5 menit (pengganti `setHeaders` 300 detik). Dashboard, sesi, dan ruang `0`.
- Root `tsconfig.json` memakai `types: ['bun']`. Root `package.json` mendapat `@types/bun` dan `@playwright/test`.
- ESLint dan Prettier mengabaikan `.worktrees/` dan `.agents/` supaya lint lokal tidak ikut memeriksa worktree fase lain.

#### Diketahui
- E2E belum masuk CI.
- Galeri `/info` (`src/lib/galeri`, 39 file, ±17,9 MB) dibawa apa adanya, dan berpengaruh pada Fast Data Transfer. Kandidat kompresi di update berikutnya.
- Belum diuji dengan backend nyata: login hash Firebase, upload/download laporan besar (Storage), AI grading, cron auto-submit, dan rate limit `/run`.
- Temuan T-43..T-71.
### Fase 7 — Backend elearning (2026-11-18)

#### Ditambahkan
- Domain elearning di backend Go: 64 route di `/api/elearning/*`, menggantikan 10 Vercel Function `elearning/api/*.ts` dan 59 query Supabase yang dulu dijalankan langsung dari browser (53 menjadi endpoint, 5 kode mati tidak dibawa, 1 keep-alive milik Fase 5). Peta lengkapnya ada di `docs/updates/elearning-rls.md` § Peta query, dan tabel endpoint di `docs/api.md` § E-Learning.
- Auth: `verify` menerima 3 format token (siakad, elearning, Go praktikum). Juga `receive-token`, `login` (lewat usecase praktikum dengan audit dan rate limit 10/menit per IP), `auth/supabase-token` (re-sign JWT Supabase, deteksi encoding secret dari `auth.ts`), dan `POST /sso` (form dari siakad → 303 ke `/sso?code=…`, code AES-256-GCM sekali pakai dengan umur ≤ 60 detik).
- `ai` (rotasi key Gemini), `grade` (konkurensi 3, Gemini atau Ollama), `report`, dan `validate-quiz`.
- Kurikulum, progres, streak, profil, pencapaian, admin users, game Bug Hunt, leaderboard (termasuk mingguan), prestige (RPC `do_prestige_reset`), playground, sesi, monitoring, activity log, dan upload `lesson-images`.
- RLS ditiru di `usecase/elearning`: 137 `CREATE POLICY` dipetakan (56 aktif). Aturannya beserta column grant `users` dan trigger `auth_role()` didokumentasikan di `docs/updates/elearning-rls.md`, dengan 124 test diizinkan/ditolak untuk 62 aturan.
- `pkg/gemini.GenerateParts` (API key lewat header `x-goog-api-key`, bukan URL) dan `pkg/ollama/native.go` (API native Ollama untuk elearning).

#### Diubah
- Respons memakai envelope `pkg/response`. Pesan dan status code lama tetap sama.
- Error baca DB yang dulu ditelan menjadi `[]`/`0`/`null` sekarang 500 (game, leaderboard, playground, sebagian progres). Error non-Postgres dibalas `Internal server error`.
- Daftar dipaginasi maksimal 100 (`/monitoring/*`, `/activity-logs`, `/admin/users`). `PUT /curriculum` dan `PUT /game/questions` berjalan dalam satu transaksi.
- `golang.org/x/text` menjadi dependency langsung (`collate` untuk urutan leaderboard mingguan).

#### Diketahui
- SQL repository hanya diuji dengan mock. Sebelum cutover, wajib diuji di salinan non-prod DB elearning, lalu `select * from pg_policies where schemaname='public'` dicocokkan dengan `docs/updates/elearning-rls.md`.
- Perilaku lama yang salah tetap ditiru dan dicatat sebagai T-23..T-40 (misalnya koordinator ditolak di `ai`/`grade`, dan `xp_earned` game tanpa batas).
### Fase 5 — Backend siakad (2026-11-18)

#### Ditambahkan
- Domain siakad di backend Go (`internal/*/siakad/`, `handler/siakad/`, `handler/cron/`, `router_siakad.go`, `config/siakad.go`, `app/siakad.go`), port dari 11 file `siakad/api/*.ts`. Pemetaan endpoint lama → baru ada di `docs/api.md` § Siakad.
- Gateway `POST /api/siakad/rpc?fn=<nama>`: alias role, `PUBLIC_RPCS`, `ROLE_PERMISSIONS`, injeksi `p_*_id` dari JWT, dan `stripCredentials`, sama dengan `rpc.ts`. Nama fungsi hanya diambil dari whitelist. Nama dan tipe argumen dibaca dari `pg_proc`, dan seluruh body dikirim sebagai satu parameter `?::json` (`json_to_record`).
- Auth siakad: login (rate limit, bcrypt lewat `login_user`, `system_settings`), logout (`revoke_token`), dan verify (cookie `siakad_token` `Domain=.algohub.web.id; SameSite=Lax` atau header `Authorization`).
- `generate-jwt` (SSO ke elearning), `analyze-schedule` (Gemini atau Bynara, Stepfun hanya menerima gambar), `send-push` (VAPID), `send-wa`, dan `sync-elearning` (membaca `DBs.Elearning`, menulis `elearning_progress` di `DBs.Siakad`, sesuai lokasi tabel di prod).
- Cron `GET /api/cron/siakad/check-shifts` (07.00) dan `GET /api/cron/keep-alive` (00.00, ping kedua project Supabase) dengan `CRON_SECRET` Bearer yang dibandingkan secara constant-time. Keduanya terdaftar di `backend/vercel.json`.
- Package `pkg/jwt/roles.go` (satu `NormalizeRole`), `pkg/gemini`, `pkg/bynara`, `pkg/webpush` (`SherClockHolmes/webpush-go`), dan `pkg/whatsapp`.
- Test matriks role × RPC, injeksi identitas, cookie, cron, push, WA, AI (httptest), sync, dan seam route praktikum + siakad pada satu engine.

#### Diubah
- `config.Load()` mewajibkan `DATABASE_URL_SIAKAD`, `DATABASE_URL_ELEARNING`, dan `SIAKAD_JWT_SECRET`. Tanpa ketiganya, server gagal start.
- Cron siakad sekarang wajib `CRON_SECRET`. Sebelumnya `check-shifts` terbuka dan `keep-alive` hanya mengecek header `x-vercel-cron`, yang bisa dipalsukan.
- Error 500 yang bukan dari Postgres (misalnya gagal konek pgx) dibalas `Internal server error`, tidak lagi memuat user, host, atau nama database.

#### Diketahui
- SQL baru hanya diuji dengan sqlmock. Beberapa kolom belum dicocokkan dengan skema prod (`schedule_assignments.*`, `users.phone`, `users.nama/class_code/major`, `sso_handshakes.user_id`, `group_assistants.schedule_id`). Wajib diuji di preview sebelum cutover.
- Satu `CRON_SECRET` dipakai bersama oleh cron praktikum (cron-job.org), cron siakad, dan bypass `send-push`/`send-wa` (`T-41`).
### Fase 2 — Backend praktikum (2026-11-17)

#### Dipindahkan
- Backend praktikum dari repo `praktikum/backend` ke `backend/` (modul `github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend`), dengan `dto`, `entity`, `repository`, `usecase`, dan `handler` di subfolder `praktikum/`. `pkg/ollama` tidak lagi meng-import config.

#### Ditambahkan
- 140 route praktikum di `/api/praktikum/*`, serta cron di `/api/cron/praktikum/auto-submit` dan `/api/cron/praktikum/rumpang`. 129 path lama tetap dilayani sebagai alias (`ALIAS-LAMA`) dengan handler, middleware, dan limiter yang sama sampai cutover. 11 route mahasiswa sudah di `/api/praktikum/*` sejak awal.
- Env `DATABASE_URL_PRAKTIKUM` (wajib, server gagal start dengan pesan jelas kalau kosong), `DATABASE_URL_SIAKAD`, dan `DATABASE_URL_ELEARNING` (dibaca, tapi pool baru dibuka di Fase 5/7).
- Test route (140 baru, 129 alias, `compile-c` 404, tanpa prefix ganda, swagger hanya di luar production, rantai auth 401, limiter login dipakai bersama path baru dan alias), test `ClientIP`, rate limit per IP dan per user, serta test config dan pool DB.
- Tabel pemetaan 140 route lama → baru di `docs/api.md`.

#### Diubah
- Koneksi DB: satu pool per DSN unik (5 open, 2 idle, idle 1 menit, lifetime 5 menit) dengan `PreferSimpleProtocol` untuk pooler transaction mode. `AutoMigrate`, seed, dan goroutine sweeper tidak dijalankan.
- Rate limit memakai IP klien dari `X-Forwarded-For`/`X-Real-IP` (diisi Vercel dan tidak bisa dipalsukan dari luar), bukan IP proxy. Di prod lama semua user berbagi satu kuota.
- `JWT_SECRET` default di production ditolak saat start lewat error (dulu `panic`), dengan pesan yang sama.
- `/swagger/*any` hanya dipasang di luar production. Docs swag di-generate ulang ke `backend/internal/delivery/http/docs`.
- Kode dirapikan sampai `golangci-lint` 0 issue (dari 73) tanpa perubahan perilaku endpoint: helper generik untuk handler dan repository jadwal/kelas/pedoman, dan request besar dioper lewat pointer. Komentar non-swag dikurangi dari ±1060 ke 647 baris, dan 14 kode audit lama tetap ada.

#### Dihapus
- `POST /compile-c` (`compile_handler.go`), `pkg/cwasm`, `pkg/serverless`, `api/index.go`, `Dockerfile`, `.dockerignore`, `database/migration`, `database/seed`, goroutine sweeper.
- Env `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_PARAMS` (digabung ke `DATABASE_URL_PRAKTIKUM`), `CWASM_CLANG`, `WASI_SYSROOT`, `CWASM_MAX_CONCURRENT`, `CWASM_QUEUE_TIMEOUT_SECONDS`, `SWEEPER_INTERVAL_SECONDS`, `APP_PORT`.

#### Keamanan
- `github.com/golang-jwt/jwt/v5` v5.2.2 (GO-2025-3553), `github.com/jackc/pgx/v5` v5.9.2 (GO-2026-5004), `filippo.io/edwards25519` v1.1.1 (GO-2026-4503). `govulncheck`: tidak ada kerentanan.

#### Diketahui
- Belum diuji terhadap DB nyata, termasuk kenaikan pgx. Smoke test preview (health, login, verify token) menunggu izin pemilik untuk memakai DB.
- Temuan bug lama yang tidak diperbaiki dicatat di `docs/updates/temuan.md`.
### Fase 3 — Runner dan UI dasar (2026-11-17)

#### Ditambahkan
- `@algohub/runner`: C (clang WASM) dan Python (Pyodide v0.27.0, tiga mirror jsDelivr, retry init 2×) di Web Worker. Input interaktif lewat SharedArrayBuffer untuk kedua bahasa, dan baris yang diketik sebelum diminta masuk antrean. Mode batch `run(code, { stdin })` juga tersedia.
- Timeout C 40 detik dimulai setelah compile, dijeda saat menunggu input, dan dimulai ulang setelah input. Saat habis, worker di-terminate lalu dibuat ulang.
- File compiler (`clang`, `lld`, `memfs`, `sysroot.tar`) diunduh sekali per halaman. Base URL diatur lewat `PUBLIC_COMPILER_BASE_URL`. Biner tidak disimpan di repo.
- Plugin Vite `@algohub/runner/vite` (`runnerWorkers()`) yang melayani dan meng-emit worker same-origin di `/runner/`.
- `@algohub/ui`: Button, Card, Dialog, Toaster/toast (shadcn-svelte + bits-ui 2), AppLayout, CodeEditor (Monaco dari CDN lewat `@monaco-editor/loader`), Terminal (xterm 6 dengan penyunting baris).
- `@algohub/config`: dukungan Svelte (eslint-plugin-svelte, prettier-plugin-svelte) dan `tailwind.css` (Tailwind 4) dengan tema warna praktikum.
- Harness dev `packages/runner/harness` (tidak dideploy) dan 10 uji Playwright Chromium: crossOriginIsolated, C dua `scanf`, error compile, Python `input()` di dalam `def`, reset setelah timeout, batch C/Python, Stop C/Python, dan Monaco dari CDN. Ditambah 92 test Vitest runner dan 4 test UI.
- `prettier-plugin-svelte` sebagai devDependency root.

#### Diubah
- Perilaku runner disatukan dari praktikum dan elearning: COEP `credentialless`, injeksi `setbuf` versi praktikum (tambah `#include <stdio.h>`), dan Stop C memakai terminate + re-init. Python `input()` memakai SAB + `pyodide.setStdin`, menggantikan regex lama yang rusak untuk `input()` di dalam `def`. Grafik matplotlib dikirim sebagai `images[]`.
- `.prettierignore`: `packages/runner/static/vendor/` (wasm-clang `shared.js`, Apache-2.0).

#### Diketahui
- Stop Python pada program CPU-bound (`while True: pass`) tidak menghentikan program, sama dengan kedua app lama.
- `Runner.stop()` dan `Runner.sendInput()` belum dibedakan per bahasa.
### Fase 1 — Kerangka monorepo (2026-11-17)

#### Ditambahkan
- Monorepo Bun workspaces (`apps/*`, `packages/*`) dengan `packageManager: bun@1.3.14`. Script root: `lint`, `format`, `check`, `test`, `build`, `dev:api`.
- `packages/config` (`@algohub/config`): `tsconfig.json` dasar, ESLint flat config (`no-console` kecuali `console.error`, `eqeqeq`, `no-debugger`), dan konfigurasi Prettier.
- Folder kosong (`.gitkeep`) untuk `apps/`, `packages/{ui,api-client,auth,query,runner}`, `scripts/`, `tests/e2e/`, `backend/migrations/`, `backend/pkg/`, `docs/updates/`.
- `backend/` modul `github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend` dengan entrypoint tunggal `cmd/server/main.go`: listen `:$PORT` di Vercel, `127.0.0.1:8080` di laptop.
- `GET /api/health` → `{"service":"algohub-api","status":"ok"}`.
- Middleware CORS (`internal/middleware/cors.go`) dipasang paling luar: allowlist eksak `CORS_ORIGINS` + regex `CORS_ORIGIN_PATTERNS` (otomatis di-anchor), kredensial, `Vary: Origin`, preflight `204` dengan `Access-Control-Max-Age: 86400`, dan header CORS ikut di response 401/404/500.
- `internal/config`: `APP_ENV`, `CORS_ORIGINS`, `CORS_ORIGIN_PATTERNS`. Tanpa `CORS_ORIGINS`, production gagal start dan non-production memakai `http://localhost:5173`–`5176`.
- Test config dan CORS (`internal/config/config_test.go`, `internal/delivery/http/cors_test.go`).
- `backend/vercel.json`: `framework: go`, region `sin1`.
- `.golangci.yml` (golangci-lint v2: standard + `dupl`, `gocritic`, `misspell`, `unconvert`, `unparam`; formatter `gofmt`, `goimports`).
- CI `.github/workflows/ci.yml`: job `web` (lint, check, test, build), job `go` (vet, golangci-lint, `go test -race`), dan job `changelog` yang gagal kalau `apps/`, `backend/`, atau `packages/` berubah tanpa `CHANGELOG.md` (kecuali label `no-changelog`).
- `.gitattributes` (semua file teks LF, supaya gofmt dan Prettier konsisten di Windows).
- `.env.example` (bagian `algohub-api`), `.gitignore`, `.prettierignore`, `README.md`, `docs/README.md`, `docs/api.md`, `docs/deploy.md`.
- Project Vercel baru `algohub-api` (Root Directory `backend`, preset Go, tanpa Git integration dan tanpa domain) dengan env preview `APP_ENV`, `CORS_ORIGINS`, `CORS_ORIGIN_PATTERNS`. Deployment preview lulus uji health dan CORS (`docs/deploy.md`).

#### Diubah
- Versi Go: `go 1.26.0` + `toolchain go1.26.11`, bukan Go 1.24 seperti di rencana dan praktikum lama. Alasannya: Go 1.24 sudah tidak mendapat patch keamanan (Go 1.27 sudah rilis, dan Go hanya mendukung dua versi mayor terakhir). Vercel memakai versi dari `go`/`toolchain` di `go.mod`, jadi versinya dipin. `govulncheck` dengan go1.26.2 + `golang.org/x/net` v0.50.0 menemukan 17 kerentanan yang terjangkau. Dengan go1.26.11 + `golang.org/x/net` v0.62.0, hasilnya "No vulnerabilities found".

#### Keamanan
- `golang.org/x/net` v0.62.0 dan `golang.org/x/crypto` v0.59.0 (indirect lewat Gin v1.10.0).

### Fase 0 — Verifikasi (2026-11-17)

#### Ditambahkan
- `docs/deploy.md`: pemetaan domain → project Vercel, project Supabase per app, scheduler eksternal, dan daftar pemakaian Vercel yang masih harus diisi pemilik.

#### Temuan
- Prod memakai **2 project Supabase**: siakad `huddecanbapwjairmaef`, sedangkan elearning dan praktikum `oidwvfcxlxfyojkyurmm`. Backend memakai `DATABASE_URL_SIAKAD` terpisah dari `DATABASE_URL_PRAKTIKUM`/`DATABASE_URL_ELEARNING`.
- Semua domain prod ada di team `lab-algoritma-pemrograman`: `algohub.web.id`/`www` → `labap-frontend`, `api` → `labap-backend`, `siakad` → `labap-siakad`, `elearning` → `labap-elearning`.
- Auto-submit praktikum dipicu cron-job.org setiap 1 menit (menurut dokumentasi lama, belum diverifikasi).
