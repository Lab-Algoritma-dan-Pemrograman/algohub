// @vitest-environment node
import { readFileSync } from 'node:fs';
import { describe, expect, it } from 'vitest';
interface Header {
  key: string;
  value: string;
}
interface VercelConfig {
  framework: string | null;
  installCommand: string;
  buildCommand: string;
  outputDirectory: string;
  regions: string[];
  rewrites: { source: string; destination: string }[];
  headers: { source: string; headers: Header[] }[];
  crons?: unknown;
}
const config: VercelConfig = JSON.parse(
  readFileSync(new URL('../vercel.json', import.meta.url), 'utf8'),
);
function headersFor(source: string): Record<string, string> {
  const entry = config.headers.find((h) => h.source === source);
  if (!entry) throw new Error(`tidak ada header untuk ${source}`);
  return Object.fromEntries(entry.headers.map((h) => [h.key, h.value]));
}
function cspDirectives(): Map<string, string[]> {
  const csp = headersFor('/(.*)')['Content-Security-Policy'];
  return new Map(
    csp
      .split(';')
      .map((part) => part.trim())
      .filter(Boolean)
      .map((part) => {
        const [name, ...values] = part.split(/\s+/);
        return [name, values];
      }),
  );
}
describe('vercel.json', () => {
  it('output statis tanpa framework preset dan tanpa cron', () => {
    expect(config.framework).toBeNull();
    expect(config.installCommand).toBe('cd ../.. && bun install --frozen-lockfile');
    expect(config.buildCommand).toBe('bun run build');
    expect(config.outputDirectory).toBe('build');
    expect(config.regions).toEqual(['sin1']);
    // Cron keep-alive pindah ke backend Go (02-backend.md), bukan di app statis.
    expect(config.crons).toBeUndefined();
  });
  it('semua path jatuh ke index.html SPA', () => {
    expect(config.rewrites).toEqual([{ source: '/(.*)', destination: '/index.html' }]);
  });
  it('COOP dan COEP terpasang di semua path agar SharedArrayBuffer tersedia', () => {
    expect(headersFor('/(.*)')).toMatchObject({
      'Cross-Origin-Opener-Policy': 'same-origin',
      'Cross-Origin-Embedder-Policy': 'credentialless',
    });
  });
  it('header keamanan dasar', () => {
    expect(headersFor('/(.*)')).toMatchObject({
      'X-Frame-Options': 'DENY',
      'X-Content-Type-Options': 'nosniff',
      'Referrer-Policy': 'strict-origin-when-cross-origin',
      'Strict-Transport-Security': 'max-age=63072000; includeSubDomains; preload',
    });
  });
  it('CSP mengikuti vercel.json lama tanpa Firebase, Judge0, dan Wandbox', () => {
    const csp = cspDirectives();
    expect(csp.get('default-src')).toEqual(["'self'"]);
    expect(csp.get('script-src')).toEqual([
      "'self'",
      "'unsafe-inline'",
      "'unsafe-eval'",
      "'wasm-unsafe-eval'",
      'https://cdn.jsdelivr.net',
      'https://gcore.jsdelivr.net',
      'https://fastly.jsdelivr.net',
      'blob:',
    ]);
    expect(csp.get('worker-src')).toEqual(["'self'", 'blob:']);
    expect(csp.get('style-src')).toEqual([
      "'self'",
      "'unsafe-inline'",
      'https://fonts.googleapis.com',
      'https://cdn.jsdelivr.net',
      'https://gcore.jsdelivr.net',
      'https://fastly.jsdelivr.net',
    ]);
    expect(csp.get('font-src')).toEqual([
      "'self'",
      'https://fonts.gstatic.com',
      'https://cdn.jsdelivr.net',
      'https://gcore.jsdelivr.net',
      'https://fastly.jsdelivr.net',
      'data:',
    ]);
    expect(csp.get('img-src')).toEqual(["'self'", 'data:', 'blob:', 'https:']);
    expect(csp.get('frame-ancestors')).toEqual(["'none'"]);
    expect(csp.get('base-uri')).toEqual(["'self'"]);
    expect(csp.get('object-src')).toEqual(["'none'"]);
    const raw = headersFor('/(.*)')['Content-Security-Policy'];
    for (const dropped of ['firebase', '*.googleapis.com', 'wandbox', 'judge0']) {
      expect(raw).not.toContain(dropped);
    }
  });
  it('connect-src memuat API Go, Supabase Realtime, jsDelivr, dan portal', () => {
    expect(cspDirectives().get('connect-src')).toEqual([
      "'self'",
      'https://api.algohub.web.id',
      'https://*.supabase.co',
      'wss://*.supabase.co',
      'https://cdn.jsdelivr.net',
      'https://gcore.jsdelivr.net',
      'https://fastly.jsdelivr.net',
      'https://algohub.web.id',
      'https://algohub-portal.vercel.app',
      'https://siakad.algohub.web.id',
    ]);
  });
  it('worker runner selalu direvalidasi, aset immutable di-cache setahun', () => {
    expect(headersFor('/runner/(.*)')).toEqual({
      'Cache-Control': 'public, max-age=0, must-revalidate',
      'Cross-Origin-Resource-Policy': 'same-origin',
    });
    expect(headersFor('/_app/immutable/(.*)')).toEqual({
      'Cache-Control': 'public, max-age=31536000, immutable',
    });
  });
  it('service worker dan manifest tidak di-cache browser', () => {
    expect(headersFor('/sw.js')['Cache-Control']).toContain('no-cache');
    expect(headersFor('/manifest.webmanifest')['Cache-Control']).toBe('no-cache');
  });
});
