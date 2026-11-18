import type { Achievement } from '@algohub/api-client/elearning';
import { beforeEach, describe, expect, it, vi } from 'vitest';
const api = vi.hoisted(() => ({
  api: {
    achievements: { check: vi.fn(), checkXp: vi.fn(), catalog: vi.fn(), mine: vi.fn() },
    progress: { grantXp: vi.fn() },
    game: { saveHistory: vi.fn() },
    streak: { touch: vi.fn() },
  },
}));
vi.mock('#lib/api.ts', () => api);
const { appStore } = await import('../src/lib/stores/app.svelte');
const { applyXpGain } = await import('../src/lib/rewards');
const { announceAchievements, checkXpOptions } = await import('../src/lib/queries/achievements');
const { grantXpOptions } = await import('../src/lib/queries/progress');
const { saveGameHistoryOptions } = await import('../src/lib/queries/game');
const { touchStreakOptions } = await import('../src/lib/queries/streak');
const { queryClient } = await import('../src/lib/query');

const ach = (id: string): Achievement => ({
  id,
  title: id,
  description: '',
  icon: 'Trophy',
  requirementType: null,
  requirementValue: null,
});
const profile = {
  nim: '1',
  nama: 'U',
  kelas: 'A',
  jurusan: null,
  shift: null,
  email: null,
  prestigeTier: null,
  prestigeMult: 1,
  prestigeCount: 0,
  xp: 10,
  level: 1,
  streak: 0,
  lastActive: null,
  createdAt: null,
  role: 'praktikan',
  assessmentAccess: null,
  levelAccessOverrides: null,
  studyTime: 0,
};

beforeEach(() => {
  appStore.reset();
  queryClient.clear();
  vi.clearAllMocks();
});

describe('antrean popup', () => {
  it('achievement tampil satu per satu sesuai urutan masuk', () => {
    appStore.pushAchievements([ach('a'), ach('b')]);
    expect(appStore.unlockedAchievement?.id).toBe('a');
    appStore.shiftAchievement();
    expect(appStore.unlockedAchievement?.id).toBe('b');
    appStore.shiftAchievement();
    expect(appStore.unlockedAchievement).toBeNull();
  });

  it('achievement yang sama tidak masuk antrean dua kali', () => {
    appStore.pushAchievement(ach('a'));
    appStore.pushAchievement(ach('a'));
    expect(appStore.achievementQueue).toHaveLength(1);
  });

  it('reset saat logout mengosongkan antrean dan level-up', () => {
    appStore.pushAchievement(ach('a'));
    appStore.setLevelUp(3);
    appStore.reset();
    expect(appStore.achievementQueue).toEqual([]);
    expect(appStore.levelUp).toBeNull();
  });
});

describe('popup setelah mutasi XP', () => {
  it('level naik memicu popup level-up dan memperbarui profil', () => {
    appStore.setProfile(profile);
    applyXpGain({ xp: 2010, level: 2, leveledUp: true, streak: 4, lastActive: 'x' });
    expect(appStore.levelUp).toBe(2);
    expect(appStore.profile).toMatchObject({ xp: 2010, level: 2, streak: 4, lastActive: 'x' });
  });

  it('tanpa naik level popup tidak muncul', () => {
    appStore.setProfile(profile);
    applyXpGain({ xp: 60, level: 1, leveledUp: false });
    expect(appStore.levelUp).toBeNull();
    expect(appStore.profile?.streak).toBe(0);
  });

  it('POST /progress/xp: level-up dan achievement baru masuk antrean', async () => {
    appStore.setProfile(profile);
    const res = { xp: 2000, level: 2, leveledUp: true, achievement: ach('xp-2000') };
    const options = grantXpOptions();
    api.api.progress.grantXp.mockResolvedValue(res);
    expect(await options.mutationFn(1990)).toBe(res);
    options.onSuccess(res);
    expect(api.api.progress.grantXp).toHaveBeenCalledWith(1990);
    expect(appStore.levelUp).toBe(2);
    expect(appStore.unlockedAchievement?.id).toBe('xp-2000');
  });

  it('POST /achievements/check-xp mengumumkan achievement bila ada', () => {
    checkXpOptions().onSuccess({ achievement: ach('x') });
    checkXpOptions().onSuccess({ achievement: null });
    expect(appStore.achievementQueue.map((a) => a.id)).toEqual(['x']);
  });

  it('Bug Hunt tersimpan: XP, level-up, dan kuota diperbarui', async () => {
    appStore.setProfile(profile);
    const invalidate = vi.spyOn(queryClient, 'invalidateQueries');
    await saveGameHistoryOptions().onSuccess({
      xp: 2100,
      level: 2,
      streak: 1,
      lastActive: 'now',
      leveledUp: true,
    });
    expect(appStore.levelUp).toBe(2);
    expect(invalidate).toHaveBeenCalledWith({ queryKey: ['elearning', 'game', 'plays'] });
    expect(invalidate).toHaveBeenCalledWith({ queryKey: ['elearning', 'leaderboard'] });
  });

  it('announceAchievements mengambil ulang daftar achievement milik user', () => {
    const invalidate = vi.spyOn(queryClient, 'invalidateQueries');
    announceAchievements([]);
    expect(invalidate).not.toHaveBeenCalled();
    announceAchievements([ach('a')]);
    expect(invalidate).toHaveBeenCalledWith({ queryKey: ['elearning', 'achievements', 'mine'] });
  });

  it('POST /streak memperbarui streak di profil', () => {
    appStore.setProfile(profile);
    touchStreakOptions().onSuccess({ streak: 5, lastActive: 'now' });
    expect(appStore.profile).toMatchObject({ streak: 5, lastActive: 'now' });
  });
});
