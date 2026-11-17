import { test as base, expect, type Page } from '@playwright/test';
import type { Role } from './fixtures/users.ts';
import { MockApi, type RpcReply } from './mock/api.ts';
import { MockRealtime } from './mock/realtime.ts';

const APP_ORIGIN = 'http://localhost:5174';
const STATIC_CDN =
  /^https:\/\/(fonts\.googleapis\.com|fonts\.gstatic\.com|cdnjs\.cloudflare\.com)\//;

export interface Mocks {
  api: MockApi;
  realtime: MockRealtime;
}

// Tidak ada request yang boleh keluar ke jaringan: font/ikon CDN dijawab kosong, sisanya diblokir.
async function isolate(page: Page) {
  await page.route('**/*', (route) => {
    const url = route.request().url();
    if (url.startsWith(APP_ORIGIN)) return route.continue();
    if (STATIC_CDN.test(url)) return route.fulfill({ status: 200, body: '' });
    return route.abort('blockedbyclient');
  });
}

export async function setupMocks(
  page: Page,
  role: Role,
  overrides: Record<string, RpcReply> = {},
): Promise<Mocks> {
  const api = new MockApi(role, overrides);
  const realtime = new MockRealtime();
  await isolate(page);
  await api.install(page);
  await realtime.install(page);
  return { api, realtime };
}

export async function loginAs(page: Page, mocks: Mocks): Promise<void> {
  await page.goto('/');
  await page.getByPlaceholder('Masukkan NIM atau Username').fill('user');
  await page.getByPlaceholder('••••••••').fill('rahasia');
  await page.getByRole('button', { name: /Masuk Sekarang/ }).click();
  await expect(page).toHaveURL(/\/beranda$/);
  expect(mocks.api.loggedIn).toBe(true);
}

// Isi semua DB IndexedDB algohub-* sebagai teks, untuk memeriksa apa yang dipersist.
export function idbAlgohubDump(page: Page): Promise<{ entries: number; text: string }> {
  return page.evaluate(async () => {
    const dbs = (await indexedDB.databases()).filter((d) => d.name?.startsWith('algohub-'));
    let entries = 0;
    let text = '';
    for (const info of dbs) {
      const db = await new Promise<IDBDatabase>((resolve, reject) => {
        const req = indexedDB.open(info.name as string);
        req.onsuccess = () => resolve(req.result);
        req.onerror = () => reject(req.error);
      });
      for (const store of Array.from(db.objectStoreNames)) {
        const values = await new Promise<unknown[]>((resolve, reject) => {
          const req = db.transaction(store, 'readonly').objectStore(store).getAll();
          req.onsuccess = () => resolve(req.result);
          req.onerror = () => reject(req.error);
        });
        entries += values.length;
        text += values.map((v) => (typeof v === 'string' ? v : JSON.stringify(v))).join('\n');
      }
      db.close();
    }
    return { entries, text };
  });
}

// Persister TanStack menunda tulis sekitar 1 detik.
export async function waitForPersist(page: Page): Promise<void> {
  await page.waitForTimeout(1500);
}

export const test = base;
export { expect };
