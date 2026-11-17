import { expect, loginAs, setupMocks, test } from './helpers.ts';

const praktikan = {
  id: 101,
  username: '202511001',
  full_name: 'Praktikan Uji',
  role: 'mahasiswa',
  class_code: 'TE B',
  major: 'Teknik Elektro',
  shift: '1',
};

const logRow = (id: number, status: string) => ({
  id,
  user_id: 101,
  status,
  notes: 'Input manual oleh Aslab',
  check_in_time: new Date().toISOString(),
  is_verified: true,
  reschedule_status: null,
  reschedule_schedule_id: null,
  recorded_by_name: 'Koordinator Uji',
  user_full_name: 'Praktikan Uji',
  user_role: 'mahasiswa',
  user_username: '202511001',
  user_major: 'Teknik Elektro',
  user_class_code: 'TE B',
  user_shift: '1',
});

test('staff input manual absensi lalu event Realtime menampilkan baris baru', async ({ page }) => {
  const logs: ReturnType<typeof logRow>[] = [];
  const mocks = await setupMocks(page, 'koordinator', {
    get_users_secure: [praktikan],
    get_schedules_secure: [],
    get_attendance_logs_secure: () => logs,
    get_system_settings_full_secure: [{ active_shift: 'all', wa_templates: {} }],
    get_deletion_history_secure: [],
    upsert_attendance_log_secure: null,
  });
  await loginAs(page, mocks);
  await page.goto('/absensi');
  await expect(page.getByText('Belum ada data absensi pada tanggal ini.')).toBeVisible();
  await mocks.realtime.waitForChannel(page, 'attendance_live_updates');

  await page.getByLabel('NIM Praktikan').fill('202511001');
  await page.getByRole('button', { name: 'Simpan Kehadiran' }).click();
  await expect(page.getByText('Hore! Aksi kamu berhasil dilaksanakan! 🥳🚀')).toBeVisible();
  const [params] = mocks.api.callsTo('upsert_attendance_log_secure');
  expect(params).toMatchObject({
    p_caller_id: 401,
    p_target_user_id: 101,
    p_status: 'Hadir',
    p_notes: 'Input manual oleh Aslab',
    p_is_verified: true,
    p_type: 'staff_manual',
    p_meeting: null,
  });

  logs.push(logRow(1, 'Hadir'));
  expect(mocks.realtime.emit('attendance_logs', 'INSERT', { id: 1, user_id: 101 })).toBeGreaterThan(
    0,
  );
  await expect(page.getByTestId('attendance-row')).toHaveCount(1);
  await expect(page.getByTestId('attendance-row')).toContainText('Praktikan Uji');
});
