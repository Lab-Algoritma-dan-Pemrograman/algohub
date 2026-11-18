import { elearningKeys } from '@algohub/api-client/elearning';
import { queryOptions } from '@tanstack/svelte-query';
import { api } from '#lib/api.ts';
import { fetchAllPages } from '#lib/paginate.ts';
// Level dan modul tidak membawa jawaban dan sama untuk semua peran, jadi aman dipersist untuk
// tampilan offline. Lesson tidak dipersist: untuk admin, kordas, dan asisten backend mengembalikan
// baris lengkap (quiz.correctAnswer dan solution), dan 03-frontend.md melarang kunci jawaban
// tersimpan di IndexedDB.
export const levelsQuery = () =>
  queryOptions({
    queryKey: elearningKeys.curriculum.levels(),
    queryFn: () => fetchAllPages((page, limit) => api.curriculum.levels({ page, limit })),
    meta: { persist: true },
  });
export const modulesQuery = () =>
  queryOptions({
    queryKey: elearningKeys.curriculum.modules(),
    queryFn: () => fetchAllPages((page, limit) => api.curriculum.modules({ page, limit })),
    meta: { persist: true },
  });
export const lessonsQuery = () =>
  queryOptions({
    queryKey: elearningKeys.curriculum.lessons(),
    queryFn: () => fetchAllPages((page, limit) => api.curriculum.lessons({ page, limit })),
  });
