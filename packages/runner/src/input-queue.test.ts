import { describe, expect, it } from 'vitest';
import { InputQueue } from './input-queue';
const fakeChannel = (limit = 10) => {
  const sent: string[] = [];
  return {
    sent,
    fits: (line: string) => line.length <= limit,
    send: (line: string) => {
      if (line.length > limit) return false;
      sent.push(line);
      return true;
    },
  };
};
describe('InputQueue', () => {
  it('mengirim langsung saat program sedang menunggu', () => {
    const channel = fakeChannel();
    const queue = new InputQueue(channel);
    queue.needInput();
    expect(queue.waiting).toBe(true);
    expect(queue.push('3')).toBe(true);
    expect(channel.sent).toEqual(['3']);
    expect(queue.waiting).toBe(false);
  });
  it('mengantre ketikan sebelum diminta lalu mengirim FIFO untuk dua scanf', () => {
    const channel = fakeChannel();
    const queue = new InputQueue(channel);
    queue.push('3');
    queue.push('4');
    expect(channel.sent).toEqual([]);
    queue.needInput();
    expect(channel.sent).toEqual(['3']);
    expect(queue.waiting).toBe(false);
    queue.needInput();
    expect(channel.sent).toEqual(['3', '4']);
    queue.needInput();
    expect(queue.waiting).toBe(true);
  });
  it('menolak baris terlalu panjang tanpa mengantre', () => {
    const channel = fakeChannel(2);
    const queue = new InputQueue(channel);
    expect(queue.push('panjang')).toBe(false);
    queue.needInput();
    expect(channel.sent).toEqual([]);
    expect(queue.push('panjang')).toBe(false);
    expect(queue.waiting).toBe(true);
  });
  it('clear membuang antrean dan status menunggu', () => {
    const channel = fakeChannel();
    const queue = new InputQueue(channel);
    queue.push('1');
    queue.clear();
    queue.needInput();
    expect(channel.sent).toEqual([]);
    queue.clear();
    expect(queue.waiting).toBe(false);
  });
});
