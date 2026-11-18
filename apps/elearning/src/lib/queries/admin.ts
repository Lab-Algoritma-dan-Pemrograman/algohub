import {
  elearningKeys,
  type GameQuestionInput,
  type GameQuestionUpdate,
  type GameSettingsUpdate,
  type PlaygroundInput,
  type PlaygroundUpdate,
} from '@algohub/api-client/elearning';
import { keepPreviousData, queryOptions } from '@tanstack/svelte-query';
import { toRequest } from '#lib/admin/curriculum-draft.ts';
import type { Level } from '#lib/admin/types.ts';
import { api } from '#lib/api.ts';
import { fetchAllPages } from '#lib/paginate.ts';
import { queryClient } from '#lib/query.ts';

export const ADMIN_PAGE_SIZE = 25;

// Daftar user dan progresnya berisi data pribadi mahasiswa dan cepat basi, jadi tidak dipersist.
export const adminUsersQuery = (q: string, page: number) =>
  queryOptions({
    queryKey: elearningKeys.admin.users(q, page),
    queryFn: () => api.admin.users({ q, page, limit: ADMIN_PAGE_SIZE }),
    placeholderData: keepPreviousData,
  });

// Key hanya membedakan per NIM; semua halaman diambil sekaligus karena detail user perlu
// seluruh lesson yang selesai, urut terbaru dulu seperti query lama.
export const adminUserProgressQuery = (nim: string) =>
  queryOptions({
    queryKey: elearningKeys.admin.userProgress(nim),
    queryFn: async () => {
      const rows = await fetchAllPages((page, limit) =>
        api.admin.userProgress(nim, { page, limit }),
      );
      return rows
        .filter((row) => row.completed)
        .sort((a, b) => (b.completed_at ?? '').localeCompare(a.completed_at ?? ''));
    },
  });

const adminUsersPrefix = ['elearning', 'admin', 'users'] as const;
const leaderboardPrefix = ['elearning', 'leaderboard'] as const;
const monitoringPrefix = ['elearning', 'monitoring'] as const;

const invalidateAll = (keys: readonly (readonly unknown[])[]) =>
  Promise.all(keys.map((queryKey) => queryClient.invalidateQueries({ queryKey })));

// Progres, XP, dan penghapusan user mengubah daftar admin, monitoring, dan papan peringkat.
const afterUserChange = (nim: string) =>
  invalidateAll([
    adminUsersPrefix,
    ['elearning', 'admin', 'user-progress', nim],
    monitoringPrefix,
    leaderboardPrefix,
  ]);

export const adminUserActions = {
  async setRole(nim: string, role: string) {
    await api.admin.setRole(nim, role);
    await invalidateAll([adminUsersPrefix, monitoringPrefix]);
  },
  async setLevelAccess(nim: string, overrides: Record<string, string>) {
    await api.admin.setLevelAccess(nim, overrides);
    await invalidateAll([adminUsersPrefix]);
  },
  async resetUser(nim: string) {
    await api.admin.resetUser(nim);
    await afterUserChange(nim);
  },
  async resetLevel(nim: string, levelId: string) {
    await api.admin.resetLevel(nim, levelId);
    await afterUserChange(nim);
  },
  async resetLesson(nim: string, lessonId: string, lessonTitle?: string) {
    await api.admin.resetLesson(nim, { lessonId, lessonTitle });
    await afterUserChange(nim);
  },
  async adjustXp(nim: string, xp: number) {
    await api.admin.adjustXp(nim, xp);
    await afterUserChange(nim);
  },
  async deleteUser(nim: string) {
    await api.admin.deleteUser(nim);
    await afterUserChange(nim);
  },
};

const curriculumKeys = [
  elearningKeys.curriculum.levels(),
  elearningKeys.curriculum.modules(),
  elearningKeys.curriculum.lessons(),
  elearningKeys.progress.lessons(),
];

// Audit kurikulum ditulis klien seperti app lama; kegagalannya tidak membatalkan penyimpanan.
const auditCurriculum = (details: string) =>
  api.activityLogs.add({ eventType: 'curriculum_modified', details }).catch(() => undefined);

export const curriculumActions = {
  async save(draft: readonly Level[], details: string) {
    await api.curriculum.save(toRequest(draft));
    await auditCurriculum(details);
    await invalidateAll(curriculumKeys);
  },
  // Dipakai impor AI: kurikulum lama dikosongkan dulu, lalu diganti seluruhnya.
  async replace(draft: readonly Level[], details: string) {
    await api.curriculum.clear();
    await api.curriculum.save(toRequest(draft));
    await auditCurriculum(details);
    await invalidateAll(curriculumKeys);
  },
  async clear() {
    await api.curriculum.clear();
    await auditCurriculum('Mengosongkan seluruh kurikulum');
    await invalidateAll(curriculumKeys);
  },
};

const GAME_LANGUAGES = ['c', 'python'] as const;

// Admin melihat seluruh bank soal sekaligus; endpoint mengacak lalu memotong sesuai limit.
export const adminGameQuestionsQuery = () =>
  queryOptions({
    queryKey: [...elearningKeys.game.questions(), 'all'],
    queryFn: async () => {
      const lists = await Promise.all(
        GAME_LANGUAGES.map((language) => api.game.questions({ language, limit: 100 })),
      );
      return lists.flat();
    },
  });

const invalidateGame = () =>
  invalidateAll([['elearning', 'game', 'questions'], elearningKeys.game.plays()]);

export const gameActions = {
  async addQuestion(body: GameQuestionInput) {
    await api.game.addQuestion(body);
    await invalidateGame();
  },
  async updateQuestion(id: string, body: GameQuestionUpdate) {
    await api.game.updateQuestion(id, body);
    await invalidateGame();
  },
  async deleteQuestion(id: string) {
    await api.game.deleteQuestion(id);
    await invalidateGame();
  },
  async replaceQuestions(questions: GameQuestionInput[]) {
    await api.game.replaceQuestions(questions);
    await invalidateGame();
  },
  async seedQuestions(questions: GameQuestionInput[]) {
    const res = await api.game.seedQuestions(questions);
    await invalidateGame();
    return res.seeded;
  },
  async updateSettings(body: GameSettingsUpdate) {
    await api.game.updateSettings(body);
    await invalidateAll([elearningKeys.game.settings()]);
  },
};

export const adminPlaygroundQuery = () =>
  queryOptions({
    queryKey: [...elearningKeys.playground.examples(), 'all'],
    queryFn: () => fetchAllPages((page, limit) => api.playground.examples({ page, limit })),
  });

const invalidatePlayground = () => invalidateAll([['elearning', 'playground', 'examples']]);

export const playgroundActions = {
  async add(body: PlaygroundInput) {
    await api.playground.add(body);
    await invalidatePlayground();
  },
  async update(id: string, body: PlaygroundUpdate) {
    await api.playground.update(id, body);
    await invalidatePlayground();
  },
  async remove(id: string) {
    await api.playground.remove(id);
    await invalidatePlayground();
  },
};
