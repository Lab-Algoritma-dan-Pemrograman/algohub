import { elearningKeys } from '@algohub/api-client/elearning';
import { queryOptions } from '@tanstack/svelte-query';
import { api } from '#lib/api.ts';
import { fetchAllPages } from '#lib/paginate.ts';

// Contoh kode dikelola admin dan hanya dibuka dari modal, jadi tidak dipersist.
export const playgroundExamplesQuery = (language: string) =>
  queryOptions({
    queryKey: elearningKeys.playground.examples(language),
    queryFn: () =>
      fetchAllPages((page, limit) => api.playground.examples({ page, limit, language })),
  });
