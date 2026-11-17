import { praktikumKeys, type AdminStatistik } from '@algohub/api-client/praktikum';
import { queryOptions } from '@tanstack/svelte-query';
import { api } from '#lib/api.ts';

export const dashboardQuery = () =>
  queryOptions({
    queryKey: praktikumKeys.admin.dashboard,
    queryFn: () => api.get<AdminStatistik>('/admin/dashboard'),
  });
