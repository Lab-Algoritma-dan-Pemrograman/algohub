import { compilerDir, runnerWorkers } from '@algohub/runner/vite';
import adapter from '@sveltejs/adapter-static';
import { sveltekit } from '@sveltejs/kit/vite';
import tailwindcss from '@tailwindcss/vite';
import { svelteTesting } from '@testing-library/svelte/vite';
import { defineConfig } from 'vitest/config';

// SharedArrayBuffer untuk input C/Python interaktif hanya tersedia saat crossOriginIsolated.
// credentialless (bukan require-corp) agar gambar Supabase dan CDN tetap dimuat tanpa CORP.
const isolationHeaders = {
  'Cross-Origin-Opener-Policy': 'same-origin',
  'Cross-Origin-Embedder-Policy': 'credentialless',
};

export default defineConfig({
  // Middleware runner wajib sebelum SvelteKit: middleware dipasang sesuai urutan plugin dan
  // SvelteKit menjawab request lebih dulu, sehingga worker tidak pernah sampai ke plugin di belakangnya.
  plugins: [
    runnerWorkers(),
    compilerDir(),
    tailwindcss(),
    sveltekit({
      adapter: adapter({
        pages: 'build',
        assets: 'build',
        fallback: 'index.html',
        precompress: false,
      }),
      // App lama tidak memeriksa versi baru secara berkala; SvelteKit 3 default-nya tiap jam.
      version: { pollInterval: 0 },
    }),
    svelteTesting(),
  ],
  server: {
    // CORS backend dev hanya mengizinkan port 5173–5176.
    port: 5173,
    headers: isolationHeaders,
  },
  preview: {
    headers: isolationHeaders,
  },
  test: {
    include: ['src/**/*.test.ts', 'tests/**/*.test.ts'],
    environment: 'jsdom',
    setupFiles: ['@testing-library/jest-dom/vitest'],
    server: {
      deps: {
        inline: [/katex/],
      },
    },
  },
});
