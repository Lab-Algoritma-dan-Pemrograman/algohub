import { api } from '#lib/api.ts';
import { realtimeEnabled, setRealtimeAuth } from './client.ts';
// Token Realtime diperbarui sebelum kedaluwarsa; batas bawah mencegah loop rapat bila server
// mengirim umur token yang sangat pendek.
const REFRESH_MARGIN_MS = 60_000;
const MIN_DELAY_MS = 30_000;
const FALLBACK_LIFETIME_MS = 3_600_000;
const RETRY_MS = 30_000;
let timer: ReturnType<typeof setTimeout> | undefined;
let running = false;
// Umur token dibaca dari klaim exp (detik); expiresIn dari server dipakai bila klaim tidak ada.
export function tokenLifetimeMs(token: string, expiresIn?: number, now = Date.now()): number {
  try {
    const part = token.split('.')[1] ?? '';
    const claims = JSON.parse(atob(part.replace(/-/g, '+').replace(/_/g, '/'))) as { exp?: number };
    if (typeof claims.exp === 'number') return claims.exp * 1000 - now;
  } catch {
    // Token bukan JWT yang bisa dibaca; pakai expiresIn.
  }
  return typeof expiresIn === 'number' ? expiresIn * 1000 : FALLBACK_LIFETIME_MS;
}
export function refreshDelayMs(lifetimeMs: number): number {
  return Math.max(MIN_DELAY_MS, lifetimeMs - REFRESH_MARGIN_MS);
}
async function refresh(): Promise<void> {
  if (!running) return;
  let delay = RETRY_MS;
  try {
    const { token, expiresIn } = await api.auth.supabaseToken();
    await setRealtimeAuth(token);
    delay = refreshDelayMs(tokenLifetimeMs(token, expiresIn));
  } catch {
    // Realtime hanya pelengkap tampilan; coba lagi nanti tanpa mengganggu halaman.
  }
  if (running) timer = setTimeout(refresh, delay);
}
export function startRealtimeToken(): void {
  if (running || !realtimeEnabled()) return;
  running = true;
  void refresh();
}
export function stopRealtimeToken(): void {
  running = false;
  clearTimeout(timer);
  timer = undefined;
}
