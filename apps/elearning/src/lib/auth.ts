import { ApiError } from '@algohub/api-client';
import { elearningKeys, type AuthPayload } from '@algohub/api-client/elearning';
import { logout, requireRole } from '@algohub/auth';
import { redirect } from '@sveltejs/kit';
import { queryOptions } from '@tanstack/svelte-query';
import { api } from './api.ts';
import { queryClient, takePersister } from './query.ts';
import { normalizeRole, numericId } from './roles.ts';
import { accessFor, FALLBACK_ROUTE } from './routes.ts';
import { session, type ElearningSessionUser } from './session.ts';
import { appStore } from './stores/app.svelte.ts';
export interface AuthState {
  user: ElearningSessionUser;
}
export function toSessionUser(payload: AuthPayload): ElearningSessionUser {
  const nim = String(payload.nim ?? '');
  return {
    id: numericId(nim),
    nim,
    nama: payload.nama ?? nim,
    kelas: payload.kelas ?? '',
    role: normalizeRole(payload.role),
    jurusan: payload.jurusan,
    email: payload.email,
    shift: payload.shift,
  };
}
async function fetchSession(): Promise<AuthState | null> {
  const token = session.token;
  if (!token) return null;
  try {
    const res = await api.auth.verify({ token });
    const user = toSessionUser(res.payload);
    session.set(res.token || token, user);
    return { user };
  } catch (error) {
    // Token ditolak: sesi lokal tidak berlaku lagi. Gangguan jaringan atau 5xx tidak boleh
    // mengeluarkan user yang sedang offline, jadi identitas tersimpan tetap dipakai.
    if (error instanceof ApiError && error.status < 500) {
      session.clear();
      return null;
    }
    const cached = session.user;
    return cached ? { user: cached } : null;
  }
}
const sessionQuery = () =>
  queryOptions({
    queryKey: elearningKeys.session(),
    queryFn: fetchSession,
    staleTime: Infinity,
  });
export function loadSession(): Promise<AuthState | null> {
  return queryClient.ensureQueryData(sessionQuery());
}
export function sessionState(): AuthState | null | undefined {
  return queryClient.getQueryData(elearningKeys.session());
}
export function startSession(token: string, payload: AuthPayload): AuthState {
  const user = toSessionUser(payload);
  // Cache memori milik user lain tidak boleh ikut dipersist ke IndexedDB user yang baru masuk.
  if (session.user?.id !== user.id) queryClient.clear();
  session.set(token, user);
  const state = { user };
  queryClient.setQueryData(elearningKeys.session(), state);
  return state;
}
// Kode dari /sso hanya berlaku sekali dan 60 detik; token panjang tidak pernah lewat URL.
// Memo per kode: startSession mengganti userId sehingga layout membuat ulang halaman /sso, dan
// pemasangan kedua tidak boleh menukar kode yang sama lagi (server menolaknya sebagai terpakai).
const ssoExchanges = new Map<string, Promise<AuthState>>();
export function exchangeSsoCode(code: string): Promise<AuthState> {
  let exchange = ssoExchanges.get(code);
  if (!exchange) {
    exchange = api.auth.receiveToken({ code }).then((res) => startSession(res.token, res.payload));
    ssoExchanges.set(code, exchange);
    exchange.catch(() => ssoExchanges.delete(code));
  }
  return exchange;
}
export async function loginDirect(identifier: string, password: string): Promise<AuthState> {
  const res = await api.auth.login({ identifier, password });
  return startSession(res.token, res.payload);
}
export async function signOut(): Promise<void> {
  // Dua panggilan ini harus selesai sebelum token lokal dihapus, tapi kegagalannya tidak
  // boleh menahan logout (jaringan putus, sesi sudah kedaluwarsa).
  const user = session.user;
  if (user) {
    await api.activityLogs
      .add({ eventType: 'logout', details: 'Melakukan logout dari sistem' })
      .catch(() => undefined);
    await api.sessions.end().catch(() => undefined);
  }
  appStore.reset();
  await logout({ queryClient, persister: takePersister(), session });
  location.href = '/';
}
export async function guardRoute(pathname: string): Promise<AuthState> {
  const state = await loadSession();
  if (!state) redirect(307, '/');
  const access = accessFor(pathname);
  if (access) requireRole({ user: state.user }, access.roles, FALLBACK_ROUTE);
  return state;
}
