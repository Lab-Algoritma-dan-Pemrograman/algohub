import { fileURLToPath } from 'node:url';
import { defineConfig, devices } from '@playwright/test';

const compilerDir = process.env.ALGOHUB_COMPILER_DIR;
if (!compilerDir) {
  throw new Error(
    'ALGOHUB_COMPILER_DIR wajib diisi: folder berisi clang, lld, memfs, dan sysroot.tar (bun run fetch-compiler .cache/compiler/v1).',
  );
}

const PORT = 4173;
const ORIGIN = `http://127.0.0.1:${PORT}`;

export default defineConfig({
  timeout: 180_000,
  // Test compiler mengunduh dan meng-compile clang (±58 MB); dijalankan berurutan
  // supaya tidak berebut CPU dan memicu timeout palsu.
  workers: 1,
  reporter: 'list',
  use: { baseURL: ORIGIN },
  projects: [
    {
      name: 'praktikum',
      testDir: './praktikum',
      use: { ...devices['Desktop Chrome'] },
    },
  ],
  webServer: {
    // Tanpa --host, Vite di Windows hanya mendengarkan ::1 sehingga 127.0.0.1 tidak terjangkau.
    command: `bun run build && bun run preview --host 127.0.0.1 --port ${PORT} --strictPort`,
    cwd: fileURLToPath(new URL('../../apps/praktikum/', import.meta.url)),
    // API diarahkan ke origin preview sendiri sehingga setiap request bisa dijawab page.route
    // dan build e2e tidak mungkin memanggil backend prod.
    env: {
      PUBLIC_API_BASE_URL: ORIGIN,
      PUBLIC_COMPILER_BASE_URL: '/compiler/v1/',
      PUBLIC_APP_VERSION: 'e2e',
      ALGOHUB_COMPILER_DIR: compilerDir,
    },
    url: ORIGIN,
    timeout: 300_000,
    reuseExistingServer: false,
  },
});
