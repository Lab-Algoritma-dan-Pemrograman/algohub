import { describe, expect, it } from 'vitest';
import { shouldVirtualize, VIRTUAL_THRESHOLD, virtualPadding } from './virtual';
describe('shouldVirtualize', () => {
  it('aktif hanya bila jumlah baris melebihi ambang', () => {
    expect(VIRTUAL_THRESHOLD).toBe(100);
    expect(shouldVirtualize(0)).toBe(false);
    expect(shouldVirtualize(100)).toBe(false);
    expect(shouldVirtualize(101)).toBe(true);
    expect(shouldVirtualize(1_200)).toBe(true);
  });
  it('menerima ambang khusus', () => {
    expect(shouldVirtualize(10, 10)).toBe(false);
    expect(shouldVirtualize(11, 10)).toBe(true);
  });
});
describe('virtualPadding', () => {
  it('nol saat tidak ada item terlihat', () => {
    expect(virtualPadding([], 5_000)).toEqual({ top: 0, bottom: 0 });
  });
  it('mengisi ruang di atas item pertama dan di bawah item terakhir', () => {
    const items = [
      { start: 480, end: 528 },
      { start: 528, end: 576 },
    ];
    expect(virtualPadding(items, 57_600)).toEqual({ top: 480, bottom: 57_024 });
  });
});
