import { elearningKeys } from '@algohub/api-client/elearning';
import { queryOptions } from '@tanstack/svelte-query';
import { api } from '#lib/api.ts';
import { fetchAllPages } from '#lib/paginate.ts';
import { queryClient } from '#lib/query.ts';
// Progres milik user tidak dipersist: berubah tiap lesson selesai dan mudah basi di perangkat bersama.
export const completedLessonsQuery = () =>
  queryOptions({
    queryKey: elearningKeys.progress.lessons(),
    queryFn: async () => {
      const rows = await fetchAllPages((page, limit) => api.progress.myLessons({ page, limit }));
      return rows.filter((row) => row.completed).map((row) => row.lesson_id);
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
