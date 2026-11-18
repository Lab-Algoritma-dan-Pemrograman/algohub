import { appStore } from '#lib/stores/app.svelte.ts';
interface XpGain {
  xp: number;
  level: number;
  leveledUp: boolean;
  streak?: number;
  lastActive?: string | null;
}
// Hasil mutasi XP (lesson, Bug Hunt, grant) selalu memperbarui profil di header dan memicu
// popup level-up; achievement dikirim terpisah oleh pemanggil karena bentuk responsnya berbeda.
export function applyXpGain(gain: XpGain): void {
  appStore.patchProfile({
    xp: gain.xp,
    level: gain.level,
    ...(gain.streak === undefined ? {} : { streak: gain.streak }),
    ...(gain.lastActive === undefined ? {} : { lastActive: gain.lastActive }),
  });
  if (gain.leveledUp) appStore.setLevelUp(gain.level);
}
