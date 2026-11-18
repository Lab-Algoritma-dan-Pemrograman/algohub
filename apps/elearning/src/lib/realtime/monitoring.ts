import {
  elearningKeys,
  type ActiveSessionRole,
  type ActivityLog,
  type ElearningUser,
  type MonitoringProgress,
  type Page,
} from '@algohub/api-client/elearning';
import type { QueryClient, QueryKey } from '@tanstack/svelte-query';
import { ACTIVITY_LOG_PAGE_SIZE } from '#lib/queries/monitoring.ts';
import { queryClient } from '#lib/query.ts';
import { subscribe, type ChangePayload } from './client.ts';

const LOG_PREFIX = ['elearning', 'activity-logs'] as const;
const INVALIDATE_DELAY_MS = 500;

const pending = new Map<string, ReturnType<typeof setTimeout>>();

// Hapus massal (reset progres) memicu puluhan event DELETE tanpa NIM; satu refetch setelah
// jeda cukup dan tidak menumpuk permintaan.
function invalidateSoon(qc: QueryClient, queryKey: QueryKey) {
  const id = JSON.stringify(queryKey);
  clearTimeout(pending.get(id));
  pending.set(
    id,
    setTimeout(() => {
      pending.delete(id);
      void qc.invalidateQueries({ queryKey });
    }, INVALIDATE_DELAY_MS),
  );
}

const rowOf = (payload: ChangePayload): Record<string, unknown> =>
  (payload.eventType === 'DELETE' ? payload.old : payload.new) as Record<string, unknown>;

const text = (value: unknown): string | null => (typeof value === 'string' ? value : null);

// Event active_sessions hanya membawa baris tabel tanpa join role. Role diambil dari entri
// lama atau daftar mahasiswa; bila tidak ketemu, hanya key sesi yang diambil ulang.
export function applySessionChange(qc: QueryClient, payload: ChangePayload): void {
  const key = elearningKeys.monitoring.sessions();
  const row = rowOf(payload);
  const nim = text(row.nim);
  if (!nim) return invalidateSoon(qc, key);

  if (payload.eventType === 'DELETE') {
    qc.setQueryData<ActiveSessionRole[]>(key, (old) => old?.filter((s) => s.nim !== nim));
    return;
  }

  const current = qc.getQueryData<ActiveSessionRole[]>(key);
  if (!current) return;
  const students = qc.getQueryData<ElearningUser[]>(elearningKeys.monitoring.students());
  const known = current.find((s) => s.nim === nim)?.users;
  const role = known?.role ?? students?.find((s) => s.nim === nim)?.role;
  if (!known && !role) return invalidateSoon(qc, key);

  const next: ActiveSessionRole = {
    nim,
    last_heartbeat: text(row.last_heartbeat),
    users: known ?? { role: role ?? null },
  };
  qc.setQueryData<ActiveSessionRole[]>(key, [...current.filter((s) => s.nim !== nim), next]);
}

// student_progress: lesson selesai ditambahkan ke daftar NIM-nya, dibatalkan dicabut. DELETE
// hanya membawa kunci primer, jadi tanpa nim dan lesson_id cache diambil ulang.
export function applyProgressChange(qc: QueryClient, payload: ChangePayload): void {
  const key = elearningKeys.monitoring.progress();
  const row = rowOf(payload);
  const nim = text(row.nim);
  const lessonId = text(row.lesson_id);
  if (!nim || !lessonId) return invalidateSoon(qc, key);

  const done = payload.eventType !== 'DELETE' && row.completed === true;
  qc.setQueryData<MonitoringProgress[]>(key, (old) => {
    if (!old) return old;
    const entry = old.find((p) => p.nim === nim);
    const rest = old.filter((p) => p.nim !== nim);
    const ids = (entry?.lessonIds ?? []).filter((id) => id !== lessonId);
    if (done) ids.push(lessonId);
    return ids.length > 0 ? [...rest, { nim, lessonIds: ids }] : rest;
  });
}

function logListKeys(qc: QueryClient) {
  return qc.getQueryCache().findAll({ queryKey: LOG_PREFIX });
}

// activity_logs: log baru masuk di atas halaman pertama; log yang dihapus dicabut dari semua
// halaman yang ada di cache. Halaman lain ditandai basi tanpa langsung diambil ulang.
export function applyActivityLogChange(qc: QueryClient, payload: ChangePayload): void {
  if (payload.eventType === 'INSERT') {
    const log = payload.new as unknown as ActivityLog;
    qc.setQueryData<Page<ActivityLog>>(elearningKeys.activityLogs.list(1), (old) =>
      old
        ? {
            ...old,
            total: old.total + 1,
            items: [log, ...old.items].slice(0, old.limit || ACTIVITY_LOG_PAGE_SIZE),
          }
        : old,
    );
    void qc.invalidateQueries({
      queryKey: LOG_PREFIX,
      predicate: (query) => query.queryKey[2] !== 1,
      refetchType: 'none',
    });
    return;
  }
  if (payload.eventType === 'DELETE') {
    const id = text(rowOf(payload).id);
    if (!id) return invalidateSoon(qc, [...LOG_PREFIX]);
    for (const query of logListKeys(qc)) {
      qc.setQueryData<Page<ActivityLog>>(query.queryKey, (old) =>
        old?.items.some((l) => l.id === id)
          ? {
              ...old,
              total: Math.max(0, old.total - 1),
              items: old.items.filter((l) => l.id !== id),
            }
          : old,
      );
    }
    return;
  }
  // UPDATE tidak dipakai app ini; ambil ulang daftar yang terbuka.
  invalidateSoon(qc, [...LOG_PREFIX]);
}

// Layar monitoring: sesi aktif dan progres belajar.
export function watchMonitoring(qc: QueryClient = queryClient): () => void {
  const stops = [
    subscribe('monitoring:active_sessions', { table: 'active_sessions' }, (p) =>
      applySessionChange(qc, p),
    ),
    subscribe('monitoring:student_progress', { table: 'student_progress' }, (p) =>
      applyProgressChange(qc, p),
    ),
  ];
  return () => stops.forEach((stop) => stop());
}

export function watchActivityLogs(qc: QueryClient = queryClient): () => void {
  return subscribe('audit:activity_logs', { table: 'activity_logs' }, (p) =>
    applyActivityLogChange(qc, p),
  );
}
