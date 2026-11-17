// @vitest-environment node
import { existsSync, readdirSync, readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { describe, expect, it } from 'vitest';
import * as rootLayout from '../src/routes/+layout';

const appDir = fileURLToPath(new URL('..', import.meta.url));
const srcFiles = readdirSync(`${appDir}src`, { recursive: true, encoding: 'utf8' }).map((file) =>
  file.replaceAll('\\', '/'),
);

describe('output statis saja', () => {
  it('tidak ada kode server SvelteKit di src', () => {
    const serverOnly = srcFiles.filter(
      (file) =>
        /(^|\/)hooks\.server\./.test(file) ||
        /(^|\/)\+(page|layout)\.server\./.test(file) ||
        /(^|\/)\+server\./.test(file) ||
        /(^|\/)server(\/|\.[jt]s$)/.test(file) ||
        /\.(server|remote)\./.test(file),
    );
    expect(serverOnly).toEqual([]);
  });

  it('adapter-static dengan fallback SPA dipasang di plugin sveltekit()', () => {
    const viteConfig = readFileSync(`${appDir}vite.config.ts`, 'utf8');
    expect(viteConfig).toMatch(/import adapter from '@sveltejs\/adapter-static'/);
    expect(viteConfig).toMatch(/adapter\(\{[^}]*fallback: 'index\.html'/);
    expect(existsSync(`${appDir}svelte.config.js`)).toBe(false);
  });

  it('root layout mematikan SSR dan prerender', () => {
    expect(rootLayout.ssr).toBe(false);
    expect(rootLayout.prerender).toBe(false);
  });
});
