import { createAppQueryClient, createIdbPersister, type AppPersister } from '@algohub/query';

export const queryClient = createAppQueryClient();

let current: { userId: number; persister: AppPersister } | null = null;

export function persisterFor(userId: number): AppPersister {
  if (current?.userId !== userId) {
    current = { userId, persister: createIdbPersister('siakad', userId) };
  }
  return current.persister;
}

// Persister yang sudah ditutup tidak bisa dipakai lagi, jadi login berikutnya harus membuat yang baru.
export function takePersister(): AppPersister | undefined {
  const persister = current?.persister;
  current = null;
  return persister;
}
