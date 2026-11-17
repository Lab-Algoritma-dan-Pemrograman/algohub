import type { Page } from '@playwright/test';
import { expect, test } from './mock-api';

const output = (page: Page) => page.locator('[data-terminal-output]');

const TWO_SCANF = `#include <stdio.h>
int main() {
    int a, b;
    printf("Masukkan a: ");
    scanf("%d", &a);
    printf("Masukkan b: ");
    scanf("%d", &b);
    printf("Hasil: %d\\n", a + b);
    return 0;
}
`;

const PY_INPUT = `nama = input("Nama: ")
print("Halo " + nama)
`;

async function setCode(page: Page, code: string) {
  await page.waitForFunction(() => (window as { monaco?: unknown }).monaco !== undefined);
  await page.evaluate((src) => {
    const { monaco } = window as unknown as {
      monaco: { editor: { getModels(): { setValue(v: string): void }[] } };
    };
    monaco.editor.getModels()[0].setValue(src);
  }, code);
}

// Run sebelum runtime siap membuat RunnableEditor menulis pesan "masih diunduh" alih-alih menjalankan.
async function run(page: Page) {
  await expect(page.getByText('Menyiapkan')).toHaveCount(0, { timeout: 150_000 });
  await page.getByRole('button', { name: /Run/ }).click();
}

async function typeLine(page: Page, text: string) {
  await page.locator('.xterm').click();
  await page.keyboard.type(text);
  await page.keyboard.press('Enter');
}

test.beforeEach(async ({ page }) => {
  await page.goto('/info/praktikum');
});

test('halaman crossOriginIsolated', async ({ page }) => {
  expect(await page.evaluate(() => crossOriginIsolated)).toBe(true);
});

test('C starter mencetak Hello, Praktikum!', async ({ page }) => {
  await run(page);
  await expect(output(page)).toContainText('Hello, Praktikum!', { timeout: 120_000 });
  await expect(output(page)).toContainText('Program selesai');
});

test('C dengan dua scanf menerima input interaktif', async ({ page }) => {
  await setCode(page, TWO_SCANF);
  await run(page);
  await expect(output(page)).toContainText('Masukkan a:', { timeout: 120_000 });
  await typeLine(page, '3');
  await expect(output(page)).toContainText('Masukkan b:');
  await typeLine(page, '4');
  await expect(output(page)).toContainText('Hasil: 7');
  await expect(output(page)).toContainText('Program selesai');
});

test('C dengan error kompilasi menampilkan diagnostik clang', async ({ page }) => {
  await setCode(page, 'int main() { return x; }\n');
  await run(page);
  await expect(output(page)).toContainText('Compilation Error', { timeout: 120_000 });
  await expect(output(page)).toContainText('error:');
});

test('Python print', async ({ page }) => {
  await page.getByRole('button', { name: 'Python', exact: true }).click();
  await run(page);
  await expect(output(page)).toContainText('Hello, Praktikum!', { timeout: 120_000 });
  await expect(output(page)).toContainText('Program selesai');
});

test('Python input() interaktif', async ({ page }) => {
  await page.getByRole('button', { name: 'Python', exact: true }).click();
  await setCode(page, PY_INPUT);
  await run(page);
  await expect(output(page)).toContainText('Nama:', { timeout: 120_000 });
  await typeLine(page, 'Budi');
  await expect(output(page)).toContainText('Halo Budi');
  await expect(output(page)).toContainText('Program selesai');
});
