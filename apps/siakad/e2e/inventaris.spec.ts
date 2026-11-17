import { expect, loginAs, setupMocks, test } from './helpers.ts';

test('asisten divisi inventaris menambah barang', async ({ page }) => {
  const items = [
    {
      id: 1,
      name: 'Proyektor',
      condition: 'Baik',
      quantity: 2,
      location: 'Lab 1',
      is_rentable: false,
      price_per_day: 0,
    },
  ];
  const mocks = await setupMocks(page, 'asisten', {
    get_inventory_items_secure: () => items,
    upsert_inventory_item_secure: (params) => {
      items.push({
        id: 2,
        name: String(params.p_name),
        condition: String(params.p_condition),
        quantity: Number(params.p_quantity),
        location: String(params.p_location),
        is_rentable: Boolean(params.p_is_rentable),
        price_per_day: Number(params.p_price_per_day),
      });
      return null;
    },
  });
  await loginAs(page, mocks);
  await page.getByLabel('Menu Utama').getByText('Inventaris', { exact: true }).click();
  await expect(page.getByTestId('inventory-row')).toHaveCount(1);

  await page.getByRole('button', { name: 'Tambah Barang' }).click();
  await page.getByPlaceholder('Nama Barang').fill('Kabel HDMI');
  await page.getByPlaceholder('Jumlah').fill('5');
  await page.getByPlaceholder('Lokasi Penyimpanan').fill('Lemari A');
  await page.getByRole('switch').click();
  await page.getByPlaceholder('Contoh: 10000').fill('15000');
  await page.getByRole('button', { name: 'Simpan' }).click();
  await expect(page.getByText('Hore! Aksi kamu berhasil dilaksanakan! 🥳🚀')).toBeVisible();

  expect(mocks.api.callsTo('upsert_inventory_item_secure')).toEqual([
    {
      p_caller_id: 301,
      p_id: '0',
      p_name: 'Kabel HDMI',
      p_condition: 'Baik',
      p_quantity: 5,
      p_location: 'Lemari A',
      p_is_rentable: true,
      p_price_per_day: 15000,
    },
  ]);
  await expect(page.getByTestId('inventory-row')).toHaveCount(2);
});
