import { afterEach, describe, expect, it, vi } from 'vitest';
import { ApiError, createApi } from './client';
function jsonResponse(status: number, body: unknown): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { 'Content-Type': 'application/json' },
  });
}
function setup(response: Response, token: string | null = 'tok') {
  const fetchMock = vi.fn(async (_url: string, _init: RequestInit) => response);
  vi.stubGlobal('fetch', fetchMock);
  const onUnauthorized = vi.fn();
  const api = createApi({
    baseUrl: 'http://api.test',
    prefix: '/api/praktikum',
    getToken: () => token,
    onUnauthorized,
  });
  const call = () => {
    const [url, init] = fetchMock.mock.calls[0];
    return { url, init, headers: init.headers as Record<string, string> };
  };
  return { api, onUnauthorized, call };
}
afterEach(() => {
  vi.unstubAllGlobals();
});
describe('createApi', () => {
  it('mengirim Bearer, prefix, dan credentials same-origin lalu mengembalikan data', async () => {
    const { api, call } = setup(jsonResponse(200, { success: true, message: 'ok', data: [1] }));
    await expect(api.get<number[]>('/info/modul')).resolves.toEqual([1]);
    const { url, init, headers } = call();
    expect(url).toBe('http://api.test/api/praktikum/info/modul');
    expect(init.method).toBe('GET');
    expect(init.credentials).toBe('same-origin');
    expect(headers['Authorization']).toBe('Bearer tok');
  });
  it('tidak mengirim Authorization tanpa token', async () => {
    const { api, call } = setup(jsonResponse(200, { success: true, message: 'ok' }), null);
    await api.get('/info/modul');
    expect(call().headers['Authorization']).toBeUndefined();
  });
  it('tidak mengirim Authorization dengan auth: false', async () => {
    const { api, call } = setup(jsonResponse(200, { success: true, message: 'ok' }));
    await api.post('/auth/login', { nim: '1' }, { auth: false });
    expect(call().headers['Authorization']).toBeUndefined();
  });
  it('mengirim body JSON dengan Content-Type', async () => {
    const { api, call } = setup(jsonResponse(200, { success: true, message: 'ok' }));
    await api.put('/profile', { nama: 'A' });
    const { init, headers } = call();
    expect(init.method).toBe('PUT');
    expect(headers['Content-Type']).toBe('application/json');
    expect(init.body).toBe('{"nama":"A"}');
  });
  it('mengirim FormData tanpa Content-Type', async () => {
    const { api, call } = setup(jsonResponse(200, { success: true, message: 'ok' }));
    const form = new FormData();
    form.append('file', 'isi');
    await api.upload('/laporan-besar', form);
    const { init, headers } = call();
    expect(init.method).toBe('POST');
    expect(init.body).toBe(form);
    expect(headers['Content-Type']).toBeUndefined();
  });
  it('success: false melempar ApiError dengan pesan server', async () => {
    const { api } = setup(jsonResponse(200, { success: false, message: 'NIM tidak ditemukan' }));
    const err = await api.del('/admin/users/1').catch((e: unknown) => e);
    expect(err).toBeInstanceOf(ApiError);
    expect(err).toMatchObject({ message: 'NIM tidak ditemukan', status: 200 });
  });
  it('status gagal tanpa pesan melempar HTTP <status>', async () => {
    const { api } = setup(jsonResponse(500, { success: false, message: '' }));
    await expect(api.get('/x')).rejects.toMatchObject({ message: 'HTTP 500', status: 500 });
  });
  it('401 memanggil onUnauthorized sekali', async () => {
    const { api, onUnauthorized } = setup(
      jsonResponse(401, { success: false, message: 'Token tidak valid' }),
    );
    await expect(api.get('/profile')).rejects.toMatchObject({
      message: 'Token tidak valid',
      status: 401,
    });
    expect(onUnauthorized).toHaveBeenCalledTimes(1);
  });
  it('401 dengan auth: false tidak memanggil onUnauthorized', async () => {
    const { api, onUnauthorized } = setup(
      jsonResponse(401, { success: false, message: 'Password salah' }),
    );
    await expect(api.post('/auth/login', {}, { auth: false })).rejects.toMatchObject({
      status: 401,
    });
    expect(onUnauthorized).not.toHaveBeenCalled();
  });
  it('response non-JSON melempar HTTP 502', async () => {
    const { api } = setup(new Response('<html>Bad Gateway</html>', { status: 502 }));
    await expect(api.get('/x')).rejects.toMatchObject({ message: 'HTTP 502', status: 502 });
  });
});
