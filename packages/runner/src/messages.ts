export const messages = {
  loadingC: 'Menyiapkan compiler C (clang+WASI)… unduhan pertama bisa ±1 menit.',
  loadingPython: 'Menyiapkan Python (Pyodide)… unduhan pertama bisa ±1 menit.',
  timeout:
    'Program tidak merespons dalam 40 detik (kemungkinan infinite loop) — compiler di-reset. Periksa kode Anda.',
  compileFailed: 'Gagal melakukan kompilasi.',
  unavailableC: 'Compiler Clang gagal dimuat atau sedang bermasalah. Silakan muat ulang halaman.',
  unavailablePython:
    'Interpreter Python gagal dimuat atau sedang bermasalah. Pastikan koneksi internet stabil dan muat ulang halaman.',
  stopped: 'Eksekusi dihentikan oleh pengguna.',
  cancelled: 'Program dibatalkan.',
  done: 'Program selesai!',
  noInteractive: 'Mode input interaktif tidak tersedia (SAB mati).',
  idle: 'Klik RUN untuk menguji kode Anda...',
} as const;
