// @vitest-environment node
import { existsSync, readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { describe, expect, it } from 'vitest';
import { createPwaOptions } from '../pwa-options';
const options = createPwaOptions('http://localhost:8080');
const runtimeCaching = options.generateSW?.runtimeCaching ?? [];
const handlerFor = (url: string) =>
  runtimeCaching.find((entry) => {
    const pattern = entry.urlPattern;
    return pattern instanceof RegExp && pattern.test(url);
  })?.handler;
describe('PWA', () => {
  it('vite.config.ts membungkus konfigurasi SvelteKit dengan withPwa dan adapter-static', () => {
    const viteConfig = readFileSync(new URL('../vite.config.ts', import.meta.url), 'utf8');
    expect(viteConfig).toMatch(/withPwa\(/);
    expect(viteConfig).toMatch(/createPwaOptions\(/);
    expect(viteConfig.indexOf('crossOriginIsolation(),')).toBeLessThan(
      viteConfig.indexOf('withPwa('),
    );
  });
  it('manifest memuat ikon 192 dan 512 serta nama aplikasi', () => {
    expect(options.manifest).toMatchObject({ name: 'AlgoHub E-Learning', display: 'standalone' });
    const sizes = (options.manifest ? options.manifest.icons : [])?.map((icon) => icon.sizes);
    expect(sizes).toEqual(['192x192', '512x512']);
  });
  it('respons API tidak pernah di-cache: selalu NetworkOnly', () => {
    expect(handlerFor('https://api.algohub.web.id/api/elearning/profile/sync')).toBe('NetworkOnly');
    expect(handlerFor('http://localhost:8080/api/elearning/leaderboard')).toBe('NetworkOnly');
  });
  it('aturan API berada sebelum aturan cache lain', () => {
    const firstCaching = runtimeCaching.findIndex((e) => e.handler !== 'NetworkOnly');
    const lastNetworkOnly = runtimeCaching.map((e) => e.handler).lastIndexOf('NetworkOnly');
    expect(lastNetworkOnly).toBeLessThan(firstCaching);
  });
  it('hanya Pyodide dari jsDelivr yang di-cache saat runtime', () => {
    const cached = runtimeCaching.filter((entry) => entry.handler !== 'NetworkOnly');
    expect(cached).toHaveLength(1);
    expect(cached[0].handler).toBe('CacheFirst');
    expect(handlerFor('https://cdn.jsdelivr.net/pyodide/v0.27.0/full/pyodide.js')).toBe(
      'CacheFirst',
    );
    expect(handlerFor('https://cdn.jsdelivr.net/npm/other/file.js')).toBeUndefined();
  });
  it('navigasi SPA jatuh ke index.html dan tidak menelan worker runner', () => {
    expect(options.kit).toMatchObject({ adapterFallback: 'index.html', spa: true });
    expect(
      options.generateSW?.navigateFallbackDenylist?.some((re) => re.test('/runner/x.js')),
    ).toBe(true);
  });
});

const buildDir = fileURLToPath(new URL('../build/', import.meta.url));
const built = existsSync(`${buildDir}sw.js`);
// Origin dari env hanya ada di sw.js bila build dibuat dengan PUBLIC_API_BASE_URL yang sama.
const API_URLS = [
  'https://api.algohub.web.id/api/elearning/profile/sync',
  ...(process.env.PUBLIC_API_BASE_URL ? [`/api/elearning/leaderboard`] : []),
];
const toRegExp = (literal: string) => {
  const end = literal.lastIndexOf('/');
  return new RegExp(literal.slice(1, end), literal.slice(end + 1));
};

// Hanya berjalan setelah `bun run build`; tanpa build, konfigurasi di atas sudah diuji.
describe.skipIf(!built)('output build PWA', () => {
  const sw = built ? readFileSync(`${buildDir}sw.js`, 'utf8') : '';

  it('manifest dan service worker ada di build', () => {
    expect(existsSync(`${buildDir}manifest.webmanifest`)).toBe(true);
    expect(existsSync(`${buildDir}index.html`)).toBe(true);
    const manifest = JSON.parse(readFileSync(`${buildDir}manifest.webmanifest`, 'utf8'));
    expect(manifest).toMatchObject({ name: 'AlgoHub E-Learning', display: 'standalone' });
  });

  it('service worker tidak punya route cache untuk API', () => {
    // Nama kelas workbox dimangling di sw.js, jadi route dikenali dari pola dan ada tidaknya opsi:
    // NetworkOnly dibuat tanpa argumen, sedangkan CacheFirst membawa cacheName.
    const routes = [
      ...sw.matchAll(/\b\w\((\/\^https?(?:[^,]|\\,)+?\/[a-z]*),new (\w+)(\(|,)/g),
    ].map((m) => ({ pattern: toRegExp(m[1]), withOptions: m[3] === '(' }));
    expect(routes.length).toBeGreaterThan(1);
    for (const url of API_URLS) {
      const matched = routes.filter((route) => route.pattern.test(url));
      expect(matched.length, url).toBeGreaterThan(0);
      expect(
        matched.every((route) => !route.withOptions),
        url,
      ).toBe(true);
    }
    expect([...sw.matchAll(/cacheName:`([^`]+)`/g)].map((m) => m[1])).toEqual(['pyodide-cache']);
  });

  it('precache tidak memuat URL API', () => {
    expect(sw).not.toMatch(/url:`[^`]*\/api\//);
  });
});
