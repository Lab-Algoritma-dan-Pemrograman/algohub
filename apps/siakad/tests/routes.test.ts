import { beforeEach, describe, expect, it, vi } from 'vitest';

const api = vi.hoisted(() => ({
  verify: vi.fn(),
  rpc: vi.fn(),
  logout: vi.fn(),
}));

vi.mock('#lib/api.ts', () => api);
vi.mock('#lib/notifications.ts', () => ({ unsubscribeWebPush: vi.fn() }));
vi.mock('$app/navigation', () => ({ invalidateAll: vi.fn() }));

const { guardRoute } = await import('../src/lib/auth');
const { queryClient } = await import('../src/lib/query');
const { ROUTE_ACCESS } = await import('../src/lib/routes');

// Tabel Halaman 05-siakad.md (kolom Route, Role, Menu key) untuk route yang butuh login.
const EXPECTED: Record<string, { roles?: string[]; menuKey?: string }> = {
  '/beranda': {},
  '/e-learning': { menuKey: '/e-learning' },
  '/profil': {},
  '/absensi': {},
  '/sewa-barang': {},
  '/kritik-saran': { roles: ['praktikan', 'koordinator'] },
  '/jadwal-saya': { roles: ['asisten', 'praktikan', 'koordinator'] },
  '/manajemen-user': { roles: ['koordinator', 'asisten'], menuKey: '/manajemen-user' },
  '/validasi-absensi': { roles: ['asisten', 'koordinator'], menuKey: '/validasi-absensi' },
  '/jadwal-jaga': { roles: ['asisten', 'koordinator'] },
  '/inventaris': { roles: ['asisten', 'koordinator'], menuKey: '/inventaris' },
  '/buat-qr': { roles: ['asisten', 'koordinator'] },
  '/manajemen-jadwal': { roles: ['asisten', 'koordinator'], menuKey: '/manajemen-jadwal' },
  '/manajemen-kelas': { roles: ['koordinator', 'asisten'], menuKey: '/manajemen-kelas' },
  '/ketersediaan': { roles: ['koordinator', 'asisten'], menuKey: '/ketersediaan' },
  '/laporan-keuangan': { roles: ['koordinator', 'asisten'], menuKey: '/laporan-keuangan' },
  '/audit-log': { roles: ['asisten', 'koordinator'], menuKey: '/audit-log' },
  '/manajemen-sewa': { roles: ['koordinator', 'asisten'], menuKey: '/inventaris' },
  '/pengaturan': { roles: ['koordinator'] },
};

function signedIn(role: string, division?: string, allowed: string[] = []) {
  api.verify.mockResolvedValue({
    ok: true,
    status: 200,
    body: { user: { id: '7', username: 'u', full_name: 'U', role, division } },
  });
  api.rpc.mockResolvedValue({ data: allowed.map((menu_key) => ({ menu_key })), error: null });
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
  vi.clearAllMocks();
});

describe('ROUTE_ACCESS', () => {
  it('sama persis dengan tabel 05-siakad.md (19 route login)', () => {
    expect(ROUTE_ACCESS).toEqual(EXPECTED);
    expect(Object.keys(ROUTE_ACCESS)).toHaveLength(19);
  });
});

describe('guardRoute', () => {
  it('tanpa sesi diarahkan ke /', async () => {
    api.verify.mockResolvedValue({ ok: false, status: 401, body: { error: 'Unauthorized' } });
    expect(await outcome('/beranda')).toBe('/');
  });

  it('role DB lama dinormalkan: mahasiswa = praktikan', async () => {
    signedIn('mahasiswa');
    expect(await outcome('/kritik-saran')).toBe('ok');
    expect(await outcome('/validasi-absensi')).toBe('/beranda');
  });

  it('penyewa tidak boleh membuka manajemen user', async () => {
    signedIn('penyewa');
    expect(await outcome('/manajemen-user')).toBe('/beranda');
    expect(await outcome('/sewa-barang')).toBe('ok');
  });

  it('koordinator lolos semua route', async () => {
    signedIn('koordinator');
    for (const path of Object.keys(EXPECTED)) expect(await outcome(path)).toBe('ok');
    expect(api.rpc).not.toHaveBeenCalled();
  });

  it('asisten tanpa menu key divisi diarahkan ke /beranda', async () => {
    signedIn('asisten', 'inventaris', ['/inventaris']);
    expect(await outcome('/inventaris')).toBe('ok');
    expect(await outcome('/manajemen-sewa')).toBe('ok');
    expect(await outcome('/manajemen-user')).toBe('/beranda');
    expect(await outcome('/jadwal-jaga')).toBe('ok');
    expect(api.rpc).toHaveBeenCalledWith('get_division_access_secure', {
      p_viewer_id: 7,
      p_division: 'inventaris',
    });
  });

  it('ALWAYS_ALLOWED_MENUS lolos untuk asisten tanpa akses divisi', async () => {
    signedIn('asisten', 'humas', []);
    expect(await outcome('/e-learning')).toBe('ok');
  });
});
