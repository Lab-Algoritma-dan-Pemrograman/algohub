import { rpcQueryOptions } from './rpc.ts';

export interface DashboardStats {
  user_phone?: string | null;
  total_kelas?: number;
  attendance_rate?: number;
  total_feedback?: number;
  total_students_under_me?: number;
  total_classes_under_me?: number;
  my_attendance_count?: number;
  upcoming_shifts_count?: number;
  total_users?: number;
  total_assistants?: number;
  total_students?: number;
  pending_attendance?: number;
  pending_assistant_requests?: number;
}

export const dashboardStatsQuery = (viewerId: number) =>
  rpcQueryOptions<DashboardStats | null>('get_dashboard_stats_secure', { p_viewer_id: viewerId });
