import { expect, loginAs, setupMocks, test } from './helpers.ts';

const row = (id: number, full_name: string, username: string) => ({
  id,
  username,
  full_name,
  role: 'mahasiswa',
  nim: username,
  class_code: 'TE B',
  shift: '1',
  phone_number: null,
  division: null,
  assistant_code: null,
  is_active: true,
});

test('koordinator membuat user baru dan tabel diperbarui', async ({ page }) => {
  const users = [row(101, 'Praktikan Uji', '202511001')];
  const mocks = await setupMocks(page, 'koordinator', {
    get_users_secure: () => users,
    get_system_settings_secure: [{ active_shift: 'all' }],
    admin_create_user: (params) => {
      users.push(row(102, String(params.p_full_name), String(params.p_username)));
      return null;
    },
  });
  await loginAs(page, mocks);
  await page.goto('/manajemen-user');
  await expect(page.getByTestId('user-row')).toHaveCount(1);

  await page.getByRole('button', { name: 'Tambah Manual' }).click();
  await page.getByLabel('Username / NIM').fill('202511002');
  await page.getByLabel('Nama Lengkap').fill('Praktikan Baru');
  await page.getByLabel('No. WhatsApp / HP').fill('081211112222');
  await page.getByPlaceholder('Masukkan password').fill('rahasia1');
  await page.getByRole('button', { name: 'Simpan', exact: true }).click();
  await expect(page.getByText('Yeay! Data berhasil disimpan dengan aman! 🎉✨')).toBeVisible();

  expect(mocks.api.callsTo('admin_create_user')).toEqual([
    {
      p_caller_id: 401,
      p_username: '202511002',
      p_password: 'rahasia1',
      p_full_name: 'Praktikan Baru',
      p_role: 'praktikan',
      p_nim: '202511002',
      p_assistant_code: null,
      p_division: null,
      p_phone_number: '081211112222',
      p_class_code: null,
      p_shift: null,
      p_is_active: true,
      p_major: null,
    },
  ]);
  await expect(page.getByTestId('user-row')).toHaveCount(2);
  await expect(page.getByText('Praktikan Baru')).toBeVisible();
});
