import { siakadKeys, type SiakadUser } from '@algohub/api-client/siakad';
import { logout, requireRole } from '@algohub/auth';
import { invalidateAll } from '$app/navigation';
import { redirect } from '@sveltejs/kit';
import { queryOptions } from '@tanstack/svelte-query';
import { logout as revokeSession, rpc, verify } from './api.ts';
import { unsubscribeWebPush } from './notifications.ts';
import { queryClient, takePersister } from './query.ts';
import { ALWAYS_ALLOWED_MENUS, canonRole } from './roles.ts';
import { ROUTE_ACCESS } from './routes.ts';
import { session } from './session.ts';

export interface AuthState {
  user: SiakadUser;
  allowedPaths: string[];
}

type RawUser = Record<string, unknown> & { id: number | string; role: string };

async function fetchAllowedPaths(user: SiakadUser): Promise<string[]> {
  if (user.role !== 'asisten' || !user.division) return [];
  const { data, error } = await rpc<(string | { menu_key: string })[]>(
    'get_division_access_secure',
    { p_viewer_id: user.id, p_division: user.division },
  );
  if (!data || error || !Array.isArray(data)) return [];
  return data.map((d) => (typeof d === 'string' ? d : d.menu_key));
}

// Hanya klaim JWT yang disimpan di localStorage; kolom lain dari login_user cukup di memori.
function toSessionUser(raw: RawUser): SiakadUser {
  return {
    id: Number(raw.id),
    username: raw.username as string,
    full_name: raw.full_name as string,
    role: canonRole(raw.role),
    nim: (raw.nim as string) || undefined,
    assistant_code: (raw.assistant_code as string) || undefined,
    division: (raw.division as string) || undefined,
  };
}

async function fetchSession(): Promise<AuthState | null> {
  try {
    const res = await verify();
    if (!res.ok) {
      session.clear();
      return null;
    }
    const raw = res.body?.user;
    if (!raw) return null;
    const user = toSessionUser(raw);
    if (session.token) session.set(session.token, user);
    return { user, allowedPaths: await fetchAllowedPaths(user) };
  } catch (error) {
    console.error('Gagal membaca sesi:', error);
    session.clear();
    return null;
  }
}

const sessionQuery = () =>
  queryOptions({
    queryKey: siakadKeys.session(),
    queryFn: fetchSession,
    staleTime: Infinity,
  });

export function loadSession(): Promise<AuthState | null> {
  return queryClient.ensureQueryData(sessionQuery());
}

export function sessionState(): AuthState | null | undefined {
  return queryClient.getQueryData(siakadKeys.session());
}

export async function startSession(token: string, raw: RawUser): Promise<AuthState> {
  const user = { ...raw, ...toSessionUser(raw), role: canonRole(raw.role) } as SiakadUser;
  // Cache memori milik user lain tidak boleh ikut dipersist ke IndexedDB user yang baru masuk.
  if (session.user?.id !== user.id) queryClient.clear();
  session.set(token, toSessionUser(raw));
  const state = { user, allowedPaths: await fetchAllowedPaths(user) };
  queryClient.setQueryData(siakadKeys.session(), state);
  return state;
}

// Load layout dijalankan ulang supaya sidebar dan halaman membaca user yang baru.
export async function updateSessionUser(patch: Record<string, unknown>): Promise<void> {
  const state = sessionState();
  if (!state) return;
  const user = { ...state.user, ...patch } as SiakadUser;
  queryClient.setQueryData(siakadKeys.session(), { ...state, user });
  if (session.token) session.set(session.token, toSessionUser(user));
  await invalidateAll();
}

export async function signOut(): Promise<void> {
  await unsubscribeWebPush(sessionState()?.user ?? session.user);
  // Dipanggil sebelum token lokal dihapus supaya fallback Bearer tetap ikut mencabut token.
  await revokeSession().catch(() => undefined);
  await logout({ queryClient, persister: takePersister(), session });
  location.href = '/';
}

export async function guardRoute(pathname: string): Promise<AuthState> {
  const state = await loadSession();
  if (!state) redirect(307, '/');
  const access = ROUTE_ACCESS[pathname];
  if (!access) return state;
  if (access.roles) {
    requireRole({ user: state.user }, [...access.roles, 'koordinator'], '/beranda');
  }
  const { menuKey } = access;
  if (
    menuKey &&
    state.user.role === 'asisten' &&
    !ALWAYS_ALLOWED_MENUS.includes(menuKey) &&
    !state.allowedPaths.includes(menuKey)
  ) {
    redirect(307, '/beranda');
  }
  return state;
}
