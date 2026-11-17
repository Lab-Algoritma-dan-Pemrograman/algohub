import type {
  AnalyzeScheduleBody,
  GenerateJwtBody,
  LoginBody,
  LoginResponse,
  RawResponse,
  RpcCall,
  RpcResult,
  SendPushBody,
  SendWaBody,
  VerifyResponse,
} from './types';

export interface SiakadApiOptions {
  baseUrl: string;
  getToken: () => string | null;
}

const isObject = (value: unknown): value is Record<string, unknown> =>
  !!value && typeof value === 'object';

function thenable<T>(promise: Promise<RpcResult<T>>): RpcCall<T> {
  return {
    then: (onFulfilled, onRejected) => promise.then(onFulfilled, onRejected),
    single: <U>() =>
      thenable<U>(
        promise.then((res) => ({
          ...res,
          data: (Array.isArray(res.data) ? res.data[0] : res.data) as U | null,
        })),
      ),
    maybeSingle: <U>() =>
      thenable<U>(
        promise.then((res) => ({
          ...res,
          data: (Array.isArray(res.data) && res.data.length > 0
            ? res.data[0]
            : res.data || null) as U | null,
        })),
      ),
  };
}

export function createSiakadApi({ baseUrl, getToken }: SiakadApiOptions) {
  function headers(): Record<string, string> {
    const result: Record<string, string> = { 'Content-Type': 'application/json' };
    const token = getToken();
    if (token) result['Authorization'] = `Bearer ${token}`;
    return result;
  }

  async function callRpc<T>(fn: string, params?: object): Promise<RpcResult<T>> {
    try {
      const res = await fetch(`${baseUrl}/api/siakad/rpc?fn=${encodeURIComponent(fn)}`, {
        method: 'POST',
        credentials: 'include',
        headers: headers(),
        body: JSON.stringify(params || {}),
      });
      const body: unknown = res.headers.get('content-type')?.includes('application/json')
        ? await res.json().catch(() => null)
        : await res.text();
      if (!res.ok) {
        const message = isObject(body) && 'error' in body ? body.error : body || 'API Proxy error';
        return { data: null, error: { message: String(message), status: res.status } };
      }
      // Fungsi DB tertentu mengembalikan objek { error } dengan status 200.
      if (isObject(body) && 'error' in body) {
        return { data: null, error: { message: String(body.error) } };
      }
      return { data: body as T, error: null };
    } catch (err) {
      return { data: null, error: { message: (err as Error)?.message || 'Network error' } };
    }
  }

  async function raw<T>(method: string, path: string, body?: unknown): Promise<RawResponse<T>> {
    const res = await fetch(`${baseUrl}/api/siakad${path}`, {
      method,
      credentials: 'include',
      headers: headers(),
      body: body === undefined ? undefined : JSON.stringify(body),
    });
    const text = await res.text();
    let parsed: T | null = null;
    if (text) {
      try {
        parsed = JSON.parse(text) as T;
      } catch {
        parsed = null;
      }
    }
    return { ok: res.ok, status: res.status, body: parsed };
  }

  return {
    rpc: <T = unknown>(fn: string, params?: object): RpcCall<T> => thenable(callRpc<T>(fn, params)),
    login: (body: LoginBody) => raw<LoginResponse>('POST', '/auth/login', body),
    logout: () => raw<null>('POST', '/auth/logout'),
    verify: () => raw<VerifyResponse>('GET', '/auth/verify'),
    generateJwt: <T = Record<string, unknown>>(body: GenerateJwtBody) =>
      raw<T>('POST', '/generate-jwt', body),
    analyzeSchedule: <T = Record<string, unknown>>(body: AnalyzeScheduleBody) =>
      raw<T>('POST', '/analyze-schedule', body),
    sendPush: <T = Record<string, unknown>>(body: SendPushBody) =>
      raw<T>('POST', '/send-push', body),
    sendWa: <T = Record<string, unknown>>(body: SendWaBody) => raw<T>('POST', '/send-wa', body),
    syncElearning: <T = Record<string, unknown>>() => raw<T>('POST', '/sync-elearning'),
  };
}

export type SiakadApi = ReturnType<typeof createSiakadApi>;
