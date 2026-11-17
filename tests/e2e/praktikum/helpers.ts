import { expect, type Page } from '@playwright/test';
import { PASSWORD, users, type Role } from './fixtures';

const HOME: Record<Role, RegExp> = {
  mahasiswa: /\/praktikum\/dashboard$/,
  asisten: /\/praktikum\/admin$/,
  koordinator: /\/praktikum\/admin$/,
};

export async function loginAs(page: Page, role: Role) {
  await page.goto('/praktikum/login');
  await page.locator('#identifier').fill(users[role].nim);
  await page.locator('#pw').fill(PASSWORD);
  await page.getByRole('button', { name: /Masuk ke Portal/ }).click();
  await expect(page).toHaveURL(HOME[role]);
}

export async function logout(page: Page) {
  await page.getByRole('button', { name: 'Logout' }).click();
  await expect(page).toHaveURL(/\/praktikum\/login$/);
}

// Persister menunda tulis berikutnya 1 detik (throttle), jadi isi IndexedDB baru final setelahnya.
export const waitForPersist = (page: Page) => page.waitForTimeout(1_500);

export interface IdbEntry {
  db: string;
  store: string;
  key: string;
  value: string;
}

export function readAlgohubIdb(page: Page): Promise<IdbEntry[]> {
  return page.evaluate(async () => {
    const promised = <T>(req: IDBRequest<T>) =>
      new Promise<T>((resolve, reject) => {
        req.onsuccess = () => resolve(req.result);
        req.onerror = () => reject(req.error);
      });
    const entries: IdbEntry[] = [];
    for (const info of await indexedDB.databases()) {
      if (!info.name?.startsWith('algohub-')) continue;
      const db = await promised(indexedDB.open(info.name));
      for (const store of Array.from(db.objectStoreNames)) {
        const os = db.transaction(store, 'readonly').objectStore(store);
        const [keys, values] = await Promise.all([
          promised(os.getAllKeys()),
          promised(os.getAll()),
        ]);
        keys.forEach((key, i) =>
          entries.push({
            db: info.name!,
            store,
            key: String(key),
            value: typeof values[i] === 'string' ? values[i] : JSON.stringify(values[i]),
          }),
        );
      }
      db.close();
    }
    return entries;
  });
}

/** Query key yang tersimpan di cache persist, mis. ['praktikum','profile']. */
export function persistedQueryKeys(entries: IdbEntry[]): string[] {
  return entries.flatMap((e) => {
    const client = JSON.parse(e.value) as {
      clientState?: { queries?: { queryKey: unknown[] }[] };
    };
    return (client.clientState?.queries ?? []).map((q) => JSON.stringify(q.queryKey));
  });
}
