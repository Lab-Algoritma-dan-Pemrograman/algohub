# Changelog

Semua perubahan penting dicatat di file ini. Format mengikuti [Keep a Changelog](https://keepachangelog.com/id-ID/1.1.0/), dan versi mengikuti [Semantic Versioning](https://semver.org/lang/id/).

## [Belum dirilis]

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
