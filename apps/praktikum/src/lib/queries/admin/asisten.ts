import { praktikumKeys, type AmpuanKelompok, type User } from '@algohub/api-client/praktikum';
import { queryOptions } from '@tanstack/svelte-query';
import { api } from '#lib/api.ts';
import { createDeleteMutation, createPostMutation, createSaveMutation, orEmpty } from './crud.ts';

export const asistenQuery = () =>
  queryOptions({
    queryKey: praktikumKeys.admin.asisten,
    queryFn: async () => orEmpty(await api.get<User[] | null>('/admin/asisten')),
  });

export const ampuanQuery = () =>
  queryOptions({
    queryKey: praktikumKeys.admin.ampuan,
    queryFn: async () => orEmpty(await api.get<AmpuanKelompok[] | null>('/admin/ampuan')),
  });

// Profil asisten dan ampuan tampil di /info/asisten dan /info/jadwal.
const asistenDependents = [
  praktikumKeys.admin.asisten,
  praktikumKeys.admin.ampuan,
  praktikumKeys.info.all,
];

export const createSaveAsistenMutation = () =>
  createSaveMutation<Record<string, unknown>>('/admin/asisten', ...asistenDependents);

export const createAddAmpuanMutation = () =>
  createPostMutation<{ asisten_id: number; kelas_id: number; kelompok: string }>(
    '/admin/ampuan',
    ...asistenDependents,
  );

export const createDeleteAmpuanMutation = () =>
  createDeleteMutation('/admin/ampuan', ...asistenDependents);

// Akun asisten dihapus lewat endpoint users, beserta ampuan miliknya.
export const createDeleteAsistenMutation = () =>
  createDeleteMutation('/admin/users', ...asistenDependents, praktikumKeys.admin.dashboard);
