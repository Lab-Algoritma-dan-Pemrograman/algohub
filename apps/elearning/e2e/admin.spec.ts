import { users } from './fixtures.ts';
import { expect, loginAs, test } from './test.ts';

test('admin mengubah XP peserta: konfirmasi lalu PUT /admin/users/{nim}/xp', async ({
  page,
  api,
}) => {
  await loginAs(page, 'admin');
  await page.goto('/admin');
  await page.getByRole('button', { name: new RegExp(users.mahasiswa.nama) }).click();
  await page.getByPlaceholder('XP baru').fill('900');
  await page.getByRole('button', { name: 'Simpan' }).click();

  const dialog = page.getByRole('dialog');
  await expect(dialog.getByText('dari 340 menjadi 900', { exact: false })).toBeVisible();
  await dialog.getByRole('button', { name: 'Lanjutkan' }).click();

  await expect(page.getByText('XP berhasil diubah menjadi 900.')).toBeVisible();
  const [call] = api.callsTo('PUT', `/admin/users/${users.mahasiswa.nim}/xp`);
  expect(call.body).toEqual({ xp: 900 });
  expect(call.authorization).toBe(`Bearer token-${users.admin.nim}`);
});

test('admin tidak bisa mengubah peran akunnya sendiri', async ({ page, api }) => {
  api.on('GET', '/admin/users', {
    items: [
      {
        nim: users.admin.nim,
        id: 'uuid-admin',
        nama: users.admin.nama,
        kelas: users.admin.kelas,
        jurusan: users.admin.jurusan,
        email: null,
        role: 'admin',
        elearning_role: 'admin',
        xp: 0,
        level: 1,
        streak: 0,
        study_time: 0,
        last_active: null,
        created_at: null,
        assessment_access: {},
        level_access_overrides: {},
        shift: null,
        prestige_tier: null,
        prestige_mult: 1,
        prestige_count: 0,
        last_prestige_at: null,
      },
    ],
    total: 1,
    page: 1,
    limit: 25,
  });
  await loginAs(page, 'admin');
  await page.goto('/admin');
  await page.getByRole('button', { name: new RegExp(users.admin.nama) }).click();
  await page.getByRole('combobox').selectOption('praktikan');
  await expect(page.getByText(/tidak dapat mengubah peran akun Anda sendiri/)).toBeVisible();
  expect(api.callsTo('PATCH', /\/role$/)).toHaveLength(0);
});
