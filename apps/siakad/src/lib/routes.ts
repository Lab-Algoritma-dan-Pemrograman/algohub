export interface RouteAccess {
  roles?: readonly string[];
  menuKey?: string;
}

// Disalin dari App.tsx lama (ProtectedRoute allowedRoles dan requiredMenuKey).
export const ROUTE_ACCESS: Record<string, RouteAccess> = {
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
