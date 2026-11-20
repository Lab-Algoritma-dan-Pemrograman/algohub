import type { SiakadUser } from '@algohub/api-client/siakad';
import { createQuery } from '@tanstack/svelte-query';
import {
  checkAndNotifyUpcomingShifts,
  sendBrowserNotification,
  subscribeToWebPush,
} from './notifications.ts';
import { invalidateRpcs, rpcQueryOptions } from './queries/rpc.ts';
import { subscribe } from './realtime/client.ts';

export interface AssignmentRow {
  id: number;
  user_id: number;
  schedule_id: number | null;
  schedule_day?: string | null;
  schedule_start?: string | null;
  schedule_end?: string | null;
  schedule_title?: string | null;
  schedule_major?: string | null;
  schedule_class_code?: string | null;
  user_full_name?: string | null;
  original_user_id?: number | null;
  original_user_full_name?: string | null;
  [key: string]: unknown;
}

export const mapAssignment = (a: AssignmentRow) => ({
  ...a,
  schedule: {
    id: a.schedule_id,
    day_of_week: a.schedule_day,
    start_time: a.schedule_start,
    end_time: a.schedule_end,
    title: a.schedule_title,
    major: a.schedule_major,
    class_code: a.schedule_class_code,
  },
  user: { id: a.user_id, full_name: a.user_full_name },
  original_user: { id: a.original_user_id, full_name: a.original_user_full_name },
});

export type Assignment = ReturnType<typeof mapAssignment>;

export const assignmentsQuery = (
  viewerId: number,
  extra?: { refetchInterval?: number; refetchIntervalInBackground?: boolean },
) =>
  rpcQueryOptions<AssignmentRow[]>(
    'get_schedule_assignments_secure',
    { p_viewer_id: viewerId },
    extra,
  );

// Dipasang di layout area login: cek shift tiap 30 detik + Realtime penugasan dan reschedule.
export function useShiftNotifications(getUser: () => SiakadUser) {
  const isStaff = () => {
    const role = getUser().role;
    return role === 'asisten' || role === 'koordinator';
  };

  const query = createQuery(() => ({
    ...assignmentsQuery(getUser().id, {
      refetchInterval: 120_000,
      refetchIntervalInBackground: false,
    }),
    enabled: isStaff(),
  }));

  $effect(() => {
    if (query.error) console.error('Gagal memeriksa notifikasi jadwal jaga:', query.error);
  });

  $effect(() => {
    const data = query.data;
    if (data && isStaff()) checkAndNotifyUpcomingShifts(getUser(), data.map(mapAssignment));
  });

  $effect(() => {
    const user = getUser();
    void subscribeToWebPush(user);
    const offAssignments = subscribe(
      'global_jadwal_jaga_notif',
      { table: 'schedule_assignments' },
      (payload) => {
        if (isStaff()) void invalidateRpcs(['get_schedule_assignments_secure']);
        const next = payload.new as Record<string, unknown> | undefined;
        if (
          (payload.eventType === 'INSERT' || payload.eventType === 'UPDATE') &&
          next?.user_id === user.id
        ) {
          const isUpdate = payload.eventType === 'UPDATE';
          void sendBrowserNotification(
            isUpdate ? '✏ Perubahan Jadwal Jaga!' : '🚀 Penugasan Jadwal Jaga Baru!',
            {
              body: `Jadwal jaga kamu sebagai ${next.task_role || 'Petugas'} untuk ${next.activity_name || 'Praktikum'} telah ${isUpdate ? 'diperbarui' : 'ditambahkan'}.`,
              onClickUrl: '/jadwal-jaga',
            },
          );
        } else if (next?.status === 'mencari_pengganti' && next?.user_id !== user.id) {
          void sendBrowserNotification('🔄 Permintaan Tukar Jadwal Jaga!', {
            body: `Ada asisten yang sedang mencari pengganti jadwal jaga (Swap). Klik untuk melihat & membantu.`,
            onClickUrl: '/jadwal-jaga',
          });
        }
      },
    );
    const offAttendance = subscribe(
      'global_reschedule_praktikum_notif',
      { table: 'attendance_logs' },
      (payload) => {
        const next = payload.new as Record<string, unknown> | undefined;
        const prev = payload.old as Record<string, unknown> | undefined;
        if (
          next?.reschedule_status === 'pending' &&
          (user.role === 'asisten' || user.role === 'koordinator')
        ) {
          void sendBrowserNotification('📅 Permohonan Reschedule Praktikum Baru!', {
            body: `Ada pengajuan reschedule jadwal praktikum baru yang membutuhkan verifikasi.`,
            onClickUrl: '/validasi-absensi',
          });
        }
        if (
          next?.user_id === user.id &&
          payload.eventType === 'UPDATE' &&
          next?.reschedule_status &&
          next.reschedule_status !== prev?.reschedule_status
        ) {
          const isApproved = next.reschedule_status === 'approved';
          void sendBrowserNotification(
            isApproved ? '✅ Reschedule Praktikum Disetujui!' : '❌ Reschedule Praktikum Ditolak',
            {
              body: isApproved
                ? `Pengajuan reschedule jadwal praktikum kamu telah disetujui!`
                : `Pengajuan reschedule jadwal praktikum kamu ditolak/dikembalikan.`,
              onClickUrl: '/absensi',
            },
          );
        }
      },
    );
    return () => {
      offAssignments();
      offAttendance();
    };
  });
}
