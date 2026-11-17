import { expect, idbAlgohubDump, loginAs, setupMocks, test, waitForPersist } from './helpers.ts';

test('data user dan audit log tidak dipersist, jadwal umum dipersist', async ({ page }) => {
  const mocks = await setupMocks(page, 'koordinator', {
    get_users_secure: [
      {
        id: 101,
        username: '202511001',
        full_name: 'Praktikan Rahasia',
        role: 'mahasiswa',
        is_active: true,
      },
    ],
    get_system_settings_secure: [{ active_shift: 'all' }],
    get_audit_logs_secure: [
      {
        id: 1,
        created_at: new Date().toISOString(),
        actor_name: 'Koordinator Uji',
        actor_id: 401,
        action_type: 'LOGIN',
        description: 'Login',
        payload: null,
      },
    ],
    get_schedules_secure: [
      {
        id: 7,
        type: 'praktikum',
        status: 'approved',
        day_of_week: 'Senin',
        start_time: '08:00:00',
        end_time: '10:00:00',
        title: 'Praktikum Algoritma',
        major: 'Teknik Elektro',
        class_code: 'B',
      },
    ],
  });
  await loginAs(page, mocks);
  await page.goto('/manajemen-user');
  await expect(page.getByText('Praktikan Rahasia')).toBeVisible();
  await page.goto('/audit-log');
  await expect(page.getByText('Koordinator Uji').first()).toBeVisible();
  await page.goto('/manajemen-jadwal');
  await expect(page.getByText('Praktikum Algoritma')).toBeVisible();
  await waitForPersist(page);

  const { text } = await idbAlgohubDump(page);
  expect(text).toContain('get_schedules_secure');
  expect(text).not.toContain('get_users_secure');
  expect(text).not.toContain('get_audit_logs_secure');
  expect(text).not.toContain('Praktikan Rahasia');
});
