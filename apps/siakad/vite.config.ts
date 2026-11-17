import adapter from '@sveltejs/adapter-static';
import { sveltekit } from '@sveltejs/kit/vite';
import tailwindcss from '@tailwindcss/vite';
import { svelteTesting } from '@testing-library/svelte/vite';
import { defineConfig } from 'vitest/config';

export default defineConfig({
  plugins: [
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
      // App lama hanya mendaftarkan service worker saat alur notifikasi push berjalan.
      serviceWorker: { register: false },
    }),
    svelteTesting(),
  ],
  server: {
    // CORS backend dev hanya mengizinkan port 5173–5176; 5173 dipakai praktikum.
    port: 5174,
  },
  test: {
    include: ['src/**/*.test.ts', 'tests/**/*.test.ts'],
    environment: 'jsdom',
    setupFiles: ['./vitest-setup.ts'],
  },
});
