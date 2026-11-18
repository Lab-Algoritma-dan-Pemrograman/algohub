import { activityLog, classmate, lessons, page as pageOf, userRow } from './fixtures.ts';
import {
  expect,
  loginAs,
  persistedQueryKeys,
  readAlgohubIdb,
  test,
  waitForPersist,
} from './test.ts';

test('hanya kurikulum dan papan total yang dipersist; lesson dan data pribadi staf tidak', async ({
  page,
  api,
}) => {
  api.on('GET', '/admin/users', () =>
    pageOf([userRow({ ...classmate, nama: 'Peserta Rahasia' }, { xp: 777 })]),
  );
  api.on('GET', '/monitoring/students', () =>
    pageOf([userRow({ ...classmate, nama: 'Monitor Rahasia' })]),
  );
  api.on('GET', '/activity-logs', () => pageOf([activityLog({ nama: 'Aktor Rahasia' })]));

  await loginAs(page, 'admin');
  await page.goto('/admin');
  await expect(page.getByText('Peserta Rahasia')).toBeVisible();
  await waitForPersist(page);
  await page.goto('/monitoring');
  await expect(page.getByText('Monitor Rahasia')).toBeVisible();
  await waitForPersist(page);
  await page.goto('/auditlog');
  await expect(page.getByText('Aktor Rahasia')).toBeVisible();
  await waitForPersist(page);
  await page.goto('/leaderboard');
  await page.getByRole('button', { name: 'Total', exact: true }).click();
  await expect(page.getByText('Citra Lestari').first()).toBeVisible();
  await waitForPersist(page);
  await page.goto('/profile');
  await expect(page.getByText('Langkah Pertama')).toBeVisible();
  await waitForPersist(page);

  const { entries } = await readAlgohubIdb(page);
  expect(entries.length).toBeGreaterThan(0);
  const keys = persistedQueryKeys(entries);
  expect(keys).toContain('["elearning","curriculum","levels"]');
  expect(keys).toContain('["elearning","curriculum","modules"]');
  expect(keys).not.toContain('["elearning","curriculum","lessons"]');
  const allowed = /^\["elearning","(curriculum","(levels|modules)"|leaderboard","list",)/;
  expect(keys.filter((key) => !allowed.test(key))).toEqual([]);

  const text = entries.map((e) => e.value).join('\n');
  for (const secret of ['Peserta Rahasia', 'Monitor Rahasia', 'Aktor Rahasia', 'token-']) {
    expect(text).not.toContain(secret);
  }
});
test('staf membuka kurikulum dan lesson: kunci jawaban dan solusi tidak masuk IndexedDB', async ({
  page,
  api,
}) => {
  api.on('GET', '/curriculum/lessons', () =>
    pageOf(
      lessons.map((lesson) => ({
        ...lesson,
        solution: 'SOLUSI-RAHASIA-STAF',
        quiz: { question: 'Q?', options: ['a', 'b'], correctAnswer: 1 },
      })),
    ),
  );
  await loginAs(page, 'admin');
  await page.goto('/courses');
  await expect(page.getByText('Menjumlahkan dengan scanf').first()).toBeVisible();
  await waitForPersist(page);
  await page.goto('/admin');
  await waitForPersist(page);
  expect(api.callsTo('GET', '/curriculum/lessons').length).toBeGreaterThan(0);
  const { entries } = await readAlgohubIdb(page);
  expect(persistedQueryKeys(entries)).toContain('["elearning","curriculum","levels"]');
  const text = entries.map((e) => e.value).join('\n');
  for (const marker of ['correctAnswer', 'SOLUSI-RAHASIA-STAF', 'scanf']) {
    expect(text).not.toContain(marker);
  }
});
