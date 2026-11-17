import { dev } from '$app/env';
import { PUBLIC_VAPID_PUBLIC_KEY } from '$app/env/public';
import { rpc, sendPush, sendWa } from './api.ts';
import { getDailyQuote } from './quotes.ts';
import { toast } from './toast.ts';

interface NotifyUser {
  id: number | string;
  role?: string;
  full_name?: string;
}

export const urlBase64ToUint8Array = (base64String: string) => {
  const padding = '='.repeat((4 - (base64String.length % 4)) % 4);
  const base64 = (base64String + padding).replace(/-/g, '+').replace(/_/g, '/');
  const rawData = window.atob(base64);
  const outputArray = new Uint8Array(rawData.length);
  for (let i = 0; i < rawData.length; ++i) outputArray[i] = rawData.charCodeAt(i);
  return outputArray;
};

export const registerServiceWorker = async () => {
  if (!('serviceWorker' in navigator)) return null;
  try {
    // Saat dev SvelteKit menyajikan service worker sebagai ES module.
    return await navigator.serviceWorker.register('/service-worker.js', {
      type: dev ? 'module' : 'classic',
    });
  } catch {
    return null;
  }
};

export const subscribeToWebPush = async (user: NotifyUser | null | undefined) => {
  if (!user || !('serviceWorker' in navigator) || !('PushManager' in window)) return;
  if (Notification.permission !== 'granted') return;
  try {
    await registerServiceWorker();
    const reg = await navigator.serviceWorker.ready;
    let sub = await reg.pushManager.getSubscription();
    if (!sub) {
      sub = await reg.pushManager.subscribe({
        userVisibleOnly: true,
        applicationServerKey: urlBase64ToUint8Array(PUBLIC_VAPID_PUBLIC_KEY),
      });
    }
    const subJson = sub.toJSON();
    if (subJson.endpoint && subJson.keys?.p256dh && subJson.keys?.auth) {
      // Satu jalur saja: gateway menyuntik p_caller_id dari JWT dan hanya RPC ini yang punya
      // GRANT di produksi; upsert langsung ke tabel dulu gagal 401 tetapi tercatat berhasil.
      const { error } = await rpc('save_push_subscription_secure', {
        p_caller_id: Number(user.id),
        p_endpoint: subJson.endpoint,
        p_p256dh: subJson.keys.p256dh,
        p_auth: subJson.keys.auth,
        p_user_agent: navigator.userAgent,
      });
      if (error) console.error('Gagal menyimpan langganan VAPID Web Push:', error.message);
    }
  } catch (err) {
    console.error('Gagal berlangganan VAPID Web Push:', err);
  }
};

export const unsubscribeWebPush = async (user?: NotifyUser | null) => {
  if (!('serviceWorker' in navigator)) return;
  try {
    const reg = await navigator.serviceWorker.getRegistration();
    const sub = await reg?.pushManager.getSubscription();
    if (!sub) return;
    // Lewat RPC karena role anon tidak punya GRANT ke tabel; tanpa ini token push lama tidak
    // pernah terhapus dan perangkat yang sama tetap menerima notifikasi pengguna sebelumnya.
    if (user?.id) {
      const { error } = await rpc('delete_push_subscription_secure', {
        p_caller_id: Number(user.id),
        p_endpoint: sub.endpoint,
      });
      if (error) console.error('Gagal menghapus langganan push dari DB:', error.message);
    }
    await sub.unsubscribe();
  } catch (err) {
    console.error('Gagal menghapus push subscription saat logout:', err);
  }
};

export const triggerServerWebPush = async (
  targetUserId: number,
  title: string,
  body: string,
  url: string = '/jadwal-jaga',
) => {
  try {
    const res = await sendPush<{ sent?: number }>({ targetUserId, title, body, url });
    const data = res.body ?? {};
    if (data.sent === 0) {
      toast.warning(
        "Belum ada token device HP terdaftar di database. Silakan klik tombol 'Notifikasi Browser Aktif' di HP kamu dulu!",
      );
    } else if (data.sent && data.sent > 0) {
      toast.success(`Server VAPID Push terkirim ke ${data.sent} device!`);
    }
    return data;
  } catch (err) {
    console.error('Gagal mengirim Server VAPID Push:', err);
  }
};

export const triggerServerWhatsApp = async (
  targetUserId: number,
  message: string,
  targetPhone?: string,
) => {
  try {
    const res = await sendWa<{ sent?: unknown; error?: string }>({
      targetUserId,
      targetPhone,
      message,
    });
    const data = res.body ?? {};
    if (data.sent) {
      toast.success('Pesan WhatsApp otomatis terkirim dari server!');
    } else if (data.error && !data.error.includes('belum dikonfigurasi')) {
      toast.error('Status WA Gateway: ' + data.error);
    }
    return data;
  } catch (err) {
    console.error('Gagal mengirim Server WhatsApp Broadcast:', err);
  }
};

export const requestNotificationPermission = async (user?: NotifyUser | null) => {
  if (!('Notification' in window)) return false;
  await registerServiceWorker();
  let isGranted = Notification.permission === 'granted';
  if (!isGranted && Notification.permission !== 'denied') {
    isGranted = (await Notification.requestPermission()) === 'granted';
  }
  if (isGranted && user) await subscribeToWebPush(user);
  return isGranted;
};

export const getNotificationPermissionState = (): NotificationPermission | 'unsupported' => {
  if (!('Notification' in window)) return 'unsupported';
  return Notification.permission;
};

export interface ShiftNotificationOptions {
  body: string;
  icon?: string;
  tag?: string;
  onClickUrl?: string;
}

export const formatLocalDateStr = (d: Date): string => {
  const year = d.getFullYear();
  const month = String(d.getMonth() + 1).padStart(2, '0');
  const day = String(d.getDate()).padStart(2, '0');
  return `${year}-${month}-${day}`;
};

// Riwayat dihapus supaya shift yang diedit ulang bisa memicu notifikasi lagi.
export const clearShiftNotifiedHistory = (assignmentId?: string | number) => {
  try {
    const keys = Object.keys(sessionStorage).filter((key) => key.startsWith('notified_shifts_'));
    if (assignmentId) {
      const idStr = assignmentId.toString();
      for (const key of keys) {
        const list: string[] = JSON.parse(sessionStorage.getItem(key) || '[]');
        sessionStorage.setItem(key, JSON.stringify(list.filter((item) => item !== idStr)));
      }
    } else {
      for (const key of keys) sessionStorage.removeItem(key);
    }
  } catch (e) {
    console.error('Gagal menghapus riwayat notifikasi:', e);
  }
};

export const playNotificationSound = () => {
  try {
    const AudioContextClass =
      window.AudioContext ||
      (window as unknown as { webkitAudioContext?: typeof AudioContext }).webkitAudioContext;
    if (!AudioContextClass) return;
    const ctx = new AudioContextClass();
    const now = ctx.currentTime;
    const tone = (freq: number, start: number, gain: number, end: number) => {
      const osc = ctx.createOscillator();
      const g = ctx.createGain();
      osc.type = 'sine';
      osc.frequency.setValueAtTime(freq, now + start);
      g.gain.setValueAtTime(gain, now + start);
      g.gain.exponentialRampToValueAtTime(0.001, now + end);
      osc.connect(g);
      g.connect(ctx.destination);
      osc.start(now + start);
      osc.stop(now + end);
    };
    tone(880, 0, 0.15, 0.15);
    tone(1320, 0.1, 0.2, 0.35);
  } catch {
    // Autoplay audio bisa ditolak browser sebelum ada interaksi; notifikasi tetap jalan.
  }
};

export const sendBrowserNotification = async (title: string, options: ShiftNotificationOptions) => {
  playNotificationSound();
  toast.info(title, { description: options.body, duration: 8000 });
  if (!('Notification' in window) || Notification.permission !== 'granted') return;
  try {
    const defaultIcon = '/logo.png';
    const notificationOptions: NotificationOptions & Record<string, unknown> = {
      body: options.body,
      icon: options.icon || defaultIcon,
      tag: options.tag || `jadwal-jaga-${Date.now()}`,
      badge: defaultIcon,
      vibrate: [200, 100, 200, 100, 200],
      requireInteraction: true,
      renotify: true,
      data: { url: options.onClickUrl || '/jadwal-jaga' },
    };
    if ('serviceWorker' in navigator) {
      const reg = await navigator.serviceWorker.getRegistration();
      if (reg?.showNotification) {
        await reg.showNotification(title, notificationOptions);
        return;
      }
    }
    const notification = new Notification(title, notificationOptions);
    notification.onclick = (event) => {
      event.preventDefault();
      window.focus();
      if (options.onClickUrl) window.location.href = options.onClickUrl;
      notification.close();
    };
  } catch (err) {
    console.error('Gagal mengirim notifikasi browser:', err);
  }
};

export interface ShiftAssignment {
  id: number | string;
  activity_date?: string | null;
  task_role?: string | null;
  activity_name?: string | null;
  user_id?: number | null;
  user?: { id?: number | null } | null;
  schedule?: {
    start_time?: string | null;
    major?: string | null;
    class_code?: string | null;
  } | null;
}

// Tahap pengingat: H-1, hari ini, 2 jam, 1 jam, dan 30 menit sebelum shift.
export const checkAndNotifyUpcomingShifts = (
  user: NotifyUser | null | undefined,
  assignments: ShiftAssignment[] | null | undefined,
) => {
  if (!user || !assignments || assignments.length === 0) return;
  const now = new Date();
  const todayStr = formatLocalDateStr(now);
  const tomorrow = new Date(now);
  tomorrow.setDate(now.getDate() + 1);
  const tomorrowStr = formatLocalDateStr(tomorrow);
  const quote = getDailyQuote(user.role || 'asisten');
  const tracker = (stage: string) => {
    const key = `notified_shifts_${stage}_${user.id}_${todayStr}`;
    const list: string[] = JSON.parse(sessionStorage.getItem(key) || '[]');
    return {
      has: (id: string) => list.includes(id),
      add: (id: string) => {
        list.push(id);
        sessionStorage.setItem(key, JSON.stringify(list));
      },
    };
  };
  const notifiedH1 = tracker('h1');
  const notifiedGeneral = tracker('today_general');
  const notified2h = tracker('2h');
  const notified1h = tracker('1h');
  const notified30m = tracker('30m');
  const myAssignments = assignments.filter((a) => a.user?.id === user.id || a.user_id === user.id);
  for (const assignment of myAssignments) {
    if (!assignment.activity_date) continue;
    const assignDateStr = assignment.activity_date.split('T')[0];
    const assignId = assignment.id.toString();
    const taskRole = assignment.task_role || 'Petugas Jaga';
    const activityName = assignment.activity_name || 'Praktikum';
    const startTimeStr = assignment.schedule?.start_time?.slice(0, 5) || '00:00';
    const scheduleInfo = `(Jam ${startTimeStr} WIB - ${assignment.schedule?.major || ''} ${assignment.schedule?.class_code || ''})`;
    if (assignDateStr === tomorrowStr && !notifiedH1.has(assignId)) {
      sendBrowserNotification(`⏰ Pengingat H-1 Jadwal Jaga`, {
        body: `Besok kamu bertugas sebagai ${taskRole} di ${activityName} ${scheduleInfo}. Siapkan materi ya!\n\n✨ "${quote}"`,
        tag: `shift-h1-${assignId}`,
        onClickUrl: '/jadwal-jaga',
      });
      notifiedH1.add(assignId);
    }
    if (assignDateStr !== todayStr) continue;
    if (!notifiedGeneral.has(assignId)) {
      sendBrowserNotification(`📌 Jadwal Jaga Hari Ini!`, {
        body: `Halo ${user.full_name || 'Asisten'}, hari ini kamu bertugas sebagai ${taskRole} di ${activityName} ${scheduleInfo}.\n\n✨ "${quote}"`,
        tag: `shift-today-gen-${assignId}`,
        onClickUrl: '/jadwal-jaga',
      });
      notifiedGeneral.add(assignId);
    }
    if (!assignment.schedule?.start_time) continue;
    const [hours, minutes] = assignment.schedule.start_time.split(':');
    const shiftTime = new Date(now);
    shiftTime.setHours(parseInt(hours, 10), parseInt(minutes, 10), 0, 0);
    const diffMinutes = Math.floor((shiftTime.getTime() - now.getTime()) / 60_000);
    if (diffMinutes <= 120 && diffMinutes > 60 && !notified2h.has(assignId)) {
      sendBrowserNotification(`⏰ 2 Jam Lagi! Jadwal Jaga`, {
        body: `Halo ${user.full_name || 'Asisten'}, 2 jam lagi kamu bertugas sebagai ${taskRole} di ${activityName} (Pukul ${startTimeStr} WIB).\n\n✨ "${quote}"`,
        tag: `shift-2h-${assignId}`,
        onClickUrl: '/jadwal-jaga',
      });
      notified2h.add(assignId);
    }
    if (diffMinutes <= 60 && diffMinutes > 30 && !notified1h.has(assignId)) {
      sendBrowserNotification(`⏳ 1 Jam Lagi! Jadwal Jaga`, {
        body: `Persiapkan diri! 1 jam lagi kamu bertugas sebagai ${taskRole} di ${activityName} (Pukul ${startTimeStr} WIB).\n\n✨ "${quote}"`,
        tag: `shift-1h-${assignId}`,
        onClickUrl: '/jadwal-jaga',
      });
      notified1h.add(assignId);
    }
    if (diffMinutes <= 30 && diffMinutes > 0 && !notified30m.has(assignId)) {
      sendBrowserNotification(`🚨 30 Menit Lagi! Jadwal Jaga Segera Dimulai`, {
        body: `Perhatian! 30 menit lagi kamu bertugas sebagai ${taskRole} di ${activityName} (Pukul ${startTimeStr} WIB). Segera menuju laboratorium!\n\n✨ "${quote}"`,
        tag: `shift-30m-${assignId}`,
        onClickUrl: '/jadwal-jaga',
      });
      notified30m.add(assignId);
    }
  }
};
