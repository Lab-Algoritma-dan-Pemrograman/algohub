import type { Page } from '@playwright/test';
import { NILAI_MARKER, OTHER_USER_NAME, SOAL_MARKER, ruang, startedRuang } from './fixtures';
import {
  loginAs,
  logout,
  persistedQueryKeys,
  readAlgohubIdb,
  waitForPersist,
  type IdbEntry,
} from './helpers';
import { expect, test } from './mock-api';

const FORBIDDEN_MARKERS = [SOAL_MARKER, 'tok-e2e-', OTHER_USER_NAME, String(NILAI_MARKER)];

async function visit(page: Page, path: string, ready: RegExp | string) {
  await page.goto(path);
  await expect(page.getByText(ready).first()).toBeVisible();
}

function expectNoForbiddenData(entries: IdbEntry[]) {
  const dump = JSON.stringify(entries);
  for (const marker of FORBIDDEN_MARKERS) expect(dump, marker).not.toContain(marker);
}

test('logout mengosongkan IndexedDB algohub-*', async ({ page }) => {
  await loginAs(page, 'mahasiswa');
  await visit(page, '/praktikum/profil', /Mahasiswa Uji|Profil/);
  await waitForPersist(page);
  const before = await readAlgohubIdb(page);
  expect(before.map((e) => `${e.db}/${e.store}/${e.key}`)).toEqual([
    'algohub-praktikum/query-cache/algohub-praktikum-101',
  ]);
  expect(persistedQueryKeys(before)).toContain('["praktikum","profile"]');

  await logout(page);
  await waitForPersist(page);
  expect(await readAlgohubIdb(page)).toEqual([]);
  expect(await page.evaluate(() => localStorage.getItem('token'))).toBeNull();
  expect(await page.evaluate(() => localStorage.getItem('user'))).toBeNull();
});

test('data terlarang tidak tersimpan di IndexedDB', async ({ page, api }) => {
  const r = startedRuang(ruang('pretest'));
  api.on('GET', '/ruang', r);
  api.on('POST', '/autosave', null);

  await loginAs(page, 'mahasiswa');
  await visit(page, '/praktikum/dashboard', 'Halo, Mahasiswa');
  await visit(page, '/praktikum/sesi', 'Sesi 1');
  await visit(page, `/praktikum/sesi/1/pretest?aktivasi=11&course=31`, SOAL_MARKER);
  await visit(page, '/praktikum/profil', /Profil/);
  await waitForPersist(page);
  let entries = await readAlgohubIdb(page);
  expect(persistedQueryKeys(entries).sort()).toEqual(['["praktikum","profile"]']);
  expectNoForbiddenData(entries);

  await logout(page);
  await loginAs(page, 'asisten');
  await visit(page, '/praktikum/admin/rekap-nilai', OTHER_USER_NAME);
  await visit(page, '/praktikum/admin/users', OTHER_USER_NAME);
  await visit(page, '/praktikum/admin/log', OTHER_USER_NAME);
  await page.goto('/praktikum/admin/rekap-jawaban');
  await expect.poll(() => api.callsTo('GET', '/admin/rekap-jawaban').length).toBeGreaterThan(0);
  await visit(page, '/praktikum/profil', /Profil/);
  await waitForPersist(page);
  entries = await readAlgohubIdb(page);
  expect(entries.map((e) => e.key)).toEqual(['algohub-praktikum-201']);
  expect(persistedQueryKeys(entries).sort()).toEqual([
    '["praktikum","admin","kelas"]',
    '["praktikum","profile"]',
  ]);
  expectNoForbiddenData(entries);
});
