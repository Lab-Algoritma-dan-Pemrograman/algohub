import { users } from './fixtures.ts';
import { expect, loginAs, test } from './test.ts';

test('profil menampilkan identitas, XP, dan katalog achievement dari backend', async ({ page }) => {
  await loginAs(page, 'mahasiswa');
  await page.goto('/profile');
  await expect(page.getByRole('heading', { level: 1, name: users.mahasiswa.nama })).toBeVisible();
  await expect(page.getByText('120 XP').first()).toBeVisible();
  await expect(page.getByText('Langkah Pertama')).toBeVisible();
  await expect(page.getByText('0 / 1 Terbuka')).toBeVisible();
  await expect(page.getByText('Semua level sudah tamat.')).toBeVisible();
});

test('prestige reset: konfirmasi, request tanpa body, lalu status berubah', async ({
  page,
  api,
}) => {
  await loginAs(page, 'mahasiswa');
  await page.goto('/profile');
  await page.getByRole('button', { name: 'Reset & Ambil x1.50' }).click();
  await page.getByRole('button', { name: 'Ya, Reset', exact: true }).click();

  await expect(page.getByRole('status').filter({ hasText: 'Prestige #1 aktif' })).toBeVisible();
  expect(api.callsTo('POST', '/prestige/reset')[0].body).toEqual({});
  await expect(page.getByRole('button', { name: 'Belum Layak Prestige' })).toBeDisabled();
  await expect(page.getByText('Master x1.50', { exact: true })).toBeVisible();
});

test('prestige yang belum layak tidak bisa dipicu', async ({ page, api }) => {
  api.on('GET', '/prestige/eligibility', {
    eligible: false,
    reason: 'Belum tamat 6/6 level.',
    levelsDone: 2,
    totalLevels: 6,
    lessonsDone: 8,
    totalLessons: 30,
    tier: 'Master',
    mult: 1.5,
    prestigeCount: 0,
    lastPrestigeAt: null,
    semesterUsed: 0,
  });
  await loginAs(page, 'mahasiswa');
  await page.goto('/profile');
  await expect(page.getByRole('button', { name: 'Belum Layak Prestige' })).toBeDisabled();
  expect(api.callsTo('POST', '/prestige/reset')).toHaveLength(0);
});
