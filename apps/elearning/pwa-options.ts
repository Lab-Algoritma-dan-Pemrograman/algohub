import type { withPwa } from '@vite-pwa/sveltekit';
// Origin API yang tidak boleh masuk cache service worker apa pun yang terjadi.
const API_ORIGINS = ['https://api.algohub.web.id'];
const escapeRegExp = (text: string) => text.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
export function apiOrigin(apiBaseUrl: string): string | null {
  try {
    return new URL(apiBaseUrl).origin;
  } catch {
    return null;
  }
}
type PwaOptions = NonNullable<Parameters<typeof withPwa>[1]>;
export function createPwaOptions(apiBaseUrl: string): PwaOptions {
  const origins = [...new Set([...API_ORIGINS, apiOrigin(apiBaseUrl)].filter((o) => o !== null))];
  return {
    registerType: 'autoUpdate',
    // Pendaftaran dilakukan dari +layout.svelte lewat virtual:pwa-register (app ini SPA tanpa SSR).
    injectRegister: false,
    kit: { adapterFallback: 'index.html', spa: true },
    // App lama tidak punya manifest (PWA-nya dimatikan); nilai di bawah baru dibuat di port ini.
    manifest: {
      name: 'AlgoHub E-Learning',
      short_name: 'E-Learning',
      description: 'Belajar Python dan Bahasa C bersama Lab Algoritma & Pemrograman.',
      lang: 'id',
      start_url: '/',
      scope: '/',
      display: 'standalone',
      theme_color: '#8A1538',
      background_color: '#FDFBF7',
      icons: [
        { src: '/icons/pwa-192.png', sizes: '192x192', type: 'image/png' },
        { src: '/icons/pwa-512.png', sizes: '512x512', type: 'image/png', purpose: 'any maskable' },
      ],
    },
    maximumFileSizeToCacheInBytes: 5 * 1024 * 1024,
    generateSW: {
      // Plugin menambah pola prerendered/** sendiri, tetapi SPA ini tidak punya halaman prerender.
      globStrict: false,
      // Worker runner dan file compiler punya aturan cache sendiri; bukan halaman SPA.
      navigateFallbackDenylist: [/^\/runner\//, /^\/compiler\//],
      runtimeCaching: [
        // Urutan penting: aturan pertama yang cocok menang, jadi API selalu NetworkOnly.
        ...origins.map((origin) => ({
          urlPattern: new RegExp(`^${escapeRegExp(origin)}/`),
          handler: 'NetworkOnly' as const,
        })),
        {
          urlPattern: /^https:\/\/cdn\.jsdelivr\.net\/pyodide\/.*$/i,
          handler: 'CacheFirst' as const,
          options: {
            cacheName: 'pyodide-cache',
            expiration: { maxEntries: 100, maxAgeSeconds: 60 * 60 * 24 * 30 },
            cacheableResponse: { statuses: [0, 200] },
          },
        },
      ],
    },
  };
}
