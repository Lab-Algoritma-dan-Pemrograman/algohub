import { praktikumKeys, type BulkResponse, type User } from '@algohub/api-client/praktikum';
import { createMutation, queryOptions } from '@tanstack/svelte-query';
import { api } from '#lib/api.ts';
import { invalidate } from '#lib/query.ts';
import { createDeleteMutation, createPostMutation, createSaveMutation, orEmpty } from './crud.ts';

export interface UserBody {
  nim: string;
  nama: string;
  kelas_id: number | null;
  shift: number;
  gelombang: number | null;
  kelompok: string | null;
}

export const usersQuery = () =>
  queryOptions({
    queryKey: praktikumKeys.admin.users,
    queryFn: async () => orEmpty(await api.get<User[] | null>('/admin/users')),
  });

// Menghapus/mereset mahasiswa mengubah statistik, peserta aktivasi, nilai, dan rekap.
const userDependents = [praktikumKeys.admin.all, praktikumKeys.info.all];

const resetPassword = (id: number) => api.post<unknown>(`/admin/users/${id}/reset-password`);

export const createSaveUserMutation = () =>
  createSaveMutation<UserBody>('/admin/users', ...userDependents);

export const createDeleteUserMutation = () =>
  createDeleteMutation('/admin/users', ...userDependents);

export function createResetPasswordMutation() {
  return createMutation(() => ({
    mutationFn: resetPassword,
    onSuccess: () => invalidate(...userDependents),
  }));
}

// Satu request per mahasiswa seperti app lama; cache baru disegarkan sekali setelah semuanya selesai.
export function createBulkUserActionMutation() {
  return createMutation(() => ({
    mutationFn: async ({ action, ids }: { action: 'delete' | 'reset_pw'; ids: number[] }) => {
      let ok = 0;
      let fail = 0;
      for (const id of ids) {
        try {
          if (action === 'delete') await api.del<unknown>(`/admin/users/${id}`);
          else await resetPassword(id);
          ok++;
        } catch {
          fail++;
        }
      }
      return { ok, fail };
    },
    onSuccess: () => invalidate(...userDependents),
  }));
}

export const createBulkImportUsersMutation = () =>
  createPostMutation<{ users: UserBody[] }, BulkResponse>('/admin/users/bulk', ...userDependents);

export function createBulkAttrsMutation() {
  return createMutation(() => ({
    mutationFn: (body: Record<string, unknown>) =>
      api.put<BulkResponse>('/admin/users/bulk-attrs', body),
    onSuccess: () => invalidate(...userDependents),
  }));
}
