import { readFileSync } from 'node:fs';
import { describe, expect, it } from 'vitest';

interface Header {
  key: string;
  value: string;
}

interface Redirect {
  source: string;
  destination: string;
  statusCode?: number;
  permanent?: boolean;
}

interface VercelConfig {
  framework: string | null;
  installCommand: string;
  buildCommand: string;
  outputDirectory: string;
  regions: string[];
  rewrites?: unknown[];
  redirects: Redirect[];
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

// Meniru pencocokan path-to-regexp Vercel untuk dua bentuk yang dipakai: path persis dan
// akhiran `/:path*` (segmen boleh kosong). Query string tidak ditulis di destination karena
// Vercel meneruskannya sendiri, jadi di sini ditempel ulang dengan cara yang sama.
function redirectFor(url: string): { status: number; location: string } | null {
  const [path, query] = url.split('?');
  for (const rule of config.redirects) {
    const wildcard = rule.source.endsWith('/:path*');
    const base = wildcard ? rule.source.slice(0, -'/:path*'.length) : rule.source;
    let rest: string;
    if (path === base) rest = '';
    else if (wildcard && path.startsWith(`${base}/`)) rest = path.slice(base.length + 1);
    else continue;
    const target = wildcard
      ? rule.destination.replace(/\/:path\*$/, rest ? `/${rest}` : '')
      : rule.destination;
    return {
      status: rule.statusCode ?? 0,
      location: query === undefined ? target : `${target}?${query}`,
    };
  }
  return null;
}

const PRAKTIKUM = 'https://praktikum.algohub.web.id';

describe('vercel.json', () => {
  it('output statis tanpa framework preset dan tanpa rewrite fallback', () => {
    expect(config.framework).toBeNull();
    expect(config.installCommand).toBe('cd ../.. && bun install --frozen-lockfile');
    expect(config.buildCommand).toBe('bun run build');
    expect(config.outputDirectory).toBe('build');
    expect(config.regions).toEqual(['sin1']);
    expect(config.rewrites).toBeUndefined();
  });

  it('semua redirect memakai status 301', () => {
    for (const rule of config.redirects) {
      expect(rule.statusCode).toBe(301);
      expect(rule.permanent).toBeUndefined();
      expect(rule.destination.startsWith(`${PRAKTIKUM}/`)).toBe(true);
      expect(rule.destination).not.toContain('?');
    }
  });

  describe.each([
    ['/info', '/info'],
    ['/info/:path*', '/info/jadwal/kelas-a'],
    ['/praktikum', '/praktikum'],
    ['/praktikum/:path*', '/praktikum/dashboard'],
    ['/playground', '/playground'],
    ['/playground-c', '/playground-c'],
    ['/layanan/:path*', '/layanan/login'],
  ])('baris %s', (_row, path) => {
    it('301 ke path yang sama di praktikum', () => {
      expect(redirectFor(path)).toEqual({ status: 301, location: `${PRAKTIKUM}${path}` });
    });

    it('query string ikut', () => {
      expect(redirectFor(`${path}?tab=2&q=a`)).toEqual({
        status: 301,
        location: `${PRAKTIKUM}${path}?tab=2&q=a`,
      });
    });
  });

  it('path portal dan path mirip tidak ikut dialihkan', () => {
    for (const path of ['/', '/404', '/compiler/v1/clang', '/information', '/playground/x']) {
      expect(redirectFor(path)).toBeNull();
    }
  });

  it('header keamanan di semua path', () => {
    expect(headersFor('/(.*)')).toMatchObject({
      'X-Frame-Options': 'DENY',
      'X-Content-Type-Options': 'nosniff',
      'Referrer-Policy': 'strict-origin-when-cross-origin',
      'Strict-Transport-Security': 'max-age=63072000; includeSubDomains; preload',
      'Permissions-Policy':
        'camera=(self), microphone=(), geolocation=(), payment=(), usb=(), interest-cohort=()',
    });
    expect(headersFor('/(.*)')['Content-Security-Policy']).toContain("frame-ancestors 'none'");
  });

  it('file compiler immutable dan boleh diambil lintas origin', () => {
    expect(headersFor('/compiler/v1/(.*)')).toEqual({
      'Cache-Control': 'public, max-age=31536000, immutable',
      'Access-Control-Allow-Origin': '*',
      'Cross-Origin-Resource-Policy': 'cross-origin',
    });
  });

  it('aset build di-cache setahun', () => {
    expect(headersFor('/_app/immutable/(.*)')).toEqual({
      'Cache-Control': 'public, max-age=31536000, immutable',
    });
  });
});
