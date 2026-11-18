import { tokenFor, users } from './fixtures.ts';
import { expect, loginAs, readAlgohubIdb, test, waitForPersist } from './test.ts';

test('logout mengosongkan IndexedDB algohub-* dan token lokal', async ({ page, api }) => {
  await loginAs(page, 'mahasiswa');
  await page.goto('/courses');
  await expect(page.getByText('Input Output C').first()).toBeVisible();
  await waitForPersist(page);

  const before = await readAlgohubIdb(page);
  expect(before.databases.length, 'database algohub-* harus ada sebelum logout').toBeGreaterThan(0);
  expect(before.entries.length, 'cache harus terisi sebelum logout').toBeGreaterThan(0);

  await page.getByRole('button', { name: /Keluar/ }).click();
  await expect(page).toHaveURL(/\/$/);
  await expect(page.getByRole('link', { name: /Mulai Sekarang/ }).first()).toBeVisible();

  const after = await readAlgohubIdb(page);
  expect(after.entries).toEqual([]);
  const storage = await page.evaluate(() => JSON.stringify(Object.entries(localStorage)));
  expect(storage).not.toContain(tokenFor(users.mahasiswa.nim));

  expect(api.callsTo('DELETE', '/sessions/me')).toHaveLength(1);
  expect(api.callsTo('POST', '/activity-logs').map((c) => c.body)).toContainEqual({
    eventType: 'logout',
    details: 'Melakukan logout dari sistem',
  });
});

test('setelah logout, route aplikasi tidak bisa dibuka tanpa login ulang', async ({ page }) => {
  await loginAs(page, 'mahasiswa');
  await page.getByRole('button', { name: /Keluar/ }).click();
  await expect(page).toHaveURL(/\/$/);
  await page.goto('/profile');
  await expect(page).toHaveURL(/\/$/);
});
