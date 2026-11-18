import { elearningKeys } from '@algohub/api-client/elearning';
import { shouldPersistQuery } from '@algohub/query';
import { dehydrate } from '@tanstack/svelte-query';
import { readdirSync, readFileSync } from 'node:fs';
import { resolve } from 'node:path';
import { beforeEach, describe, expect, it, vi } from 'vitest';
const api = vi.hoisted(() => ({
  api: {
    auth: { verify: vi.fn() },
    curriculum: { levels: vi.fn(), modules: vi.fn(), lessons: vi.fn() },
    progress: { myLessons: vi.fn(), recent: vi.fn() },
    playground: { examples: vi.fn() },
    leaderboard: { list: vi.fn(), weekly: vi.fn(), rank: vi.fn() },
    validateQuiz: vi.fn(),
    admin: { users: vi.fn(), userProgress: vi.fn() },
    monitoring: { sessions: vi.fn(), students: vi.fn(), progress: vi.fn() },
    activityLogs: { list: vi.fn() },
    game: { questions: vi.fn() },
  },
}));
vi.mock('#lib/api.ts', () => api);
const { levelsQuery, modulesQuery, lessonsQuery } = await import('../src/lib/queries/curriculum');
const { completedLessonsQuery, recentProgressQuery } = await import('../src/lib/queries/progress');
const { playgroundExamplesQuery } = await import('../src/lib/queries/playground');
const { leaderboardListQuery, leaderboardRankQuery, weeklyLeaderboardQuery } =
  await import('../src/lib/queries/leaderboard');
const { validateQuizOptions } = await import('../src/lib/queries/quiz');
const { adminUsersQuery, adminUserProgressQuery, adminGameQuestionsQuery, adminPlaygroundQuery } =
  await import('../src/lib/queries/admin');
const {
  activityLogsQuery,
  monitoringProgressQuery,
  monitoringSessionsQuery,
  monitoringStudentsQuery,
} = await import('../src/lib/queries/monitoring');
const { loadSession } = await import('../src/lib/auth');
const { queryClient } = await import('../src/lib/query');
const { session } = await import('../src/lib/session');
// 03-frontend.md § Persist: hanya data referensi yang tidak rahasia; sisanya wajib hilang saat
// cache disimpan ke IndexedDB.
const ALLOWED = [
  elearningKeys.curriculum.levels(),
  elearningKeys.curriculum.modules(),
  elearningKeys.curriculum.lessons(),
];
const FORBIDDEN = [
  elearningKeys.session(),
  elearningKeys.progress.lessons(),
  elearningKeys.progress.recent(),
  elearningKeys.activityLogs.list(),
  elearningKeys.monitoring.sessions(),
  elearningKeys.monitoring.students(),
  elearningKeys.monitoring.progress('2022110083'),
  elearningKeys.admin.users(),
  elearningKeys.admin.userProgress('2022110083'),
  elearningKeys.game.questions('c'),
  elearningKeys.game.plays(),
  elearningKeys.leaderboard.weekly(),
  elearningKeys.leaderboard.weekly({ kelas: 'A1', limit: 3 }),
  elearningKeys.leaderboard.rank(100, { kelas: 'A1' }),
  elearningKeys.leaderboard.filters(),
  elearningKeys.game.settings(),
  elearningKeys.achievements.catalog(),
  elearningKeys.achievements.mine(),
  elearningKeys.prestige.eligibility(),
  elearningKeys.playground.examples('c'),
  ['elearning', 'quiz', 'validate', 'lesson-1'],
];
const page = <T>(items: T[]) => ({ items, total: items.length, page: 1, limit: 100 });
const dehydratedKeys = () =>
  dehydrate(queryClient, { shouldDehydrateQuery: shouldPersistQuery }).queries.map((q) =>
    JSON.stringify(q.queryKey),
  );
beforeEach(() => {
  queryClient.clear();
  localStorage.clear();
  session.clear();
  vi.clearAllMocks();
});
describe('persist', () => {
  it('kurikulum ikut terdehydrate, sesi dan progres tidak', async () => {
    api.api.curriculum.levels.mockResolvedValue(page([{ id: 'l1' }]));
    api.api.curriculum.modules.mockResolvedValue(page([{ id: 'm1' }]));
    api.api.curriculum.lessons.mockResolvedValue(page([{ id: 'les1' }]));
    api.api.progress.myLessons.mockResolvedValue(
      page([
        { lesson_id: 'les1', completed: true },
        { lesson_id: 'les2', completed: false },
      ]),
    );
    api.api.progress.recent.mockResolvedValue([{ id: 'p1', lesson_id: 'les1' }]);
    session.set('tok', {
      id: 1,
      nim: '1',
      nama: 'U',
      kelas: '',
      role: 'praktikan',
      jurusan: null,
      email: null,
      shift: null,
    });
    api.api.auth.verify.mockResolvedValue({
      token: 'tok',
      payload: {
        nim: '1',
        nama: 'U',
        kelas: '',
        shift: null,
        jurusan: null,
        email: null,
        role: 'x',
      },
    });
    await queryClient.fetchQuery(levelsQuery());
    await queryClient.fetchQuery(modulesQuery());
    await queryClient.fetchQuery(lessonsQuery());
    await queryClient.fetchQuery(completedLessonsQuery());
    await queryClient.fetchQuery(recentProgressQuery());
    await loadSession();
    expect(await queryClient.fetchQuery(completedLessonsQuery())).toEqual(['les1']);
    expect(dehydratedKeys().sort()).toEqual(ALLOWED.map((key) => JSON.stringify(key)).sort());
  });
  it.each(FORBIDDEN.map((key) => [JSON.stringify(key), key] as const))(
    'query %s tidak dipersist walau sukses',
    async (_name, key) => {
      await queryClient.fetchQuery({ queryKey: key, queryFn: () => ({ rahasia: true }) });
      expect(dehydratedKeys()).toEqual([]);
    },
  );
  it('query gagal tidak dipersist walau meta persist menyala', async () => {
    api.api.curriculum.levels.mockRejectedValue(new Error('x'));
    await queryClient.fetchQuery({ ...levelsQuery(), retry: false }).catch(() => undefined);
    expect(dehydratedKeys()).toEqual([]);
  });
  it('contoh playground tidak dipersist walau berhasil dimuat', async () => {
    api.api.playground.examples.mockResolvedValue(page([{ id: 'ex1' }]));
    await queryClient.fetchQuery(playgroundExamplesQuery('c'));
    expect(dehydratedKeys()).toEqual([]);
  });

  it('validasi kuis (kunci jawaban) tidak masuk cache yang dipersist', async () => {
    api.api.validateQuiz.mockResolvedValue({ isCorrect: true, correctAnswer: 2 });
    const options = validateQuizOptions('lesson-1');
    expect(options.mutationKey).toEqual(['elearning', 'quiz', 'validate', 'lesson-1']);
    await queryClient.getMutationCache().build(queryClient, options).execute(2);
    expect(api.api.validateQuiz).toHaveBeenCalledWith({ lessonId: 'lesson-1', selectedOption: 2 });
    const state = dehydrate(queryClient, { shouldDehydrateQuery: shouldPersistQuery });
    expect(state.queries).toEqual([]);
    expect(JSON.stringify(state)).not.toContain('correctAnswer');
  });

  it('papan total ikut dipersist per filter, papan mingguan dan peringkat pribadi tidak', async () => {
    api.api.leaderboard.list.mockResolvedValue(page([{ nim: '1', nama: 'Ani', xp: 10 }]));
    api.api.leaderboard.weekly.mockResolvedValue(page([{ nim: '1', nama: 'Ani', weeklyXp: 10 }]));
    api.api.leaderboard.rank.mockResolvedValue({ rank: 3 });
    await queryClient.fetchQuery(leaderboardListQuery({ kelas: 'A1', limit: 20 }));
    await queryClient.fetchQuery(weeklyLeaderboardQuery({ kelas: 'A1', limit: 3 }));
    await queryClient.fetchQuery(leaderboardRankQuery(10, { kelas: 'A1' }));
    expect(dehydratedKeys()).toEqual([
      JSON.stringify(elearningKeys.leaderboard.list({ kelas: 'A1', limit: 20 })),
    ]);
  });

  it('query admin, monitoring, dan audit log (data pribadi) tidak dipersist walau sukses', async () => {
    api.api.admin.users.mockResolvedValue(page([{ nim: '1', nama: 'Ani' }]));
    api.api.admin.userProgress.mockResolvedValue(page([{ lesson_id: 'les1', completed: true }]));
    api.api.monitoring.sessions.mockResolvedValue(
      page([{ nim: '1', users: { role: 'praktikan' } }]),
    );
    api.api.monitoring.students.mockResolvedValue(page([{ nim: '1', nama: 'Ani' }]));
    api.api.monitoring.progress.mockResolvedValue(page([{ nim: '1', lessonIds: ['les1'] }]));
    api.api.activityLogs.list.mockResolvedValue(page([{ id: 'a', nim: '1', nama: 'Ani' }]));
    api.api.game.questions.mockResolvedValue([{ id: 'q1', code: 'x' }]);
    api.api.playground.examples.mockResolvedValue(page([{ id: 'ex1' }]));
    await queryClient.fetchQuery(adminUsersQuery('', 1));
    await queryClient.fetchQuery(adminUserProgressQuery('1'));
    await queryClient.fetchQuery(monitoringSessionsQuery());
    await queryClient.fetchQuery(monitoringStudentsQuery());
    await queryClient.fetchQuery(monitoringProgressQuery());
    await queryClient.fetchQuery(activityLogsQuery(1));
    await queryClient.fetchQuery(adminGameQuestionsQuery());
    await queryClient.fetchQuery(adminPlaygroundQuery());
    expect(queryClient.getQueryCache().getAll()).toHaveLength(8);
    expect(dehydratedKeys()).toEqual([]);
  });
  it('hanya queries/curriculum.ts dan queries/leaderboard.ts yang menandai persist: true', () => {
    const srcDir = resolve(process.cwd(), 'src');
    const marked = readdirSync(srcDir, { recursive: true, encoding: 'utf8' })
      .map((file) => file.replaceAll('\\', '/'))
      .filter((file) => /\.(ts|svelte)$/.test(file))
      .filter((file) => /persist:\s*true/.test(readFileSync(`${srcDir}/${file}`, 'utf8')));
    expect(marked.sort()).toEqual(['lib/queries/curriculum.ts', 'lib/queries/leaderboard.ts']);
  });
});
