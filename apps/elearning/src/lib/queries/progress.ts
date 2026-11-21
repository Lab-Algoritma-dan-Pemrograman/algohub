import { elearningKeys } from '@algohub/api-client/elearning';
import { queryOptions } from '@tanstack/svelte-query';
import { api } from '#lib/api.ts';
import { fetchAllPages } from '#lib/paginate.ts';
import { queryClient } from '#lib/query.ts';
import { applyXpGain } from '#lib/rewards.ts';
import { announceAchievements } from './achievements.ts';
// Progres milik user tidak dipersist: berubah tiap lesson selesai dan mudah basi di perangkat bersama.
export const completedLessonsQuery = () =>
  queryOptions({
    queryKey: elearningKeys.progress.lessons(),
    queryFn: async () => {
      // Backend mengembalikan Page<string> berisi lesson_id yang sudah difilter completed = true.
      return fetchAllPages<string>((page, limit) => api.progress.myLessons({ page, limit }));
    },
  });
export const recentProgressQuery = (limit = 5) =>
  queryOptions({
    queryKey: elearningKeys.progress.recent(),
    queryFn: () => api.progress.recent(limit),
  });

// Setelah lesson selesai, daftar selesai dan riwayat terbaru diambil ulang; mutasi baru selesai
// setelah refetch supaya penguncian level berikutnya sudah memakai data terbaru.
export const completeLessonOptions = () => ({
  mutationFn: ({ lessonId, xpReward }: { lessonId: string; xpReward: number }) =>
    api.progress.completeLesson(lessonId, xpReward),
  onSuccess: () =>
    Promise.all([
      queryClient.invalidateQueries({ queryKey: elearningKeys.progress.lessons() }),
      queryClient.invalidateQueries({ queryKey: elearningKeys.progress.recent() }),
    ]),
});

// Pemberian XP manual (POST /progress/xp). Level naik memicu popup, achievement XP baru
// ikut diumumkan dari respons yang sama.
export const grantXpOptions = () => ({
  mutationFn: (amount: number) => api.progress.grantXp(amount),
  onSuccess: (res: Awaited<ReturnType<typeof api.progress.grantXp>>) => {
    applyXpGain(res);
    if (res.achievement) announceAchievements([res.achievement]);
  },
});
