import { PASSWORD, tokenFor, users, type Role } from './fixtures.ts';
import { expect, loginAs, test } from './test.ts';

test('login langsung: kredensial dikirim dan token dipakai untuk request berikutnya', async ({
  page,
  api,
}) => {
  await loginAs(page, 'mahasiswa');
  await expect(page.getByRole('heading', { name: /Selamat datang kembali, Budi/ })).toBeVisible();
  expect(api.callsTo('POST', '/login')[0].body).toEqual({
    identifier: users.mahasiswa.nim,
    password: PASSWORD,
  });
  await expect
    .poll(() => api.callsTo('POST', '/profile/sync')[0]?.authorization)
    .toBe(`Bearer ${tokenFor(users.mahasiswa.nim)}`);
});

test('login salah menampilkan pesan dari server dan tetap di halaman login', async ({ page }) => {
  await page.goto('/login');
  await page.locator('#identifier').fill(users.mahasiswa.nim);
  await page.locator('#pw').fill('salah');
  await page.getByRole('button', { name: /Masuk ke E-Learning/ }).click();
  await expect(page.getByRole('alert')).toContainText('NIM/email atau password salah');
  await expect(page).toHaveURL(/\/login$/);
});

test('SSO /sso?code= menukar kode sekali pakai; pemakaian kedua ditolak server', async ({
  page,
  api,
}) => {
  await page.goto('/sso?code=kode-sekali');
  await expect(page).toHaveURL(/\/dashboard$/);
  expect(api.callsTo('POST', '/receive-token').map((c) => c.body)).toEqual([
    { code: 'kode-sekali' },
  ]);
  // Token panjang tidak boleh lewat URL; hanya kode pendek yang ada di /sso.
  expect(page.url()).not.toContain('token');

  await page.evaluate(() => localStorage.clear());
  await page.goto('/sso?code=kode-sekali');
  await expect(page).toHaveURL(/\/login$/);
  await expect(page.getByRole('alert')).toContainText('Kode sudah dipakai');
  expect(api.callsTo('POST', '/receive-token')).toHaveLength(2);
});

test('SSO tanpa kode kembali ke login dengan pesan', async ({ page, api }) => {
  await page.goto('/sso');
  await expect(page).toHaveURL(/\/login$/);
  await expect(page.getByRole('alert')).toContainText('Kode masuk tidak ditemukan');
  expect(api.callsTo('POST', '/receive-token')).toHaveLength(0);
});

test('tanpa sesi, route aplikasi dialihkan ke landing', async ({ page }) => {
  await page.goto('/dashboard');
  await expect(page).toHaveURL(/\/$/);
  await expect(page.getByRole('link', { name: /Mulai Sekarang/ }).first()).toBeVisible();
});

const GUARDED = ['/admin', '/monitoring', '/auditlog'] as const;
const ALLOWED: Record<Role, readonly string[]> = {
  mahasiswa: [],
  asisten: ['/monitoring'],
  kordas: GUARDED,
  admin: GUARDED,
};
const HEADING: Record<string, string> = {
  '/admin': 'Admin Dashboard',
  '/monitoring': 'Monitoring Mahasiswa',
  '/auditlog': 'Audit Log',
};

for (const role of ['mahasiswa', 'asisten', 'kordas', 'admin'] as const) {
  test(`role ${role}: route staf hanya terbuka untuk role yang berhak`, async ({ page }) => {
    await loginAs(page, role);
    for (const path of GUARDED) {
      await page.goto(path);
      if (ALLOWED[role].includes(path)) {
        await expect(page).toHaveURL(new RegExp(`${path}$`));
        await expect(page.getByRole('heading', { name: HEADING[path] })).toBeVisible();
      } else {
        await expect(page).toHaveURL(/\/dashboard$/);
      }
    }
  });
}
