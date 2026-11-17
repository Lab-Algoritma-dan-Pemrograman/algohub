import type { Page, Route } from '@playwright/test';
import { claimsOf, USERS, type Role } from '../fixtures/users.ts';
import { defaultRpc } from '../fixtures/rpc.ts';

const API = 'http://api.siakad.test';
const APP_ORIGIN = 'http://localhost:5174';

export type RpcParams = Record<string, unknown>;
export type RpcReply =
  | ((params: RpcParams) => unknown)
  | Record<string, unknown>
  | unknown[]
  | string
  | number
  | boolean
  | null;

export interface RpcCall {
  fn: string;
  params: RpcParams;
}

const cors = {
  'Access-Control-Allow-Origin': APP_ORIGIN,
  'Access-Control-Allow-Credentials': 'true',
  'Access-Control-Allow-Headers': 'Content-Type, Authorization',
  'Access-Control-Allow-Methods': 'GET, POST, OPTIONS',
};

// Mock backend Go siakad: bentuk respons mengikuti handler/siakad dan rpc_repository.go
// (setof → array, scalar → nilai, void → null; error { error, details? }).
export class MockApi {
  readonly calls: RpcCall[] = [];
  readonly endpoints: { path: string; body: unknown }[] = [];
  readonly rpc: Record<string, RpcReply>;
  role: Role;
  loggedIn = false;

  constructor(role: Role, overrides: Record<string, RpcReply> = {}) {
    this.role = role;
    this.rpc = { ...defaultRpc(role), ...overrides };
  }

  callsTo(fn: string): RpcParams[] {
    return this.calls.filter((c) => c.fn === fn).map((c) => c.params);
  }

  async install(page: Page): Promise<void> {
    await page.route(`${API}/**`, (route) => this.handle(route));
  }

  private json(route: Route, status: number, body: unknown) {
    return route.fulfill({
      status,
      headers: { ...cors, 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    });
  }

  private async handle(route: Route) {
    const req = route.request();
    if (req.method() === 'OPTIONS') return route.fulfill({ status: 204, headers: cors });
    const url = new URL(req.url());
    const path = url.pathname.replace(/^\/api\/siakad/, '');
    const body = req.postData() ? JSON.parse(req.postData() as string) : undefined;
    if (path === '/rpc') {
      const fn = url.searchParams.get('fn') ?? '';
      const params = (body ?? {}) as RpcParams;
      this.calls.push({ fn, params });
      const reply = fn in this.rpc ? this.rpc[fn] : [];
      const data =
        typeof reply === 'function' ? (reply as (p: RpcParams) => unknown)(params) : reply;
      if (data instanceof RpcFailure) return this.json(route, data.status, { error: data.message });
      return this.json(route, 200, data ?? null);
    }
    this.endpoints.push({ path, body });
    switch (path) {
      case '/auth/login': {
        const user = USERS[this.role];
        this.loggedIn = true;
        return this.json(route, 200, { user, token: `token-${this.role}` });
      }
      case '/auth/logout':
        this.loggedIn = false;
        return route.fulfill({ status: 204, headers: cors });
      case '/auth/verify':
        if (!this.loggedIn) return this.json(route, 401, { error: 'Unauthorized: Missing token' });
        return this.json(route, 200, { user: claimsOf(this.role) });
      default: {
        const reply = this.rpc[`endpoint:${path}`];
        const data = typeof reply === 'function' ? (reply as (p: unknown) => unknown)(body) : reply;
        return this.json(route, 200, data ?? {});
      }
    }
  }
}

export class RpcFailure {
  constructor(
    readonly message: string,
    readonly status = 400,
  ) {}
}
