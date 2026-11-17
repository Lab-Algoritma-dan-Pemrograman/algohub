// @vitest-environment node
import { existsSync, readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { describe, expect, it } from 'vitest';
import { PUBLIC_PATHS } from '../src/lib/auth';

const routesDir = fileURLToPath(new URL('../src/routes/', import.meta.url));

// Daftar route § Halaman di 07-praktikum.md, sama dengan URL app lama.
const ROUTES = [
  '',
  'info',
  'info/asisten',
  'info/jadwal',
  'info/laporan',
  'info/modul',
  'info/praktikum',
  'layanan/login',
  'playground',
  'playground-c',
  'praktikum/login',
  'praktikum/register',
  'praktikum/lupa-password',
  'praktikum/reset-password',
  'praktikum/dashboard',
  'praktikum/profil',
  'praktikum/laporan-besar',
  'praktikum/sesi',
  'praktikum/sesi/[sesiId]',
  'praktikum/sesi/[sesiId]/pretest',
  'praktikum/sesi/[sesiId]/posttest',
  'praktikum/sesi/[sesiId]/keterampilan',
  'praktikum/sesi/[sesiId]/ujian',
  'praktikum/admin',
  ...[
    'aktivasi',
    'asisten',
    'jadwal',
    'kelas',
    'laporan-besar',
    'log',
    'modul',
    'pedoman',
    'pengaturan',
    'penilaian',
    'rekap-jawaban',
    'rekap-nilai',
    'sesi',
    'users',
  ].map((p) => `praktikum/admin/${p}`),
  'praktikum/superadmin/pengaturan',
];

const read = (file: string) => readFileSync(`${routesDir}${file}`, 'utf8');

describe('route dan guard', () => {
  it.each(ROUTES)('/%s punya +page.svelte', (route) => {
    expect(existsSync(`${routesDir}${route ? `${route}/` : ''}+page.svelte`)).toBe(true);
  });

  it('praktikum/ mengecualikan 4 path publik dari cek token', () => {
    expect(PUBLIC_PATHS).toEqual([
      '/praktikum/login',
      '/praktikum/register',
      '/praktikum/lupa-password',
      '/praktikum/reset-password',
    ]);
    const layout = read('praktikum/+layout.ts');
    expect(layout).toMatch(/import \{[^}]*PUBLIC_PATHS[^}]*\} from '#lib\/auth\.ts'/);
    expect(layout).toMatch(/requireToken\(session, LOGIN_PATH\)/);
  });

  it('admin/ hanya untuk asisten dan koordinator', () => {
    const layout = read('praktikum/admin/+layout.ts');
    expect(layout).toMatch(/await parent\(\)/);
    expect(layout).toMatch(
      /requireRole\(session, \['asisten', 'koordinator'\], '\/praktikum\/dashboard'\)/,
    );
  });

  it('superadmin/ hanya untuk koordinator', () => {
    const layout = read('praktikum/superadmin/+layout.ts');
    expect(layout).toMatch(/await parent\(\)/);
    expect(layout).toMatch(/requireRole\(session, \['koordinator'\], '\/praktikum\/dashboard'\)/);
  });
});
