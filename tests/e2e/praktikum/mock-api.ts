import { test as base, expect, type Route } from '@playwright/test';
import {
  PASSWORD,
  aktivasi,
  auditLogs,
  dashboard,
  kelas,
  otherUser,
  permissions,
  peserta,
  profile,
  rekapKelas,
  sesiUser,
  statistik,
  tokenFor,
  users,
  type Role,
} from './fixtures';

const API_PREFIX = '/api/praktikum';

export interface ApiCall {
  method: string;
  path: string;
  search: URLSearchParams;
  body: unknown;
}

type Reply = unknown | ((call: ApiCall) => unknown);

interface Handler {
  method: string;
  path: string | RegExp;
  reply: Reply;
  status: number;
}

export class ApiError {
  constructor(
    readonly status: number,
    readonly message: string,
  ) {}
}

export class MockApi {
  readonly calls: ApiCall[] = [];
  readonly unmocked: string[] = [];
  private handlers: Handler[] = [];
  private role: Role = 'mahasiswa';

  constructor() {
    this.defaults();
  }

  /** Handler yang didaftarkan belakangan menang, jadi test bisa menimpa jawaban default. */
  on(method: string, path: string | RegExp, reply: Reply, status = 200) {
    this.handlers.unshift({ method, path, reply, status });
    return this;
  }

  callsTo(method: string, path: string | RegExp) {
    return this.calls.filter((c) => c.method === method && matches(path, c.path));
  }

  async handle(route: Route) {
    const req = route.request();
    const url = new URL(req.url());
    const path = url.pathname.startsWith(API_PREFIX)
      ? url.pathname.slice(API_PREFIX.length)
      : url.pathname;
    const raw = req.postData();
    let body: unknown = raw;
    if (raw && req.headers()['content-type']?.includes('application/json')) body = JSON.parse(raw);
    const call: ApiCall = { method: req.method(), path, search: url.searchParams, body };
    this.calls.push(call);
    const handler = url.pathname.startsWith(API_PREFIX)
      ? this.handlers.find((h) => h.method === call.method && matches(h.path, path))
      : undefined;
    if (!handler) {
      this.unmocked.push(`${call.method} ${url.pathname}${url.search}`);
      return route.fulfill({ status: 500, json: { success: false, message: 'tidak di-mock' } });
    }
    let data: unknown;
    try {
      data = typeof handler.reply === 'function' ? handler.reply(call) : handler.reply;
    } catch (e) {
      if (e instanceof ApiError) {
        return route.fulfill({ status: e.status, json: { success: false, message: e.message } });
      }
      throw e;
    }
    return route.fulfill({ status: handler.status, json: { success: true, message: 'OK', data } });
  }

  private defaults() {
    const ok = null;
    this.on('POST', '/auth/login', ({ body }) => {
      const { identifier, password } = body as { identifier: string; password: string };
      const user = Object.values(users).find((u) => u.nim === identifier);
      if (!user || password !== PASSWORD) throw new ApiError(401, 'NIM/email atau password salah');
      this.role = user.role;
      return { token: tokenFor(user.role), user };
    });
    this.on('POST', '/auth/logout', ok);
    this.on('GET', '/info/announcements', {});
    this.on('GET', '/dashboard', dashboard);
    this.on('GET', '/sesi', sesiUser);
    this.on('GET', '/profile', () => profile(this.role));
    this.on('GET', '/admin/dashboard', statistik);
    this.on('GET', '/admin/kelas', kelas);
    this.on('GET', '/admin/jadwal', [
      {
        id: 1,
        kelas_id: 1,
        shift: 1,
        hari: 'Senin',
        jam_mulai: '08:00:00',
        jam_selesai: '10:00:00',
        keterangan: '',
        kelas: kelas[0],
      },
    ]);
    this.on('GET', '/admin/konfigurasi', [{ key: 'jadwal_mode', value: 'tabel' }]);
    this.on('GET', '/admin/konfigurasi/ai-model', {
      effective: 'qwen',
      override: '',
      default: 'qwen',
    });
    this.on('GET', '/admin/asisten', [users.asisten]);
    this.on('GET', '/admin/ampuan', []);
    this.on('GET', '/admin/pedoman', [
      { id: 1, nama_dokumen: 'Pedoman', file_url: 'https://example.test/p.pdf' },
    ]);
    this.on('GET', '/admin/sesi', [
      { id: 1, judul_sesi: 'Sesi 1', deskripsi: '', urutan: 1, is_ujian_praktik: false },
    ]);
    this.on('GET', '/admin/aktivasi', [aktivasi]);
    this.on('GET', /^\/admin\/aktivasi\/\d+$/, aktivasi);
    this.on('GET', /^\/admin\/aktivasi\/\d+\/susulan$/, []);
    this.on('GET', /^\/admin\/aktivasi\/\d+\/peserta$/, peserta);
    this.on('GET', '/admin/users', [users.mahasiswa, otherUser]);
    this.on('GET', '/admin/laporan-besar', null);
    this.on('GET', '/admin/audit-logs', auditLogs);
    this.on('GET', /^\/admin\/rekap\/kelas\/\d+$/, rekapKelas(3));
    this.on('GET', '/admin/rekap-jawaban', { items: [], total: 0 });
    this.on('GET', '/superadmin/permissions', permissions);
  }
}

function matches(path: string | RegExp, actual: string) {
  return typeof path === 'string' ? path === actual : path.test(actual);
}

export const test = base.extend<{ api: MockApi }>({
  api: [
    async ({ page, context }, use) => {
      const api = new MockApi();
      const prod: string[] = [];
      const pageErrors: string[] = [];
      await context.route(/api\.algohub\.web\.id/, (route) => {
        prod.push(route.request().url());
        return route.abort();
      });
      await page.route('**/api/**', (route) => api.handle(route));
      page.on('pageerror', (e) => pageErrors.push(e.message));
      await use(api);
      expect(api.unmocked, 'request API tak ter-mock').toEqual([]);
      expect(prod, 'request ke api.algohub.web.id').toEqual([]);
      expect(pageErrors, 'pageerror').toEqual([]);
    },
    { auto: true },
  ],
});

export { expect };
