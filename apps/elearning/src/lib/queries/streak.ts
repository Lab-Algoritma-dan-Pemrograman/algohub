import { api } from '#lib/api.ts';
import { appStore } from '#lib/stores/app.svelte.ts';
export const touchStreakOptions = () => ({
  mutationFn: () => api.streak.touch(),
  onSuccess: (res: { streak: number; lastActive: string }) =>
    appStore.patchProfile({ streak: res.streak, lastActive: res.lastActive }),
});
