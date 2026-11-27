import {
  praktikumKeys,
  type DashboardResponse,
  type LaporanBesarStatus,
  type Profile,
  type SesiUserItem,
  type UploadUrlResponse,
} from '@algohub/api-client/praktikum';
import { createMutation, queryOptions } from '@tanstack/svelte-query';
import { api } from '#lib/api.ts';
import { queryClient } from '#lib/query.ts';

// Pengganti swrGet lama: isi cache tampil seketika, lalu selalu diambil ulang saat halaman dibuka.
// staleTime 0 = data tetap dianggap basi tiap mount (revalidasi jalan), sedangkan IDB persist
// (meta.persist) memberi first paint instan dari cache kunjungan sebelumnya sehingga skeleton
// penuh tidak muncul lagi. Friskness data tidak berubah.
const ALWAYS_REVALIDATE = 0;
const persist = { persist: true };

export const mahasiswaQueries = {
  dashboard: () =>
    queryOptions({
      queryKey: praktikumKeys.mahasiswa.dashboard,
      queryFn: () => api.get<DashboardResponse>('/dashboard'),
      staleTime: ALWAYS_REVALIDATE,
      meta: persist,
    }),
  sesi: () =>
    queryOptions({
      queryKey: praktikumKeys.mahasiswa.sesi,
      queryFn: async () => (await api.get<SesiUserItem[] | null>('/sesi')) ?? [],
      staleTime: ALWAYS_REVALIDATE,
      meta: persist,
    }),
  profile: () =>
    queryOptions({
      queryKey: praktikumKeys.profile,
      queryFn: () => api.get<Profile>('/profile'),
      meta: { persist: true },
    }),
  laporanBesar: () =>
    queryOptions({
      queryKey: praktikumKeys.mahasiswa.laporanBesar,
      queryFn: () => api.get<LaporanBesarStatus>('/laporan-besar'),
    }),
};

export function createUpdateProfileMutation() {
  return createMutation(() => ({
    mutationFn: (body: Record<string, unknown>) => api.put<unknown>('/profile', body),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: praktikumKeys.profile }),
  }));
}

const refreshLaporanBesar = () =>
  queryClient.invalidateQueries({ queryKey: praktikumKeys.mahasiswa.laporanBesar });

export function createUploadLaporanMutation() {
  return createMutation(() => ({
    mutationFn: async (input: {
      jenis: 'laporan' | 'analisa';
      ext: string;
      file: File;
      contentType: string;
    }) => {
      const res = await api.post<UploadUrlResponse>('/laporan-besar/upload-url', {
        jenis: input.jenis,
        ext: input.ext,
      });
      // URL signed Supabase berada di luar backend, jadi tidak lewat klien API (tanpa Bearer/envelope).
      const put = await fetch(res.upload_url, {
        method: 'PUT',
        body: input.file,
        headers: { 'Content-Type': input.contentType },
      });
      if (!put.ok) {
        throw new Error(`Unggah gagal ke penyimpanan (HTTP ${put.status}).`);
      }
      return res.path;
    },
  }));
}

export function createKumpulLaporanMutation() {
  return createMutation(() => ({
    mutationFn: (body: { laporan_path: string | null; analisa_path: string | null }) =>
      api.post<unknown>('/laporan-besar/kumpul', body),
    onSuccess: refreshLaporanBesar,
  }));
}

export function createBatalKumpulLaporanMutation() {
  return createMutation(() => ({
    mutationFn: () => api.del<unknown>('/laporan-besar/kumpul'),
    onSuccess: refreshLaporanBesar,
  }));
}
