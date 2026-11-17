import {
  praktikumKeys,
  type AktivasiSesi,
  type PesertaAktivasi,
  type Susulan,
} from '@algohub/api-client/praktikum';
import { createMutation, queryOptions } from '@tanstack/svelte-query';
import { api } from '#lib/api.ts';
import { invalidate } from '#lib/query.ts';
import { createDeleteMutation, createPostMutation, orEmpty } from './crud.ts';
import { NILAI_KEYS } from './penilaian.ts';

const aktivasi = praktikumKeys.admin.aktivasi;

// Detail, susulan, dan peserta berubah karena praktikan dan asisten lain, jadi setiap kali dibuka
// selalu diambil ulang seperti app lama.
const LIVE = 0;

export const aktivasiListQuery = () =>
  queryOptions({
    queryKey: aktivasi.list,
    queryFn: async () => orEmpty(await api.get<AktivasiSesi[] | null>('/admin/aktivasi')),
  });

export const aktivasiDetailQuery = (id: number) =>
  queryOptions({
    queryKey: aktivasi.detail(id),
    queryFn: () => api.get<AktivasiSesi | null>(`/admin/aktivasi/${id}`),
    staleTime: LIVE,
  });

export const susulanQuery = (id: number) =>
  queryOptions({
    queryKey: aktivasi.susulan(id),
    queryFn: async () => orEmpty(await api.get<Susulan[] | null>(`/admin/aktivasi/${id}/susulan`)),
    staleTime: LIVE,
  });

export const pesertaQuery = (id: number) =>
  queryOptions({
    queryKey: aktivasi.peserta(id),
    queryFn: async () =>
      orEmpty(await api.get<PesertaAktivasi[] | null>(`/admin/aktivasi/${id}/peserta`)),
    staleTime: LIVE,
  });

export interface AktivasiBody {
  sesi_praktikum_id: number;
  kelas_id: number;
  shift: number;
  gelombang: number | null;
  gacha_pilihan: string;
}

export const createAktivasiMutation = () =>
  createPostMutation<AktivasiBody>('/admin/aktivasi', aktivasi.all, praktikumKeys.admin.dashboard);

// Menghapus aktivasi ikut menghapus jawaban, nilai, susulan, dan log praktikan.
export const createDeleteAktivasiMutation = () =>
  createDeleteMutation('/admin/aktivasi', praktikumKeys.admin.all);

export function createTokenMutation() {
  return createMutation(() => ({
    mutationFn: (id: number) => api.post<unknown>(`/admin/aktivasi/${id}/token`),
    onSuccess: () => invalidate(aktivasi.all),
  }));
}

// Menutup course men-submit otomatis jawaban praktikan, jadi nilai dan rekap ikut berubah.
export const createBukaTutupCourseMutation = () =>
  createPostMutation<{ aktivasi_course_id: number; is_open?: boolean; unlock_only?: boolean }>(
    '/admin/aktivasi-course/buka-tutup',
    aktivasi.all,
    ...NILAI_KEYS,
  );

export const createBukaTutupPesertaMutation = () =>
  createPostMutation<{
    aktivasi_course_id: number;
    mahasiswa_ids: number[];
    aksi: 'unlock' | 'submit';
  }>('/admin/aktivasi-course/buka-tutup-peserta', aktivasi.all, ...NILAI_KEYS);

export function createAddSusulanMutation() {
  return createMutation(() => ({
    mutationFn: ({ id, body }: { id: number; body: { mahasiswa_id: number; alasan: string } }) =>
      api.post<unknown>(`/admin/aktivasi/${id}/susulan`, body),
    onSuccess: (_data, { id }) => invalidate(aktivasi.susulan(id)),
  }));
}

export function createRemoveSusulanMutation() {
  return createMutation(() => ({
    mutationFn: ({ id, mahasiswaId }: { id: number; mahasiswaId: number }) =>
      api.del<unknown>(`/admin/aktivasi/${id}/susulan/${mahasiswaId}`),
    onSuccess: (_data, { id }) => invalidate(aktivasi.susulan(id)),
  }));
}
