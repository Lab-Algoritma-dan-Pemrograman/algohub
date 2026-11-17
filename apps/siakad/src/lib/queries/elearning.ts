import { ApiError } from '@algohub/api-client';
import { siakadKeys } from '@algohub/api-client/siakad';
import { queryOptions } from '@tanstack/svelte-query';
import { syncElearning } from '#lib/api.ts';

export interface ElearningProgress {
  [key: string]: unknown;
}

// Progres mahasiswa tidak pernah dipersist (meta tanpa persist).
export const elearningProgressQuery = (nim: string) =>
  queryOptions({
    queryKey: siakadKeys.elearningProgress(nim),
    queryFn: async () => {
      const res = await syncElearning<ElearningProgress & { error?: string }>();
      if (!res.ok) throw new ApiError(res.body?.error || `HTTP ${res.status}`, res.status);
      return res.body;
    },
  });
