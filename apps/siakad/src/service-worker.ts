/// <reference no-default-lib="true"/>
/// <reference lib="esnext" />
/// <reference lib="webworker" />
import { handleNotificationClick, handlePush } from '#lib/sw/handlers.ts';

const sw = self as unknown as ServiceWorkerGlobalScope;

// Tidak ada handler fetch: respons API tidak boleh di-cache service worker.
sw.addEventListener('install', () => {
  void sw.skipWaiting();
});

sw.addEventListener('activate', (event) => {
  event.waitUntil(sw.clients.claim());
});

sw.addEventListener('push', (event) => handlePush(event, sw));

sw.addEventListener('notificationclick', (event) => handleNotificationClick(event, sw));
