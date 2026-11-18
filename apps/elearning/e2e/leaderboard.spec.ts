import { classmate, classmate2, users } from './fixtures.ts';
import { expect, loginAs, test } from './test.ts';

test('mahasiswa melihat top 3 mingguan kelasnya dan XP-nya sendiri', async ({ page, api }) => {
  await loginAs(page, 'mahasiswa');
  await page.goto('/leaderboard');
  await expect(page.getByText(classmate.nama).first()).toBeVisible();
  await expect(page.getByText(classmate2.nama).first()).toBeVisible();
  await expect(page.getByText(users.mahasiswa.nama).first()).toBeVisible();
  await expect(page.getByText('XP minggumu: 340 XP')).toBeVisible();
  await expect(page.getByText(/Top 3 Penjelajah di Kelas A1/)).toBeVisible();
  const params = api.callsTo('GET', '/leaderboard/weekly').map((c) => ({
    kelas: c.search.get('kelas'),
    jurusan: c.search.get('jurusan'),
    shift: c.search.get('shift'),
  }));
  expect(params[0]).toEqual({ kelas: 'A1', jurusan: 'Teknik Elektro', shift: '1' });
});

test('kordas memilih papan Total dan memfilter kelas lewat server', async ({ page, api }) => {
  await loginAs(page, 'kordas');
  await page.goto('/leaderboard');
  await page.getByRole('button', { name: 'Total', exact: true }).click();
  await expect(page.getByText(classmate.nama).first()).toBeVisible();
  expect(api.callsTo('GET', '/leaderboard')).not.toHaveLength(0);

  await expect(page.getByRole('option', { name: 'B1' })).toBeAttached();
  await page.getByLabel('Filter kelas').selectOption('B1');
  await expect
    .poll(() => api.callsTo('GET', '/leaderboard').some((c) => c.search.get('kelas') === 'B1'))
    .toBe(true);
  await expect(page.getByRole('button', { name: 'Reset Filter' })).toBeVisible();
});
