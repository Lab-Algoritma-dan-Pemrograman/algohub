import { render, screen, waitFor } from '@testing-library/svelte';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import type { ChangePayload } from '../src/lib/realtime/client';

const api = vi.hoisted(() => ({
  api: {
    monitoring: {
      sessions: vi.fn(),
      students: vi.fn(),
      progress: vi.fn(),
      cleanupStale: vi.fn(),
    },
    curriculum: { levels: vi.fn(), modules: vi.fn(), lessons: vi.fn() },
    admin: { adjustXp: vi.fn(), resetUser: vi.fn(), deleteUser: vi.fn() },
    activityLogs: { list: vi.fn(), remove: vi.fn() },
  },
}));
const realtime = vi.hoisted(() => ({
  handlers: new Map<string, (payload: unknown) => void>(),
  subscribe: vi.fn(),
}));
realtime.subscribe.mockImplementation(
  (name: string, _filter: unknown, onChange: (payload: unknown) => void) => {
    realtime.handlers.set(name, onChange);
    return () => realtime.handlers.delete(name);
  },
);

vi.mock('#lib/api.ts', () => api);
vi.mock('#lib/realtime/client.ts', () => realtime);

const { default: QueryHost } = await import('./QueryHost.svelte');
const { default: Monitoring } = await import('../src/routes/(app)/monitoring/+page.svelte');
const { default: AuditLog } = await import('../src/routes/(app)/auditlog/+page.svelte');
const { elearningKeys } = await import('@algohub/api-client/elearning');
const { queryClient } = await import('../src/lib/query');
const { applyActivityLogChange, applyProgressChange, applySessionChange } =
  await import('../src/lib/realtime/monitoring');

const page = <T>(items: T[]) => ({ items, total: items.length, page: 1, limit: 100 });
const data = (role: string) =>
  ({
    auth: {
      user: {
        id: 1,
        nim: '999',
        nama: 'Staf',
        kelas: '',
        role,
        jurusan: null,
        email: null,
        shift: null,
      },
    },
  }) as never;

const student = (nim: string, nama: string, xp = 100) => ({
  nim,
  id: null,
  nama,
  kelas: 'A1',
  jurusan: 'TI',
  email: null,
  role: 'praktikan',
  elearning_role: null,
  xp,
  level: 1,
  streak: 0,
  study_time: 0,
  last_active: null,
  created_at: null,
  assessment_access: {},
  level_access_overrides: {},
  shift: null,
  prestige_tier: null,
  prestige_mult: 1,
  prestige_count: 0,
  last_prestige_at: null,
});

const change = (
  eventType: 'INSERT' | 'UPDATE' | 'DELETE',
  row: Record<string, unknown>,
): ChangePayload =>
  ({
    eventType,
    new: eventType === 'DELETE' ? {} : row,
    old: eventType === 'DELETE' ? row : {},
  }) as never;

const emit = (channel: string, payload: ChangePayload) => {
  const handler = realtime.handlers.get(channel);
  if (!handler) throw new Error(`channel ${channel} belum disubscribe`);
  handler(payload);
};

beforeEach(() => {
  queryClient.clear();
  realtime.handlers.clear();
  vi.clearAllMocks();
  // Pembersihan sesi basi tidak diselesaikan supaya tidak memicu refetch sesi di luar event.
  api.api.monitoring.cleanupStale.mockReturnValue(new Promise(() => undefined));
  api.api.monitoring.students.mockResolvedValue(
    page([student('1', 'Ani', 300), student('2', 'Beni', 200)]),
  );
  api.api.monitoring.sessions.mockResolvedValue(page([]));
  api.api.monitoring.progress.mockResolvedValue(page([]));
  api.api.curriculum.levels.mockResolvedValue(
    page([{ id: 'lv1', title: 'Level 1 C', description: '', sort_order: 1 }]),
  );
  api.api.curriculum.modules.mockResolvedValue(
    page([{ id: 'm1', level_id: 'lv1', title: 'Modul 1', sort_order: 1 }]),
  );
  api.api.curriculum.lessons.mockResolvedValue(
    page([
      { id: 'les1', module_id: 'm1', title: 'Hello', sort_order: 1 },
      { id: 'les2', module_id: 'm1', title: 'Variabel', sort_order: 2 },
    ]),
  );
});

describe('/monitoring', () => {
  it('event Realtime memperbarui sesi dan progres tanpa memuat ulang data', async () => {
    render(QueryHost, { page: Monitoring, data: data('admin') });
    expect(await screen.findByText('Ani')).toBeInTheDocument();
    expect(screen.getByText('0 Aktif')).toBeInTheDocument();
    expect(screen.getAllByText('0%').length).toBe(2);
    expect([...realtime.handlers.keys()].sort()).toEqual([
      'monitoring:active_sessions',
      'monitoring:student_progress',
    ]);
    await waitFor(() => {
      expect(queryClient.getQueryData(elearningKeys.monitoring.sessions())).toBeDefined();
      expect(queryClient.getQueryData(elearningKeys.monitoring.progress())).toBeDefined();
    });
    const before = {
      sessions: api.api.monitoring.sessions.mock.calls.length,
      students: api.api.monitoring.students.mock.calls.length,
      progress: api.api.monitoring.progress.mock.calls.length,
    };

    // Role diambil dari daftar mahasiswa di cache, jadi tidak perlu refetch sesi.
    emit(
      'monitoring:active_sessions',
      change('INSERT', { nim: '1', last_heartbeat: new Date().toISOString() }),
    );
    await waitFor(() => expect(screen.getByText('1 Aktif')).toBeInTheDocument());

    emit(
      'monitoring:student_progress',
      change('INSERT', { id: 'p1', nim: '2', lesson_id: 'les1', completed: true }),
    );
    await waitFor(() => expect(screen.getByText('50%')).toBeInTheDocument());

    emit('monitoring:active_sessions', change('DELETE', { nim: '1' }));
    await waitFor(() => expect(screen.getByText('0 Aktif')).toBeInTheDocument());

    expect(api.api.monitoring.sessions).toHaveBeenCalledTimes(before.sessions);
    expect(api.api.monitoring.students).toHaveBeenCalledTimes(before.students);
    expect(api.api.monitoring.progress).toHaveBeenCalledTimes(before.progress);
  });

  it('asisten hanya membaca: tidak ada aksi XP, reset, atau hapus', async () => {
    render(QueryHost, { page: Monitoring, data: data('asisten') });
    const card = await screen.findByText('Ani');
    card.click();
    await screen.findByText(/Terakhir Aktif/);
    expect(screen.queryByRole('button', { name: /Atur XP/i })).toBeNull();
  });

  it('admin melihat aksi pengelolaan setelah kartu dibuka', async () => {
    render(QueryHost, { page: Monitoring, data: data('admin') });
    (await screen.findByText('Ani')).click();
    expect(await screen.findByRole('button', { name: /Atur XP/i })).toBeInTheDocument();
  });
});

describe('applySessionChange', () => {
  const key = elearningKeys.monitoring.sessions();

  it('role yang tak dikenal hanya menandai key sesi untuk diambil ulang', () => {
    vi.useFakeTimers();
    queryClient.setQueryData(key, []);
    const spy = vi.spyOn(queryClient, 'invalidateQueries');
    applySessionChange(queryClient, change('INSERT', { nim: '9', last_heartbeat: 'x' }));
    expect(spy).not.toHaveBeenCalled();
    vi.advanceTimersByTime(600);
    expect(spy).toHaveBeenCalledTimes(1);
    expect(spy).toHaveBeenCalledWith({ queryKey: key });
    vi.useRealTimers();
  });

  it('heartbeat dari sesi yang sudah ada memakai role lama', () => {
    queryClient.setQueryData(key, [
      { nim: '1', last_heartbeat: 'lama', users: { role: 'asisten' } },
    ]);
    applySessionChange(queryClient, change('UPDATE', { nim: '1', last_heartbeat: 'baru' }));
    expect(queryClient.getQueryData(key)).toEqual([
      { nim: '1', last_heartbeat: 'baru', users: { role: 'asisten' } },
    ]);
  });
});

describe('applyProgressChange', () => {
  const key = elearningKeys.monitoring.progress();

  it('menambah dan mencabut lesson selesai per NIM', () => {
    queryClient.setQueryData(key, [{ nim: '1', lessonIds: ['a'] }]);
    applyProgressChange(
      queryClient,
      change('INSERT', { nim: '1', lesson_id: 'b', completed: true }),
    );
    expect(queryClient.getQueryData(key)).toEqual([{ nim: '1', lessonIds: ['a', 'b'] }]);
    applyProgressChange(
      queryClient,
      change('UPDATE', { nim: '1', lesson_id: 'a', completed: false }),
    );
    expect(queryClient.getQueryData(key)).toEqual([{ nim: '1', lessonIds: ['b'] }]);
    applyProgressChange(queryClient, change('DELETE', { nim: '1', lesson_id: 'b' }));
    expect(queryClient.getQueryData(key)).toEqual([]);
  });
});

describe('/auditlog', () => {
  const log = (id: string, nama: string) => ({
    id,
    nim: '1',
    nama,
    event_type: 'login',
    timestamp: '2026-05-01T03:00:00Z',
    details: 'Masuk',
    ip_address: null,
  });

  it('log baru dan log yang dihapus lewat Realtime langsung tampil tanpa refetch', async () => {
    api.api.activityLogs.list.mockResolvedValue({
      items: [log('a', 'Ani')],
      total: 1,
      page: 1,
      limit: 50,
    });
    render(QueryHost, { page: AuditLog, data: data('admin') });
    expect(await screen.findByText('Ani')).toBeInTheDocument();
    const calls = api.api.activityLogs.list.mock.calls.length;

    emit('audit:activity_logs', change('INSERT', log('b', 'Beni')));
    expect(await screen.findByText('Beni')).toBeInTheDocument();

    emit('audit:activity_logs', change('DELETE', { id: 'a' }));
    await waitFor(() => expect(screen.queryByText('Ani')).toBeNull());
    expect(api.api.activityLogs.list).toHaveBeenCalledTimes(calls);
  });

  it('kordas tidak melihat pilihan hapus', async () => {
    api.api.activityLogs.list.mockResolvedValue({
      items: [log('a', 'Ani')],
      total: 1,
      page: 1,
      limit: 50,
    });
    render(QueryHost, { page: AuditLog, data: data('kordas') });
    await screen.findByText('Ani');
    expect(screen.queryByLabelText('Pilih log')).toBeNull();
  });
});

describe('applyActivityLogChange', () => {
  it('log yang dihapus dicabut dari halaman manapun di cache', () => {
    const entry = (id: string) => ({ id, nim: '1', nama: 'x', event_type: 'login' });
    queryClient.setQueryData(elearningKeys.activityLogs.list(2), {
      items: [entry('z')],
      total: 51,
      page: 2,
      limit: 50,
    });
    applyActivityLogChange(queryClient, change('DELETE', { id: 'z' }));
    expect(queryClient.getQueryData(elearningKeys.activityLogs.list(2))).toEqual({
      items: [],
      total: 50,
      page: 2,
      limit: 50,
    });
  });
});
