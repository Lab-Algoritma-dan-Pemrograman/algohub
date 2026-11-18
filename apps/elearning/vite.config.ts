import { compilerDir, runnerWorkers } from '@algohub/runner/vite';
import adapter from '@sveltejs/adapter-static';
import tailwindcss from '@tailwindcss/vite';
import { svelteTesting } from '@testing-library/svelte/vite';
import { withPwa } from '@vite-pwa/sveltekit';
import type { Connect, Plugin } from 'vite';
import { defineConfig } from 'vitest/config';
import { createPwaOptions } from './pwa-options.ts';
// SharedArrayBuffer untuk input C/Python interaktif hanya tersedia saat crossOriginIsolated.
// credentialless (bukan require-corp) agar gambar Supabase dan CDN tetap dimuat tanpa CORP.
// Dipasang sebagai middleware karena halaman dijawab SvelteKit, bukan oleh server statis Vite
// yang membaca server.headers/preview.headers; di production header datang dari vercel.json.
const setIsolationHeaders: Connect.NextHandleFunction = (_req, res, next) => {
  res.setHeader('Cross-Origin-Opener-Policy', 'same-origin');
  res.setHeader('Cross-Origin-Embedder-Policy', 'credentialless');
  next();
};
function crossOriginIsolation(): Plugin {
  return {
    name: 'elearning-cross-origin-isolation',
    configureServer(server) {
      server.middlewares.use(setIsolationHeaders);
    },
    configurePreviewServer(server) {
      server.middlewares.use(setIsolationHeaders);
    },
  };
}
export default defineConfig({
  // Middleware wajib sebelum SvelteKit: middleware dipasang sesuai urutan plugin dan SvelteKit
  // menjawab request lebih dulu, sehingga plugin di belakangnya tidak pernah dilewati.
  plugins: [
    crossOriginIsolation(),
    runnerWorkers(),
    compilerDir(),
    tailwindcss(),
    // withPwa membungkus sveltekit() agar service worker dibuat sebelum adapter-static berjalan.
    withPwa(
      {
        adapter: adapter({
          pages: 'build',
          assets: 'build',
          fallback: 'index.html',
          precompress: false,
        }),
        // App lama tidak memeriksa versi baru secara berkala; SvelteKit 3 default-nya tiap jam.
        version: { pollInterval: 0 },
      },
      createPwaOptions(process.env.PUBLIC_API_BASE_URL ?? ''),
    ),
    svelteTesting(),
  ],
  server: {
    // CORS backend dev hanya mengizinkan port 5173–5176; 5173 praktikum, 5174 siakad.
    port: 5175,
  },
  test: {
    include: ['src/**/*.test.ts', 'tests/**/*.test.ts'],
    environment: 'jsdom',
    setupFiles: ['./vitest-setup.ts'],
    server: {
      deps: {
        inline: [/katex/],
      },
    },
  },
});
