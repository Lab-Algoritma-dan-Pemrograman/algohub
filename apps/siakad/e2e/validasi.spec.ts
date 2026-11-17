import { expect, loginAs, setupMocks, test } from './helpers.ts';

const izin = (id: number, name: string) => ({
  id,
  user_id: 100 + id,
  status: 'Izin',
  notes: '[Kelas B - Senin] Acara keluarga',
  check_in_time: new Date().toISOString(),
  verification_status: 'pending',
  is_verified: false,
  reschedule_status: null,
  reschedule_schedule_id: null,
  user_full_name: name,
  user_role: 'mahasiswa',
  user_phone_number: '081200000777',
});

test('asisten menyetujui izin dan event Realtime menambah permintaan baru', async ({ page }) => {
  const logs = [izin(1, 'Praktikan Satu')];
  const mocks = await setupMocks(page, 'asisten', {
    get_attendance_logs_secure: () => logs,
    admin_verify_attendance_secure: null,
  });
  await loginAs(page, mocks);
  await page.goto('/validasi-absensi');
  await expect(page.getByTestId('pending-row')).toHaveCount(1);
  await mocks.realtime.waitForChannel(page, 'validasi_live_updates');

  await page.getByRole('button', { name: 'Setuju' }).click();
  await expect(page.getByText('Izin Disetujui. Praktikan bisa pilih jadwal. 🎉')).toBeVisible();
  expect(mocks.api.callsTo('admin_verify_attendance_secure')).toEqual([
    { p_caller_id: 301, p_log_id: 1, p_is_approved: true, p_type: 'license' },
  ]);
  await expect(page.getByTestId('pending-row')).toHaveCount(0);

  logs[0] = { ...logs[0], verification_status: 'approved', is_verified: true };
  logs.push(izin(2, 'Praktikan Dua'));
  mocks.realtime.emit('attendance_logs', 'INSERT', { id: 2 });
  await expect(page.getByTestId('pending-row')).toHaveCount(1);
  await expect(page.getByTestId('pending-row')).toContainText('Praktikan Dua');
  await expect(page.getByTestId('history-row')).toContainText('Praktikan Satu');
});
