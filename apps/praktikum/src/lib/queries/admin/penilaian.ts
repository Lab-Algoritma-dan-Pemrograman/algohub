import {
  praktikumKeys,
  type AiGradeTargets,
  type PenilaianRekapResponse,
} from '@algohub/api-client/praktikum';
import { createMutation, queryOptions } from '@tanstack/svelte-query';
import { api } from '#lib/api.ts';
import { invalidate, queryClient } from '#lib/query.ts';
import { createPostMutation } from './crud.ts';

const penilaian = praktikumKeys.admin.penilaian;

// Key yang isinya berubah setiap kali nilai atau status jawaban berubah.
export const NILAI_KEYS = [
  penilaian.all,
  praktikumKeys.admin.rekap.all,
  praktikumKeys.admin.dashboard,
];

// Rekap dan target AI dipakai tepat setelah penilaian berjalan; data lama membuat jawaban yang
// sudah dinilai terkirim lagi ke AI.
export const penilaianRekapQuery = (aktivasiSesiId: number, courseId: number) =>
  queryOptions({
    queryKey: penilaian.rekap(aktivasiSesiId, courseId),
    queryFn: () =>
      api.get<PenilaianRekapResponse>(
        `/admin/penilaian/rekap?aktivasi_sesi_id=${aktivasiSesiId}&course_id=${courseId}`,
      ),
    staleTime: 0,
  });

export const fetchAiTargets = (aktivasiSesiId: number, courseId: number) =>
  queryClient.fetchQuery({
    queryKey: penilaian.aiTargets(aktivasiSesiId, courseId),
    queryFn: () =>
      api.get<AiGradeTargets>(
        `/admin/penilaian/ai-grade/targets?aktivasi_sesi_id=${aktivasiSesiId}&course_id=${courseId}`,
      ),
    staleTime: 0,
  });

export const createSaveNilaiMutation = () =>
  createPostMutation<{ jawaban_id: number; nilai: number; feedback: string | null }>(
    '/admin/penilaian',
    ...NILAI_KEYS,
  );

// Dipanggil berulang selama AI grading; cache disegarkan sekali lewat refreshNilai() di akhir.
export function createAiGradeOneMutation() {
  return createMutation(() => ({
    mutationFn: (jawabanId: number) =>
      api.post<unknown>('/admin/penilaian/ai-grade/one', { jawaban_id: jawabanId }),
  }));
}

export const refreshNilai = () => invalidate(...NILAI_KEYS);

export const createBulkActionMutation = () =>
  createPostMutation<{ action: 'reset_nilai' | 'delete' | 'buka_kunci'; jawaban_ids: number[] }>(
    '/admin/penilaian/bulk-action',
    praktikumKeys.admin.aktivasi.all,
    ...NILAI_KEYS,
  );

export const createKeaktifanMutation = () =>
  createPostMutation<{ items: { pengerjaan_id: number; nilai: number }[] }>(
    '/admin/keaktifan',
    ...NILAI_KEYS,
  );

export const createInjectJawabanMutation = () =>
  createPostMutation<{
    mahasiswa_id: number;
    soal_terpilih_id: number;
    jawaban_teks: string;
    auto_submit: boolean;
  }>('/admin/jawaban/inject', praktikumKeys.admin.aktivasi.all, ...NILAI_KEYS);
