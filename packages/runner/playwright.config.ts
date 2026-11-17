import { defineConfig, devices } from '@playwright/test';
if (!process.env.ALGOHUB_COMPILER_DIR) {
  throw new Error(
    'ALGOHUB_COMPILER_DIR wajib diisi: folder berisi clang, lld, memfs, dan sysroot.tar (mis. praktikum/frontend/static).',
  );
}
const PORT = 5179;
export default defineConfig({
  testDir: 'e2e',
  timeout: 180_000,
  // Setiap test mengunduh dan meng-compile clang (±58 MB); dijalankan berurutan
  // supaya tidak berebut CPU dan memicu timeout palsu.
  workers: 1,
  reporter: 'list',
  use: { baseURL: `http://localhost:${PORT}` },
  projects: [{ name: 'chromium', use: { ...devices['Desktop Chrome'] } }],
  webServer: {
    command: `bun run dev --port ${PORT} --strictPort`,
    url: `http://localhost:${PORT}`,
    reuseExistingServer: false,
  },
});
