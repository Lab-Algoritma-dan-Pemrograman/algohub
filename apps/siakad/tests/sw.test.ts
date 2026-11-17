// @vitest-environment node
import { describe, expect, it, vi } from 'vitest';
import { handleNotificationClick, handlePush } from '../src/lib/sw/handlers';

function pushEvent(data: { json(): unknown; text(): string } | null) {
  const waits: Promise<unknown>[] = [];
  return { event: { data, waitUntil: (p: Promise<unknown>) => waits.push(p) }, waits };
}

function pushScope() {
  const showNotification = vi.fn(async () => undefined);
  return { scope: { registration: { showNotification } }, showNotification };
}

describe('handlePush', () => {
  it('payload JSON dipakai untuk judul, isi, dan url', async () => {
    const { event, waits } = pushEvent({
      json: () => ({ title: 'Shift', body: 'Besok jaga', url: '/absensi' }),
      text: () => '',
    });
    const { scope, showNotification } = pushScope();
    handlePush(event, scope);
    await Promise.all(waits);
    expect(showNotification).toHaveBeenCalledWith('Shift', {
      body: 'Besok jaga',
      icon: '/logo.png',
      badge: '/logo.png',
      vibrate: [200, 100, 200],
      data: { url: '/absensi' },
      actions: [{ action: 'open', title: 'Buka Aplikasi' }],
    });
  });

  it('payload teks menimpa isi bawaan', () => {
    const { event } = pushEvent({
      json: () => {
        throw new SyntaxError('bukan JSON');
      },
      text: () => 'Halo teks',
    });
    const { scope, showNotification } = pushScope();
    handlePush(event, scope);
    expect(showNotification).toHaveBeenCalledWith(
      '🔔 Notifikasi Lab AP',
      expect.objectContaining({ body: 'Halo teks', data: { url: '/jadwal-jaga' } }),
    );
  });

  it('tanpa data memakai judul dan isi bawaan sw.js', () => {
    const { event } = pushEvent(null);
    const { scope, showNotification } = pushScope();
    handlePush(event, scope);
    expect(showNotification).toHaveBeenCalledWith(
      '🔔 Notifikasi Lab AP',
      expect.objectContaining({
        body: 'Ada pembaruan jadwal jaga!',
        data: { url: '/jadwal-jaga' },
      }),
    );
  });

  it('JSON tanpa url jatuh ke /jadwal-jaga', () => {
    const { event } = pushEvent({ json: () => ({ title: 'T', body: 'B' }), text: () => '' });
    const { scope, showNotification } = pushScope();
    handlePush(event, scope);
    expect(showNotification).toHaveBeenCalledWith(
      'T',
      expect.objectContaining({ data: { url: '/jadwal-jaga' } }),
    );
  });
});

function clickEvent(data: unknown) {
  const waits: Promise<unknown>[] = [];
  const close = vi.fn();
  return {
    event: { notification: { close, data }, waitUntil: (p: Promise<unknown>) => waits.push(p) },
    waits,
    close,
  };
}

describe('handleNotificationClick', () => {
  it('memfokuskan jendela yang URL-nya memuat target', async () => {
    const focus = vi.fn(async () => undefined);
    const openWindow = vi.fn(async () => null);
    const { event, waits, close } = clickEvent({ url: '/absensi' });
    handleNotificationClick(event, {
      clients: {
        matchAll: async () => [
          { url: 'https://siakad.test/beranda' },
          { url: 'https://siakad.test/absensi', focus },
        ],
        openWindow,
      },
    });
    await Promise.all(waits);
    expect(close).toHaveBeenCalled();
    expect(focus).toHaveBeenCalled();
    expect(openWindow).not.toHaveBeenCalled();
  });

  it('membuka jendela baru kalau tidak ada yang cocok, default /jadwal-jaga', async () => {
    const openWindow = vi.fn(async () => null);
    const matchAll = vi.fn(async () => [{ url: 'https://siakad.test/beranda', focus: vi.fn() }]);
    const { event, waits, close } = clickEvent(undefined);
    handleNotificationClick(event, { clients: { matchAll, openWindow } });
    await Promise.all(waits);
    expect(close).toHaveBeenCalled();
    expect(matchAll).toHaveBeenCalledWith({ type: 'window', includeUncontrolled: true });
    expect(openWindow).toHaveBeenCalledWith('/jadwal-jaga');
  });
});
