import { defineConfig, devices } from '@playwright/test';

const PORT = 5174;

export default defineConfig({
  testDir: 'e2e',
  timeout: 60_000,
  // Satu dev server dan mock per test; berurutan supaya hasil stabil di PC lemah.
  workers: 1,
  reporter: 'list',
  use: { baseURL: `http://localhost:${PORT}`, serviceWorkers: 'block' },
  projects: [{ name: 'chromium', use: { ...devices['Desktop Chrome'] } }],
  webServer: {
    // Build produksi + preview: dev server memuat ulang halaman saat Vite mengoptimasi
    // dependency baru, sehingga test pertama gagal acak.
    command: `bun run build && bun run preview --port ${PORT} --strictPort`,
    url: `http://localhost:${PORT}`,
    reuseExistingServer: false,
    timeout: 300_000,
    // Host .test tidak pernah ada di jaringan; semua request ke sana dijawab mock Playwright.
    env: {
      PUBLIC_API_BASE_URL: 'http://api.siakad.test',
      PUBLIC_SUPABASE_URL: 'http://supabase.siakad.test',
      PUBLIC_SUPABASE_PUBLISHABLE_KEY: 'e2e-anon',
    },
  },
});
