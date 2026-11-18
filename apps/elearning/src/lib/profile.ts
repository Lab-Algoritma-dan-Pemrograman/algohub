import { api } from './api.ts';
import { appStore } from './stores/app.svelte.ts';
// Menggantikan inisialisasi user di SupabaseProvider.tsx lama: backend membuat atau memperbarui
// baris user, menghitung level, dan mengembalikan achievement yang baru terbuka.
export async function loadProfile(): Promise<void> {
  const { profile, newlyUnlocked } = await api.profile.sync();
  appStore.setProfile(profile);
  appStore.pushAchievements(newlyUnlocked ?? []);
}
// Port formatStudyTime di Dashboard.tsx lama.
export function formatStudyTime(seconds: number): string {
  if (seconds <= 0) return '0d';
  const hrs = Math.floor(seconds / 3600);
  const mins = Math.floor((seconds % 3600) / 60);
  const secs = seconds % 60;
  if (hrs > 0) return `${hrs}j ${mins}m`;
  if (mins > 0) return `${mins}m ${secs}s`;
  return `${secs}s`;
}
