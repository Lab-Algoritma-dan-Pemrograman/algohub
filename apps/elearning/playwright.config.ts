import { defineConfig, devices } from '@playwright/test';
import { SUPABASE_HOST } from './e2e/mock-realtime.ts';

const PORT = 5175;
const ORIGIN = `http://127.0.0.1:${PORT}`;

export default defineConfig({
  testDir: 'e2e',
  // Spec compiler mengunduh dan meng-compile clang (±58 MB) serta Pyodide.
  timeout: 180_000,
  // Berurutan supaya tidak berebut CPU dan memicu timeout palsu di PC lemah.
  workers: 1,
  reporter: 'list',
  use: { baseURL: ORIGIN, serviceWorkers: 'block' },
  projects: [{ name: 'chromium', use: { ...devices['Desktop Chrome'] } }],
  webServer: {
    // Build produksi + preview: dev server memuat ulang halaman saat Vite mengoptimasi
    // dependency baru, sehingga test pertama gagal acak. Tanpa --host, Vite di Windows hanya
    // mendengarkan ::1 sehingga 127.0.0.1 tidak terjangkau.
    command: `bun run build && bun run preview --host 127.0.0.1 --port ${PORT} --strictPort`,
    url: ORIGIN,
    reuseExistingServer: false,
    timeout: 300_000,
    // API diarahkan ke origin preview sendiri sehingga setiap request bisa dijawab page.route dan
    // build e2e tidak mungkin memanggil backend prod. ALGOHUB_COMPILER_DIR diwarisi dari shell.
    env: {
      PUBLIC_API_BASE_URL: ORIGIN,
      PUBLIC_COMPILER_BASE_URL: '/compiler/v1/',
      PUBLIC_APP_VERSION: 'e2e',
      PUBLIC_SUPABASE_URL: `http://${SUPABASE_HOST}`,
      PUBLIC_SUPABASE_PUBLISHABLE_KEY: 'e2e-anon',
    },
  },
});
