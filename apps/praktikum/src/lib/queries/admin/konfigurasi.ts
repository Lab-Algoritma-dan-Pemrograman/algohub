import { praktikumKeys, type AiModelConfig, type Konfigurasi } from '@algohub/api-client/praktikum';
import { queryOptions } from '@tanstack/svelte-query';
import { api } from '#lib/api.ts';
import { queryClient } from '#lib/query.ts';
import { createPostMutation, orEmpty } from './crud.ts';

export const konfigurasiQuery = () =>
  queryOptions({
    queryKey: praktikumKeys.admin.konfigurasi,
    queryFn: async () => orEmpty(await api.get<Konfigurasi[] | null>('/admin/konfigurasi')),
  });

// Halaman pengaturan lama memuat konfigurasi tanpa cache (+page.ts "always fresh"), dan form
// jadwal/modul mengisi nilai awalnya sekali dari server.
export const fetchKonfigurasiFresh = () =>
  queryClient.fetchQuery({ ...konfigurasiQuery(), staleTime: 0 });

export const konfigurasiValue = (list: Konfigurasi[], key: string) =>
  list.find((k) => k.key === key)?.value;

export const aiModelQuery = () =>
  queryOptions({
    queryKey: praktikumKeys.admin.aiModel,
    queryFn: () => api.get<AiModelConfig>('/admin/konfigurasi/ai-model'),
  });

// Konfigurasi dibaca halaman info (jadwal, modul, pengumuman) dan model AI aktif.
export const createSaveKonfigurasiMutation = () =>
  createPostMutation<{ key: string; value: string }>(
    '/admin/konfigurasi',
    praktikumKeys.admin.konfigurasi,
    praktikumKeys.admin.aiModel,
    praktikumKeys.info.all,
  );
