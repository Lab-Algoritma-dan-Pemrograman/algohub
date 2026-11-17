import { readdirSync, readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { describe, expect, it } from 'vitest';
import * as rootLayout from '../src/routes/+layout';

const appDir = fileURLToPath(new URL('..', import.meta.url));
const srcFiles = readdirSync(`${appDir}src`, { recursive: true, encoding: 'utf8' }).map((file) =>
  file.replaceAll('\\', '/'),
);

describe('portal statis penuh', () => {
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

  it('adapter-static tanpa fallback SPA', () => {
    const viteConfig = readFileSync(`${appDir}vite.config.ts`, 'utf8');
    expect(viteConfig).toMatch(/import adapter from '@sveltejs\/adapter-static'/);
    expect(viteConfig).not.toMatch(/fallback/);
  });

  it('semua halaman di-prerender', () => {
    expect(rootLayout.prerender).toBe(true);
  });

  it('folder unduhan compiler di-ignore git', () => {
    expect(readFileSync(`${appDir}.gitignore`, 'utf8').split(/\r?\n/)).toContain(
      '/static/compiler/',
    );
  });
});
