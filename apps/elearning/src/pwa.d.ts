// Modul virtual dari @vite-pwa/sveltekit; tipenya tidak diekspor oleh paket.
declare module 'virtual:pwa-register' {
  export function registerSW(options?: {
    immediate?: boolean;
  }): (reloadPage?: boolean) => Promise<void>;
}
