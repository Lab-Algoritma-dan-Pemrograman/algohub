import { describe, expect, it } from 'vitest';
import { confirm, confirmState, settleConfirm } from '../src/lib/stores/confirm.svelte';

describe('confirm', () => {
  it('membuka dialog dan menyelesaikan promise dengan jawaban user', async () => {
    const answer = confirm('Hapus barang ini?');
    expect(confirmState.open).toBe(true);
    expect(confirmState.message).toBe('Hapus barang ini?');
    settleConfirm(true);
    await expect(answer).resolves.toBe(true);
    expect(confirmState.open).toBe(false);
  });

  it('pemanggilan beruntun membatalkan dialog sebelumnya', async () => {
    const first = confirm('Pertama');
    const second = confirm('Kedua');
    await expect(first).resolves.toBe(false);
    expect(confirmState.message).toBe('Kedua');
    settleConfirm(false);
    await expect(second).resolves.toBe(false);
  });

  it('emoji bawaan pesan lama dibuang', () => {
    void confirm('🗑 ⚠ Hapus data? 🌟');
    expect(confirmState.message).toBe('Hapus data?');
    settleConfirm(false);
  });
});
