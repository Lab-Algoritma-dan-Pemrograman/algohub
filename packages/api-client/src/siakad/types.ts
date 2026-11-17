export interface SiakadUser {
  id: number;
  username: string;
  full_name: string;
  role: string;
  nim?: string;
  assistant_code?: string;
  division?: string;
  [key: string]: unknown;
}

export interface LoginBody {
  username: string;
  password: string;
}

export interface LoginResponse {
  user: Record<string, unknown> & { id: number | string; role: string };
  token: string;
}

export interface VerifyResponse {
  user: Record<string, unknown> & { id: number | string; role: string };
}

export interface SiakadErrorBody {
  error: string;
  details?: string;
}

export interface RpcError {
  message: string;
  status?: number;
}

export interface RpcResult<T> {
  data: T | null;
  error: RpcError | null;
}

export interface RpcCall<T> extends PromiseLike<RpcResult<T>> {
  single<U = T extends readonly (infer E)[] ? E : T>(): RpcCall<U>;
  maybeSingle<U = T extends readonly (infer E)[] ? E : T>(): RpcCall<U>;
}

export interface RawResponse<T> {
  ok: boolean;
  status: number;
  body: T | null;
}

export interface GenerateJwtBody {
  handshake_code: string;
}

export interface AnalyzeScheduleBody {
  fileBase64: string;
  mimeType: string;
  model: string;
}

export interface SendPushBody {
  targetUserId: number | string;
  title: string;
  body: string;
  url: string;
}

export interface SendWaBody {
  targetUserId?: number | string;
  targetPhone?: string;
  message: string;
}
