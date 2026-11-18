// @vitest-environment node
import { readFileSync } from 'node:fs';
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
