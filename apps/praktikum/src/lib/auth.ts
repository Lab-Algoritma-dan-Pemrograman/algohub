import { logout } from '@algohub/auth';
import { clearAllAnswerDrafts } from './answer-drafts.ts';
import { queryClient, takePersister } from './query.ts';
import { session } from './session.ts';

export const LOGIN_PATH = '/praktikum/login';

export const PUBLIC_PATHS = [
  LOGIN_PATH,
  '/praktikum/register',
  '/praktikum/lupa-password',
  '/praktikum/reset-password',
];

export function homeFor(role: string): string {
  return role === 'asisten' || role === 'koordinator' ? '/praktikum/admin' : '/praktikum/dashboard';
}

function clearSession(): Promise<void> {
  return logout({ queryClient, persister: takePersister(), session });
}

export function signOut(): Promise<void> {
  clearAllAnswerDrafts();
  return clearSession();
}

export async function expireSession(): Promise<void> {
  // Jangan hapus draft jawaban di sini. 401 bisa terjadi di tengah ujian
  // (token kedaluwarsa), dan draft per user+sesi harus dipulihkan setelah
  // login ulang. Menghapusnya di sini justru menghilangkan jawaban terakhir
  // yang belum sempat tersimpan ke server.
  await clearSession();
  if (location.pathname !== LOGIN_PATH) location.href = LOGIN_PATH;
}
