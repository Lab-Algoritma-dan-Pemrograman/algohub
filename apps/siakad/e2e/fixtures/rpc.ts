import { DIVISION_ACCESS, type Role } from './users.ts';

// Respons bawaan per RPC; spec menimpa yang dibutuhkan. Bentuk baris dari types.ts siakad lama.
export function defaultRpc(role: Role): Record<string, unknown> {
  return {
    get_public_settings: [{ active_shift: 'all', announcement: 'Selamat datang di uji e2e' }],
    get_division_access_secure: DIVISION_ACCESS[role].map((menu_key) => ({ menu_key })),
    check_pj_absen_today: false,
    get_dashboard_stats_secure: {
      user_phone: '081200000000',
      total_kelas: 1,
      attendance_rate: 90,
      total_feedback: 2,
      total_students_under_me: 12,
      total_classes_under_me: 1,
      my_attendance_count: 3,
      upcoming_shifts_count: 1,
      total_users: 50,
      total_assistants: 10,
      total_students: 40,
      pending_attendance: 4,
      pending_assistant_requests: 1,
    },
    get_schedule_assignments_secure: [],
    get_attendance_logs_secure: [],
    get_personal_schedules_secure: [],
    save_push_subscription_secure: null,
    delete_push_subscription_secure: null,
    update_user_profile_secure: [{ full_name: 'Nama Baru' }],
    update_password: null,
  };
}
