export type AppRole = 'admin' | 'kordas' | 'asisten' | 'praktikan';
// Port dari normalizeRole di services/tokenService.ts lama. Role tak dikenal jatuh ke praktikan
// (hak paling kecil), sama seperti app lama.
export function normalizeRole(role: string | null | undefined): AppRole {
  if (!role) return 'praktikan';
  const clean = String(role).toLowerCase().trim();
  if (['admin', 'superadmin', 'super_admin', 'administrator', 'super admin'].includes(clean)) {
    return 'admin';
  }
  if (['kordas', 'koordinator', 'korda', 'coordinator'].includes(clean)) return 'kordas';
  if (['asisten', 'assistant', 'laboran', 'ast'].includes(clean)) return 'asisten';
  return 'praktikan';
}
export const isAdminOrKordas = (role: string | null | undefined): boolean => {
  const r = normalizeRole(role);
  return r === 'admin' || r === 'kordas';
};
export const isStaff = (role: string | null | undefined): boolean =>
  isAdminOrKordas(role) || normalizeRole(role) === 'asisten';
// SessionUser dari @algohub/auth mewajibkan id numerik, sedangkan identitas elearning adalah NIM
// (string, bisa berupa username). Hash FNV-1a 32-bit menghasilkan id stabil per NIM untuk
// kunci IndexedDB dan penanda ganti user.
export function numericId(nim: string): number {
  let hash = 0x811c9dc5;
  for (let i = 0; i < nim.length; i++) {
    hash ^= nim.charCodeAt(i);
    hash = Math.imul(hash, 0x01000193);
  }
  return hash >>> 0;
}
