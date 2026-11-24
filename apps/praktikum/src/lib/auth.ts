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

export function signOut(): Promise<void> {
  clearAllAnswerDrafts();
  return logout({ queryClient, persister: takePersister(), session });
}

export async function expireSession(): Promise<void> {
  await signOut();
  if (location.pathname !== LOGIN_PATH) location.href = LOGIN_PATH;
}
