---
title: Monorepo Architecture
description: Architecture of the SvelteKit SPA and Go Gin monorepo, Bun and Go toolchains, and shared TypeScript packages (@algohub/*).
---

# Monorepo Architecture

AlgoHub consolidates all laboratory computing subsystems into a unified **Monorepo**. The architecture couples 4 SvelteKit Single Page Applications (SPAs), 1 centralized Go Gin REST API service, and 6 shared TypeScript workspace packages.

---

## 1. Project Directory Layout

```
algohub/
├── apps/                                 ← Frontend SPAs (1 folder = 1 Vercel subdomain)
│   ├── portal/                           ← algohub.web.id (Prerendered Landing & Compiler CDN)
│   ├── siakad/                           ← siakad.algohub.web.id (Academic Operations)
│   ├── elearning/                        ← elearning.algohub.web.id (Gamified E-Learning)
│   └── praktikum/                        ← praktikum.algohub.web.id (Core Practicum Engine)
│
├── backend/                              ← Unified Go API Service (api.algohub.web.id)
│   ├── cmd/server/main.go                ← Shared entrypoint for Vercel Go runtime & local dev
│   ├── internal/
│   │   ├── app/                          ← Dependency injection & bootstrap assembler
│   │   ├── config/                       ← Environment variable schema parser
│   │   ├── database/                     ← GORM connection pool manager
│   │   ├── middleware/                   ← CORS, Auth, Role Guards, Rate Limiters
│   │   ├── delivery/http/                ← HTTP Routing & Domain Handlers
│   │   │   └── handler/{siakad,elearning,praktikum,cron}/
│   │   ├── dto/{siakad,elearning,praktikum}/
│   │   ├── entity/{siakad,elearning,praktikum}/
│   │   ├── repository/{siakad,elearning,praktikum}/
│   │   └── usecase/{siakad,elearning,praktikum}/
│   └── pkg/                              ← Reusable utility libraries (jwt, gemini, webpush, response)
│
├── packages/                             ← Shared TypeScript packages (Bun Workspaces)
│   ├── config/                           ← Tailwind themes, shared TSConfig, ESLint, Prettier
│   ├── ui/                               ← Shared Svelte UI components (Monaco, xterm, DataTable)
│   ├── api-client/                       ← Domain HTTP API Clients & Query Key Factories
│   ├── auth/                             ← Browser sessions, role normalizers & layout guards
│   ├── query/                            ← Central QueryClient & IndexedDB persisters
│   └── runner/                           ← Clang WASM & Pyodide Web Worker runner
│
├── docs/                                 ← Architectural specifications, ERD, and runbooks
├── tests/e2e/                            ← Comprehensive Playwright end-to-end tests
├── package.json                          ← Bun workspaces manifest
└── bun.lock                              ← Frozen lockfile
```

---

## 2. Toolchain & Package Management

AlgoHub strictly enforces versioned toolchain requirements to eliminate environment inconsistencies:

### Bun 1.3.14 (Strict Package Manager)
- Locked via `"packageManager": "bun@1.3.14"`.
- **Exclusive Use of Bun**: Tools like `npm`, `npx`, `yarn`, or `pnpm` are blocked to prevent dependency resolution divergences.
- Accelerates installation pipelines and resolves internal package workspace symlinks (`apps/*` and `packages/*`).

### Go Toolchain & Linter
- **Go 1.26**: Backend runs module `github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend` specifying `go 1.26.0` and `toolchain go1.26.11`.
- **GOTOOLCHAIN=auto**: The exact Go toolchain version is fetched on-demand during CI and container execution.
- **golangci-lint v2**: Enforces static analysis across linter suites: `unused`, `gocritic`, `govet`, `errcheck`, and `dupl`.

---

## 3. SvelteKit Client-Side SPA Architecture

All four frontend applications in `apps/*` are configured as **Pure Single Page Applications (SPA)**:
- **Static Adapter**: Governed by `@sveltejs/adapter-static` with fallback index routing:
  ```javascript
  // svelte.config.js
  export default {
    kit: {
      adapter: adapter({ fallback: 'index.html' }),
    },
  };
  ```
- **Zero Node.js Server Runtime**: Apps omit `+page.server.ts`, `+server.ts`, and `hooks.server.ts`. No Node.js process executes at runtime.
- **Client Route Guards**: Session validation and RBAC checks run inside `+layout.ts` leveraging `requireRole()` from `@algohub/auth`:
  ```typescript
  // src/routes/+layout.ts
  export const ssr = false;
  export const prerender = false; // (portal: prerender = true)

  export const load = async () => {
    requireRole(['mahasiswa', 'asisten', 'koordinator']);
  };
  ```

---

## 4. Shared TypeScript Packages (`packages/*`)

Logic shared across multiple frontends is isolated into discrete `@algohub/*` workspace packages:

| Package | Package Name | Scope & Core Contents |
|---|---|---|
| `packages/config` | `@algohub/config` | Common `tsconfig.json`, `eslint.config.js`, `.prettierrc`, and universal Tailwind CSS color themes. |
| `packages/ui` | `@algohub/ui` | Component library based on Bits UI: **DataTable** (TanStack Table + Virtualization), **CodeEditor** (Monaco), **Terminal** (xterm.js), and **RichTextEditor / RichTextRenderer** (Tiptap). |
| `packages/api-client` | `@algohub/api-client` | Fetch-based typed API client sending `credentials: 'include'`. Houses isolated domain modules (`praktikum.ts`, `siakad.ts`, `elearning.ts`) and **Query Key Factories**. |
| `packages/auth` | `@algohub/auth` | Browser token storage, role normalization (`NormalizeRole`), layout guards (`requireRole`), and coordinated logout sweeps (purging memory, TanStack caches, and IndexedDB). |
| `packages/query` | `@algohub/query` | Centralized `createAppQueryClient()` factory, IndexedDB persistence adapters, and query caching whitelist rules. |
| `packages/runner` | `@algohub/runner` | Client-side execution engine managing Web Workers for Clang WebAssembly and Pyodide, handling I/O over `SharedArrayBuffer` with runtime kill switches. |

---

## 5. Development Workflows & Commands

### Monorepo Root Commands
```bash
bun install             # Install dependencies across all workspaces
bun run dev:portal      # Start Portal locally on port 5173
bun run dev:siakad      # Start SIAKAD locally on port 5174
bun run dev:elearning   # Start E-Learning locally on port 5175
bun run dev:praktikum   # Start Praktikum locally on port 5176
bun run dev:api         # Start Go API locally at http://127.0.0.1:8080
bun run lint            # Run linter across all workspaces
bun run check           # Perform SvelteKit TypeScript typecheck
bun run test            # Execute Vitest unit tests
bun run build           # Build production bundles for all apps
```

### Backend Commands (from `backend/`)
```bash
go vet ./...            # Run standard Go static analysis
golangci-lint run ./... # Run comprehensive linter suite
go test ./...           # Execute Go backend tests
```
