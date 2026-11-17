// Tipe struktural supaya modul ini bisa diuji di luar ServiceWorkerGlobalScope.
export interface PushScope {
  registration: Pick<ServiceWorkerRegistration, 'showNotification'>;
}

interface WindowClientLike {
  url: string;
  focus?: () => Promise<unknown>;
}

export interface ClickScope {
  clients: {
    matchAll(options: {
      type: 'window';
      includeUncontrolled: boolean;
    }): Promise<readonly WindowClientLike[]>;
    openWindow?: (url: string) => Promise<unknown>;
  };
}

interface PushLikeEvent {
  data: { json(): unknown; text(): string } | null;
  waitUntil(promise: Promise<unknown>): void;
}

interface ClickLikeEvent {
  notification: Pick<Notification, 'close' | 'data'>;
  waitUntil(promise: Promise<unknown>): void;
}

const DEFAULT_URL = '/jadwal-jaga';

export function handlePush(event: PushLikeEvent, scope: PushScope): void {
  let payload: { title: string; body: string; url?: string } = {
    title: '🔔 Notifikasi Lab AP',
    body: 'Ada pembaruan jadwal jaga!',
    url: DEFAULT_URL,
  };
  if (event.data) {
    try {
      payload = event.data.json() as typeof payload;
    } catch {
      payload.body = event.data.text();
    }
  }
  event.waitUntil(
    scope.registration.showNotification(payload.title, {
      body: payload.body,
      icon: '/logo.png',
      badge: '/logo.png',
      vibrate: [200, 100, 200],
      data: { url: payload.url || DEFAULT_URL },
      actions: [{ action: 'open', title: 'Buka Aplikasi' }],
    } as NotificationOptions),
  );
}

export function handleNotificationClick(event: ClickLikeEvent, scope: ClickScope): void {
  event.notification.close();
  const targetUrl: string = event.notification.data?.url || DEFAULT_URL;
  event.waitUntil(
    scope.clients
      .matchAll({ type: 'window', includeUncontrolled: true })
      .then((clientList): Promise<unknown> | undefined => {
        for (const client of clientList) {
          if (client.url.includes(targetUrl) && client.focus) return client.focus();
        }
        if (scope.clients.openWindow) return scope.clients.openWindow(targetUrl);
      }),
  );
}
