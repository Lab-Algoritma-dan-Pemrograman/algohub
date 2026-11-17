import { shouldPersistQuery } from '@algohub/query';
import { dehydrate } from '@tanstack/svelte-query';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const api = vi.hoisted(() => ({
  verify: vi.fn(),
  rpc: vi.fn(),
  logout: vi.fn(),
  syncElearning: vi.fn(),
}));

vi.mock('#lib/api.ts', () => api);
vi.mock('#lib/notifications.ts', () => ({ unsubscribeWebPush: vi.fn() }));
vi.mock('$app/navigation', () => ({ invalidateAll: vi.fn() }));

const { PERSISTED_RPCS, rpcQueryOptions } = await import('../src/lib/queries/rpc');
const { elearningProgressQuery } = await import('../src/lib/queries/elearning');
const { loadSession } = await import('../src/lib/auth');
const { queryClient } = await import('../src/lib/query');

// 03-frontend.md § Persist: hanya data referensi yang tidak rahasia.
const ALLOWED = [
  'get_public_settings',
  'get_schedules_secure',
  'get_external_links_secure',
  'get_renter_items_secure',
];

const FORBIDDEN = [
  'get_users_secure',
  'get_attendance_logs_secure',
  'get_audit_logs_secure',
  'get_financial_records_secure',
  'get_feedback_secure',
  'get_system_settings_full_secure',
  'get_dashboard_stats_secure',
  'get_division_access_secure',
  'get_group_members_secure',
  'get_schedule_assignments_secure',
  'get_admin_rentals_secure',
  'get_my_rentals_secure',
  'get_personal_schedules_secure',
  'get_qr_session_secure',
];

const dehydratedKeys = () =>
  dehydrate(queryClient, { shouldDehydrateQuery: shouldPersistQuery }).queries.map((q) =>
    JSON.stringify(q.queryKey),
  );

beforeEach(() => {
  queryClient.clear();
  vi.clearAllMocks();
});

describe('persist', () => {
  it('daftar persist hanya berisi data yang boleh', () => {
    expect([...PERSISTED_RPCS].sort()).toEqual([...ALLOWED].sort());
  });

  it.each(FORBIDDEN)('%s tidak ditandai persist', (fn) => {
    expect(rpcQueryOptions(fn).meta?.persist).not.toBe(true);
  });

  it('RPC terlarang, sesi, dan sync-elearning tidak ikut terdehydrate', async () => {
    api.rpc.mockResolvedValue({ data: [{ id: 1 }], error: null });
    api.verify.mockResolvedValue({
      ok: true,
      status: 200,
      body: { user: { id: 1, username: 'k', full_name: 'K', role: 'koordinator' } },
    });
    api.syncElearning.mockResolvedValue({ ok: true, status: 200, body: { progress: 10 } });
    for (const fn of [...FORBIDDEN, ...ALLOWED]) {
      await queryClient.fetchQuery(rpcQueryOptions(fn, { p_viewer_id: 1 }));
    }
    await loadSession();
    await queryClient.fetchQuery(elearningProgressQuery('2025001'));
    const keys = dehydratedKeys();
    for (const fn of FORBIDDEN) expect(keys.some((k) => k.includes(`"${fn}"`))).toBe(false);
    expect(keys.some((k) => k.includes('"session"'))).toBe(false);
    expect(keys.some((k) => k.includes('"sync-elearning"'))).toBe(false);
    expect(keys).toHaveLength(ALLOWED.length);
  });

  it('query gagal tidak dipersist walau fn boleh', async () => {
    api.rpc.mockResolvedValue({ data: null, error: { message: 'x', status: 500 } });
    await queryClient
      .fetchQuery({ ...rpcQueryOptions('get_public_settings'), retry: false })
      .catch(() => undefined);
    expect(dehydratedKeys()).toEqual([]);
  });
});
