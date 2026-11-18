import { fireEvent, render, screen, waitFor } from '@testing-library/svelte';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const api = vi.hoisted(() => ({
  api: {
    admin: {
      users: vi.fn(),
      userProgress: vi.fn(),
      setLevelAccess: vi.fn(),
      setRole: vi.fn(),
      resetUser: vi.fn(),
      resetLevel: vi.fn(),
      resetLesson: vi.fn(),
      adjustXp: vi.fn(),
      deleteUser: vi.fn(),
    },
    curriculum: {
      levels: vi.fn(),
      modules: vi.fn(),
      lessons: vi.fn(),
      save: vi.fn(),
      clear: vi.fn(),
    },
    activityLogs: { add: vi.fn() },
    game: {
      questions: vi.fn(),
      addQuestion: vi.fn(),
      updateQuestion: vi.fn(),
      deleteQuestion: vi.fn(),
      replaceQuestions: vi.fn(),
      seedQuestions: vi.fn(),
      updateSettings: vi.fn(),
    },
    playground: { examples: vi.fn(), add: vi.fn(), update: vi.fn(), remove: vi.fn() },
    storage: { uploadLessonImage: vi.fn() },
  },
}));
vi.mock('#lib/api.ts', () => api);

const { default: QueryHost } = await import('./QueryHost.svelte');
const { default: Admin } = await import('../src/routes/(app)/admin/+page.svelte');
const { adminUserActions, curriculumActions, gameActions, playgroundActions } =
  await import('../src/lib/queries/admin');
const { uploadLessonImage } = await import('../src/lib/admin/upload');
const { queryClient } = await import('../src/lib/query');

const page = <T>(items: T[], total = items.length) => ({ items, total, page: 1, limit: 25 });
const data = {
  auth: {
    user: {
      id: 1,
      nim: '999',
      nama: 'Admin',
      kelas: '',
      role: 'admin',
      jurusan: null,
      email: null,
      shift: null,
    },
  },
} as never;

const user = (nim: string, nama: string) => ({
  nim,
  id: null,
  nama,
  kelas: 'A1',
  jurusan: 'TI',
  email: null,
  role: 'praktikan',
  elearning_role: null,
  xp: 100,
  level: 1,
  streak: 0,
  study_time: 0,
  last_active: null,
  created_at: null,
  assessment_access: {},
  level_access_overrides: {},
  shift: null,
  prestige_tier: null,
  prestige_mult: 1,
  prestige_count: 0,
  last_prestige_at: null,
});

// jsdom tidak punya Element.animate yang dipakai transisi dialog.
Element.prototype.animate = vi.fn(() => ({
  cancel: vi.fn(),
  finish: vi.fn(),
  onfinish: null,
})) as never;

beforeEach(() => {
  queryClient.clear();
  vi.clearAllMocks();
  api.api.curriculum.levels.mockResolvedValue(page([]));
  api.api.curriculum.modules.mockResolvedValue(page([]));
  api.api.curriculum.lessons.mockResolvedValue(page([]));
  api.api.admin.userProgress.mockResolvedValue(page([]));
});

describe('/admin tabel peserta', () => {
  it('halaman dan pencarian dikirim ke server lewat ?q= dan page', async () => {
    api.api.admin.users.mockImplementation(async (p: { page: number; q: string }) =>
      page([user(`n${p.page}`, `Peserta ${p.page}${p.q ? ` (${p.q})` : ''}`)], 60),
    );
    render(QueryHost, { page: Admin, data });

    expect(await screen.findByText('Peserta 1')).toBeInTheDocument();
    expect(api.api.admin.users).toHaveBeenCalledWith({ q: '', page: 1, limit: 25 });
    expect(screen.getByText(/Halaman 1 dari 3/)).toBeInTheDocument();

    await fireEvent.click(screen.getByRole('button', { name: /Selanjutnya/ }));
    expect(await screen.findByText('Peserta 2')).toBeInTheDocument();
    expect(api.api.admin.users).toHaveBeenCalledWith({ q: '', page: 2, limit: 25 });

    await fireEvent.input(screen.getByPlaceholderText(/Cari nama atau email/), {
      target: { value: 'budi' },
    });
    await waitFor(
      () => expect(api.api.admin.users).toHaveBeenCalledWith({ q: 'budi', page: 1, limit: 25 }),
      { timeout: 2000 },
    );
    expect(await screen.findByText('Peserta 1 (budi)')).toBeInTheDocument();
  });

  it('mengubah XP lewat konfirmasi memanggil PUT xp dan mengambil ulang daftar', async () => {
    api.api.admin.users.mockResolvedValue(page([user('1', 'Ani')]));
    api.api.admin.adjustXp.mockResolvedValue({});
    render(QueryHost, { page: Admin, data });

    await fireEvent.click(await screen.findByText('Ani'));
    await fireEvent.input(await screen.findByPlaceholderText('XP baru'), {
      target: { value: '500' },
    });
    await fireEvent.click(screen.getByRole('button', { name: 'Simpan' }));
    await fireEvent.click(await screen.findByRole('button', { name: 'Lanjutkan' }));

    await waitFor(() => expect(api.api.admin.adjustXp).toHaveBeenCalledWith('1', 500));
    await waitFor(() => expect(api.api.admin.users.mock.calls.length).toBeGreaterThan(1));
  });
});

describe('aksi admin memanggil endpoint yang sesuai dan membatalkan query terkait', () => {
  const spy = () => vi.spyOn(queryClient, 'invalidateQueries');
  const invalidated = (s: ReturnType<typeof spy>) =>
    s.mock.calls.map(([filter]) => JSON.stringify(filter?.queryKey));

  it.each([
    ['setRole', () => adminUserActions.setRole('1', 'asisten'), 'setRole', ['1', 'asisten']],
    [
      'setLevelAccess',
      () => adminUserActions.setLevelAccess('1', { lv1: 'locked' }),
      'setLevelAccess',
      ['1', { lv1: 'locked' }],
    ],
    ['resetUser', () => adminUserActions.resetUser('1'), 'resetUser', ['1']],
    ['resetLevel', () => adminUserActions.resetLevel('1', 'lv1'), 'resetLevel', ['1', 'lv1']],
    [
      'resetLesson',
      () => adminUserActions.resetLesson('1', 'les1', 'Hello'),
      'resetLesson',
      ['1', { lessonId: 'les1', lessonTitle: 'Hello' }],
    ],
    ['adjustXp', () => adminUserActions.adjustXp('1', 10), 'adjustXp', ['1', 10]],
    ['deleteUser', () => adminUserActions.deleteUser('1'), 'deleteUser', ['1']],
  ])('%s', async (_name, run, method, args) => {
    const s = spy();
    api.api.admin[method as keyof typeof api.api.admin].mockResolvedValue({});
    await run();
    expect(api.api.admin[method as keyof typeof api.api.admin]).toHaveBeenCalledWith(...args);
    expect(invalidated(s)).toContain(JSON.stringify(['elearning', 'admin', 'users']));
  });

  it('reset progres juga menyegarkan progres user, monitoring, dan papan peringkat', async () => {
    const s = spy();
    api.api.admin.resetUser.mockResolvedValue({});
    await adminUserActions.resetUser('1');
    expect(invalidated(s)).toEqual(
      expect.arrayContaining([
        JSON.stringify(['elearning', 'admin', 'user-progress', '1']),
        JSON.stringify(['elearning', 'monitoring']),
        JSON.stringify(['elearning', 'leaderboard']),
      ]),
    );
  });

  it('simpan kurikulum memanggil PUT /curriculum, mencatat audit, dan menyegarkan kurikulum', async () => {
    const s = spy();
    api.api.curriculum.save.mockResolvedValue({});
    api.api.activityLogs.add.mockResolvedValue({});
    await curriculumActions.save(
      [
        {
          id: 'lv1',
          title: 'Level 1',
          description: 'd',
          modules: [
            {
              id: 'm1',
              title: 'M',
              lessons: [
                {
                  id: 'l1',
                  title: 'L',
                  explanation: 'e',
                  codeExample: '',
                  initialCode: '',
                  solution: '',
                  hint: '',
                  quiz: { question: 'q', options: ['a', 'b'], correctAnswer: 1 },
                  testCases: [],
                },
              ],
            },
          ],
        },
      ],
      'ubah',
    );
    expect(api.api.curriculum.save).toHaveBeenCalledWith({
      levels: [
        expect.objectContaining({
          id: 'lv1',
          accessMode: 'auto',
          locked: false,
          modules: [
            expect.objectContaining({
              lessons: [
                expect.objectContaining({
                  quiz: { question: 'q', options: ['a', 'b'], correctAnswer: 1 },
                  validationRules: [],
                }),
              ],
            }),
          ],
        }),
      ],
    });
    expect(api.api.activityLogs.add).toHaveBeenCalledWith({
      eventType: 'curriculum_modified',
      details: 'ubah',
    });
    expect(invalidated(s)).toContain(JSON.stringify(['elearning', 'curriculum', 'levels']));
  });

  it('kurikulum AI mengosongkan dulu sebelum menyimpan', async () => {
    const order: string[] = [];
    api.api.curriculum.clear.mockImplementation(async () => order.push('clear'));
    api.api.curriculum.save.mockImplementation(async () => order.push('save'));
    api.api.activityLogs.add.mockResolvedValue({});
    await curriculumActions.replace([], 'ai');
    expect(order).toEqual(['clear', 'save']);
  });

  it('soal game dan contoh playground memakai endpoint admin masing-masing', async () => {
    const s = spy();
    for (const fn of Object.values(api.api.game)) fn.mockResolvedValue({ seeded: true });
    for (const fn of Object.values(api.api.playground)) fn.mockResolvedValue({});
    await gameActions.addQuestion({ title: 't' });
    await gameActions.updateQuestion('q1', { title: 'u' });
    await gameActions.deleteQuestion('q1');
    await gameActions.replaceQuestions([]);
    expect(await gameActions.seedQuestions([])).toBe(true);
    await gameActions.updateSettings({ bugHuntWeeklyLimit: 2 });
    await playgroundActions.add({ title: 'p' });
    await playgroundActions.update('p1', { title: 'q' });
    await playgroundActions.remove('p1');
    expect(api.api.game.updateQuestion).toHaveBeenCalledWith('q1', { title: 'u' });
    expect(api.api.game.updateSettings).toHaveBeenCalledWith({ bugHuntWeeklyLimit: 2 });
    expect(api.api.playground.remove).toHaveBeenCalledWith('p1');
    expect(invalidated(s)).toEqual(
      expect.arrayContaining([
        JSON.stringify(['elearning', 'game', 'questions']),
        JSON.stringify(['elearning', 'playground', 'examples']),
      ]),
    );
  });
});

describe('unggah gambar pelajaran', () => {
  it('menolak format di luar JPG, PNG, GIF, dan WEBP', async () => {
    const file = new File(['x'], 'a.svg', { type: 'image/svg+xml' });
    await expect(uploadLessonImage(file)).rejects.toThrow('Format file tidak didukung');
    expect(api.api.storage.uploadLessonImage).not.toHaveBeenCalled();
  });

  it('mengirim file ke POST /storage/lesson-images dan mengembalikan URL publik', async () => {
    api.api.storage.uploadLessonImage.mockResolvedValue({ publicUrl: 'https://cdn/x.png' });
    const file = new File(['x'], 'a.png', { type: 'image/png' });
    expect(await uploadLessonImage(file)).toBe('https://cdn/x.png');
    expect(api.api.storage.uploadLessonImage).toHaveBeenCalledWith(file);
  });
});
