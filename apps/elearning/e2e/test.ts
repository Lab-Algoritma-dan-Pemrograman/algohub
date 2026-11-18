import { test as base, expect, type Page } from '@playwright/test';
import { PASSWORD, users, type Role } from './fixtures.ts';
import { MockApi } from './mock-api.ts';
import { MockRealtime } from './mock-realtime.ts';

export const ORIGIN = 'http://127.0.0.1:5175';

const FONT_CDN = /^https:\/\/(fonts\.googleapis\.com|fonts\.gstatic\.com)\//;
// Monaco dan Pyodide dimuat dari CDN publik; hanya spec compiler yang boleh menjangkaunya.
const PUBLIC_CDN = /^https:\/\/(cdn|gcore|fastly)\.jsdelivr\.net\//;

const PROD_HOST = /\/\/([\w-]+\.)*algohub\.web\.id|supabase\.co/;

interface Fixtures {
  api: MockApi;
  realtime: MockRealtime;
  // true: file clang dari ALGOHUB_COMPILER_DIR dan Pyodide dari jsDelivr benar-benar dimuat.
  realCompiler: boolean;
}

export const test = base.extend<Fixtures>({
  realCompiler: [false, { option: true }],
  realtime: [
    async ({ page }, use) => {
      const realtime = new MockRealtime();
      await realtime.install(page);
      await use(realtime);
    },
    { auto: true },
  ],
  api: [
    async ({ page, realCompiler }, use) => {
      const api = new MockApi();
      const prod: string[] = [];
      const pageErrors: string[] = [];
      // Request yang tidak dijawab mock tidak boleh keluar jaringan.
      await page.route('**/*', (route) => {
        const url = route.request().url();
        if (PROD_HOST.test(url)) prod.push(url);
        if (url.startsWith(ORIGIN)) {
          const { pathname } = new URL(url);
          if (pathname.startsWith('/compiler/') && !realCompiler) return route.abort();
          return route.continue();
        }
        if (FONT_CDN.test(url)) return route.fulfill({ status: 200, body: '' });
        if (realCompiler && PUBLIC_CDN.test(url)) return route.continue();
        return route.abort('blockedbyclient');
      });
      await page.route('**/api/elearning/**', (route) => api.handle(route));
      // Loader Monaco menolak Promise dengan Event saat CDN diblokir; itu bukan galat aplikasi.
      page.on('pageerror', (e) => {
        if (!realCompiler && e.message === 'Event') return;
        pageErrors.push(e.message);
      });
      await use(api);
      expect(api.unmocked, 'request API tak ter-mock').toEqual([]);
      expect(prod, 'request ke domain prod').toEqual([]);
      expect(pageErrors, 'pageerror').toEqual([]);
    },
    { auto: true },
  ],
});

export { expect };

export async function loginAs(page: Page, role: Role) {
  await page.goto('/login');
  await page.locator('#identifier').fill(users[role].nim);
  await page.locator('#pw').fill(PASSWORD);
  await page.getByRole('button', { name: /Masuk ke E-Learning/ }).click();
  await expect(page).toHaveURL(/\/dashboard$/);
}

// Persister menunda tulis berikutnya sekitar 1 detik (throttle), jadi isi IndexedDB baru final setelahnya.
export const waitForPersist = (page: Page) => page.waitForTimeout(1_500);

export interface IdbEntry {
  db: string;
  store: string;
  key: string;
  value: string;
}

// Isi semua database IndexedDB algohub-* sebagai teks; database tanpa entri tidak menambah apa pun.
export function readAlgohubIdb(page: Page): Promise<{ databases: string[]; entries: IdbEntry[] }> {
  return page.evaluate(async () => {
    const promised = <T>(req: IDBRequest<T>) =>
      new Promise<T>((resolve, reject) => {
        req.onsuccess = () => resolve(req.result);
        req.onerror = () => reject(req.error);
      });
    const databases: string[] = [];
    const entries: IdbEntry[] = [];
    for (const info of await indexedDB.databases()) {
      if (!info.name?.startsWith('algohub-')) continue;
      databases.push(info.name);
      const db = await promised(indexedDB.open(info.name));
      for (const store of Array.from(db.objectStoreNames)) {
        const os = db.transaction(store, 'readonly').objectStore(store);
        const [keys, values] = await Promise.all([
          promised(os.getAllKeys()),
          promised(os.getAll()),
        ]);
        keys.forEach((key, i) =>
          entries.push({
            db: info.name as string,
            store,
            key: String(key),
            value: typeof values[i] === 'string' ? values[i] : JSON.stringify(values[i]),
          }),
        );
      }
      db.close();
    }
    return { databases, entries };
  });
}

// Query key yang tersimpan di cache persist, mis. ["elearning","curriculum","levels"].
export function persistedQueryKeys(entries: IdbEntry[]): string[] {
  return entries.flatMap((e) => {
    const client = JSON.parse(e.value) as {
      clientState?: { queries?: { queryKey: unknown[] }[] };
    };
    return (client.clientState?.queries ?? []).map((q) => JSON.stringify(q.queryKey));
  });
}
