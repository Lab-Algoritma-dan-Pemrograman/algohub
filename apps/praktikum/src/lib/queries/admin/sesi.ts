import { praktikumKeys, type Course, type Sesi, type Soal } from '@algohub/api-client/praktikum';
import { createMutation, queryOptions } from '@tanstack/svelte-query';
import { api } from '#lib/api.ts';
import { invalidate } from '#lib/query.ts';
import { createDeleteMutation, createSaveMutation, orEmpty } from './crud.ts';

export interface SesiBody {
  judul_sesi: string;
  deskripsi: string;
  urutan: number;
  is_ujian_praktik: boolean;
}

export interface CourseBody {
  jenis: string;
  judul: string;
  deskripsi: string;
  durasi_menit: number;
  kuota_easy: number | null;
  kuota_medium: number | null;
  kuota_hard: number | null;
}

export const sesiListQuery = () =>
  queryOptions({
    queryKey: praktikumKeys.admin.sesi.list,
    queryFn: async () => orEmpty(await api.get<Sesi[] | null>('/admin/sesi')),
  });

export const coursesQuery = (sesiId: number) =>
  queryOptions({
    queryKey: praktikumKeys.admin.sesi.courses(sesiId),
    queryFn: async () => orEmpty(await api.get<Course[] | null>(`/admin/sesi/${sesiId}/course`)),
  });

export const soalQuery = (courseId: number) =>
  queryOptions({
    queryKey: praktikumKeys.admin.soal(courseId),
    queryFn: async () => orEmpty(await api.get<Soal[] | null>(`/admin/soal?course_id=${courseId}`)),
  });

export const createSaveSesiMutation = () =>
  createSaveMutation<SesiBody>('/admin/sesi', praktikumKeys.admin.sesi.all);

// Sesi dan course terhapus beserta soal, aktivasi, dan jawaban di dalamnya.
export const createDeleteSesiMutation = () =>
  createDeleteMutation('/admin/sesi', praktikumKeys.admin.all);

export function createSaveCourseMutation() {
  return createMutation(() => ({
    mutationFn: ({ id, sesiId, body }: { id: number | null; sesiId: number; body: CourseBody }) =>
      id
        ? api.put<unknown>(`/admin/course/${id}`, body)
        : api.post<unknown>(`/admin/sesi/${sesiId}/course`, body),
    onSuccess: () => invalidate(praktikumKeys.admin.sesi.all),
  }));
}

export const createDeleteCourseMutation = () =>
  createDeleteMutation('/admin/course', praktikumKeys.admin.all);

export function createSaveSoalMutation() {
  return createMutation(() => ({
    mutationFn: ({ id, body }: { id: number | null; body: Record<string, unknown> }) =>
      id ? api.put<unknown>(`/admin/soal/${id}`, body) : api.post<unknown>('/admin/soal', body),
    onSuccess: (_data, { body }) => invalidate(praktikumKeys.admin.soal(Number(body.course_id))),
  }));
}

export function createDeleteSoalMutation() {
  return createMutation(() => ({
    mutationFn: ({ id }: { id: number; courseId: number }) => api.del<unknown>(`/admin/soal/${id}`),
    onSuccess: (_data, { courseId }) => invalidate(praktikumKeys.admin.soal(courseId)),
  }));
}
