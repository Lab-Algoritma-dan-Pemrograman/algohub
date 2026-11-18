import { elearningKeys } from '@algohub/api-client/elearning';
import { queryOptions } from '@tanstack/svelte-query';
import { api } from '#lib/api.ts';
import { queryClient } from '#lib/query.ts';
import { applyXpGain } from '#lib/rewards.ts';
// Soal, pengaturan, dan kuota main berubah oleh admin atau permainan sendiri; tidak dipersist.
export const gameSettingsQuery = () =>
  queryOptions({
    queryKey: elearningKeys.game.settings(),
    queryFn: () => api.game.settings(),
  });
export const myPlaysQuery = () =>
  queryOptions({
    queryKey: elearningKeys.game.plays(),
    queryFn: () => api.game.myPlays(),
  });
// Soal diacak backend per permintaan, jadi hasil lama tidak dipakai ulang untuk permainan berikutnya.
export const gameQuestionsQuery = (language: string, limit: number) =>
  queryOptions({
    queryKey: elearningKeys.game.questions(language),
    queryFn: () => api.game.questions({ language, limit }),
    staleTime: 0,
    gcTime: 0,
  });
// Backend menjawab 400 WEEKLY_LIMIT_REACHED bila kuota habis; error itu dilempar apa adanya
// supaya BugHunt menampilkan status batas tercapai tanpa XP.
export const saveGameHistoryOptions = () => ({
  mutationFn: (body: { gameType: string; xpEarned: number }) => api.game.saveHistory(body),
  onSuccess: async (res: Awaited<ReturnType<typeof api.game.saveHistory>>) => {
    applyXpGain(res);
    await Promise.all([
      queryClient.invalidateQueries({ queryKey: elearningKeys.game.plays() }),
      queryClient.invalidateQueries({ queryKey: ['elearning', 'leaderboard'] }),
    ]);
  },
});
