import { OTHER_USER_NAME, otherUser, permissions } from './fixtures';
import { loginAs } from './helpers';
import { expect, test } from './mock-api';

const ADMIN_PAGES: [path: string, request: string | RegExp][] = [
  ['/praktikum/admin', '/admin/dashboard'],
  ['/praktikum/admin/aktivasi', '/admin/aktivasi'],
  ['/praktikum/admin/asisten', '/admin/ampuan'],
  ['/praktikum/admin/jadwal', '/admin/jadwal'],
  ['/praktikum/admin/kelas', '/admin/kelas'],
  ['/praktikum/admin/laporan-besar', '/admin/laporan-besar'],
  ['/praktikum/admin/log', '/admin/audit-logs'],
  ['/praktikum/admin/modul', '/admin/konfigurasi'],
  ['/praktikum/admin/pedoman', '/admin/pedoman'],
  ['/praktikum/admin/pengaturan', '/admin/konfigurasi'],
  ['/praktikum/admin/penilaian', '/admin/konfigurasi/ai-model'],
  ['/praktikum/admin/rekap-jawaban', '/admin/rekap-jawaban'],
  ['/praktikum/admin/rekap-nilai', /^\/admin\/rekap\/kelas\/1$/],
  ['/praktikum/admin/sesi', '/admin/sesi'],
  ['/praktikum/admin/users', '/admin/users'],
  ['/praktikum/superadmin/pengaturan', '/superadmin/permissions'],
];

test('semua halaman admin dan superadmin tampil sebagai koordinator', async ({ page, api }) => {
  await loginAs(page, 'koordinator');
  for (const [path, request] of ADMIN_PAGES) {
    await page.goto(path);
    await expect(page).toHaveURL(new RegExp(`${path}$`));
    await expect
      .poll(() => api.callsTo('GET', request).length, { message: path })
      .toBeGreaterThan(0);
    await page.waitForLoadState('networkidle');
  }
});

test('halaman admin tampil sebagai asisten', async ({ page, api }) => {
  await loginAs(page, 'asisten');
  for (const [path, request] of ADMIN_PAGES.slice(0, -1)) {
    await page.goto(path);
    await expect(page).toHaveURL(new RegExp(`${path}$`));
    await expect
      .poll(() => api.callsTo('GET', request).length, { message: path })
      .toBeGreaterThan(0);
    await page.waitForLoadState('networkidle');
  }
});

test('buat kelas', async ({ page, api }) => {
  api.on('POST', '/admin/kelas', null);
  await loginAs(page, 'asisten');
  await page.goto('/praktikum/admin/kelas');
  await page.locator('#n').fill('TTL C');
  await page.getByRole('button', { name: 'Simpan' }).click();
  await expect(page.getByText('Tersimpan.')).toBeVisible();
  expect(api.callsTo('POST', '/admin/kelas').map((c) => c.body)).toEqual([{ nama_kelas: 'TTL C' }]);
  await expect.poll(() => api.callsTo('GET', '/admin/kelas').length).toBeGreaterThan(1);
});

test('buat user', async ({ page, api }) => {
  api.on('POST', '/admin/users', null);
  await loginAs(page, 'asisten');
  await page.goto('/praktikum/admin/users');
  await page.getByRole('button', { name: /Tambah Mahasiswa/ }).click();
  await page.locator('#nim').fill('2027100099');
  await page.locator('#nama').fill('Mahasiswa Baru');
  await page.locator('#kelas').selectOption({ label: 'TTL B' });
  await page.locator('#shift').selectOption({ label: 'Shift 2' });
  await page.locator('#gelombang').selectOption({ label: 'Gelombang 1' });
  await page.locator('#kelompok').fill('B');
  await page.getByRole('button', { name: 'Simpan', exact: true }).click();
  await expect(page.getByText('Tersimpan.')).toBeVisible();
  expect(api.callsTo('POST', '/admin/users').map((c) => c.body)).toEqual([
    {
      nim: '2027100099',
      nama: 'Mahasiswa Baru',
      kelas_id: 2,
      shift: 2,
      gelombang: 1,
      kelompok: 'B',
    },
  ]);
});

test('simpan keaktifan rekap nilai', async ({ page, api }) => {
  api.on('POST', '/admin/keaktifan', null);
  await loginAs(page, 'asisten');
  await page.goto('/praktikum/admin/rekap-nilai');
  await page.locator('tbody input[type="number"]').first().fill('5');
  await page.getByRole('button', { name: 'Simpan Keaktifan (1)' }).click();
  await expect
    .poll(() => api.callsTo('POST', '/admin/keaktifan').map((c) => c.body))
    .toEqual([{ items: [{ pengerjaan_id: 10_001, nilai: 5 }] }]);
  await expect.poll(() => api.callsTo('GET', '/admin/rekap/kelas/1').length).toBeGreaterThan(1);
});

test('buka kunci peserta aktivasi', async ({ page, api }) => {
  api.on('POST', '/admin/aktivasi-course/buka-tutup-peserta', null);
  await loginAs(page, 'asisten');
  await page.goto('/praktikum/admin/aktivasi');
  await page.getByRole('button', { name: 'Detail' }).click();
  await page.getByLabel(`Pilih ${OTHER_USER_NAME}`).check();
  await page.getByRole('button', { name: 'Buka Kunci (1)' }).click();
  await page.getByRole('button', { name: 'Ya, Lanjutkan' }).click();
  await expect
    .poll(() => api.callsTo('POST', '/admin/aktivasi-course/buka-tutup-peserta').map((c) => c.body))
    .toEqual([{ aktivasi_course_id: 21, mahasiswa_ids: [otherUser.id], aksi: 'unlock' }]);
});

test('simpan permissions superadmin', async ({ page, api }) => {
  api.on('PUT', '/superadmin/permissions', null);
  await loginAs(page, 'koordinator');
  await page.goto('/praktikum/superadmin/pengaturan');
  await page.locator('label', { hasText: 'kelas:write' }).getByRole('button').click();
  await page.getByRole('button', { name: 'Simpan' }).click();
  await expect(page.getByText('Tersimpan.')).toBeVisible();
  expect(api.callsTo('PUT', '/superadmin/permissions').map((c) => c.body)).toEqual([
    { asisten: { ...permissions.asisten, 'kelas:write': false } },
  ]);
});
