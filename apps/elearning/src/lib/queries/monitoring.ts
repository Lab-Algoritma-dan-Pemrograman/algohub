import { elearningKeys } from '@algohub/api-client/elearning';
import { queryOptions } from '@tanstack/svelte-query';
import { api } from '#lib/api.ts';
import { fetchAllPages } from '#lib/paginate.ts';
import { queryClient } from '#lib/query.ts';

// Data monitoring berisi identitas mahasiswa dan diperbarui Realtime, jadi tidak dipersist.
export const monitoringSessionsQuery = () =>
  queryOptions({
    queryKey: elearningKeys.monitoring.sessions(),
    queryFn: () => fetchAllPages((page, limit) => api.monitoring.sessions({ page, limit })),
  });

export const monitoringStudentsQuery = () =>
  queryOptions({
    queryKey: elearningKeys.monitoring.students(),
    queryFn: () => fetchAllPages((page, limit) => api.monitoring.students({ page, limit })),
  });

// Seluruh progres per NIM sekaligus (kolom persen dan urut progres), bukan per mahasiswa.
export const monitoringProgressQuery = () =>
  queryOptions({
    queryKey: elearningKeys.monitoring.progress(),
    queryFn: () =>
      fetchAllPages((page, limit) => api.monitoring.progress(undefined, { page, limit })),
  });

export const ACTIVITY_LOG_PAGE_SIZE = 50;

export const activityLogsQuery = (page: number) =>
  queryOptions({
    queryKey: elearningKeys.activityLogs.list(page),
    queryFn: () => api.activityLogs.list({ page, limit: ACTIVITY_LOG_PAGE_SIZE }),
  });

// Sesi lebih tua dari ambang dibersihkan backend; dipanggil saat refresh dan tiap 45 detik.
export const STALE_SESSION_MINUTES = 2;
export async function cleanupStaleSessions(): Promise<void> {
  await api.monitoring.cleanupStale(STALE_SESSION_MINUTES);
  await queryClient.invalidateQueries({ queryKey: elearningKeys.monitoring.sessions() });
}

export const monitoringActions = {
  async adjustXp(nim: string, xp: number) {
    await api.admin.adjustXp(nim, xp);
    await queryClient.invalidateQueries({ queryKey: elearningKeys.monitoring.students() });
  },
  async resetUser(nim: string) {
    await api.admin.resetUser(nim);
    await Promise.all([
      queryClient.invalidateQueries({ queryKey: elearningKeys.monitoring.students() }),
      queryClient.invalidateQueries({ queryKey: elearningKeys.monitoring.progress() }),
    ]);
  },
  async deleteUser(nim: string) {
    await api.admin.deleteUser(nim);
    await Promise.all([
      queryClient.invalidateQueries({ queryKey: elearningKeys.monitoring.students() }),
      queryClient.invalidateQueries({ queryKey: elearningKeys.monitoring.progress() }),
    ]);
  },
};

export const activityLogActions = {
  // Backend menghapus satu log per permintaan; daftar diambil ulang sekali setelah semuanya selesai.
  async remove(ids: readonly string[]) {
    try {
      for (const id of ids) await api.activityLogs.remove(id);
    } finally {
      await queryClient.invalidateQueries({ queryKey: ['elearning', 'activity-logs'] });
    }
  },
};
