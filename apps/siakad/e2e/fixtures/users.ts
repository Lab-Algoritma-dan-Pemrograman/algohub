export type Role = 'praktikan' | 'penyewa' | 'asisten' | 'koordinator';

// Baris login_user tanpa kolom password (dibuang auth_usecase.go sebelum dikirim).
export const USERS: Record<Role, Record<string, unknown> & { id: number; role: string }> = {
  praktikan: {
    id: 101,
    username: '202511001',
    full_name: 'Praktikan Uji',
    role: 'mahasiswa',
    nim: '202511001',
    assistant_code: null,
    division: null,
    phone_number: '081200000101',
    class_code: 'A',
    shift: null,
    is_active: true,
  },
  penyewa: {
    id: 201,
    username: 'penyewa1',
    full_name: 'Penyewa Uji',
    role: 'penyewa',
    nim: null,
    assistant_code: null,
    division: null,
    phone_number: '081200000201',
    is_active: true,
  },
  asisten: {
    id: 301,
    username: 'asisten1',
    full_name: 'Asisten Uji',
    role: 'asisten',
    nim: null,
    assistant_code: 'L01',
    division: 'inventaris',
    phone_number: '081200000301',
    is_active: true,
  },
  koordinator: {
    id: 401,
    username: 'koor1',
    full_name: 'Koordinator Uji',
    role: 'koordinator',
    nim: null,
    assistant_code: 'P01',
    division: null,
    phone_number: '081200000401',
    is_active: true,
  },
};

const CLAIM_KEYS = ['id', 'username', 'role', 'division', 'nim', 'assistant_code', 'full_name'];

// Klaim JWT yang dikembalikan /auth/verify (sessionClaimKeys di auth_usecase.go).
export function claimsOf(role: Role): Record<string, unknown> {
  const user = USERS[role];
  return Object.fromEntries(CLAIM_KEYS.filter((k) => k in user).map((k) => [k, user[k]]));
}

export const DIVISION_ACCESS: Record<Role, string[]> = {
  praktikan: [],
  penyewa: [],
  asisten: ['/inventaris', '/validasi-absensi'],
  koordinator: [],
};
