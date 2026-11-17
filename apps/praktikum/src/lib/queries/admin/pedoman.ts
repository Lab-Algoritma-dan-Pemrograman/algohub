import { praktikumKeys, type Pedoman } from '@algohub/api-client/praktikum';
import { queryOptions } from '@tanstack/svelte-query';
import { api } from '#lib/api.ts';
import { createDeleteMutation, createSaveMutation, orEmpty } from './crud.ts';

export const pedomanQuery = () =>
  queryOptions({
    queryKey: praktikumKeys.admin.pedoman,
    queryFn: async () => orEmpty(await api.get<Pedoman[] | null>('/admin/pedoman')),
  });

const pedomanDependents = [praktikumKeys.admin.pedoman, praktikumKeys.info.laporan];

export const createSavePedomanMutation = () =>
  createSaveMutation<{ nama_dokumen: string; file_url: string }>(
    '/admin/pedoman',
    ...pedomanDependents,
  );

export const createDeletePedomanMutation = () =>
  createDeleteMutation('/admin/pedoman', ...pedomanDependents);
