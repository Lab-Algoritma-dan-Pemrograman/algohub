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
    expect(config.crons).toBeUndefined();
  });

  it('semua path jatuh ke index.html SPA', () => {
    expect(config.rewrites).toEqual([{ source: '/(.*)', destination: '/index.html' }]);
  });

  it('header keamanan sama dengan siakad lama, tanpa isolasi lintas origin', () => {
    const headers = headersFor('/(.*)');
    expect(headers).toMatchObject({
      'X-Frame-Options': 'DENY',
      'X-Content-Type-Options': 'nosniff',
      'Referrer-Policy': 'strict-origin-when-cross-origin',
      'Strict-Transport-Security': 'max-age=63072000; includeSubDomains; preload',
      'Permissions-Policy':
        'camera=(self), microphone=(), geolocation=(), payment=(), usb=(), interest-cohort=()',
    });
    expect(headers['Cross-Origin-Opener-Policy']).toBeUndefined();
    expect(headers['Cross-Origin-Embedder-Policy']).toBeUndefined();
  });

  it('CSP sama dengan vercel.json lama', () => {
    const csp = cspDirectives();
    expect(csp.get('default-src')).toEqual(["'self'"]);
    expect(csp.get('script-src')).toEqual(["'self'", "'unsafe-inline'", 'cdnjs.cloudflare.com']);
    expect(csp.get('style-src')).toEqual([
      "'self'",
      "'unsafe-inline'",
      'fonts.googleapis.com',
      'cdnjs.cloudflare.com',
    ]);
    expect(csp.get('font-src')).toEqual([
      "'self'",
      'fonts.gstatic.com',
      'cdnjs.cloudflare.com',
      'data:',
    ]);
    expect(csp.get('img-src')).toEqual([
      "'self'",
      'data:',
      'blob:',
      'https://*.public.blob.vercel-storage.com',
    ]);
    expect(csp.get('media-src')).toEqual(["'self'", 'blob:']);
    expect(csp.get('worker-src')).toEqual(["'self'", 'blob:']);
    expect(csp.get('frame-ancestors')).toEqual(["'none'"]);
  });

  it('connect-src memuat API Go, Supabase Realtime, dan origin lama', () => {
    expect(cspDirectives().get('connect-src')).toEqual([
      "'self'",
      'https://api.algohub.web.id',
      'https://algohub-api.vercel.app',
      'https://*.supabase.co',
      'wss://*.supabase.co',
      'https://generativelanguage.googleapis.com',
      'https://algohub.web.id',
    ]);
  });

  it('aset immutable di-cache setahun', () => {
    expect(headersFor('/_app/immutable/(.*)')).toEqual({
      'Cache-Control': 'public, max-age=31536000, immutable',
    });
  });
});
