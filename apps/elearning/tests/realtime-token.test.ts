import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

const api = vi.hoisted(() => ({ api: { auth: { supabaseToken: vi.fn() } } }));
const realtime = vi.hoisted(() => ({
  realtimeEnabled: vi.fn(() => true),
  setRealtimeAuth: vi.fn(async () => undefined),
}));
vi.mock('#lib/api.ts', () => api);
vi.mock('#lib/realtime/client.ts', () => realtime);

const { refreshDelayMs, startRealtimeToken, stopRealtimeToken, tokenLifetimeMs } =
  await import('../src/lib/realtime/token');

const NOW = Date.UTC(2026, 4, 1, 0, 0, 0);
const jwt = (exp: number) => `x.${btoa(JSON.stringify({ exp })).replace(/=+$/, '')}.y`;

beforeEach(() => {
  vi.useFakeTimers({ now: NOW });
  vi.clearAllMocks();
  realtime.realtimeEnabled.mockReturnValue(true);
});

afterEach(() => {
  stopRealtimeToken();
  vi.useRealTimers();
});

describe('umur token', () => {
  it('membaca klaim exp dan jatuh ke expiresIn bila token bukan JWT', () => {
    expect(tokenLifetimeMs(jwt(NOW / 1000 + 600), undefined, NOW)).toBe(600_000);
    expect(tokenLifetimeMs('bukan-jwt', 120)).toBe(120_000);
    expect(tokenLifetimeMs('bukan-jwt')).toBe(3_600_000);
  });

  it('refresh dijadwalkan semenit sebelum kedaluwarsa dengan batas bawah 30 detik', () => {
    expect(refreshDelayMs(3_600_000)).toBe(3_540_000);
    expect(refreshDelayMs(40_000)).toBe(30_000);
  });
});

describe('startRealtimeToken', () => {
  it('menukar token, memasangnya ke Realtime, lalu memperbarui sebelum kedaluwarsa', async () => {
    api.api.auth.supabaseToken.mockImplementation(async () => ({
      token: jwt((Date.now() + 3_600_000) / 1000),
    }));
    startRealtimeToken();
    await vi.advanceTimersByTimeAsync(0);
    expect(api.api.auth.supabaseToken).toHaveBeenCalledTimes(1);
    expect(realtime.setRealtimeAuth).toHaveBeenCalledTimes(1);

    await vi.advanceTimersByTimeAsync(3_540_000 - 1);
    expect(api.api.auth.supabaseToken).toHaveBeenCalledTimes(1);
    await vi.advanceTimersByTimeAsync(2);
    expect(api.api.auth.supabaseToken).toHaveBeenCalledTimes(2);
    expect(realtime.setRealtimeAuth).toHaveBeenCalledTimes(2);
  });

  it('gagal menukar token: mencoba lagi setelah 30 detik tanpa melempar', async () => {
    api.api.auth.supabaseToken.mockRejectedValueOnce(new Error('offline'));
    api.api.auth.supabaseToken.mockResolvedValue({ token: 'abc', expiresIn: 3600 });
    startRealtimeToken();
    await vi.advanceTimersByTimeAsync(0);
    expect(realtime.setRealtimeAuth).not.toHaveBeenCalled();
    await vi.advanceTimersByTimeAsync(30_000);
    expect(api.api.auth.supabaseToken).toHaveBeenCalledTimes(2);
    expect(realtime.setRealtimeAuth).toHaveBeenCalledWith('abc');
  });

  it('dipanggil dua kali tidak menggandakan jadwal, dan stop membatalkannya', async () => {
    api.api.auth.supabaseToken.mockResolvedValue({ token: 'abc', expiresIn: 3600 });
    startRealtimeToken();
    startRealtimeToken();
    await vi.advanceTimersByTimeAsync(0);
    expect(api.api.auth.supabaseToken).toHaveBeenCalledTimes(1);
    stopRealtimeToken();
    await vi.advanceTimersByTimeAsync(7_200_000);
    expect(api.api.auth.supabaseToken).toHaveBeenCalledTimes(1);
  });

  it('tanpa konfigurasi Supabase tidak ada permintaan token', async () => {
    realtime.realtimeEnabled.mockReturnValue(false);
    startRealtimeToken();
    await vi.advanceTimersByTimeAsync(0);
    expect(api.api.auth.supabaseToken).not.toHaveBeenCalled();
  });
});
