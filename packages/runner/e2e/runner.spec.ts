import { existsSync, readdirSync, statSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { expect, test, type Page } from '@playwright/test';
const DIST = fileURLToPath(new URL('../harness/dist/', import.meta.url));
const output = (page: Page) => page.locator('[data-terminal-output]');
async function runSample(page: Page, sample: string) {
  await page.getByTestId('sample').selectOption(sample);
  await page.getByTestId('run').click();
}
async function typeLine(page: Page, text: string) {
  await page.getByTestId('terminal').locator('.xterm').click();
  await page.keyboard.type(text);
  await page.keyboard.press('Enter');
}
async function runTwoScanf(page: Page) {
  await runSample(page, 'c-two-scanf');
  await expect(output(page)).toContainText('Masukkan a:', { timeout: 120_000 });
  await typeLine(page, '3');
  await expect(output(page)).toContainText('Masukkan b:');
  await typeLine(page, '4');
  await expect(output(page)).toContainText('Hasil: 7');
  await expect(output(page)).toContainText('Program selesai');
}
test.beforeEach(async ({ page }) => {
  await page.goto('/');
});
test('halaman crossOriginIsolated', async ({ page }) => {
  expect(await page.evaluate(() => crossOriginIsolated)).toBe(true);
  await expect(page.getByTestId('isolated')).toHaveText('true');
});
test('C dengan dua scanf menerima input interaktif', async ({ page }) => {
  await runTwoScanf(page);
});
test('C dengan error kompilasi menampilkan diagnostik clang', async ({ page }) => {
  await runSample(page, 'c-compile-error');
  await expect(output(page)).toContainText('Compilation Error', { timeout: 120_000 });
  await expect(output(page)).toContainText('error:');
});
test('Python input() juga di dalam def', async ({ page }) => {
  await runSample(page, 'python-input');
  await expect(output(page)).toContainText('Nama:', { timeout: 120_000 });
  await typeLine(page, 'Budi');
  await expect(output(page)).toContainText('Kota:');
  await typeLine(page, 'Bandung');
  await expect(output(page)).toContainText('Halo Budi dari Bandung');
  await expect(output(page)).toContainText('Program selesai');
});
test('C infinite loop di-reset setelah 40 detik lalu program berikutnya jalan', async ({
  page,
}) => {
  await runSample(page, 'c-infinite-loop');
  await expect(output(page)).toContainText('Mulai loop', { timeout: 120_000 });
  const started = Date.now();
  await expect(output(page)).toContainText('tidak merespons dalam 40 detik', { timeout: 60_000 });
  expect(Date.now() - started).toBeGreaterThanOrEqual(39_000);
  await expect(page.getByTestId('run')).toBeEnabled();
  await runTwoScanf(page);
});
test('C mode batch membaca stdin tanpa input interaktif', async ({ page }) => {
  await page.getByTestId('stdin').fill('3\n4');
  await runSample(page, 'c-two-scanf');
  await expect(output(page)).toContainText('Hasil: 7', { timeout: 120_000 });
  await expect(output(page)).toContainText('Program selesai');
});
test('Python mode batch membaca stdin baris per baris', async ({ page }) => {
  await page.getByTestId('stdin').fill('Budi\nBandung');
  await runSample(page, 'python-input');
  await expect(output(page)).toContainText('Halo Budi dari Bandung', { timeout: 120_000 });
  await expect(output(page)).toContainText('Program selesai');
});
test('Stop menghentikan C yang berputar lalu program berikutnya jalan', async ({ page }) => {
  await runSample(page, 'c-infinite-loop');
  await expect(output(page)).toContainText('Mulai loop', { timeout: 120_000 });
  await page.getByTestId('stop').click();
  await expect(output(page)).toContainText('dihentikan oleh pengguna');
  await expect(page.getByTestId('run')).toBeEnabled();
  await runTwoScanf(page);
});
test('Stop membatalkan Python yang menunggu input()', async ({ page }) => {
  await runSample(page, 'python-input');
  await expect(output(page)).toContainText('Nama:', { timeout: 120_000 });
  await page.getByTestId('stop').click();
  await expect(output(page)).toContainText('dihentikan oleh pengguna');
  await expect(page.getByTestId('run')).toBeEnabled();
});
test('Monaco dimuat dari CDN jsDelivr dan tidak ikut bundle', async ({ page }) => {
  const monacoRequest = page.waitForRequest(/cdn\.jsdelivr\.net\/npm\/monaco-editor/);
  await page.reload();
  await monacoRequest;
  await expect(page.locator('.monaco-editor').first()).toBeVisible({ timeout: 60_000 });
  expect(existsSync(DIST), 'jalankan `bun run build` dulu agar harness/dist ada').toBe(true);
  const files = readdirSync(DIST, { recursive: true, encoding: 'utf8' }).map((f) =>
    f.replaceAll('\\', '/'),
  );
  const allowed = (f: string) =>
    f === 'runner/pyodide-worker.js' || /^assets\/monaco-loader-[\w-]+\.js$/.test(f);
  const bundled = files.filter((f) => /monaco|pyodide|\.wasm$|\.whl$/i.test(f) && !allowed(f));
  expect(bundled).toEqual([]);
  for (const f of files.filter((name) => name.endsWith('.js'))) {
    expect(statSync(DIST + f).size, f).toBeLessThan(1_000_000);
  }
});
