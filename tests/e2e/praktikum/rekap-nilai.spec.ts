import { readFile } from 'node:fs/promises';
import { buildRekapCsv } from '../../../apps/praktikum/src/lib/rekap-nilai';
import { rekapKelas } from './fixtures';
import { loginAs } from './helpers';
import { expect, test } from './mock-api';

const ROWS = 1_200;

test('1.200 baris dirender virtual dan CSV unduhan sama', async ({ page, api }) => {
  const rekap = rekapKelas(ROWS);
  api.on('GET', /^\/admin\/rekap\/kelas\/\d+$/, rekap);
  await loginAs(page, 'asisten');
  await page.goto('/praktikum/admin/rekap-nilai');

  const first = rekap.data[0];
  const last = rekap.data[ROWS - 1];
  await expect(page.getByRole('cell', { name: first.nim, exact: true })).toBeVisible();
  expect(api.callsTo('GET', '/admin/rekap/kelas/1')).toHaveLength(1);
  expect(await page.locator('tbody tr').count()).toBeLessThan(200);

  const scroller = page.locator('div.table-wrap.overflow-auto');
  await scroller.evaluate((el) => (el.scrollTop = el.scrollHeight));
  await expect(page.getByRole('cell', { name: last.nim, exact: true })).toBeVisible();
  expect(await page.locator('tbody tr').count()).toBeLessThan(200);

  const download = page.waitForEvent('download');
  await page.getByRole('button', { name: /Export CSV/ }).click();
  const file = await download;
  expect(file.suggestedFilename()).toBe('rekap_nilai_kelas_1.csv');
  const csv = await readFile(await file.path(), 'utf8');
  expect(csv).toBe(buildRekapCsv(rekap, rekap.data));
  expect(csv.split('\n')).toHaveLength(ROWS + 1);
});
