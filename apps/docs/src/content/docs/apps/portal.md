---
title: Aplikasi Portal
description: Arsitektur landing page, distribusi biner compiler C WebAssembly, dan pengalihan rute warisan pada Portal AlgoHub.
---

# Aplikasi Portal (`apps/portal`)

**Portal AlgoHub** (`algohub.web.id`) adalah landing page statis dan gerbang publik utama bagi seluruh ekosistem Laboratorium Algoritma & Pemrograman. Selain menyajikan profil laboratorium, Portal memegang peran infrastruktur penting: mendistribusikan berkas biner compiler C berbasis WebAssembly dan menjaga kompatibilitas pranala lama melalui redirect permanen.

---

## 1. Arsitektur & Desain Visual

Portal dibangun menggunakan **SvelteKit** dengan adapter statis (`@sveltejs/adapter-static`). Seluruh halaman di-prerender penuh pada saat proses build (`prerender = true`), sehingga aplikasi berjalan murni sebagai file statis HTML/CSS/JS tanpa memakan kuota Serverless Function di Vercel (0 invocations).

### Karakteristik Visual
Desain antarmuka Portal mengadopsi spesifikasi desain presisi dari ekspor Framer:
- **Skema Warna**:
  - `Ground / Page`: `#f4f4f4`
  - `Ink / Primary`: `#0a0a0a`
  - `Ink / Muted`: `#605f5f`
  - `Ink / Faint`: `#8a8a8a`
- **Hero Section**: Layout 2 kolom asimetris (teks lead dinamis di sisi kiri dipadukan dengan panel visual interaktif di sisi kanan).
- **TypeRotator**: Animasi teks ketik berputar yang dibangun menggunakan **Svelte 5 Runes** (`$state`, `$effect`).
- **Kartu Platform**: Kartu visual ber-cover rasio 1.333 dengan radius sudut 20px yang mengarahkan ke SIAKAD, E-Learning, dan Praktikum.
- **Footer Panel**: Panel gelap (`#0a0a0a`, radius 32px) dengan tipografi ekspansif dan animasi wave wordmark *AlgoHub*.

---

## 2. Distribusi Compiler C WebAssembly (`/compiler/v1/`)

Agar aplikasi Praktikum dan E-Learning dapat mengompilasi kode bahasa C secara instan di sisi peramban tanpa membebani server, Portal bertindak sebagai Content Delivery Network (CDN) internal untuk file biner compiler Clang WebAssembly.

### Berkas Biner Compiler
Direktori `/compiler/v1/` menyajikan empat berkas inti compiler:

| Berkas | Ukuran Est. | Deskripsi |
|---|---|---|
| `clang` | ~24 MB | Executable LLVM/Clang yang dikompilasi ke WebAssembly (`.wasm`) |
| `lld` | ~14 MB | LLVM Linker versi WebAssembly untuk menghubungkan objek C |
| `memfs` | ~1 MB | In-memory virtual file system WebAssembly |
| `sysroot.tar` | ~19 MB | Standard Library C (`libc`, `stdio.h`, `math.h`, dll) terkompresi |

### Mekanisme Pengunduhan & Integritas Build
Berkas compiler berukuran besar (~58 MB) tidak disimpan di repositori Git. Sebagai gantinya:
1. Skrip `scripts/fetch-compiler.ts` dijalankan secara otomatis saat build:
   ```bash
   bun run fetch-compiler
   ```
2. Skrip mengunduh rilis `compiler-v1` resmi dari GitHub Release AlgoHub.
3. Integritas setiap biner divalidasi menggunakan perbandingan checksum terhadap berkas `SHA256SUMS`.
4. Jika ada berkas yang korup atau tidak cocok hash-nya, proses build langsung dibatalkan (fail-safe).

### Header Caching & Keamanan CORS
Dikonfigurasi pada `apps/portal/vercel.json`:

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

- **`max-age=31536000, immutable`**: Peramban dan CDN Vercel menyimpan berkas di cache lokal selama 1 tahun. Pengguna hanya perlu mengunduh compiler sekali saja.
- **`Access-Control-Allow-Origin: *`**: Memungkinkan `praktikum.algohub.web.id` dan `elearning.algohub.web.id` mengambil biner compiler tanpa kendala pembatasan CORS.
- **`Cross-Origin-Resource-Policy: cross-origin`**: Diperlukan agar biner dapat dibaca di lingkungan Web Worker yang terisolasi (`crossOriginIsolated`).

---

## 3. Mekanisme Pengalihan Rute Warisan (301 Redirects)

Sebelum pemisahan arsitektur monorepo, domain `algohub.web.id` melayani sistem praktikum lama. Agar bookmark mahasiswa dan asisten tidak rusak, Portal secara otomatis mengalihkan seluruh rute lama ke subdomain baru menggunakan aturan redirect permanen HTTP 301 di `vercel.json`:

| Rute Lama (`algohub.web.id`) | Target Pengalihan Baru | Status Kode |
|---|---|---|
| `/info` & `/info/:path*` | `https://praktikum.algohub.web.id/info/:path*` | `301 Permanent` |
| `/praktikum` & `/praktikum/:path*` | `https://praktikum.algohub.web.id/praktikum/:path*` | `301 Permanent` |
| `/playground` | `https://praktikum.algohub.web.id/playground` | `301 Permanent` |
| `/playground-c` | `https://praktikum.algohub.web.id/playground-c` | `301 Permanent` |
| `/layanan/:path*` | `https://praktikum.algohub.web.id/layanan/:path*` | `301 Permanent` |
| `www.algohub.web.id/*` | `https://algohub.web.id/*` | `301 Permanent` |

Seluruh parameter query (misal `?sesi=123`) dipertahankan selama proses pengalihan.
