import { afterEach, describe, expect, it, vi } from 'vitest';
import { createSiakadApi } from './client';

function jsonResponse(status: number, body: unknown): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { 'Content-Type': 'application/json' },
  });
}

function setup(response: Response | Error, token: string | null = 'tok') {
  const fetchMock = vi.fn(async (_url: string, _init: RequestInit) => {
    if (response instanceof Error) throw response;
    return response;
  });
  vi.stubGlobal('fetch', fetchMock);
  const api = createSiakadApi({ baseUrl: 'http://api.test', getToken: () => token });
  const call = () => {
    const [url, init] = fetchMock.mock.calls[0];
    return { url, init, headers: init.headers as Record<string, string> };
  };
  return { api, call };
}

afterEach(() => {
  vi.unstubAllGlobals();
});

describe('rpc', () => {
  it('POST ke gateway dengan nama fungsi ter-encode, body {}, Bearer, dan credentials include', async () => {
    const { api, call } = setup(jsonResponse(200, [{ id: 1 }]));
    await expect(api.rpc('get_users_secure')).resolves.toEqual({ data: [{ id: 1 }], error: null });
    const { url, init, headers } = call();
    expect(url).toBe('http://api.test/api/siakad/rpc?fn=get_users_secure');
    expect(init.method).toBe('POST');
    expect(init.credentials).toBe('include');
    expect(init.body).toBe('{}');
    expect(headers['Content-Type']).toBe('application/json');
    expect(headers['Authorization']).toBe('Bearer tok');
  });

  it('mengirim params sebagai JSON dan meng-encode nama fungsi', async () => {
    const { api, call } = setup(jsonResponse(200, null));
    await api.rpc('a&b', { p_id: 2 });
    expect(call().url).toBe('http://api.test/api/siakad/rpc?fn=a%26b');
    expect(call().init.body).toBe('{"p_id":2}');
  });

  it('tanpa token tidak mengirim Authorization', async () => {
    const { api, call } = setup(jsonResponse(200, null), null);
    await api.rpc('get_public_settings');
    expect(call().headers['Authorization']).toBeUndefined();
  });

  it('status gagal dengan body {error} memakai pesan server dan status', async () => {
    const { api } = setup(jsonResponse(403, { error: 'Forbidden' }));
    await expect(api.rpc('x')).resolves.toEqual({
      data: null,
      error: { message: 'Forbidden', status: 403 },
    });
  });

  it('status gagal dengan body teks memakai teks itu', async () => {
    const { api } = setup(new Response('Bad Gateway', { status: 502 }));
    await expect(api.rpc('x')).resolves.toEqual({
      data: null,
      error: { message: 'Bad Gateway', status: 502 },
    });
  });

  it('status gagal tanpa body memakai pesan bawaan', async () => {
    const { api } = setup(new Response('', { status: 500 }));
    await expect(api.rpc('x')).resolves.toEqual({
      data: null,
      error: { message: 'API Proxy error', status: 500 },
    });
  });

  it('200 dengan objek {error} dianggap gagal tanpa status', async () => {
    const { api } = setup(jsonResponse(200, { error: 'Username sudah dipakai' }));
    await expect(api.rpc('register_user')).resolves.toEqual({
      data: null,
      error: { message: 'Username sudah dipakai' },
    });
  });

  it('fetch gagal mengembalikan pesan error atau Network error', async () => {
    await expect(setup(new Error('offline')).api.rpc('x')).resolves.toEqual({
      data: null,
      error: { message: 'offline' },
    });
    await expect(setup(new Error('')).api.rpc('x')).resolves.toEqual({
      data: null,
      error: { message: 'Network error' },
    });
  });

  it.each([
    [[{ id: 1 }, { id: 2 }], { id: 1 }, { id: 1 }],
    // Proxy lama: [] truthy, jadi maybeSingle() mengembalikan [] apa adanya.
    [[], undefined, []],
    [{ id: 3 }, { id: 3 }, { id: 3 }],
    [null, null, null],
  ])('single/maybeSingle untuk %j', async (body, single, maybeSingle) => {
    const first = setup(jsonResponse(200, body)).api;
    expect((await first.rpc('x').single()).data).toEqual(single);
    const second = setup(jsonResponse(200, body)).api;
    expect((await second.rpc('x').maybeSingle()).data).toEqual(maybeSingle);
  });
});

describe('endpoint lain', () => {
  it('logout 204 mengembalikan body null', async () => {
    const { api, call } = setup(new Response(null, { status: 204 }));
    await expect(api.logout()).resolves.toEqual({ ok: true, status: 204, body: null });
    expect(call().url).toBe('http://api.test/api/siakad/auth/logout');
    expect(call().init.method).toBe('POST');
    expect(call().init.credentials).toBe('include');
  });

  it('verify 401 tidak melempar dan membawa body error', async () => {
    const { api, call } = setup(jsonResponse(401, { error: 'Unauthorized: Missing token' }));
    await expect(api.verify()).resolves.toEqual({
      ok: false,
      status: 401,
      body: { error: 'Unauthorized: Missing token' },
    });
    expect(call().init.method).toBe('GET');
  });

  it('body bukan JSON menjadi null', async () => {
    const { api } = setup(new Response('<html>', { status: 502 }));
    await expect(api.syncElearning()).resolves.toEqual({ ok: false, status: 502, body: null });
  });

  it.each([
    ['login', '/auth/login', { username: 'u', password: 'p' }],
    ['generateJwt', '/generate-jwt', { handshake_code: 'h' }],
    [
      'analyzeSchedule',
      '/analyze-schedule',
      { fileBase64: 'x', mimeType: 'image/png', model: 'm' },
    ],
    ['sendPush', '/send-push', { targetUserId: 1, title: 't', body: 'b', url: '/u' }],
    ['sendWa', '/send-wa', { targetUserId: 1, targetPhone: '62', message: 'm' }],
  ] as const)('%s POST ke %s dengan body JSON', async (name, path, body) => {
    const { api, call } = setup(jsonResponse(200, { ok: true }));
    const fn = api[name] as (b: unknown) => Promise<unknown>;
    await fn(body);
    expect(call().url).toBe(`http://api.test/api/siakad${path}`);
    expect(call().init.method).toBe('POST');
    expect(call().init.body).toBe(JSON.stringify(body));
    expect(call().headers['Authorization']).toBe('Bearer tok');
  });

  it('syncElearning POST tanpa body', async () => {
    const { api, call } = setup(jsonResponse(200, {}));
    await api.syncElearning();
    expect(call().url).toBe('http://api.test/api/siakad/sync-elearning');
    expect(call().init.body).toBeUndefined();
  });
});
