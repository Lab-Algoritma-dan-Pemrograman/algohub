import { expect, loginAs, setupMocks, test } from './helpers.ts';

const schedule = {
  id: 7,
  type: 'praktikum',
  status: 'approved',
  day_of_week: 'Senin',
  start_time: '08:00:00',
  end_time: '10:00:00',
  title: 'Praktikum Algoritma',
  major: 'Teknik Elektro',
  class_code: 'B',
};

const assignment = {
  id: 55,
  user_id: 301,
  schedule_id: 7,
  schedule_day: 'Senin',
  schedule_start: '08:00:00',
  schedule_end: '10:00:00',
  schedule_title: 'Praktikum Algoritma',
  schedule_major: 'Teknik Elektro',
  schedule_class_code: 'B',
  user_full_name: 'Asisten Uji',
  original_user_id: null,
  original_user_full_name: null,
  task_role: 'Operator',
  activity_name: 'Modul 2',
  activity_date: '2099-01-05',
  status: 'aktif',
};

test('koordinator menugaskan asisten lalu event Realtime memperbarui tabel', async ({ page }) => {
  const assignments: (typeof assignment)[] = [];
  const mocks = await setupMocks(page, 'koordinator', {
    check_menu_access_secure: true,
    get_system_settings_full_secure: [{ wa_templates: {} }],
    get_schedule_assignments_secure: () => assignments,
    get_schedules_secure: [schedule],
    get_users_secure: [
      { id: 301, full_name: 'Asisten Uji', role: 'asisten', phone_number: '0812' },
      { id: 401, full_name: 'Koordinator Uji', role: 'koordinator', phone_number: '0813' },
    ],
    upsert_schedule_assignment_secure: null,
    'endpoint:/send-push': { sent: 1 },
  });
  await loginAs(page, mocks);
  await page.goto('/jadwal-jaga');
  await expect(page.getByText('Tidak ada data.')).toBeVisible();
  await mocks.realtime.waitForChannel(page, 'jadwal_jaga_live_updates');

  await page.getByRole('button', { name: 'Tambah Petugas' }).click();
  await page.getByLabel('Tanggal Bertugas').fill('2099-01-05');
  const pick = async (label: string, option: string) => {
    await page.getByRole('combobox', { name: label }).click();
    await page.getByRole('option', { name: option }).click();
  };
  await pick('Kegiatan', 'Modul 2');
  await pick('Target Kelas', 'Senin, 08:00 - Praktikum Algoritma (Teknik Elektro Kelas B)');
  await pick('Asisten', 'Asisten Uji');
  await pick('Tugas', 'Operator');
  await page.getByRole('button', { name: 'Simpan Perubahan' }).click();
  await expect(page.getByText('Petugas ditambahkan! 🎉')).toBeVisible();
  expect(mocks.api.callsTo('upsert_schedule_assignment_secure')).toEqual([
    {
      p_caller_id: 401,
      p_id: 0,
      p_schedule_id: 7,
      p_user_id: 301,
      p_task_role: 'Operator',
      p_activity_name: 'Modul 2',
      p_activity_date: '2099-01-05',
    },
  ]);
  await expect.poll(() => mocks.api.endpoints.some((e) => e.path === '/send-push')).toBe(true);

  assignments.push(assignment);
  mocks.realtime.emit('schedule_assignments', 'INSERT', { id: 55, user_id: 301 });
  await expect(page.getByTestId('assignment-row')).toHaveCount(1);
  await expect(page.getByTestId('assignment-row')).toContainText('Asisten Uji');
});
