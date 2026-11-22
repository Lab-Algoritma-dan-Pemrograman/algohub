---
title: Portal Application
description: Landing page architecture, C WebAssembly compiler binary distribution, and legacy redirect mechanics on AlgoHub Portal.
---

# Portal Application (`apps/portal`)

**AlgoHub Portal** (`algohub.web.id`) serves as the official public entry point to the Algorithm & Programming Laboratory ecosystem. Beyond acting as an institutional landing page, Portal fulfills critical infrastructure duties: delivering pre-compiled C WebAssembly toolchain binaries and maintaining legacy URL backward compatibility through edge 301 redirects.

---

## 1. Architecture & Visual Design

Portal is built with **SvelteKit** targeting `@sveltejs/adapter-static`. All routes are fully pre-rendered during the build process (`prerender = true`), operating as pure static assets with zero Serverless Function invocations on Vercel.

### Design System
Portal UI is built around a Framer-exported design system:
- **Color Palette**:
  - `Ground / Page`: `#f4f4f4`
  - `Ink / Primary`: `#0a0a0a`
  - `Ink / Muted`: `#605f5f`
  - `Ink / Faint`: `#8a8a8a`
- **Hero Section**: Two-column asymmetric layout pairing dynamic lead copy with an interactive visual panel.
- **TypeRotator**: Character-cycling headline component implemented using **Svelte 5 Runes** (`$state`, `$effect`).
- **Platform Cards**: High-contrast cards featuring an aspect ratio of 1.333 and 20px border radius linking directly to SIAKAD, E-Learning, and Praktikum.
- **Dark Footer**: Curved dark container (`#0a0a0a`, radius 32px) featuring animated wave typography for the *AlgoHub* brand wordmark.

---

## 2. C WebAssembly Compiler Distribution (`/compiler/v1/`)

To support instant, client-side C compilation in Praktikum and E-Learning without server load, Portal serves as an internal CDN for Clang WebAssembly binaries.

### Compiler Binaries
The `/compiler/v1/` route serves four foundational toolchain assets:

| Asset | Size Est. | Purpose |
|---|---|---|
| `clang` | ~24 MB | Clang C compiler frontend compiled to WebAssembly (`.wasm`) |
| `lld` | ~14 MB | LLVM Linker targeting WebAssembly object files |
| `memfs` | ~1 MB | Virtual in-memory POSIX filesystem layer |
| `sysroot.tar` | ~19 MB | Compressed standard C headers and static library (`libc`, `stdio.h`, `math.h`) |

### Automated Fetch & Verification Pipeline
Due to their large footprint (~58 MB), compiler binaries are excluded from Git version control. Instead:
1. `scripts/fetch-compiler.ts` executes automatically during build setup:
   ```bash
   bun run fetch-compiler
   ```
2. Assets are downloaded from GitHub Release `compiler-v1`.
3. Checksums are verified against `SHA256SUMS`.
4. Any mismatch aborts the build process to guarantee binary integrity.

### Cache Headers & CORS Configuration
Configured in `apps/portal/vercel.json`:

```json
{
  "source": "/compiler/v1/(.*)",
  "headers": [
    {
      "key": "Cache-Control",
      "value": "public, max-age=31536000, immutable"
    },
    {
      "key": "Access-Control-Allow-Origin",
      "value": "*"
    },
    {
      "key": "Cross-Origin-Resource-Policy",
      "value": "cross-origin"
    }
  ]
}
```

- **`max-age=31536000, immutable`**: Assets are cached for 1 year by browsers and Vercel Edge CDN. Clients download the compiler exactly once.
- **`Access-Control-Allow-Origin: *`**: Allows `praktikum.algohub.web.id` and `elearning.algohub.web.id` to consume binaries cross-origin.
- **`Cross-Origin-Resource-Policy: cross-origin`**: Ensures compatibility inside `crossOriginIsolated` Web Worker contexts.

---

## 3. Legacy Route Migration (301 Permanent Redirects)

Prior to the monorepo architecture, `algohub.web.id` hosted the legacy practicum app. To protect existing bookmarks and syllabus links, Portal executes edge-level 301 redirects to the new `praktikum` subdomain:

| Legacy Route (`algohub.web.id`) | Destination | Status |
|---|---|---|
| `/info` & `/info/:path*` | `https://praktikum.algohub.web.id/info/:path*` | `301 Permanent` |
| `/praktikum` & `/praktikum/:path*` | `https://praktikum.algohub.web.id/praktikum/:path*` | `301 Permanent` |
| `/playground` | `https://praktikum.algohub.web.id/playground` | `301 Permanent` |
| `/playground-c` | `https://praktikum.algohub.web.id/playground-c` | `301 Permanent` |
| `/layanan/:path*` | `https://praktikum.algohub.web.id/layanan/:path*` | `301 Permanent` |
| `www.algohub.web.id/*` | `https://algohub.web.id/*` | `301 Permanent` |

All query strings (e.g. `?sesi=123`) are preserved during redirects.
