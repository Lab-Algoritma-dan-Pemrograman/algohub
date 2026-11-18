import type { Page } from '@playwright/test';
import { expect, loginAs, test } from './test.ts';

const output = (page: Page) => page.locator('[data-terminal-output]');

async function typeLine(page: Page, text: string) {
  await page.locator('.xterm').click();
  await page.keyboard.type(text);
  await page.keyboard.press('Enter');
}

test('contoh kode dimuat per bahasa dan "Gunakan" menutup modal', async ({ page, api }) => {
  await loginAs(page, 'mahasiswa');
  await page.goto('/playground');
  await page.getByRole('button', { name: 'Contoh Kode' }).click();
  const dialog = page.getByRole('dialog', { name: /Contoh Kode/ });
  await expect(dialog.getByText('Halo Dunia Python')).toBeVisible();
  expect(api.callsTo('GET', '/playground/examples')[0].search.get('language')).toBe('python');
  await dialog.getByRole('button', { name: 'Gunakan' }).click();
  await expect(dialog).toBeHidden();
});

test.describe('compiler asli', () => {
  test.use({ realCompiler: true });
  test.skip(
    !process.env.ALGOHUB_COMPILER_DIR,
    'ALGOHUB_COMPILER_DIR belum diisi (bun run fetch-compiler .cache/compiler/v1)',
  );

  async function runtimeReady(page: Page) {
    await expect(page.getByText(/Menyiapkan/)).toHaveCount(0, { timeout: 150_000 });
  }

  test('Python input() interaktif di terminal', async ({ page }) => {
    await loginAs(page, 'mahasiswa');
    await page.goto('/playground');
    expect(await page.evaluate(() => crossOriginIsolated)).toBe(true);
    await runtimeReady(page);
    await page.getByRole('button', { name: /Jalankan/ }).click();
    await expect(output(page)).toContainText('Siapa nama kamu?', { timeout: 120_000 });
    await typeLine(page, 'Budi');
    await expect(output(page)).toContainText('Halo, Budi! Selamat datang!');
  });

  test('C scanf interaktif di terminal', async ({ page }) => {
    await loginAs(page, 'mahasiswa');
    await page.goto('/playground');
    expect(await page.evaluate(() => crossOriginIsolated)).toBe(true);
    await page
      .getByRole('group', { name: 'Bahasa' })
      .getByRole('button', { name: 'C', exact: true })
      .click();
    await runtimeReady(page);
    await page.getByRole('button', { name: /Jalankan/ }).click();
    await expect(output(page)).toContainText('Siapa nama kamu?', { timeout: 120_000 });
    await typeLine(page, 'Budi');
    await expect(output(page)).toContainText('Halo, Budi! Selamat datang!');
  });
});
