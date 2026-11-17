import { createMutation, type QueryKey } from '@tanstack/svelte-query';
import { api } from '#lib/api.ts';
import { invalidate } from '#lib/query.ts';

// Pola halaman admin lama: id terisi = PUT /<path>/<id>, kosong = POST /<path>.
export function createSaveMutation<B>(path: string, ...keys: QueryKey[]) {
  return createMutation(() => ({
    mutationFn: ({ id, body }: { id: number | null; body: B }) =>
      id ? api.put<unknown>(`${path}/${id}`, body) : api.post<unknown>(path, body),
    onSuccess: () => invalidate(...keys),
  }));
}

export function createDeleteMutation(path: string, ...keys: QueryKey[]) {
  return createMutation(() => ({
    mutationFn: (id: number) => api.del<unknown>(`${path}/${id}`),
    onSuccess: () => invalidate(...keys),
  }));
}

export function createPostMutation<B, R = unknown>(path: string, ...keys: QueryKey[]) {
  return createMutation(() => ({
    mutationFn: (body: B) => api.post<R>(path, body),
    onSuccess: () => invalidate(...keys),
  }));
}

export const orEmpty = <T>(list: T[] | null | undefined): T[] => list ?? [];
