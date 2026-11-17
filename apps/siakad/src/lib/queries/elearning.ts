import { ApiError } from '@algohub/api-client';
import { siakadKeys } from '@algohub/api-client/siakad';
import { queryOptions } from '@tanstack/svelte-query';
import { syncElearning } from '#lib/api.ts';

export interface ElearningProgress {
  nim: string;
  student_name: string | null;
  completed_lessons: number;
  total_lessons: number;
  completion_percentage: number;
  is_completed: boolean;
  completed_levels: string[] | null;
  current_level: string | null;
  last_accessed_at: string | null;
}

// Progres mahasiswa tidak pernah dipersist (meta tanpa persist).
export const elearningProgressQuery = (nim: string) =>
  queryOptions({
    queryKey: siakadKeys.elearningProgress(nim),
    queryFn: async () => {
      const res = await syncElearning<ElearningProgress & { error?: string }>();
      if (!res.ok) {
        const message = res.body ? res.body.error || `Server error: ${res.status}` : 'Server error';
        throw new ApiError(message, res.status);
      }
      return res.body as ElearningProgress;
    },
    enabled: !!nim,
  });
