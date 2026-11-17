import type { Page } from '@playwright/test';
import type { RuangCourse } from '../../../packages/api-client/src/praktikum/types';
import { JENIS, ruang, startedRuang, type JenisRoute } from './fixtures';
import { loginAs } from './helpers';
import { expect, test, type MockApi } from './mock-api';

const essay = (page: Page) => page.getByPlaceholder('Ketikkan jawaban essay Anda di sini...');

function mockRuang(api: MockApi, initial: RuangCourse) {
  let state = initial;
  api.on('GET', '/ruang', () => state);
  api.on('POST', '/mulai', () => (state = startedRuang(state)));
  api.on('POST', '/autosave', null);
  api.on('POST', '/submit', () => {
    state = { ...state, status: 'selesai' };
    return null;
  });
}

const openRuang = (page: Page, jenis: JenisRoute, r: RuangCourse) =>
  page.goto(`/praktikum/sesi/1/${jenis}?aktivasi=${r.aktivasi_sesi_id}&course=${r.course_id}`);

for (const jenis of JENIS) {
  test(`${jenis}: mulai, autosave, submit`, async ({ page, api }) => {
    const r = ruang(jenis);
    mockRuang(api, r);
    await loginAs(page, 'mahasiswa');
    await openRuang(page, jenis, r);

    if (jenis === 'ujian') {
      await expect(page.getByText('Sesi Ujian Terkunci')).toBeVisible();
      await page.getByPlaceholder('PIN 6 DIGIT').fill('abc123');
      await page.getByRole('button', { name: 'Mulai Ujian' }).click();
    }
    await expect
      .poll(() => api.callsTo('POST', '/mulai').map((c) => c.body))
      .toEqual([
        {
          aktivasi_sesi_id: r.aktivasi_sesi_id,
          course_id: r.course_id,
          token: jenis === 'ujian' ? 'ABC123' : null,
        },
      ]);

    await essay(page).fill(`Jawaban ${jenis}`);
    await essay(page).blur();
    await expect
      .poll(() => api.callsTo('POST', '/autosave').at(-1)?.body)
      .toEqual({ soal_terpilih_id: 501, jawaban_teks: `Jawaban ${jenis}` });

    await page.getByRole('button', { name: 'Submit Jawaban' }).click();
    await expect(page.getByText('Kumpulkan Jawaban?')).toBeVisible();
    await page.getByRole('button', { name: 'Ya, Lanjutkan' }).click();
    await expect(page).toHaveURL(/\/praktikum\/dashboard$/);
    expect(api.callsTo('POST', '/submit').map((c) => c.body)).toEqual([
      { aktivasi_sesi_id: r.aktivasi_sesi_id, course_id: r.course_id },
    ]);
  });
}

test('timer habis: pesan auto-submit, autosave terkirim, kembali ke dashboard', async ({
  page,
  api,
}) => {
  const r = ruang('pretest');
  let started: RuangCourse | null = null;
  api.on('GET', '/ruang', () => (started ??= startedRuang(r, 6_000)));
  api.on('POST', '/autosave', null);
  await loginAs(page, 'mahasiswa');
  await openRuang(page, 'pretest', r);
  // Pesan waktu habis langsung ditimpa "Tersimpan otomatis" oleh flush autosave (sama dengan
  // app lama), jadi ditangkap observer mutasi DOM, bukan polling visibilitas.
  const expiredShown = page.evaluate(
    (text) =>
      new Promise<boolean>((resolve) => {
        new MutationObserver((_, observer) => {
          if (document.body.textContent?.includes(text)) {
            observer.disconnect();
            resolve(true);
          }
        }).observe(document.body, { subtree: true, childList: true, characterData: true });
      }),
    'Waktu habis. Jawaban otomatis ter-submit.',
  );
  await essay(page).fill('Jawaban sebelum waktu habis');
  expect(await expiredShown).toBe(true);
  await expect(page).toHaveURL(/\/praktikum\/dashboard$/);
  expect(api.callsTo('POST', '/autosave').map((c) => c.body)).toContainEqual({
    soal_terpilih_id: 501,
    jawaban_teks: 'Jawaban sebelum waktu habis',
  });
  expect(api.callsTo('POST', '/submit')).toEqual([]);
});

test('course ditutup asisten: layar terkunci dan editor readonly', async ({ page, api }) => {
  const r = startedRuang(ruang('posttest'));
  mockRuang(api, { ...r, is_open: false });
  await loginAs(page, 'mahasiswa');
  await openRuang(page, 'posttest', r);
  await expect(page.getByText('Sesi ini sedang ditutup oleh asisten.')).toBeVisible();
  await expect(essay(page)).not.toBeEditable();
  await expect(page.getByRole('button', { name: 'Submit Jawaban' })).toHaveCount(0);
  expect(api.callsTo('POST', '/mulai')).toEqual([]);
});
