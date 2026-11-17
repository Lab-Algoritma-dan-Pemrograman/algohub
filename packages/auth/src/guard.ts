import { redirect } from '@sveltejs/kit';
import type { Session, SessionUser } from './session.svelte';
export function requireToken(session: Pick<Session<SessionUser>, 'hasToken'>, redirectTo: string) {
  if (!session.hasToken) redirect(307, redirectTo);
}
export function requireRole(
  session: Pick<Session<SessionUser>, 'user'>,
  roles: readonly string[],
  redirectTo: string,
) {
  const role = session.user?.role;
  if (!role || !roles.includes(role)) redirect(307, redirectTo);
}
