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
    csp.split(';').map((part) => {
      const [name, ...values] = part.trim().split(/\s+/);
      return [name, values];
    }),
  );
}

describe('vercel.json', () => {
  it('output statis tanpa framework preset', () => {
    expect(config.framework).toBeNull();
    expect(config.installCommand).toBe('cd ../.. && bun install --frozen-lockfile');
    expect(config.buildCommand).toBe('bun run build');
    expect(config.outputDirectory).toBe('build');
    expect(config.regions).toEqual(['sin1']);
  });

  it('semua path jatuh ke index.html SPA', () => {
    expect(config.rewrites).toEqual([{ source: '/(.*)', destination: '/index.html' }]);
  });

  it('header keamanan dan isolasi lintas origin di semua path', () => {
    expect(headersFor('/(.*)')).toMatchObject({
      'X-Frame-Options': 'DENY',
      'X-Content-Type-Options': 'nosniff',
      'Referrer-Policy': 'strict-origin-when-cross-origin',
      'Strict-Transport-Security': 'max-age=63072000; includeSubDomains; preload',
      'Permissions-Policy':
        'camera=(self), microphone=(), geolocation=(), payment=(), usb=(), interest-cohort=()',
      'Cross-Origin-Opener-Policy': 'same-origin',
      'Cross-Origin-Embedder-Policy': 'credentialless',
    });
  });

  it('tidak membawa header debug lama', () => {
    for (const entry of config.headers) {
      expect(entry.headers.map((h) => h.key)).not.toContain('X-Hooks-Test');
    }
    expect(JSON.stringify(config)).not.toContain('X-Hooks-Test');
  });

  it('CSP sama dengan hooks.server.ts lama ditambah origin baru', () => {
    const csp = cspDirectives();
    expect(csp.get('default-src')).toEqual(["'self'"]);
    expect(csp.get('script-src')).toEqual([
      "'self'",
      "'unsafe-inline'",
      "'wasm-unsafe-eval'",
      'https://cdn.jsdelivr.net',
      'https://esm.sh',
      'blob:',
      'https://gcore.jsdelivr.net',
      'https://fastly.jsdelivr.net',
    ]);
    expect(csp.get('worker-src')).toEqual(["'self'", 'blob:']);
    expect(csp.get('style-src')).toEqual([
      "'self'",
      "'unsafe-inline'",
      'https://fonts.googleapis.com',
      'https://cdn.jsdelivr.net',
    ]);
    expect(csp.get('font-src')).toEqual([
      "'self'",
      'https://fonts.gstatic.com',
      'https://cdn.jsdelivr.net',
      'data:',
    ]);
    expect(csp.get('img-src')).toEqual([
      "'self'",
      'data:',
      'blob:',
      'https://*.supabase.co',
      'https://ui-avatars.com',
    ]);
    expect(csp.get('media-src')).toEqual(["'self'", 'blob:']);
    expect(csp.get('frame-src')).toEqual([
      "'self'",
      'https://*.supabase.co',
      'https://docs.google.com',
      'https://drive.google.com',
      'https://maps.google.com',
      'https://www.google.com',
    ]);
    expect(csp.get('frame-ancestors')).toEqual(["'none'"]);
    expect(csp.get('base-uri')).toEqual(["'self'"]);
    expect(csp.get('object-src')).toEqual(["'none'"]);
  });

  it('connect-src memuat API, Supabase, compiler, dan mirror Pyodide', () => {
    expect(cspDirectives().get('connect-src')).toEqual([
      "'self'",
      'https://api.algohub.web.id',
      'https://algohub-api.vercel.app',
      'https://*.supabase.co',
      'wss://*.supabase.co',
      'https://cdn.jsdelivr.net',
      'https://esm.sh',
      'https://algohub.web.id',
      'https://algohub-portal-gamma.vercel.app',
      'https://gcore.jsdelivr.net',
      'https://fastly.jsdelivr.net',
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
});
