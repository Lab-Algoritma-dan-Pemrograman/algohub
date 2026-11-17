import { users } from './fixtures';
import { loginAs } from './helpers';
import { expect, test } from './mock-api';

test('mahasiswa login masuk ke dashboard', async ({ page }) => {
  await loginAs(page, 'mahasiswa');
  await expect(page.getByText('Halo, Mahasiswa')).toBeVisible();
  expect(await page.evaluate(() => localStorage.getItem('token'))).toBe('tok-e2e-mahasiswa');
});

test('asisten login masuk ke admin', async ({ page }) => {
  await loginAs(page, 'asisten');
  const user = await page.evaluate(() => JSON.parse(localStorage.getItem('user') ?? 'null'));
  expect(user).toMatchObject({ id: users.asisten.id, role: 'asisten' });
});

test('koordinator login masuk ke admin lalu superadmin tampil', async ({ page, api }) => {
  await loginAs(page, 'koordinator');
  await page.goto('/praktikum/superadmin/pengaturan');
  await expect(page).toHaveURL(/\/praktikum\/superadmin\/pengaturan$/);
  await expect.poll(() => api.callsTo('GET', '/superadmin/permissions').length).toBe(1);
});

for (const path of ['/praktikum/dashboard', '/praktikum/admin/users']) {
  test(`anonim ke ${path} dialihkan ke login`, async ({ page, api }) => {
    await page.goto(path);
    await expect(page).toHaveURL(/\/praktikum\/login$/);
    expect(api.calls).toEqual([]);
  });
}

test('mahasiswa ke admin dialihkan ke dashboard', async ({ page }) => {
  await loginAs(page, 'mahasiswa');
  await page.goto('/praktikum/admin');
  await expect(page).toHaveURL(/\/praktikum\/dashboard$/);
});

test('asisten ke superadmin dialihkan ke dashboard', async ({ page }) => {
  await loginAs(page, 'asisten');
  await page.goto('/praktikum/superadmin/pengaturan');
  await expect(page).toHaveURL(/\/praktikum\/dashboard$/);
});

test('401 di halaman login menampilkan pesan server tanpa redirect', async ({ page, api }) => {
  await page.goto('/praktikum/login');
  await page.locator('#identifier').fill(users.mahasiswa.nim);
  await page.locator('#pw').fill('salah');
  await page.getByRole('button', { name: /Masuk ke Portal/ }).click();
  await expect(page.getByText('NIM/email atau password salah')).toBeVisible();
  await expect(page).toHaveURL(/\/praktikum\/login$/);
  expect(api.callsTo('POST', '/auth/login')[0].body).toEqual({
    identifier: users.mahasiswa.nim,
    password: 'salah',
  });
  expect(await page.evaluate(() => localStorage.getItem('token'))).toBeNull();
});
