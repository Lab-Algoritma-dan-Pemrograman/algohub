import { toast as baseToast } from '@algohub/ui';

type Message = Parameters<typeof baseToast>[0];
type Options = Parameters<typeof baseToast>[1];

// App lama meng-alias "sonner" ke pembungkus ini, jadi semua pesan toast melewati makeCheerful.
export function makeCheerful<M>(msg: M): M | string {
  if (typeof msg !== 'string') return msg;
  const lower = msg.toLowerCase();
  if (
    lower.includes('berhasil disimpan') ||
    lower.includes('sukses disimpan') ||
    lower.includes('saved successfully')
  ) {
    return `Yeay! Data berhasil disimpan dengan aman! 🎉✨`;
  }
  if (lower.includes('dihapus') || lower.includes('deleted')) {
    return `Sip! Data telah berhasil dihapus dari sistem! 🗑👋`;
  }
  if (lower.includes('berhasil diperbarui') || lower.includes('berhasil diupdate')) {
    return `Mantap! Pembaruan berhasil diterapkan! 🚀🌟`;
  }
  if (lower.includes('berhasil') || lower.includes('sukses') || lower.includes('success')) {
    return `Hore! Aksi kamu berhasil dilaksanakan! 🥳🚀`;
  }
  if (lower.includes('gagal') || lower.includes('error') || lower.includes('failed')) {
    const cleanMsg = msg.replace(/gagal( melakukan)?\s*:\s*/gi, '').replace(/error\s*:\s*/gi, '');
    return `Oops! Ada sedikit kendala: "${cleanMsg}" 🥺 Coba cek kembali ya!`;
  }
  if (lower.includes('selamat datang') || lower.includes('welcome')) {
    return `Halo! Selamat datang kembali di Lab AP! Semangat belajar dan berkarya! 💻🌟`;
  }
  return `${msg} 🎉`;
}

const wrap =
  (fn: (message: Message, options?: Options) => string | number) =>
  (message: Message, options?: Options) =>
    fn(makeCheerful(message) as Message, options);

export const toast = Object.assign(wrap(baseToast), {
  success: wrap(baseToast.success),
  error: wrap(baseToast.error),
  warning: wrap(baseToast.warning),
  info: wrap(baseToast.info),
  custom: baseToast.custom,
  dismiss: baseToast.dismiss,
  promise: baseToast.promise,
});
