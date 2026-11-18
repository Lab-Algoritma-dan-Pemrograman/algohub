import { ApiError } from '@algohub/api-client';
import { fireEvent, render, screen, waitFor } from '@testing-library/svelte';
import { beforeEach, describe, expect, it, vi } from 'vitest';
const api = vi.hoisted(() => ({
  api: {
    game: { settings: vi.fn(), myPlays: vi.fn(), questions: vi.fn(), saveHistory: vi.fn() },
    achievements: { check: vi.fn(), checkXp: vi.fn() },
  },
}));
vi.mock('#lib/api.ts', () => api);
const { default: BugHunt } = await import('../src/lib/components/BugHunt.svelte');
const { queryClient } = await import('../src/lib/query');
const { appStore } = await import('../src/lib/stores/app.svelte');

const question = {
  id: 'q1',
  language: 'c',
  difficulty: 'easy',
  title: 'Salah titik koma',
  code: 'int main() {\\n  int x = 1\\n  return 0;\\n}',
  bugLine: 1,
  explanation: 'Titik koma hilang.',
};
const settings = {
  bugHuntActive: true,
  bugHuntCActive: true,
  bugHuntPythonActive: true,
  bugHuntWeeklyLimit: 3,
  bugHuntQuestionCount: 1,
};

beforeEach(() => {
  queryClient.clear();
  appStore.reset();
  vi.clearAllMocks();
  api.api.game.settings.mockResolvedValue(settings);
  api.api.game.myPlays.mockResolvedValue({ allowed: true, playsUsed: 1, limit: 3 });
  api.api.game.questions.mockResolvedValue([question]);
  api.api.achievements.check.mockResolvedValue([]);
  api.api.achievements.checkXp.mockResolvedValue({ achievement: null });
});

// Klik baris 2 (indeks 1) menjawab dengan benar, lalu lanjut ke hasil akhir.
async function playToEnd() {
  await fireEvent.click(await screen.findByRole('button', { name: /int x = 1/ }));
  await fireEvent.click(await screen.findByRole('button', { name: /Lihat Hasil Akhir/ }));
}

describe('BugHunt', () => {
  it('meminta soal sesuai bahasa dan jumlah dari pengaturan', async () => {
    render(BugHunt, { language: 'c', onclose: vi.fn() });
    await screen.findByText('Temukan baris yang mengandung bug!');
    expect(api.api.game.questions).toHaveBeenCalledWith({ language: 'c', limit: 1 });
  });

  it('kuota mingguan habis di server memblokir permainan sebelum soal dimuat', async () => {
    api.api.game.myPlays.mockResolvedValue({ allowed: false, playsUsed: 3, limit: 3 });
    render(BugHunt, { language: 'python', onclose: vi.fn() });
    expect(await screen.findByText('Limit Tercapai!')).toBeInTheDocument();
    expect(api.api.game.questions).not.toHaveBeenCalled();
  });

  it('tanpa soal menampilkan pesan kosong', async () => {
    api.api.game.questions.mockResolvedValue([]);
    render(BugHunt, { language: 'c', onclose: vi.fn() });
    expect(await screen.findByText('Tidak Ada Soal')).toBeInTheDocument();
  });

  it('jawaban benar menyimpan XP lalu memeriksa achievement', async () => {
    api.api.game.saveHistory.mockResolvedValue({
      xp: 100,
      level: 1,
      streak: 1,
      lastActive: 'now',
      leveledUp: false,
    });
    api.api.achievements.check.mockResolvedValue([
      {
        id: 'bug-1',
        title: 'Pemburu',
        description: '',
        icon: 'Bug',
        requirementType: null,
        requirementValue: null,
      },
    ]);
    render(BugHunt, { language: 'c', onclose: vi.fn() });
    await playToEnd();
    await waitFor(() => expect(api.api.game.saveHistory).toHaveBeenCalled());
    const body = api.api.game.saveHistory.mock.calls[0][0];
    expect(body.gameType).toBe('bug_hunt');
    expect(body.xpEarned).toBeGreaterThanOrEqual(20);
    await waitFor(() =>
      expect(api.api.achievements.check).toHaveBeenCalledWith({
        xp: 100,
        gamesPlayed: 1,
        perfectGames: 1,
      }),
    );
    await waitFor(() => expect(appStore.unlockedAchievement?.id).toBe('bug-1'));
  });

  it('400 WEEKLY_LIMIT_REACHED menampilkan limit tercapai tanpa XP', async () => {
    api.api.game.saveHistory.mockRejectedValue(new ApiError('WEEKLY_LIMIT_REACHED', 400));
    render(BugHunt, { language: 'c', onclose: vi.fn() });
    await playToEnd();
    expect(await screen.findByText(/Limit tercapai — XP tidak disimpan/)).toBeInTheDocument();
    expect(screen.getByText('—')).toBeInTheDocument();
    expect(api.api.achievements.check).not.toHaveBeenCalled();
  });
});
