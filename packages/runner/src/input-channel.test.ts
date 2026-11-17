import { afterEach, describe, expect, it, vi } from 'vitest';
import { SAB_SIZE } from './constants';
import { createSab, InputChannel } from './input-channel';
const setup = () => {
  const sab = new SharedArrayBuffer(SAB_SIZE);
  return {
    channel: new InputChannel(sab),
    meta: new Int32Array(sab, 0, 2),
    data: new Uint8Array(sab, 8),
  };
};
describe('createSab', () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });
  it('null bila halaman tidak cross-origin isolated', () => {
    vi.stubGlobal('crossOriginIsolated', false);
    expect(createSab()).toBeNull();
  });
  it('membuat SharedArrayBuffer berukuran SAB_SIZE saat isolated', () => {
    vi.stubGlobal('crossOriginIsolated', true);
    expect(createSab()?.byteLength).toBe(SAB_SIZE);
  });
});
describe('InputChannel', () => {
  it('menulis baris + newline mulai byte 8, panjang, lalu flag 1', () => {
    const { channel, meta, data } = setup();
    expect(channel.send('é1')).toBe(true);
    expect(Array.from(meta)).toEqual([1, 4]);
    expect(new TextDecoder().decode(data.slice(0, 4))).toBe('é1\n');
  });
  it('menolak baris yang melebihi SAB_SIZE - 8 byte', () => {
    const { channel, meta } = setup();
    const max = 'a'.repeat(SAB_SIZE - 8 - 1);
    expect(channel.fits(max)).toBe(true);
    expect(channel.fits(`${max}a`)).toBe(false);
    expect(channel.send(`${max}a`)).toBe(false);
    expect(Array.from(meta)).toEqual([0, 0]);
    expect(channel.send(max)).toBe(true);
    expect(meta[1]).toBe(SAB_SIZE - 8);
  });
  it('cancel memasang flag 2 dan reset mengosongkan', () => {
    const { channel, meta } = setup();
    channel.send('x');
    channel.cancel();
    expect(meta[0]).toBe(2);
    channel.reset();
    expect(Array.from(meta)).toEqual([0, 0]);
  });
});
