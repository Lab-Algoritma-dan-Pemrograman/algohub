import { isRedirect } from '@sveltejs/kit';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { requireRole, requireToken } from './guard';
import { logout } from './logout';
import { Session } from './session.svelte';
type User = { id: number; role: string; nama: string };
const mahasiswa: User = { id: 1, role: 'mahasiswa', nama: 'A' };
function redirectOf(fn: () => void) {
  try {
    fn();
  } catch (e) {
    if (isRedirect(e)) return { status: e.status, location: e.location };
    throw e;
  }
  return null;
}
beforeEach(() => {
  localStorage.clear();
});
afterEach(() => {
  vi.restoreAllMocks();
});
describe('Session', () => {
  it('membaca localStorage hanya saat diakses', () => {
    localStorage.setItem('user', JSON.stringify(mahasiswa));
    const getItem = vi.spyOn(Storage.prototype, 'getItem');
    const session = new Session<User>();
    expect(getItem).not.toHaveBeenCalled();
    expect(session.user).toEqual(mahasiswa);
    expect(getItem).toHaveBeenCalledWith('user');
  });
  it('set dan clear menyimpan serta menghapus token dan user', () => {
    const session = new Session<User>();
    expect(session.user).toBeNull();
    expect(session.hasToken).toBe(false);
    session.set('tok', mahasiswa);
    expect(localStorage.getItem('token')).toBe('tok');
    expect(JSON.parse(localStorage.getItem('user')!)).toEqual(mahasiswa);
    expect(session.user).toEqual(mahasiswa);
    expect(session.token).toBe('tok');
    expect(session.hasToken).toBe(true);
    session.clear();
    expect(localStorage.getItem('token')).toBeNull();
    expect(localStorage.getItem('user')).toBeNull();
    expect(session.user).toBeNull();
    expect(session.hasToken).toBe(false);
  });
});
describe('guard', () => {
  it('requireToken melempar redirect 307 tanpa token', () => {
    const session = new Session<User>();
    expect(redirectOf(() => requireToken(session, '/praktikum/login'))).toEqual({
      status: 307,
      location: '/praktikum/login',
    });
    session.set('tok', mahasiswa);
    expect(redirectOf(() => requireToken(session, '/praktikum/login'))).toBeNull();
  });
  it('requireRole melempar redirect 307 bila role tidak cocok atau belum login', () => {
    const session = new Session<User>();
    const admin = ['asisten', 'koordinator'];
    expect(redirectOf(() => requireRole(session, admin, '/praktikum/dashboard'))).toEqual({
      status: 307,
      location: '/praktikum/dashboard',
    });
    session.set('tok', mahasiswa);
    expect(redirectOf(() => requireRole(session, admin, '/praktikum/dashboard'))).toEqual({
      status: 307,
      location: '/praktikum/dashboard',
    });
  });
  it('requireRole tidak melempar bila role cocok', () => {
    const session = new Session<User>();
    session.set('tok', { ...mahasiswa, role: 'koordinator' });
    expect(redirectOf(() => requireRole(session, ['koordinator'], '/x'))).toBeNull();
  });
});
describe('logout', () => {
  it('urutan: close → queryClient.clear → removeClient → session.clear', async () => {
    const calls: string[] = [];
    let release!: () => void;
    const removed = new Promise<void>((resolve) => (release = resolve));
    const done = logout({
      persister: {
        close: () => calls.push('close'),
        removeClient: () => {
          calls.push('removeClient');
          return removed;
        },
      },
      queryClient: { clear: () => calls.push('clear') },
      session: { clear: () => calls.push('session.clear') },
    });
    expect(calls).toEqual(['close', 'clear', 'removeClient']);
    release();
    await done;
    expect(calls).toEqual(['close', 'clear', 'removeClient', 'session.clear']);
  });
  it('tanpa persister tetap membersihkan cache dan sesi', async () => {
    const calls: string[] = [];
    await logout({
      queryClient: { clear: () => calls.push('clear') },
      session: { clear: () => calls.push('session.clear') },
    });
    expect(calls).toEqual(['clear', 'session.clear']);
  });
});
