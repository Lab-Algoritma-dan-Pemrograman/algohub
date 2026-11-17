import { praktikumKeys, type Jadwal } from '@algohub/api-client/praktikum';
import { queryOptions } from '@tanstack/svelte-query';
import { api } from '#lib/api.ts';
import { createDeleteMutation, createSaveMutation, orEmpty } from './crud.ts';

export interface JadwalBody {
  kelas_id: number | null;
  shift: number;
  hari: string;
  jam_mulai: string;
  jam_selesai: string;
  keterangan: string;
}

export const jadwalQuery = () =>
  queryOptions({
    queryKey: praktikumKeys.admin.jadwal,
    queryFn: async () => orEmpty(await api.get<Jadwal[] | null>('/admin/jadwal')),
  });

const jadwalDependents = [praktikumKeys.admin.jadwal, praktikumKeys.info.jadwal];

export const createSaveJadwalMutation = () =>
  createSaveMutation<JadwalBody>('/admin/jadwal', ...jadwalDependents);

export const createDeleteJadwalMutation = () =>
  createDeleteMutation('/admin/jadwal', ...jadwalDependents);
