<script lang="ts">
  import type { Html5Qrcode } from 'html5-qrcode';

  let {
    onScan,
    onCameraError,
  }: { onScan: (text: string) => void; onCameraError: (message: string) => void } = $props();

  // html5-qrcode hanya dimuat di halaman scan; jeda 500 ms sama dengan kode lama supaya
  // elemen #qr-reader sudah ter-render sebelum kamera dibuka.
  $effect(() => {
    let scanner: Html5Qrcode | null = null;
    let cancelled = false;
    const timer = setTimeout(async () => {
      if (!document.getElementById('qr-reader')) return;
      try {
        const { Html5Qrcode } = await import('html5-qrcode');
        if (cancelled) return;
        scanner = new Html5Qrcode('qr-reader');
        await scanner
          .start(
            { facingMode: 'environment' },
            { fps: 10, qrbox: { width: 250, height: 250 } },
            async (decodedText) => {
              if (scanner?.isScanning) await scanner.stop();
              onScan(decodedText);
            },
            () => {},
          )
          .catch(() => {
            onCameraError('Gagal mengakses kamera. Pastikan browser memiliki izin kamera.');
          });
      } catch (initError) {
        console.error('Scanner init error:', initError);
      }
    }, 500);
    return () => {
      cancelled = true;
      clearTimeout(timer);
      if (scanner?.isScanning) scanner.stop().catch(console.error);
    };
  });
</script>

<div
  id="qr-reader"
  class="w-full border-none [&_video]:h-[300px] [&_video]:object-cover [&>div]:border-none"
></div>
