import { users } from './fixtures.ts';
import { expect, loginAs, test } from './test.ts';

test('event Realtime active_sessions memperbarui monitoring tanpa reload', async ({
  page,
  api,
  realtime,
}) => {
  await loginAs(page, 'asisten');
  await page.goto('/monitoring');
  await expect(page.getByText(users.mahasiswa.nama)).toBeVisible();
  await expect(page.getByText('0 Aktif', { exact: true })).toBeVisible();
  await realtime.waitForChannel(page, 'monitoring:active_sessions');
  await page.evaluate(() => {
    (window as unknown as { __tanpaReload: boolean }).__tanpaReload = true;
  });

  // Daftar sesi awal harus sudah masuk cache sebelum event diterapkan; emit diulang sampai tampil.
  await expect(async () => {
    realtime.emit('active_sessions', 'INSERT', {
      nim: users.mahasiswa.nim,
      last_heartbeat: new Date().toISOString(),
    });
    await expect(page.getByText('1 Aktif', { exact: true })).toBeVisible({ timeout: 1_000 });
  }).toPass({ timeout: 15_000 });

  expect(
    await page.evaluate(() => (window as unknown as { __tanpaReload?: boolean }).__tanpaReload),
  ).toBe(true);
  expect(api.callsTo('GET', '/monitoring/students')).toHaveLength(1);
});

test('asisten hanya melihat: tidak ada kontrol ubah XP, reset, atau hapus', async ({ page }) => {
  await loginAs(page, 'asisten');
  await page.goto('/monitoring');
  await page.getByText(users.mahasiswa.nama).click();
  await expect(page.getByRole('button', { name: /Hapus/ })).toHaveCount(0);
  await expect(page.getByRole('button', { name: /Reset/ })).toHaveCount(0);
});
