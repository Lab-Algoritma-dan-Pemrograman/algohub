import type { Page } from '@playwright/test';
import { expect, loginAs, test } from './test.ts';

test.use({ realCompiler: true });
test.skip(
  !process.env.ALGOHUB_COMPILER_DIR,
  'ALGOHUB_COMPILER_DIR belum diisi (bun run fetch-compiler .cache/compiler/v1)',
);

const runButton = (page: Page) => page.getByRole('button', { name: /Jalankan/ });

// Jalankan sebelum runtime siap hanya mengembalikan pesan "masih diunduh", jadi tunggu status hilang.
async function runtimeReady(page: Page) {
  await expect(page.getByText(/Menyiapkan/)).toHaveCount(0, { timeout: 150_000 });
  await expect(runButton(page)).toBeEnabled();
}

test('lesson C: scanf membaca stdin panel sandbox dengan compiler asli', async ({ page }) => {
  await loginAs(page, 'mahasiswa');
  await page.goto('/lesson/les-c');
  expect(await page.evaluate(() => crossOriginIsolated)).toBe(true);
  await runtimeReady(page);
  await page.locator('#sandbox-stdin').fill('3 4');
  await runButton(page).click();
  await expect(page.getByText('Hasil: 7')).toBeVisible({ timeout: 120_000 });
});

test('lesson Python: input() sampai kuis, latihan, dan selesai memberi XP', async ({
  page,
  api,
}) => {
  await loginAs(page, 'mahasiswa');
  await page.goto('/lesson/les-py');
  expect(await page.evaluate(() => crossOriginIsolated)).toBe(true);
  await runtimeReady(page);
  await page.locator('#sandbox-stdin').fill('Budi');
  await runButton(page).click();
  await expect(page.getByText('Nama: Halo Budi')).toBeVisible({ timeout: 120_000 });

  await page.getByRole('button', { name: /Ikuti Kuis/ }).click();
  await page.getByRole('button', { name: 'read()', exact: true }).click();
  await page.getByRole('button', { name: 'Periksa Jawaban' }).click();
  await expect(page.getByText(/Kurang tepat/)).toContainText('input()');
  await page.getByRole('button', { name: 'Coba Lagi' }).click();
  await page.getByRole('button', { name: 'input()', exact: true }).click();
  await page.getByRole('button', { name: 'Periksa Jawaban' }).click();
  await expect(page.getByText('Benar! Bagus sekali.')).toBeVisible();

  await expect(page.getByRole('heading', { name: 'Latihan Pemrograman' })).toBeVisible();
  await runButton(page).click();
  await expect(page.getByText('Berhasil!', { exact: true })).toBeVisible({ timeout: 120_000 });
  await page.getByRole('button', { name: 'Selesaikan Pelajaran' }).click();

  await expect(page.getByRole('dialog', { name: 'Kurikulum selesai' })).toBeVisible();
  expect(api.callsTo('POST', '/progress/lessons/les-py/complete')[0].body).toEqual({
    xpReward: 60,
  });
});
