---
title: Arsitektur Monorepo
description: Struktur monorepo SvelteKit SPA dan Go Gin, toolchain Bun dan Go, serta paket TypeScript bersama (@algohub/*).
---

# Arsitektur Monorepo

AlgoHub menyatukan seluruh subsistem perangkat lunak laboratorium ke dalam satu repositori terpadu (**Monorepo**). Arsitektur ini menggabungkan 4 aplikasi web frontend Single Page Application (SPA) berbasis SvelteKit, 1 layanan backend terpusat berbasis Go Gin, dan 6 paket pustaka bersama (*shared packages*).

---

## 1. Struktur Direktori Proyek

```
algohub/
├── apps/                                 ← Frontend SPA (1 folder = 1 subdomain Vercel)
│   ├── portal/                           ← algohub.web.id (Prerendered Landing & CDN Compiler)
│   ├── siakad/                           ← siakad.algohub.web.id (SIAKAD Laboratorium)
│   ├── elearning/                        ← elearning.algohub.web.id (E-Learning Gamifikasi)
│   └── praktikum/                        ← praktikum.algohub.web.id (Sistem Praktikum Lab)
│
├── backend/                              ← Layanan API Go Tunggal (api.algohub.web.id)
│   ├── cmd/server/main.go                ← Entrypoint tunggal Vercel Go Runtime & lokal
│   ├── internal/
│   │   ├── app/                          ← Komposisi modul & perakitan dependensi
│   │   ├── config/                       ← Parser variabel lingkungan
│   │   ├── database/                     ← Manajemen connection pool GORM
│   │   ├── middleware/                   ← CORS, Auth, Role Guard, Rate Limiter
│   │   ├── delivery/http/                ← Router utama & HTTP handler
│   │   │   └── handler/{siakad,elearning,praktikum,cron}/
│   │   ├── dto/{siakad,elearning,praktikum}/
│   │   ├── entity/{siakad,elearning,praktikum}/
│   │   ├── repository/{siakad,elearning,praktikum}/
│   │   └── usecase/{siakad,elearning,praktikum}/
│   └── pkg/                              ← Library utilitas independen (jwt, gemini, webpush, response)
│
├── packages/                             ← Kode TypeScript bersama (Bun Workspaces)
│   ├── config/                           ← Tema Tailwind, konfigurasi TS, ESLint, Prettier
│   ├── ui/                               ← Komponen antarmuka Svelte (Monaco, xterm, DataTable)
│   ├── api-client/                       ← HTTP Client per domain & Query Key Factory
│   ├── auth/                             ← Sesi peramban, role normalizer & guard layout
│   ├── query/                            ← QueryClient bersama & persister IndexedDB
│   └── runner/                           ← Web Worker runner Clang WASM & Pyodide
│
├── docs/                                 ← Dokumentasi teknis, ERD, dan catatan arsitektur
├── tests/e2e/                            ← Pengujian otomatis Playwright terintegrasi
├── package.json                          ← Definisi workspace Bun
└── bun.lock                              ← Lockfile dependensi Bun
```

---

## 2. Toolchain & Pengelola Paket

AlgoHub memberlakukan aturan toolchain yang ketat untuk menjamin konsistensi reproduksi build:

### Bun 1.3.14 (Strict Package Manager)
- Proyek dikonfigurasi dengan `"packageManager": "bun@1.3.14"`.
- **Wajib menggunakan Bun**: Penggunaan `npm`, `npx`, `yarn`, atau `pnpm` dilarang keras untuk mencegah perbedaan pohon dependensi (*dependency drift*).
- Menjalankan instalasi dependensi secara cepat dan mengelola symlink Bun Workspaces (`apps/*` dan `packages/*`).

### Go Toolchain & Linter
- **Go 1.26**: Backend menggunakan modul `github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend` dengan `go 1.26.0` dan `toolchain go1.26.11`.
- **GOTOOLCHAIN=auto**: Go toolchain diunduh secara otomatis sesuai versi yang tertera di `backend/go.mod`.
- **golangci-lint v2**: Analisis statis kode Go dengan linter aktif: `unused`, `gocritic`, `govet`, `errcheck`, dan `dupl`.

---

## 3. Arsitektur Frontend SvelteKit SPA

Keempat aplikasi di `apps/*` dibangun menggunakan SvelteKit dengan konfigurasi **Pure Single Page Application (SPA)**:
- **Adapter Statis**: Menggunakan `@sveltejs/adapter-static` dengan konfigurasi:
  ```javascript
  // svelte.config.js
  export default {
    kit: {
      adapter: adapter({ fallback: 'index.html' }),
    },
  };
  ```
- **Ketiadaan Server Node.js**: Frontend tidak memiliki file `+page.server.ts`, `+server.ts`, atau `hooks.server.ts`. Tidak ada kode yang dieksekusi di Node.js server saat runtime.
- **Guard Rute Klien**: Proteksi hak akses login dan otorisasi peran ditangani pada `+layout.ts` menggunakan fungsi `requireRole()` dari `@algohub/auth`:
  ```typescript
  // src/routes/+layout.ts
  export const ssr = false;
  export const prerender = false; // (portal: prerender = true)

  export const load = async () => {
    requireRole(['mahasiswa', 'asisten', 'koordinator']);
  };
  ```

---

## 4. Paket TypeScript Bersama (`packages/*`)

Untuk mencegah duplikasi logika di antara keempat aplikasi frontend, kode bersama dikonsolidasi ke dalam paket `@algohub/*`:

| Paket | Nama Package | Fungsi & Kandungan Utama |
|---|---|---|
| `packages/config` | `@algohub/config` | Konfigurasi dasar `tsconfig.json`, `eslint.config.js`, `.prettierrc`, dan tema dasar Tailwind CSS yang dipakai seragam. |
| `packages/ui` | `@algohub/ui` | Komponen Svelte berbasis Bits UI: **DataTable** (TanStack Table + Virtualization), **CodeEditor** (Monaco Editor), **Terminal** (xterm.js), dan **RichTextEditor / RichTextRenderer** (Tiptap). |
| `packages/api-client` | `@algohub/api-client` | Klien HTTP berbasis Fetch dengan `credentials: 'include'`. Menyediakan submodul terpisah (`praktikum.ts`, `siakad.ts`, `elearning.ts`) dan **Query Key Factories** untuk cache TanStack Query. |
| `packages/auth` | `@algohub/auth` | Manajemen sesi di browser, normalisasi peran pengguna (`NormalizeRole`), guard autentikasi `requireRole()`, dan prosedur logout terkoordinasi (membersihkan memori, cache Query, dan IndexedDB). |
| `packages/query` | `@algohub/query` | Konfigurasi sentral `createAppQueryClient()`, integrasi persister IndexedDB per user (`algohub-<app>-<id>`), dan whitelist query yang diizinkan untuk disimpan offline. |
| `packages/runner` | `@algohub/runner` | Mesin eksekusi kode di peramban. Mengelola Web Worker Clang WebAssembly dan Pyodide Python, sinkronisasi I/O via `SharedArrayBuffer`, serta mekanisme *force kill timeout*. |

---

## 5. Alur Data & Perintah Pengembangan

### Perintah Utama (Root)
```bash
bun install             # Pasang seluruh dependensi workspace
bun run dev:portal      # Jalankan Portal lokal di port 5173
bun run dev:siakad      # Jalankan SIAKAD lokal di port 5174
bun run dev:elearning   # Jalankan E-Learning lokal di port 5175
bun run dev:praktikum   # Jalankan Praktikum lokal di port 5176
bun run dev:api         # Jalankan Go API lokal di http://127.0.0.1:8080
bun run lint            # Pemeriksaan linter seluruh workspace
bun run check           # Pemeriksaan tipe TypeScript SvelteKit
bun run test            # Menjalankan unit test Vitest
bun run build           # Membangun build produksi seluruh aplikasi
```

### Backend (dari direktori `backend/`)
```bash
go vet ./...            # Pemeriksaan kode bawaan Go
golangci-lint run ./... # Linter komprehensif
go test ./...           # Menjalankan suite pengujian Go
```
