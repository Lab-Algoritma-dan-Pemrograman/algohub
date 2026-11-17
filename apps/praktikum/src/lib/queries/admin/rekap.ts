import {
  praktikumKeys,
  type RekapJawabanResponse,
  type RekapKelasResponse,
} from '@algohub/api-client/praktikum';
import { queryOptions } from '@tanstack/svelte-query';
import { api } from '#lib/api.ts';

export const rekapKelasQuery = (kelasId: number) =>
  queryOptions({
    queryKey: praktikumKeys.admin.rekap.kelas(kelasId),
    queryFn: () => api.get<RekapKelasResponse>(`/admin/rekap/kelas/${kelasId}`),
  });

export interface RekapJawabanFilter {
  kelas_id: number;
  sesi_id: number;
  jenis: string;
  search: string;
}

// Parameter kosong tidak dikirim, sama dengan URLSearchParams yang dibangun halaman lama.
function rekapJawabanSearch(f: RekapJawabanFilter): string {
  const params = new URLSearchParams();
  if (f.kelas_id > 0) params.append('kelas_id', f.kelas_id.toString());
  if (f.sesi_id > 0) params.append('sesi_id', f.sesi_id.toString());
  if (f.jenis) params.append('jenis', f.jenis);
  if (f.search) params.append('search', f.search);
  return params.toString();
}

export const rekapJawabanQuery = (filter: RekapJawabanFilter) =>
  queryOptions({
    queryKey: praktikumKeys.admin.rekap.jawaban({ ...filter }),
    queryFn: () =>
      api.get<RekapJawabanResponse>(`/admin/rekap-jawaban?${rekapJawabanSearch(filter)}`),
  });
