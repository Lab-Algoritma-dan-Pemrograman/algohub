import type { UploadResponse } from '@algohub/api-client/praktikum';
import { createMutation } from '@tanstack/svelte-query';
import { api } from '#lib/api.ts';

function toForm(input: FormData | { file: File; folder: string }): FormData {
  if (input instanceof FormData) return input;
  const fd = new FormData();
  fd.append('file', input.file);
  fd.append('folder', input.folder);
  return fd;
}

// RichTextEditor menyusun FormData sendiri (folder soal-images), halaman lain cukup file + folder.
export function createAdminUploadMutation() {
  return createMutation(() => ({
    mutationFn: (input: FormData | { file: File; folder: string }) =>
      api.upload<UploadResponse>('/admin/upload', toForm(input)),
  }));
}
