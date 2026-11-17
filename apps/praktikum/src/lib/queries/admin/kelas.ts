import { praktikumKeys, type Kelas } from '@algohub/api-client/praktikum';
import { queryOptions } from '@tanstack/svelte-query';
import { api } from '#lib/api.ts';
import { createDeleteMutation, createSaveMutation, orEmpty } from './crud.ts';

// Data referensi yang jarang berubah dan dipakai hampir semua halaman admin.
export const kelasQuery = () =>
  queryOptions({
    queryKey: praktikumKeys.admin.kelas,
    queryFn: async () => orEmpty(await api.get<Kelas[] | null>('/admin/kelas')),
    staleTime: 10 * 60_000,
    meta: { persist: true },
  });

// Nama kelas ikut tampil di users, jadwal, aktivasi, rekap, dan halaman info.
const kelasDependents = [praktikumKeys.admin.all, praktikumKeys.info.all];

export const createSaveKelasMutation = () =>
  createSaveMutation<{ nama_kelas: string }>('/admin/kelas', ...kelasDependents);

export const createDeleteKelasMutation = () =>
  createDeleteMutation('/admin/kelas', ...kelasDependents);
