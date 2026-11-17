import { describe, expect, it, vi } from 'vitest';

const base = vi.hoisted(() =>
  Object.assign(vi.fn(), {
    success: vi.fn(),
    error: vi.fn(),
    warning: vi.fn(),
    info: vi.fn(),
    custom: vi.fn(),
    dismiss: vi.fn(),
    promise: vi.fn(),
  }),
);

vi.mock('@algohub/ui', () => ({ toast: base }));

const { makeCheerful, toast } = await import('../src/lib/toast');

describe('toast', () => {
  it('semua varian melewati makeCheerful dan meneruskan opsi', () => {
    toast.error('Gagal: timeout', { description: 'd' });
    expect(base.error).toHaveBeenCalledWith(
      'Oops! Ada sedikit kendala: "timeout" 🥺 Coba cek kembali ya!',
      { description: 'd' },
    );
    toast.success('Login Berhasil!');
    expect(base.success).toHaveBeenCalledWith(
      'Hore! Aksi kamu berhasil dilaksanakan! 🥳🚀',
      undefined,
    );
    expect(toast.dismiss).toBe(base.dismiss);
  });
});

describe('makeCheerful', () => {
  it.each([
    ['Data berhasil disimpan', 'Yeay! Data berhasil disimpan dengan aman! 🎉✨'],
    ['Barang dihapus', 'Sip! Data telah berhasil dihapus dari sistem! 🗑👋'],
    ['Profil berhasil diperbarui!', 'Mantap! Pembaruan berhasil diterapkan! 🚀🌟'],
    ['Login Berhasil!', 'Hore! Aksi kamu berhasil dilaksanakan! 🥳🚀'],
    [
      'Gagal update: koneksi',
      'Oops! Ada sedikit kendala: "Gagal update: koneksi" 🥺 Coba cek kembali ya!',
    ],
    ['Gagal: timeout', 'Oops! Ada sedikit kendala: "timeout" 🥺 Coba cek kembali ya!'],
    [
      'Selamat datang',
      'Halo! Selamat datang kembali di Lab AP! Semangat belajar dan berkarya! 💻🌟',
    ],
    ['Terjadi Kesalahan', 'Terjadi Kesalahan 🎉'],
  ])('%s', (input, expected) => {
    expect(makeCheerful(input)).toBe(expected);
  });

  it('bukan string dikembalikan apa adanya', () => {
    const node = { a: 1 };
    expect(makeCheerful(node)).toBe(node);
  });
});
