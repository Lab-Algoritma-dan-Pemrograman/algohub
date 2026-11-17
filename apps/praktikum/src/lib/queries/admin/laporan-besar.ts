import {
  praktikumKeys,
  type BerkasUrl,
  type DownloadBatchResult,
  type LaporanBesar,
  type LaporanBesarTarget,
  type PengumpulanLaporanBesar,
} from '@algohub/api-client/praktikum';
import { createMutation } from '@tanstack/svelte-query';
import { api } from '#lib/api.ts';
import { invalidate, queryClient } from '#lib/query.ts';

const keys = praktikumKeys.admin.laporanBesar;
const base = '/admin/laporan-besar';

export type JenisBerkas = 'laporan' | 'analisa';

export const fetchLaporanBesar = () =>
  queryClient.fetchQuery({
    queryKey: keys.current,
    queryFn: async () => (await api.get<LaporanBesar | null>(base)) ?? null,
  });

export const fetchPengumpulan = (id: number) =>
  queryClient.fetchQuery({
    queryKey: keys.pengumpulan(id),
    queryFn: async () => {
      const res = await api.get<{ data: PengumpulanLaporanBesar[] } | PengumpulanLaporanBesar[]>(
        `${base}/${id}/pengumpulan`,
      );
      return Array.isArray(res) ? res : (res?.data ?? []);
    },
  });

// Signed URL berumur pendek, jadi tidak pernah dipakai ulang dari cache.
export const fetchBerkas = (
  id: number,
  mahasiswaId: number,
  aksi: 'preview' | 'download',
  jenis: JenisBerkas,
) =>
  queryClient.fetchQuery({
    queryKey: keys.berkas(id, mahasiswaId, aksi, jenis),
    queryFn: () =>
      api.get<BerkasUrl>(`${base}/${id}/pengumpulan/${mahasiswaId}/${aksi}?jenis=${jenis}`),
    staleTime: 0,
    gcTime: 0,
  });

const refresh = () => invalidate(keys.all);

export function createSaveLaporanBesarMutation() {
  return createMutation(() => ({
    mutationFn: ({ id, body }: { id: number | null; body: Record<string, unknown> }) =>
      id ? api.put<unknown>(`${base}/${id}`, body) : api.post<unknown>(base, body),
    onSuccess: refresh,
  }));
}

export function createSaveTargetsMutation() {
  return createMutation(() => ({
    mutationFn: ({ id, targets }: { id: number; targets: Omit<LaporanBesarTarget, 'kelas'>[] }) =>
      api.put<unknown>(`${base}/${id}/target`, { targets }),
    onSuccess: refresh,
  }));
}

export function createMarkPengumpulanMutation() {
  return createMutation(() => ({
    mutationFn: ({
      id,
      mahasiswaId,
      body,
    }: {
      id: number;
      mahasiswaId: number;
      body: { status: 'terkumpul' | 'belum'; catatan: string | undefined };
    }) => api.put<unknown>(`${base}/${id}/pengumpulan/${mahasiswaId}/mark`, body),
    onSuccess: refresh,
  }));
}

export function createSetOverrideMutation() {
  return createMutation(() => ({
    mutationFn: ({
      id,
      mahasiswaId,
      alasan,
    }: {
      id: number;
      mahasiswaId: number;
      alasan: string;
    }) => api.put<unknown>(`${base}/${id}/pengumpulan/${mahasiswaId}/override`, { alasan }),
    onSuccess: refresh,
  }));
}

export function createRevokeOverrideMutation() {
  return createMutation(() => ({
    mutationFn: ({ id, mahasiswaId }: { id: number; mahasiswaId: number }) =>
      api.del<unknown>(`${base}/${id}/pengumpulan/${mahasiswaId}/override`),
    onSuccess: refresh,
  }));
}

export function createDownloadBatchMutation() {
  return createMutation(() => ({
    mutationFn: ({ id, items }: { id: number; items: { mhs_id: number; jenis: string }[] }) =>
      api.post<{ data: DownloadBatchResult[] }>(`${base}/${id}/download-batch`, { items }),
  }));
}
