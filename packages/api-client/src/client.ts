export interface Envelope<T = unknown> {
  success: boolean;
  message: string;
  data?: T;
  error?: unknown;
}
export class ApiError extends Error {
  status: number;
  constructor(message: string, status: number) {
    super(message);
    this.status = status;
  }
}
export interface RequestOptions {
  auth?: boolean;
}
export interface ApiOptions {
  baseUrl: string;
  prefix?: string;
  getToken: () => string | null;
  onUnauthorized?: () => void;
  credentials?: RequestCredentials;
}
export function createApi({
  baseUrl,
  prefix = '',
  getToken,
  onUnauthorized,
  credentials = 'same-origin',
}: ApiOptions) {
  async function request<T>(
    method: string,
    path: string,
    body: unknown,
    { auth = true }: RequestOptions = {},
  ): Promise<T> {
    const headers: Record<string, string> = {};
    const token = auth ? getToken() : null;
    if (token) headers['Authorization'] = `Bearer ${token}`;
    let payload: BodyInit | undefined;
    if (body instanceof FormData) {
      payload = body;
    } else if (body !== undefined) {
      headers['Content-Type'] = 'application/json';
      payload = JSON.stringify(body);
    }
    const res = await fetch(`${baseUrl}${prefix}${path}`, {
      method,
      headers,
      body: payload,
      credentials,
    });
    let json: Envelope<T>;
    try {
      json = await res.json();
    } catch {
      throw new ApiError(`HTTP ${res.status}`, res.status);
    }
    if (!res.ok || !json.success) {
      if (res.status === 401 && auth) onUnauthorized?.();
      throw new ApiError(json.message || `HTTP ${res.status}`, res.status);
    }
    return json.data as T;
  }
  return {
    get: <T>(path: string, opts?: RequestOptions) => request<T>('GET', path, undefined, opts),
    post: <T>(path: string, body?: unknown, opts?: RequestOptions) =>
      request<T>('POST', path, body, opts),
    put: <T>(path: string, body?: unknown, opts?: RequestOptions) =>
      request<T>('PUT', path, body, opts),
    patch: <T>(path: string, body?: unknown, opts?: RequestOptions) =>
      request<T>('PATCH', path, body, opts),
    del: <T>(path: string, opts?: RequestOptions) => request<T>('DELETE', path, undefined, opts),
    upload: <T>(path: string, form: FormData, opts?: RequestOptions) =>
      request<T>('POST', path, form, opts),
  };
}
export type Api = ReturnType<typeof createApi>;
