import { createMutation } from '@tanstack/svelte-query';
import { api } from '#lib/api.ts';

export function createLogoutMutation() {
  return createMutation(() => ({
    mutationFn: () => api.post<unknown>('/auth/logout'),
  }));
}
