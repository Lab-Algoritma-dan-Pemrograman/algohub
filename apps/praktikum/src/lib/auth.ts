import { logout } from '@algohub/auth';
import { queryClient, takePersister } from './query.ts';
import { session } from './session.ts';

export const LOGIN_PATH = '/praktikum/login';

export function signOut(): Promise<void> {
  return logout({ queryClient, persister: takePersister(), session });
}

export async function expireSession(): Promise<void> {
  await signOut();
  if (location.pathname !== LOGIN_PATH) location.href = LOGIN_PATH;
}
