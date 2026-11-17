import { praktikumKeys, type RolePermissions } from '@algohub/api-client/praktikum';
import { createMutation } from '@tanstack/svelte-query';
import { api } from '#lib/api.ts';
import { invalidate, queryClient } from '#lib/query.ts';

const permissions = praktikumKeys.superadmin.permissions;

// Toggle izin diedit lokal lalu disimpan sekaligus, jadi nilai awal diambil sekali dari server.
export const fetchPermissions = () =>
  queryClient.fetchQuery({
    queryKey: permissions,
    queryFn: () => api.get<RolePermissions | null>('/superadmin/permissions'),
    staleTime: 0,
  });

export function createSavePermissionsMutation() {
  return createMutation(() => ({
    mutationFn: (body: RolePermissions) => api.put<unknown>('/superadmin/permissions', body),
    onSuccess: () => invalidate(permissions),
  }));
}
