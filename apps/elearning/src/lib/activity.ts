import { PUBLIC_API_BASE_URL } from '$app/env/public';
import { api } from './api.ts';
import { session, type ElearningSessionUser } from './session.ts';
import { appStore } from './stores/app.svelte.ts';
const HEARTBEAT_MS = 60_000;
const TAB_STALE_MS = 4_000;
const TAB_KEY = 'study_time_active_tab';
const TAB_TIME_KEY = 'study_time_active_tab_timestamp';
// Nama aktivitas sama dengan nilai `page` di app lama supaya halaman monitoring tetap terbaca.
export function activityName(pathname: string): string {
  const segment = pathname.split('/').filter(Boolean)[0];
  if (!segment) return 'dashboard';
  return segment === 'lesson' ? 'lesson' : segment;
}
// Waktu belajar hanya dihitung di daftar kursus dan halaman lesson, seperti Layout.tsx lama.
export function isStudyActivity(activity: string): boolean {
  return activity === 'courses' || activity === 'lesson';
}
export interface ActivityOptions {
  user: ElearningSessionUser;
  pathname: () => string;
}
// Menggantikan blok useEffect di Layout.tsx lama: heartbeat tiap 30 detik, hitung waktu belajar
// hanya di tab aktif, dan kirim sisa detik saat halaman ditutup.
export function startActivity({ user, pathname }: ActivityOptions): () => void {
  const tabId = Math.random().toString(36).slice(2);
  let pending = 0;
  let sending = false;
  const claimTab = () => {
    const active = localStorage.getItem(TAB_KEY);
    const stamp = Number(localStorage.getItem(TAB_TIME_KEY) ?? 0);
    if (!active || active === tabId || Date.now() - stamp > TAB_STALE_MS) {
      localStorage.setItem(TAB_KEY, tabId);
      localStorage.setItem(TAB_TIME_KEY, String(Date.now()));
    }
    return localStorage.getItem(TAB_KEY) === tabId;
  };
  const tick = () => {
    // Tab hidden = user tidak sedang belajar: jangan tambah detik agar waktu belajar tak menggelembung palsu.
    if (document.hidden) return;
    if (claimTab() && isStudyActivity(activityName(pathname()))) {
      pending += 1;
      appStore.patchProfile({ studyTime: (appStore.profile?.studyTime ?? 0) + 1 });
    }
  };
  const heartbeat = async () => {
    if (sending) return;
    sending = true;
    const seconds = pending;
    try {
      const res = await api.sessions.heartbeat({
        currentActivity: activityName(pathname()),
        nama: user.nama || 'Anonymous',
        kelas: user.kelas || 'Unknown',
        touchStreak: true,
        studySeconds: seconds,
      });
      pending -= seconds;
      if (res.streak) {
        appStore.patchProfile({ streak: res.streak.streak, lastActive: res.streak.lastActive });
      }
      if (typeof res.studyTime?.studyTime === 'number') {
        appStore.patchProfile({ studyTime: res.studyTime.studyTime });
      }
    } catch {
      // Detik yang belum terkirim tetap di `pending` dan ikut heartbeat berikutnya.
    } finally {
      sending = false;
    }
  };
  const flushOnHide = () => {
    const token = session.token;
    if (!token) return;
    void fetch(`${PUBLIC_API_BASE_URL}/api/elearning/sessions/me`, {
      method: 'PUT',
      keepalive: true,
      headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
      body: JSON.stringify({
        currentActivity: activityName(pathname()),
        studySeconds: pending,
      }),
    }).catch(() => undefined);
  };
  claimTab();
  void heartbeat();
  const tickTimer = setInterval(tick, 1000);
  let beatTimer = setInterval(() => void heartbeat(), HEARTBEAT_MS);
  // Tab hidden = pause heartbeat & tick; kembali visible = restart + kirim satu heartbeat.
  const onVisibility = () => {
    if (document.hidden) {
      clearInterval(beatTimer);
    } else {
      clearInterval(beatTimer);
      beatTimer = setInterval(() => void heartbeat(), HEARTBEAT_MS);
      void heartbeat();
    }
  };
  document.addEventListener('visibilitychange', onVisibility);
  window.addEventListener('pagehide', flushOnHide);
  return () => {
    clearInterval(tickTimer);
    clearInterval(beatTimer);
    document.removeEventListener('visibilitychange', onVisibility);
    window.removeEventListener('pagehide', flushOnHide);
    if (localStorage.getItem(TAB_KEY) === tabId) {
      localStorage.removeItem(TAB_KEY);
      localStorage.removeItem(TAB_TIME_KEY);
    }
    // Setelah logout token sudah hilang; mengirim lagi akan membuat sesi aktif baru.
    if (session.hasToken) void heartbeat();
  };
}
