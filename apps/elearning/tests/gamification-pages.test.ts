import { ApiError } from '@algohub/api-client';
import { fireEvent, render, screen, waitFor } from '@testing-library/svelte';
import { beforeEach, describe, expect, it, vi } from 'vitest';
const api = vi.hoisted(() => ({
  api: {
    leaderboard: { list: vi.fn(), weekly: vi.fn(), rank: vi.fn(), filters: vi.fn() },
    achievements: { catalog: vi.fn(), mine: vi.fn() },
    prestige: { eligibility: vi.fn(), reset: vi.fn() },
    progress: { myLessons: vi.fn() },
    profile: { sync: vi.fn() },
  },
}));
const realtime = vi.hoisted(() => ({ subscribe: vi.fn(() => () => {}) }));
vi.mock('#lib/api.ts', () => api);
vi.mock('#lib/realtime/client.ts', () => realtime);
const { default: QueryHost } = await import('./QueryHost.svelte');
const { default: Leaderboard } = await import('../src/routes/(app)/leaderboard/+page.svelte');
const { default: Profile } = await import('../src/routes/(app)/profile/+page.svelte');
const { queryClient } = await import('../src/lib/query');
const { appStore } = await import('../src/lib/stores/app.svelte');

const user = (role: string) => ({
  id: 1,
  nim: '2022110083',
  nama: 'Budi',
  kelas: 'A1',
  role,
  jurusan: 'TI',
  email: null,
  shift: '2',
});
const data = (role: string) => ({ auth: { user: user(role) } }) as never;
const entry = (nim: string, nama: string, xp: number, weeklyXp = xp) => ({
  nim,
  nama,
  kelas: 'A1',
  jurusan: 'TI',
  shift: '2',
  email: null,
  xp,
  level: 2,
  streak: 0,
  lastActive: '',
  createdAt: '',
  role: 'mahasiswa',
  assessmentAccess: {},
  weeklyXp,
  lessonXp: weeklyXp,
  bughuntXp: 0,
  lessonsCompleted: 0,
  isCarry: false,
  carryXp: 0,
});
const page = <T>(items: T[]) => ({ items, total: items.length, page: 1, limit: 100 });
const eligibility = {
  eligible: true,
  reason: 'Layak prestige: tamat penuh 6/6.',
  levelsDone: 6,
  totalLevels: 6,
  lessonsDone: 60,
  totalLessons: 60,
  tier: 'Master',
  mult: 1.5,
  prestigeCount: 0,
  lastPrestigeAt: null,
  semesterUsed: 0,
};

// jsdom tidak punya Element.animate yang dipakai transisi Svelte.
Element.prototype.animate = vi.fn(() => ({
  cancel: vi.fn(),
  finish: vi.fn(),
  onfinish: null,
})) as never;

beforeEach(() => {
  queryClient.clear();
  appStore.reset();
  vi.clearAllMocks();
  api.api.leaderboard.filters.mockResolvedValue({ kelas: ['A1'], jurusan: ['TI'], shift: ['2'] });
  api.api.achievements.catalog.mockResolvedValue([
    {
      id: 'a1',
      title: 'Langkah Pertama',
      description: 'Selesaikan lesson',
      icon: 'Star',
      requirementType: null,
      requirementValue: null,
    },
    {
      id: 'a2',
      title: 'Pemburu Bug',
      description: 'Main Bug Hunt',
      icon: 'Bug',
      requirementType: null,
      requirementValue: null,
    },
  ]);
  api.api.achievements.mine.mockResolvedValue([
    {
      id: 'a1',
      title: 'Langkah Pertama',
      description: '',
      icon: 'Star',
      requirementType: null,
      requirementValue: null,
    },
  ]);
  api.api.progress.myLessons.mockResolvedValue(page(['l1']));
  api.api.prestige.eligibility.mockResolvedValue(eligibility);
});

describe('/leaderboard', () => {
  it('mahasiswa melihat top 3 mingguan kelasnya dan peringkat sendiri di luar top 3', async () => {
    const top = [entry('1', 'Ani', 300), entry('2', 'Beni', 200), entry('3', 'Cici', 100)];
    api.api.leaderboard.weekly.mockImplementation(async (p: { limit: number }) =>
      page(p.limit === 0 ? [...top, entry('2022110083', 'Budi', 50)] : top),
    );
    render(QueryHost, { page: Leaderboard, data: data('mahasiswa') });
    expect(await screen.findByText('Papan Peringkat Mingguan')).toBeInTheDocument();
    expect((await screen.findAllByText('Ani')).length).toBeGreaterThan(0);
    expect(api.api.leaderboard.weekly).toHaveBeenCalledWith({
      limit: 3,
      kelas: 'A1',
      jurusan: 'TI',
      shift: '2',
    });
    expect(await screen.findByText('XP minggumu: 50 XP')).toBeInTheDocument();
    expect(await screen.findByText('4')).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'Total' })).toBeNull();
    expect(api.api.leaderboard.filters).not.toHaveBeenCalled();
  });

  it('staf memilih tab Total dan filter kelas yang dipetakan ke query', async () => {
    api.api.leaderboard.weekly.mockResolvedValue(page([entry('1', 'Ani', 300)]));
    api.api.leaderboard.list.mockResolvedValue(page([entry('1', 'Ani', 900)]));
    api.api.leaderboard.rank.mockResolvedValue({ rank: 7 });
    render(QueryHost, { page: Leaderboard, data: data('asisten') });
    await fireEvent.click(await screen.findByRole('button', { name: 'Total' }));
    await waitFor(() => expect(api.api.leaderboard.list).toHaveBeenCalledWith({ limit: 20 }));
    await fireEvent.change(await screen.findByLabelText('Filter kelas'), {
      target: { value: 'A1' },
    });
    await waitFor(() =>
      expect(api.api.leaderboard.list).toHaveBeenCalledWith({ kelas: 'A1', limit: 20 }),
    );
    expect(await screen.findByRole('button', { name: /Reset Filter/ })).toBeInTheDocument();
  });
});

describe('/profile', () => {
  const sessionProfile = {
    nim: '2022110083',
    nama: 'Budi',
    kelas: 'A1',
    jurusan: 'TI',
    shift: '2',
    email: null,
    prestigeTier: null,
    prestigeMult: 1,
    prestigeCount: 0,
    xp: 4500,
    level: 3,
    streak: 2,
    lastActive: null,
    createdAt: '2026-03-01T00:00:00Z',
    role: 'mahasiswa',
    assessmentAccess: null,
    levelAccessOverrides: null,
    studyTime: 0,
  };

  it('menampilkan achievement terbuka, statistik, dan subscribe realtime milik user', async () => {
    appStore.setProfile(sessionProfile);
    const { unmount } = render(QueryHost, { page: Profile, data: data('mahasiswa') });
    expect(await screen.findByText('1 / 2 Terbuka')).toBeInTheDocument();
    expect(screen.getByText('Langkah Pertama')).toBeInTheDocument();
    expect(await screen.findByText('Layak prestige: tamat penuh 6/6.')).toBeInTheDocument();
    expect(screen.getByText('500 / 2000 XP')).toBeInTheDocument();
    expect(realtime.subscribe).toHaveBeenCalledWith(
      'profile-achievements-2022110083',
      { table: 'unlocked_achievements', filter: 'nim=eq.2022110083' },
      expect.any(Function),
    );
    unmount();
  });

  it('reset prestige memanggil POST /prestige/reset setelah konfirmasi', async () => {
    appStore.setProfile(sessionProfile);
    api.api.prestige.reset.mockResolvedValue({ tier: 'Master', mult: 1.5, prestigeNo: 1 });
    api.api.profile.sync.mockResolvedValue({
      profile: { ...sessionProfile, xp: 0 },
      newlyUnlocked: [],
    });
    render(QueryHost, { page: Profile, data: data('mahasiswa') });
    await fireEvent.click(await screen.findByRole('button', { name: /Reset & Ambil x1.50/ }));
    await fireEvent.click(screen.getByRole('button', { name: 'Ya, Reset' }));
    await waitFor(() => expect(api.api.prestige.reset).toHaveBeenCalledWith(undefined));
    expect(await screen.findByText(/Prestige #1 aktif: tier Master x1.50/)).toBeInTheDocument();
    await waitFor(() => expect(appStore.profile?.xp).toBe(0));
  });

  it('403 "Akses ditolak." ditampilkan apa adanya', async () => {
    appStore.setProfile(sessionProfile);
    api.api.prestige.reset.mockRejectedValue(new ApiError('Akses ditolak.', 403));
    render(QueryHost, { page: Profile, data: data('mahasiswa') });
    await fireEvent.click(await screen.findByRole('button', { name: /Reset & Ambil/ }));
    await fireEvent.click(screen.getByRole('button', { name: 'Ya, Reset' }));
    expect(await screen.findByText('Akses ditolak.')).toBeInTheDocument();
  });

  it('belum layak: tombol mati dan alasan dari server tampil', async () => {
    appStore.setProfile(sessionProfile);
    api.api.prestige.eligibility.mockResolvedValue({
      ...eligibility,
      eligible: false,
      reason: 'Syarat belum terpenuhi: tamat 2/6 level.',
    });
    render(QueryHost, { page: Profile, data: data('mahasiswa') });
    expect(await screen.findByText('Syarat belum terpenuhi: tamat 2/6 level.')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Belum Layak Prestige' })).toBeDisabled();
  });
});
