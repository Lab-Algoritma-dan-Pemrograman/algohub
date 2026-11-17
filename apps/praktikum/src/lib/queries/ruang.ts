import { praktikumKeys, type RuangCourse } from '@algohub/api-client/praktikum';
import { createMutation, queryOptions } from '@tanstack/svelte-query';
import { api } from '#lib/api.ts';
import { queryClient } from '#lib/query.ts';

// Soal dan jawaban berubah selama pengerjaan (timer, buka/tutup asisten), jadi tidak pernah
// dianggap segar dan tidak dipersist (tanpa meta.persist).
export const ruangQuery = (aktivasiSesiId: number, courseId: number) =>
  queryOptions({
    queryKey: praktikumKeys.mahasiswa.ruang(aktivasiSesiId, courseId),
    queryFn: () =>
      api.get<RuangCourse>(`/ruang?aktivasi_sesi_id=${aktivasiSesiId}&course_id=${courseId}`),
    staleTime: 0,
  });

export function createMulaiMutation() {
  return createMutation(() => ({
    mutationFn: (body: { aktivasi_sesi_id: number; course_id: number; token: string | null }) =>
      api.post<RuangCourse>('/mulai', body),
    onSuccess: (data, body) =>
      queryClient.setQueryData(
        praktikumKeys.mahasiswa.ruang(body.aktivasi_sesi_id, body.course_id),
        data,
      ),
  }));
}

export function createAutosaveMutation() {
  return createMutation(() => ({
    mutationFn: (body: { soal_terpilih_id: number; jawaban_teks: string }) =>
      api.post<unknown>('/autosave', body),
  }));
}

export function createSubmitMutation() {
  return createMutation(() => ({
    mutationFn: (body: { aktivasi_sesi_id: number; course_id: number }) =>
      api.post<unknown>('/submit', body),
    onSuccess: (_data, body) =>
      Promise.all([
        queryClient.invalidateQueries({
          queryKey: praktikumKeys.mahasiswa.ruang(body.aktivasi_sesi_id, body.course_id),
        }),
        queryClient.invalidateQueries({ queryKey: praktikumKeys.mahasiswa.dashboard }),
        queryClient.invalidateQueries({ queryKey: praktikumKeys.mahasiswa.sesi }),
      ]),
  }));
}
