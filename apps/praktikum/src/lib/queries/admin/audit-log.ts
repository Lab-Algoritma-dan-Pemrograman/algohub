import { praktikumKeys, type AuditLogsResponse } from '@algohub/api-client/praktikum';
import { keepPreviousData, queryOptions } from '@tanstack/svelte-query';
import { api } from '#lib/api.ts';

export interface AuditLogFilter {
  search: string;
  role: string;
  action: string;
  page: number;
  limit: number;
}

export const auditLogsQuery = (filter: AuditLogFilter) =>
  queryOptions({
    queryKey: praktikumKeys.admin.auditLogs({ ...filter }),
    queryFn: () => {
      const query = new URLSearchParams({
        search: filter.search.trim(),
        role: filter.role,
        action: filter.action,
        page: filter.page.toString(),
        limit: filter.limit.toString(),
      });
      return api.get<AuditLogsResponse | null>(`/admin/audit-logs?${query.toString()}`);
    },
    // Baris halaman sebelumnya tetap tampil di bawah overlay "Memuat" selama halaman baru diambil.
    placeholderData: keepPreviousData,
  });
