import { createHash } from 'node:crypto';
import { readFileSync } from 'node:fs';
import { expect, test } from '@playwright/test';

const SUMS_URL =
  'https://github.com/Lab-Algoritma-dan-Pemrograman/algohub/releases/download/compiler-v1/SHA256SUMS';

test('landing menampilkan link ke ketiga app', async ({ page }) => {
  await page.goto('/');
  await expect(
    page.getByRole('heading', { level: 1, name: 'Laboratorium Algoritma & Pemrograman' }),
  ).toBeVisible();
  for (const [name, href] of [
    ['SIAKAD', 'https://siakad.algohub.web.id'],
    ['E-Learning', 'https://elearning.algohub.web.id'],
    ['Praktikum', 'https://praktikum.algohub.web.id'],
  ]) {
    await expect(page.getByRole('link', { name: new RegExp(`^${name}`) })).toHaveAttribute(
      'href',
      href,
    );
  }
});

test('halaman 404 punya link kembali ke beranda', async ({ page }) => {
  await page.goto('/404');
  await expect(page.getByRole('heading', { name: 'Halaman tidak ditemukan' })).toBeVisible();
  await page.getByRole('link', { name: 'Kembali ke Beranda' }).click();
  await expect(page).toHaveURL('/');
});

test('404.html memakai path absolut dan satu judul', async ({ page }) => {
  // vite preview tidak menyajikan 404.html di path bersarang seperti Vercel, jadi file hasil build yang diperiksa.
  const html = readFileSync(new URL('../build/404.html', import.meta.url), 'utf8');
  expect(html).not.toMatch(/(href|src)="\.\//);
  expect(html.match(/<title>/g)).toHaveLength(1);

  await page.goto('/404');
  await expect(page).toHaveTitle(/^Halaman Tidak Ditemukan/);
});

test('clang dilayani dengan byte yang cocok dengan SHA256SUMS', async ({ request }) => {
  const sums = await (await fetch(SUMS_URL)).text();
  const expected = /^([0-9a-f]{64})\s+\*?clang$/im.exec(sums)?.[1].toLowerCase();
  expect(expected).toBeDefined();

  const res = await request.get('/compiler/v1/clang');
  expect(res.status()).toBe(200);
  expect(
    createHash('sha256')
      .update(await res.body())
      .digest('hex'),
  ).toBe(expected);
});
