import { elearningKeys } from '@algohub/api-client/elearning';
import { queryOptions } from '@tanstack/svelte-query';
import { api } from '#lib/api.ts';
import { loadProfile } from '#lib/profile.ts';
import { queryClient } from '#lib/query.ts';
// Kelayakan dihitung backend (6/6 level, cooldown 14 hari, 2x per semester) dan berubah setiap
// lesson selesai, jadi tidak dipersist.
export const prestigeEligibilityQuery = (nim?: string) =>
  queryOptions({
    queryKey: elearningKeys.prestige.eligibility(nim),
    queryFn: () => api.prestige.eligibility(nim),
  });
// Reset mengarsipkan progres dan mengembalikan XP ke 0: semua yang bergantung pada progres
// dan papan peringkat diambil ulang, profil disinkronkan dari server.
export const prestigeResetOptions = (nim?: string) => ({
  mutationFn: () => api.prestige.reset(nim),
  onSuccess: async () => {
    await Promise.all([
      queryClient.invalidateQueries({ queryKey: elearningKeys.prestige.eligibility(nim) }),
      queryClient.invalidateQueries({ queryKey: elearningKeys.progress.lessons() }),
      queryClient.invalidateQueries({ queryKey: elearningKeys.progress.recent() }),
      queryClient.invalidateQueries({ queryKey: ['elearning', 'leaderboard'] }),
      loadProfile().catch(() => undefined),
    ]);
  },
});
