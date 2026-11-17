// Kolom users.role di DB memakai kosakata yang berbeda dari kode.
// - `admin` bukan role tersendiri di SIAKAD: user ber-role 'admin' sebenarnya asisten
//   (punya assistant_code dan terdaftar di group_assistants). Di E-Learning mereka tetap 'admin'.
// - `superadmin` adalah USERNAME (superadmin_ap, role koordinator), sengaja tidak dipetakan
//   supaya role tak dikenal gagal (fail closed).
export const ROLE_ALIASES: Record<string, string> = {
  praktikan: 'praktikan',
  mhs: 'praktikan',
  mahasiswa: 'praktikan',
  peminjam: 'penyewa',
  penyewa: 'penyewa',
  kordas: 'koordinator',
  korda: 'koordinator',
  coordinator: 'koordinator',
  admin: 'asisten',
};

export const CANON_ROLES = ['koordinator', 'asisten', 'praktikan', 'penyewa'] as const;

export const canonRole = (r?: string | null): string =>
  ROLE_ALIASES[(r ?? '').toLowerCase().trim()] ?? (r ?? '').toLowerCase().trim();

export const roleAllowed = (
  role: string | undefined | null,
  allowed: readonly string[],
): boolean => {
  const r = canonRole(role);
  if (r === 'koordinator') return true;
  return allowed.map(canonRole).includes(r);
};

export const isStaffRole = (role?: string | null): boolean =>
  ['koordinator', 'asisten'].includes(canonRole(role));

// Alat kerja dasar semua asisten, dikecualikan dari pembatasan division_access. Dipakai sidebar
// dan guard route supaya keduanya tidak bisa berbeda pendapat.
export const ALWAYS_ALLOWED_MENUS = ['/absensi', '/buat-qr', '/e-learning'];
