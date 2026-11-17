import { createAppQueryClient, createIdbPersister, type AppPersister } from '@algohub/query';
import type { QueryKey } from '@tanstack/svelte-query';

export const queryClient = createAppQueryClient();

export function invalidate(...queryKeys: QueryKey[]): Promise<unknown> {
  return Promise.all(queryKeys.map((queryKey) => queryClient.invalidateQueries({ queryKey })));
}

let current: { userId: number; persister: AppPersister } | null = null;

export function persisterFor(userId: number): AppPersister {
  if (current?.userId !== userId) {
    current = { userId, persister: createIdbPersister('praktikum', userId) };
  }
  return current.persister;
}

// Persister yang sudah ditutup tidak bisa dipakai lagi, jadi login berikutnya harus membuat yang baru.
export function takePersister(): AppPersister | undefined {
  const persister = current?.persister;
  current = null;
  return persister;
}
