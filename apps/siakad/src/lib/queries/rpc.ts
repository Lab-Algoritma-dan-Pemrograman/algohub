import { ApiError } from '@algohub/api-client';
import { siakadKeys, type RpcResult } from '@algohub/api-client/siakad';
import {
  createMutation,
  mutationOptions,
  queryOptions,
  type CreateQueryOptions,
} from '@tanstack/svelte-query';
import { rpc } from '#lib/api.ts';
import { queryClient } from '#lib/query.ts';

// Hanya data referensi yang tidak rahasia (03-frontend.md § Persist): pengaturan publik,
// jadwal/kelas umum, tautan, dan katalog barang. Data user lain, absensi, audit log, dan
// keuangan tidak boleh tersimpan di IndexedDB PC lab yang dipakai bergantian.
export const PERSISTED_RPCS: ReadonlySet<string> = new Set([
  'get_public_settings',
  'get_schedules_secure',
  'get_external_links_secure',
  'get_renter_items_secure',
]);

export async function rpcData<T>(fn: string, params?: object): Promise<T> {
  const { data, error } = await rpc<T>(fn, params);
  if (error) throw new ApiError(error.message, error.status ?? 0);
  return data as T;
}

type RpcKey = ReturnType<typeof siakadKeys.rpc>;

export type RpcQueryExtra<T> = Pick<
  CreateQueryOptions<T, Error, T, RpcKey>,
  | 'enabled'
  | 'staleTime'
  | 'refetchInterval'
  | 'refetchIntervalInBackground'
  | 'placeholderData'
  | 'retry'
>;

export function rpcQueryOptions<T>(fn: string, params?: object, extra?: RpcQueryExtra<T>) {
  return queryOptions<T, Error, T, RpcKey>({
    ...extra,
    queryKey: siakadKeys.rpc(fn, params),
    queryFn: () => rpcData<T>(fn, params),
    meta: { persist: PERSISTED_RPCS.has(fn) },
  });
}

// Sama dengan .maybeSingle() Proxy lama: array kosong tetap dikembalikan apa adanya.
export function maybeSingle<T>(data: T[] | T | null | undefined): T | T[] | null {
  return Array.isArray(data) && data.length > 0 ? data[0] : data || null;
}

// Pembacaan sekali jalan di tengah alur tulis (misal cek username sebelum daftar) selalu segar.
export function fetchRpc<T>(fn: string, params?: object): Promise<T> {
  return queryClient.fetchQuery({ ...rpcQueryOptions<T>(fn, params), staleTime: 0 });
}

export function invalidateRpcs(fns: readonly string[]): Promise<void[]> {
  return Promise.all(
    fns.map((f) => queryClient.invalidateQueries({ queryKey: siakadKeys.rpcFn(f) })),
  );
}

// Mutation mengembalikan { data, error } apa adanya karena halaman lama membaca keduanya
// (misal pesan error dari DB untuk toast) alih-alih menangkap exception.
export function rpcMutation<T = unknown, P extends object = Record<string, unknown>>(
  fn: string,
  { invalidates = [] }: { invalidates?: readonly string[] } = {},
) {
  return mutationOptions<RpcResult<T>, Error, P | undefined>({
    mutationKey: siakadKeys.rpcFn(fn),
    mutationFn: async (params) => await rpc<T>(fn, params),
    onSuccess: async (result) => {
      if (!result.error && invalidates.length) await invalidateRpcs(invalidates);
    },
  });
}

export function createRpcMutation<T = unknown, P extends object = Record<string, unknown>>(
  fn: string,
  invalidates: readonly string[] = [],
) {
  return createMutation(() => rpcMutation<T, P>(fn, { invalidates }));
}
