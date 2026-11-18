import { ApiError } from '@algohub/api-client';
import { beforeEach, describe, expect, it, vi } from 'vitest';
const api = vi.hoisted(() => ({
  api: {
    auth: { verify: vi.fn(), receiveToken: vi.fn(), login: vi.fn() },
    sessions: { end: vi.fn() },
    activityLogs: { add: vi.fn() },
  },
}));
vi.mock('#lib/api.ts', () => api);
const { exchangeSsoCode, loginDirect } = await import('../src/lib/auth');
const { queryClient } = await import('../src/lib/query');
const { session } = await import('../src/lib/session');
const { authErrorMessage, ssoRedirectMessage, NETWORK_ERROR } =
  await import('../src/lib/auth-messages');
const payload = {
  nim: '2022110083',
  nama: 'Budi',
  kelas: 'A1',
  shift: '1',
  jurusan: 'TI',
  email: null,
  role: 'mahasiswa',
};
beforeEach(() => {
  queryClient.clear();
  localStorage.clear();
  session.clear();
  vi.clearAllMocks();
});
describe('pertukaran kode SSO', () => {
  it('menukar kode lewat receive-token dengan body {code} saja dan memulai sesi', async () => {
    api.api.auth.receiveToken.mockResolvedValue({
      success: true,
      token: 'long.jwt.token',
      redirectUrl: '/',
      payload,
    });
    const state = await exchangeSsoCode('kode-1');
    expect(api.api.auth.receiveToken).toHaveBeenCalledWith({ code: 'kode-1' });
    expect(session.token).toBe('long.jwt.token');
    expect(state.user).toMatchObject({ nim: '2022110083', nama: 'Budi', role: 'praktikan' });
    expect(session.user?.id).toBe(state.user.id);
  });
  it('kode yang sama tidak ditukar dua kali saat halaman /sso dipasang ulang', async () => {
    api.api.auth.receiveToken.mockResolvedValue({ token: 't', payload });
    await Promise.all([exchangeSsoCode('kode-2'), exchangeSsoCode('kode-2')]);
    expect(api.api.auth.receiveToken).toHaveBeenCalledTimes(1);
  });
  it('kode terpakai ditolak dengan pesan API dan tidak membuat sesi', async () => {
    api.api.auth.receiveToken.mockRejectedValue(new ApiError('Kode sudah digunakan.', 400));
    const error = await exchangeSsoCode('kode-3').catch((e: unknown) => e);
    expect(authErrorMessage(error, 'fallback')).toBe('Kode sudah digunakan.');
    expect(session.hasToken).toBe(false);
  });
  it('token tidak pernah masuk ke URL', async () => {
    api.api.auth.receiveToken.mockResolvedValue({ token: 'rahasia', payload });
    await exchangeSsoCode('kode-4');
    expect(location.href).not.toContain('rahasia');
    expect(location.search).toBe('');
  });
});
describe('login langsung', () => {
  it('memakai token dari respons /login sebagai Bearer', async () => {
    api.api.auth.login.mockResolvedValue({
      token: 'supabase.jwt',
      backendToken: 'go.jwt',
      payload,
      user: {},
    });
    await loginDirect('2022110083', 'pw');
    expect(api.api.auth.login).toHaveBeenCalledWith({ identifier: '2022110083', password: 'pw' });
    expect(session.token).toBe('supabase.jwt');
  });
  it('ganti user membersihkan cache memori user sebelumnya', async () => {
    api.api.auth.login.mockResolvedValue({ token: 'a', payload });
    await loginDirect('2022110083', 'pw');
    queryClient.setQueryData(['elearning', 'progress', 'lessons'], ['x']);
    api.api.auth.login.mockResolvedValue({
      token: 'b',
      payload: { ...payload, nim: '2022110099' },
    });
    await loginDirect('2022110099', 'pw');
    expect(queryClient.getQueryData(['elearning', 'progress', 'lessons'])).toBeUndefined();
  });
});
describe('pesan kesalahan', () => {
  it('kegagalan jaringan memakai pesan lama, kegagalan API memakai pesan API', () => {
    expect(authErrorMessage(new TypeError('x'), 'Login gagal. Coba lagi.')).toBe(NETWORK_ERROR);
    expect(authErrorMessage(new ApiError('NIM/email atau password salah.', 401), 'x')).toBe(
      'NIM/email atau password salah.',
    );
  });
  it('kode error redirect SSO dari backend dipetakan, kode asing diabaikan', () => {
    expect(ssoRedirectMessage('invalid_token')).toContain('tidak valid');
    expect(ssoRedirectMessage('missing_fields')).toContain('tidak lengkap');
    expect(ssoRedirectMessage('server_error')).toContain('kesalahan server');
    expect(ssoRedirectMessage('<script>')).toBeNull();
    expect(ssoRedirectMessage(null)).toBeNull();
  });
});
