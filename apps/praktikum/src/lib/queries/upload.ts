import type { UploadResponse } from '@algohub/api-client/praktikum';
import { createMutation } from '@tanstack/svelte-query';
import { api } from '#lib/api.ts';

export function createAdminUploadMutation() {
  return createMutation(() => ({
    mutationFn: ({ file, folder }: { file: File; folder: string }) => {
      const fd = new FormData();
      fd.append('file', file);
      fd.append('folder', folder);
      return api.upload<UploadResponse>('/admin/upload', fd);
    },
  }));
}
