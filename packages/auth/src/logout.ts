export interface LogoutOptions {
  queryClient: { clear(): void };
  persister?: { close(): void; removeClient(): void | PromiseLike<void> };
  session: { clear(): void };
}
// Urutan wajib: persister ditutup dulu agar tulis tertunda tidak mengisi ulang IndexedDB
// setelah cache memori dan data tersimpan dihapus.
export async function logout({ queryClient, persister, session }: LogoutOptions): Promise<void> {
  persister?.close();
  queryClient.clear();
  await persister?.removeClient();
  session.clear();
}
