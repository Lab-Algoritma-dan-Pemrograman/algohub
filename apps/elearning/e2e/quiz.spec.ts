import { ApiFailure } from './mock-api.ts';
import { expect, loginAs, test } from './test.ts';

async function openQuiz(page: import('@playwright/test').Page) {
  await loginAs(page, 'mahasiswa');
  await page.goto('/lesson/les-c');
  await page.getByRole('button', { name: /Ikuti Kuis/ }).click();
  await expect(page.getByText('Format specifier untuk int?')).toBeVisible();
}

test('kuis divalidasi server: salah lalu benar membuka latihan', async ({ page, api }) => {
  await openQuiz(page);
  await page.getByRole('button', { name: '%s', exact: true }).click();
  await page.getByRole('button', { name: 'Periksa Jawaban' }).click();
  await expect(page.getByText(/Kurang tepat/)).toContainText('%d');

  await page.getByRole('button', { name: 'Coba Lagi' }).click();
  await page.getByRole('button', { name: '%d', exact: true }).click();
  await page.getByRole('button', { name: 'Periksa Jawaban' }).click();
  await expect(page.getByText('Benar! Bagus sekali.')).toBeVisible();
  await expect(page.getByRole('heading', { name: 'Latihan Pemrograman' })).toBeVisible();

  expect(api.callsTo('POST', '/validate-quiz').map((c) => c.body)).toEqual([
    { lessonId: 'les-c', selectedOption: 0 },
    { lessonId: 'les-c', selectedOption: 1 },
  ]);
});

test('memilih opsi tidak memanggil server; jawaban baru dikirim saat diperiksa', async ({
  page,
  api,
}) => {
  await openQuiz(page);
  await page.getByRole('button', { name: '%d', exact: true }).click();
  expect(api.callsTo('POST', '/validate-quiz')).toHaveLength(0);
  await page.getByRole('button', { name: 'Periksa Jawaban' }).click();
  await expect(page.getByText('Benar! Bagus sekali.')).toBeVisible();
  expect(api.callsTo('POST', '/validate-quiz')).toHaveLength(1);
});

test('soal kuis tidak ditemukan menampilkan pesan khusus', async ({ page, api }) => {
  api.on('POST', '/validate-quiz', () => {
    throw new ApiFailure(404, 'not found');
  });
  await openQuiz(page);
  await page.getByRole('button', { name: '%d', exact: true }).click();
  await page.getByRole('button', { name: 'Periksa Jawaban' }).click();
  await expect(page.getByRole('alert')).toContainText(
    'Soal kuis untuk pelajaran ini tidak ditemukan.',
  );
});
