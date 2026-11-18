import { existsSync } from 'node:fs';
import { resolve } from 'node:path';
import { beforeEach, describe, expect, it, vi } from 'vitest';
const api = vi.hoisted(() => ({
  api: {
    auth: { verify: vi.fn(), receiveToken: vi.fn(), login: vi.fn() },
    sessions: { end: vi.fn() },
    activityLogs: { add: vi.fn() },
  },
}));
vi.mock('#lib/api.ts', () => api);
const { guardRoute } = await import('../src/lib/auth');
const { queryClient } = await import('../src/lib/query');
const { ROUTE_ACCESS } = await import('../src/lib/routes');
const { session } = await import('../src/lib/session');
// Tabel Frontend di 06-elearning.md: hanya tiga route yang dibatasi role.
const EXPECTED = {
  '/admin': { roles: ['admin', 'kordas'] },
  '/auditlog': { roles: ['admin', 'kordas'] },
  '/monitoring': { roles: ['admin', 'kordas', 'asisten'] },
};
function signedIn(role: string) {
  session.set('tok', {
    id: 1,
    nim: '2022110083',
    nama: 'U',
    kelas: 'A',
    role: 'praktikan',
    jurusan: null,
    email: null,
    shift: null,
  });
  api.api.auth.verify.mockResolvedValue({
    token: 'tok2',
    payload: {
      nim: '2022110083',
      nama: 'U',
      kelas: 'A',
      shift: null,
      jurusan: null,
      email: null,
      role,
    },
  });
}
async function outcome(pathname: string): Promise<string> {
  try {
    await guardRoute(pathname);
    return 'ok';
  } catch (err) {
    return (err as { location?: string }).location ?? String(err);
  }
}
beforeEach(() => {
  queryClient.clear();
  localStorage.clear();
  session.clear();
  vi.clearAllMocks();
});
describe('ROUTE_ACCESS', () => {
  it('sama dengan App.tsx lama: admin dan auditlog admin/kordas, monitoring juga asisten', () => {
    expect(ROUTE_ACCESS).toEqual(EXPECTED);
  });
});
describe('route FEAT-002', () => {
  const routesDir = resolve(process.cwd(), 'src/routes') + '/';
  it('/, /login, /sso, dan /dashboard punya halaman', () => {
    const pages = [
      '+page.svelte',
      'login/+page.svelte',
      'sso/+page.svelte',
      '(app)/dashboard/+page.svelte',
    ];
    expect(pages.filter((page) => !existsSync(`${routesDir}${page}`))).toEqual([]);
  });
});
describe('guardRoute', () => {
  it('tanpa token diarahkan ke / tanpa memanggil API', async () => {
    expect(await outcome('/dashboard')).toBe('/');
    expect(api.api.auth.verify).not.toHaveBeenCalled();
  });
  it('token ditolak server: sesi dihapus dan diarahkan ke /', async () => {
    const { ApiError } = await import('@algohub/api-client');
    session.set('bad', {
      id: 1,
      nim: '1',
      nama: 'U',
      kelas: '',
      role: 'praktikan',
      jurusan: null,
      email: null,
      shift: null,
    });
    api.api.auth.verify.mockRejectedValue(new ApiError('Unauthorized', 401));
    expect(await outcome('/dashboard')).toBe('/');
    expect(session.hasToken).toBe(false);
  });
  it('jaringan putus tidak mengeluarkan user yang sudah punya identitas tersimpan', async () => {
    signedIn('praktikan');
    api.api.auth.verify.mockRejectedValue(new TypeError('Failed to fetch'));
    expect(await outcome('/dashboard')).toBe('ok');
  });
  it('praktikan ditolak di route staf dan jatuh ke /dashboard', async () => {
    signedIn('mahasiswa');
    expect(await outcome('/dashboard')).toBe('ok');
    expect(await outcome('/admin')).toBe('/dashboard');
    expect(await outcome('/auditlog')).toBe('/dashboard');
    expect(await outcome('/monitoring')).toBe('/dashboard');
  });
  it('asisten hanya boleh monitoring', async () => {
    signedIn('asisten');
    expect(await outcome('/monitoring')).toBe('ok');
    expect(await outcome('/admin')).toBe('/dashboard');
    expect(await outcome('/auditlog')).toBe('/dashboard');
  });
  it.each(['admin', 'koordinator'])('%s lolos semua route staf', async (role) => {
    signedIn(role);
    for (const path of Object.keys(EXPECTED)) expect(await outcome(path)).toBe('ok');
  });
  it('sub-path route staf ikut dibatasi', async () => {
    signedIn('praktikan');
    expect(await outcome('/admin/users')).toBe('/dashboard');
  });
});

const layouts = {
  '/admin': await import('../src/routes/(app)/admin/+layout'),
  '/monitoring': await import('../src/routes/(app)/monitoring/+layout'),
  '/auditlog': await import('../src/routes/(app)/auditlog/+layout'),
};

// Layout per route staf (FEAT-005) dipanggil lewat load() seperti SvelteKit: yang tak berhak
// mendapat redirect 307 ke /dashboard, yang berhak lolos tanpa error.
describe('layout route staf', () => {
  async function enter(path: keyof typeof layouts): Promise<unknown> {
    try {
      await layouts[path].load({ url: new URL(`http://localhost${path}`) } as never);
      return 'ok';
    } catch (err) {
      return err;
    }
  }

  it.each([
    ['praktikan', { '/admin': false, '/monitoring': false, '/auditlog': false }],
    ['asisten', { '/admin': false, '/monitoring': true, '/auditlog': false }],
    ['kordas', { '/admin': true, '/monitoring': true, '/auditlog': true }],
    ['admin', { '/admin': true, '/monitoring': true, '/auditlog': true }],
  ] as const)('%s', async (role, allowed) => {
    signedIn(role);
    for (const [path, ok] of Object.entries(allowed)) {
      const result = await enter(path as keyof typeof layouts);
      if (ok) expect(result).toBe('ok');
      else expect(result).toMatchObject({ status: 307, location: '/dashboard' });
    }
  });

  it('tanpa sesi diarahkan ke /', async () => {
    expect(await enter('/admin')).toMatchObject({ status: 307, location: '/' });
  });
});
