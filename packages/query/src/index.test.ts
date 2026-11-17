import 'fake-indexeddb/auto';
import { dehydrate, QueryClient } from '@tanstack/svelte-query';
import { createStore, get, keys } from 'idb-keyval';
import { describe, expect, it } from 'vitest';
import {
  createAppQueryClient,
  createIdbPersister,
  persistOptions,
  shouldPersistQuery,
} from './index';
function snapshot(client: QueryClient) {
  return { timestamp: Date.now(), buster: 'v1', clientState: dehydrate(client) };
}
async function clientWithPersistedQuery() {
  const client = new QueryClient();
  await client.fetchQuery({
    queryKey: ['praktikum', 'info', 'modul'],
    queryFn: () => [{ id: 1 }],
    meta: { persist: true },
  });
  return client;
}
describe('createAppQueryClient', () => {
  it('memakai opsi default bersama', () => {
    const queries = createAppQueryClient().getDefaultOptions().queries;
    expect(queries).toMatchObject({
      staleTime: 30_000,
      gcTime: 86_400_000,
      refetchOnWindowFocus: false,
    });
  });
  it('retry sekali untuk error server dan tidak untuk error klien', () => {
    const retry = createAppQueryClient().getDefaultOptions().queries?.retry as (
      failureCount: number,
      error: unknown,
    ) => boolean;
    expect(retry(0, { status: 404 })).toBe(false);
    expect(retry(0, { status: 400 })).toBe(false);
    expect(retry(0, { status: 500 })).toBe(true);
    expect(retry(1, { status: 500 })).toBe(false);
    expect(retry(0, new Error('jaringan'))).toBe(true);
  });
});
describe('shouldPersistQuery', () => {
  it('hanya true untuk query sukses dengan meta.persist === true', async () => {
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    await client.fetchQuery({ queryKey: ['ok'], queryFn: () => 1, meta: { persist: true } });
    await client.fetchQuery({ queryKey: ['tanpa-meta'], queryFn: () => 1 });
    await client.fetchQuery({ queryKey: ['meta-lain'], queryFn: () => 1, meta: { persist: 1 } });
    await client
      .fetchQuery({
        queryKey: ['gagal'],
        queryFn: () => Promise.reject(new Error('x')),
        meta: { persist: true },
      })
      .catch(() => undefined);
    const cache = client.getQueryCache();
    cache.build(client, { queryKey: ['pending'], meta: { persist: true } });
    const result = Object.fromEntries(
      cache.getAll().map((query) => [query.queryKey[0], shouldPersistQuery(query)]),
    );
    expect(result).toEqual({
      ok: true,
      'tanpa-meta': false,
      'meta-lain': false,
      gagal: false,
      pending: false,
    });
  });
  it('dipakai persistOptions sebagai filter dehydrate', () => {
    const persister = createIdbPersister('test-options', 1);
    expect(persistOptions(persister, 'v1')).toEqual({
      persister,
      maxAge: 86_400_000,
      buster: 'v1',
      dehydrateOptions: { shouldDehydrateQuery: shouldPersistQuery },
    });
  });
});
describe('createIdbPersister', () => {
  it('persist lalu restore menghasilkan data yang sama, removeClient menghapus key', async () => {
    const persister = createIdbPersister('test-roundtrip', 7);
    const store = createStore('algohub-test-roundtrip', 'query-cache');
    const persisted = snapshot(await clientWithPersistedQuery());
    await persister.persistClient(persisted);
    expect(await keys(store)).toEqual(['algohub-test-roundtrip-7']);
    expect(await persister.restoreClient()).toEqual(persisted);
    await persister.removeClient();
    expect(await get('algohub-test-roundtrip-7', store)).toBeUndefined();
  });
  it('setelah close(), persistClient tidak menulis apa pun ke IndexedDB', async () => {
    const persister = createIdbPersister('test-close', 7);
    const store = createStore('algohub-test-close', 'query-cache');
    const persisted = snapshot(await clientWithPersistedQuery());
    await persister.persistClient(persisted);
    const throttled = persister.persistClient(persisted);
    persister.close();
    await persister.removeClient();
    await throttled;
    await persister.persistClient(persisted);
    expect(await keys(store)).toEqual([]);
  });
});
