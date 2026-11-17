import { expect, idbAlgohubDump, loginAs, setupMocks, test, waitForPersist } from './helpers.ts';
import type { Role } from './fixtures/users.ts';

const MENU: Record<Role, { visible: string[]; hidden: string[] }> = {
  praktikan: {
    visible: ['Beranda', 'Absensi', 'Kritik & Saran', 'E-Learning', 'Jadwal Saya'],
    hidden: ['Manajemen User', 'Validasi Absensi', 'Pengaturan', 'Sewa & Pinjam'],
  },
  penyewa: {
    visible: ['Beranda', 'Profil Saya', 'Sewa & Pinjam'],
    hidden: ['Absensi', 'Manajemen User', 'Jadwal Saya'],
  },
  asisten: {
    visible: ['Absensi', 'Buat QR Absen', 'Inventaris', 'Validasi Absensi', 'Jadwal Jaga'],
    hidden: ['Manajemen User', 'Audit Log', 'Pengaturan', 'Kritik & Saran'],
  },
  koordinator: {
    // Kode lama menampilkan seluruh menu untuk koordinator, termasuk Sewa & Pinjam.
    visible: ['Manajemen User', 'Audit Log', 'Pengaturan', 'Kritik & Saran', 'Sewa & Pinjam'],
    hidden: [],
  },
};

const sidebar = (page: import('@playwright/test').Page) => page.getByLabel('Menu Utama');

for (const role of Object.keys(MENU) as Role[]) {
  test(`login ${role} masuk ke /beranda dengan menu sesuai role`, async ({ page }) => {
    const mocks = await setupMocks(page, role);
    await loginAs(page, mocks);
    for (const title of MENU[role].visible) {
      await expect(sidebar(page).getByText(title, { exact: true })).toBeVisible();
    }
    for (const title of MENU[role].hidden) {
      await expect(sidebar(page).getByText(title, { exact: true })).toHaveCount(0);
    }
  });
}

test('belum login membuka /beranda diarahkan ke /', async ({ page }) => {
  await setupMocks(page, 'praktikan');
  await page.goto('/beranda');
  await expect(page).toHaveURL(/\/$/);
  await expect(page.getByRole('button', { name: /Masuk Sekarang/ })).toBeVisible();
});

test('sudah login membuka / diarahkan ke /beranda', async ({ page }) => {
  const mocks = await setupMocks(page, 'praktikan');
  await loginAs(page, mocks);
  await page.goto('/');
  await expect(page).toHaveURL(/\/beranda$/);
});

test('role tidak berhak diarahkan ke /beranda', async ({ page }) => {
  const penyewa = await setupMocks(page, 'penyewa');
  await loginAs(page, penyewa);
  await page.goto('/manajemen-user');
  await expect(page).toHaveURL(/\/beranda$/);
});

test('praktikan tidak boleh membuka /validasi-absensi', async ({ page }) => {
  const mocks = await setupMocks(page, 'praktikan');
  await loginAs(page, mocks);
  await page.goto('/validasi-absensi');
  await expect(page).toHaveURL(/\/beranda$/);
});

test('asisten tanpa menu key divisi diarahkan, dengan akses tampil', async ({ page }) => {
  const mocks = await setupMocks(page, 'asisten');
  await loginAs(page, mocks);
  await page.goto('/manajemen-user');
  await expect(page).toHaveURL(/\/beranda$/);
  await page.goto('/inventaris');
  await expect(page).toHaveURL(/\/inventaris$/);
  expect(mocks.api.callsTo('get_division_access_secure')[0]).toMatchObject({
    p_division: 'inventaris',
  });
});

test('ganti password di /profil memanggil update_password', async ({ page }) => {
  const mocks = await setupMocks(page, 'praktikan');
  await loginAs(page, mocks);
  await sidebar(page).getByText('Profil Saya', { exact: true }).click();
  await expect(page).toHaveURL(/\/profil$/);
  await page.getByPlaceholder('Kosongkan jika tidak ingin mengganti password').fill('baru123');
  await page.getByRole('button', { name: 'Simpan Perubahan' }).click();
  await expect(page.getByText('Mantap! Pembaruan berhasil diterapkan! 🚀🌟')).toBeVisible();
  expect(mocks.api.callsTo('update_password')).toEqual([
    { p_caller_id: 101, p_target_id: 101, p_new_password: 'baru123' },
  ]);
});

test('logout mencabut token dan mengosongkan IndexedDB algohub-*', async ({ page }) => {
  const mocks = await setupMocks(page, 'koordinator');
  await loginAs(page, mocks);
  await waitForPersist(page);
  expect((await idbAlgohubDump(page)).entries).toBeGreaterThan(0);
  await page.getByRole('button', { name: 'Keluar Aplikasi' }).click();
  await expect(page).toHaveURL(/\/$/);
  expect(mocks.api.endpoints.some((e) => e.path === '/auth/logout')).toBe(true);
  expect((await idbAlgohubDump(page)).entries).toBe(0);
  expect(await page.evaluate(() => localStorage.getItem('token'))).toBeNull();
});

test('event Realtime penugasan baru memunculkan toast tanpa reload', async ({ page }) => {
  const mocks = await setupMocks(page, 'asisten');
  await loginAs(page, mocks);
  await mocks.realtime.waitForChannel(page, 'global_jadwal_jaga_notif');
  const before = mocks.api.callsTo('get_schedule_assignments_secure').length;
  await page.evaluate(() => ((window as unknown as { __marker: number }).__marker = 1));
  mocks.realtime.emit('schedule_assignments', 'INSERT', {
    id: 9,
    user_id: 301,
    task_role: 'Petugas Jaga',
    activity_name: 'Praktikum',
  });
  await expect(page.getByText('🚀 Penugasan Jadwal Jaga Baru! 🎉')).toBeVisible();
  await expect
    .poll(() => mocks.api.callsTo('get_schedule_assignments_secure').length)
    .toBeGreaterThan(before);
  expect(await page.evaluate(() => (window as unknown as { __marker?: number }).__marker)).toBe(1);
});

test('data dashboard tidak dipersist ke IndexedDB', async ({ page }) => {
  const mocks = await setupMocks(page, 'praktikan');
  await loginAs(page, mocks);
  await expect(page.getByText('Kontak WhatsApp Aktif')).toBeVisible();
  await waitForPersist(page);
  expect((await idbAlgohubDump(page)).text).not.toContain('get_dashboard_stats_secure');
});
