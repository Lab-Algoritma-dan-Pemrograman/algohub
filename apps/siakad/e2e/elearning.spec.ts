import { expect, loginAs, setupMocks, test } from './helpers.ts';

const progress = (completion_percentage: number) => ({
  nim: '202511001',
  student_name: 'Praktikan Uji',
  completed_lessons: completion_percentage / 10,
  total_lessons: 10,
  completion_percentage,
  is_completed: false,
  completed_levels: [],
  current_level: 'Level 1',
  last_accessed_at: null,
});

test('event Realtime elearning_progress memicu sync ulang tanpa reload', async ({ page }) => {
  let percent = 20;
  const mocks = await setupMocks(page, 'praktikan', {
    'endpoint:/sync-elearning': () => progress(percent),
  });
  await loginAs(page, mocks);
  await page.getByLabel('Menu Utama').getByText('E-Learning', { exact: true }).click();
  await expect(page.getByTestId('elearning-percent')).toContainText('20');
  await mocks.realtime.waitForChannel(page, 'elearning-student-progress-sync');
  const syncs = () => mocks.api.endpoints.filter((e) => e.path === '/sync-elearning').length;
  const before = syncs();
  percent = 60;
  mocks.realtime.emit('elearning_progress', 'UPDATE', { nim: '202511001' });
  await expect(page.getByTestId('elearning-percent')).toContainText('60');
  expect(syncs()).toBeGreaterThan(before);
});
