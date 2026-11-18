import { elearningKeys, type Achievement } from '@algohub/api-client/elearning';
import { queryOptions } from '@tanstack/svelte-query';
import { api } from '#lib/api.ts';
import { queryClient } from '#lib/query.ts';
import { appStore } from '#lib/stores/app.svelte.ts';
// Katalog sama untuk semua user, tetapi yang terbuka milik pribadi; keduanya tidak dipersist
// supaya perangkat bersama tidak menyimpan capaian orang lain.
export const achievementCatalogQuery = () =>
  queryOptions({
    queryKey: elearningKeys.achievements.catalog(),
    queryFn: () => api.achievements.catalog(),
    staleTime: 5 * 60_000,
  });
export const myAchievementsQuery = () =>
  queryOptions({
    queryKey: elearningKeys.achievements.mine(),
    queryFn: () => api.achievements.mine(),
  });
export const unlockedIds = (mine: readonly Achievement[] | undefined): Set<string> =>
  new Set((mine ?? []).map((achievement) => achievement.id));

// Achievement baru masuk antrean popup dan daftar "terbuka" di profil ikut diambil ulang.
export function announceAchievements(achievements: readonly Achievement[]): void {
  if (achievements.length === 0) return;
  appStore.pushAchievements(achievements);
  void queryClient.invalidateQueries({ queryKey: elearningKeys.achievements.mine() });
}
export const checkAchievementsOptions = () => ({
  mutationFn: (body: Parameters<typeof api.achievements.check>[0]) => api.achievements.check(body),
  onSuccess: (unlocked: Achievement[]) => announceAchievements(unlocked ?? []),
});
export const checkXpOptions = () => ({
  mutationFn: (xp: number) => api.achievements.checkXp(xp),
  onSuccess: (res: { achievement: Achievement | null }) => {
    if (res.achievement) announceAchievements([res.achievement]);
  },
});
