import { createAsyncStoragePersister } from '@tanstack/query-async-storage-persister';
import { QueryClient, type Query } from '@tanstack/svelte-query';
import type { Persister, PersistQueryClientOptions } from '@tanstack/svelte-query-persist-client';
import { createStore, del, get, set } from 'idb-keyval';
const DAY = 86_400_000;
function isClientError(error: unknown): boolean {
  const status = (error as { status?: unknown } | null)?.status;
  return typeof status === 'number' && status >= 400 && status < 500;
}
export function createAppQueryClient(): QueryClient {
  return new QueryClient({
    defaultOptions: {
      queries: {
        staleTime: 30_000,
        gcTime: DAY,
        refetchOnWindowFocus: false,
        retry: (failureCount, error) => failureCount < 1 && !isClientError(error),
      },
    },
  });
}
export function shouldPersistQuery(query: Query): boolean {
  return query.state.status === 'success' && query.meta?.persist === true;
}
export interface AppPersister extends Persister {
  close(): void;
}
export function createIdbPersister(app: string, userId: number | string): AppPersister {
  const store = createStore(`algohub-${app}`, 'query-cache');
  let closed = false;
  const persister = createAsyncStoragePersister({
    key: `algohub-${app}-${userId}`,
    storage: {
      getItem: (key) => get<string>(key, store),
      // Persister menunda tulis (throttle 1 detik); setelah logout tulisan tertunda itu
      // tidak boleh mengisi ulang IndexedDB yang baru dihapus.
      setItem: async (key, value: string) => {
        if (!closed) await set(key, value, store);
      },
      removeItem: (key) => del(key, store),
    },
  });
  return {
    ...persister,
    close: () => {
      closed = true;
    },
  };
}
export function persistOptions(
  persister: Persister,
  buster: string,
): Omit<PersistQueryClientOptions, 'queryClient'> {
  return {
    persister,
    maxAge: DAY,
    buster,
    dehydrateOptions: { shouldDehydrateQuery: shouldPersistQuery },
  };
}
