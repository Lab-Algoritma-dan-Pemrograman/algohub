import type { Achievement, Profile } from '@algohub/api-client/elearning';
// Pengganti store/useStore.ts dan useProgress.ts. Tidak ada `page`: navigasi adalah URL.
// Kurikulum dan progres lesson ada di cache TanStack Query, bukan di sini.
class AppStore {
  profile = $state<Profile | null>(null);
  sidebarOpen = $state(true);
  soundEnabled = $state(true);
  achievementQueue = $state<Achievement[]>([]);
  levelUp = $state<number | null>(null);
  // Achievement yang sedang ditampilkan adalah kepala antrean; popup level-up menunggu
  // sampai antrean kosong.
  unlockedAchievement = $derived(this.achievementQueue[0] ?? null);
  setProfile(profile: Profile | null) {
    this.profile = profile;
  }
  patchProfile(patch: Partial<Profile>) {
    if (this.profile) this.profile = { ...this.profile, ...patch };
  }
  pushAchievement(achievement: Achievement) {
    if (this.achievementQueue.some((a) => a.id === achievement.id)) return;
    this.achievementQueue = [...this.achievementQueue, achievement];
  }
  pushAchievements(achievements: readonly Achievement[]) {
    for (const achievement of achievements) this.pushAchievement(achievement);
  }
  shiftAchievement() {
    this.achievementQueue = this.achievementQueue.slice(1);
  }
  setLevelUp(level: number | null) {
    this.levelUp = level;
  }
  toggleSidebar() {
    this.sidebarOpen = !this.sidebarOpen;
  }
  toggleSound() {
    this.soundEnabled = !this.soundEnabled;
  }
  // Dipanggil saat logout supaya data user lama tidak tampil untuk user berikutnya.
  reset() {
    this.profile = null;
    this.achievementQueue = [];
    this.levelUp = null;
  }
}
export const appStore = new AppStore();
