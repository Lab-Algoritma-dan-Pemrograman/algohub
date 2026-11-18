import { describe, expect, it } from 'vitest';
import { elearningKeys } from './keys';

describe('elearningKeys', () => {
  it('semua key berawalan elearning', () => {
    const keys = [
      elearningKeys.session(),
      elearningKeys.curriculum.levels(),
      elearningKeys.curriculum.modules(),
      elearningKeys.curriculum.lessons(),
      elearningKeys.progress.lessons(),
      elearningKeys.progress.recent(),
      elearningKeys.achievements.catalog(),
      elearningKeys.achievements.mine(),
      elearningKeys.leaderboard.list(),
      elearningKeys.leaderboard.rank(10),
      elearningKeys.leaderboard.filters(),
      elearningKeys.leaderboard.weekly(),
      elearningKeys.game.questions(),
      elearningKeys.game.settings(),
      elearningKeys.game.plays(),
      elearningKeys.playground.examples(),
      elearningKeys.monitoring.sessions(),
      elearningKeys.monitoring.students(),
      elearningKeys.monitoring.progress(),
      elearningKeys.activityLogs.list(),
      elearningKeys.admin.users(),
      elearningKeys.admin.userProgress('1'),
      elearningKeys.prestige.eligibility(),
    ];
    expect(elearningKeys.all).toEqual(['elearning']);
    for (const key of keys) expect(key[0]).toBe('elearning');
    expect(new Set(keys.map((k) => JSON.stringify(k))).size).toBe(keys.length);
  });

  it('parameter masuk ke key sehingga cache terpisah per filter dan halaman', () => {
    expect(elearningKeys.leaderboard.list({ kelas: 'A' })).not.toEqual(
      elearningKeys.leaderboard.list({ kelas: 'B' }),
    );
    expect(elearningKeys.admin.users('budi', 2)).toEqual([
      'elearning',
      'admin',
      'users',
      'budi',
      2,
    ]);
    expect(elearningKeys.monitoring.progress('123')).toEqual([
      'elearning',
      'monitoring',
      'progress',
      '123',
    ]);
    expect(elearningKeys.activityLogs.list(3)).toEqual(['elearning', 'activity-logs', 3]);
  });
});
