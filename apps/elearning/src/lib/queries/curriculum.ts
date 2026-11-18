import { elearningKeys } from '@algohub/api-client/elearning';
import { queryOptions } from '@tanstack/svelte-query';
import { api } from '#lib/api.ts';
import { fetchAllPages } from '#lib/paginate.ts';
// Kurikulum sama untuk semua mahasiswa dan tidak rahasia (kunci jawaban kuis dikosongkan backend
// untuk peran P), jadi aman dipersist untuk tampilan offline.
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
    meta: { persist: true },
  });
