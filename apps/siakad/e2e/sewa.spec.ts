import { expect, loginAs, setupMocks, test } from './helpers.ts';

test('penyewa mengajukan sewa barang', async ({ page }) => {
  const mocks = await setupMocks(page, 'penyewa', {
    get_renter_items_secure: [
      {
        id: 5,
        name: 'Proyektor',
        quantity: 3,
        price_per_day: 50000,
        location: 'Lab 1',
        condition: 'Baik',
      },
    ],
    get_my_rentals_secure: [],
    get_smart_validation_contact_secure: [{ phone_number: '081200000999' }],
    submit_rental_request_secure: null,
  });
  await loginAs(page, mocks);
  await page.getByLabel('Menu Utama').getByText('Sewa & Pinjam', { exact: true }).click();
  await expect(page).toHaveURL(/\/sewa-barang$/);
  await page.getByRole('button', { name: 'Ajukan', exact: true }).click();
  await expect(page.getByText('Form Pengajuan Proyektor')).toBeVisible();
  await page.getByLabel('Catatan / Tujuan (Opsional)').fill('Seminar');
  const rentalsBefore = mocks.api.callsTo('get_my_rentals_secure').length;
  await page.getByRole('button', { name: 'Ajukan Permintaan' }).click();
  await expect(page.getByText('Hore! Aksi kamu berhasil dilaksanakan! 🥳🚀')).toBeVisible();
  expect(mocks.api.callsTo('submit_rental_request_secure')[0]).toMatchObject({
    p_user_id: 201,
    p_item_id: 5,
    p_quantity: 1,
    p_type: 'sewa',
    p_notes: 'Seminar',
  });
  await expect
    .poll(() => mocks.api.callsTo('get_my_rentals_secure').length)
    .toBeGreaterThan(rentalsBefore);
});
