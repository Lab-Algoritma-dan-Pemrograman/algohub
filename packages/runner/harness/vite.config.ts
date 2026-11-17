import { fileURLToPath } from 'node:url';
import { svelte } from '@sveltejs/vite-plugin-svelte';
import tailwindcss from '@tailwindcss/vite';
import { defineConfig } from 'vite';
import { compilerDir, runnerWorkers } from '../src/vite';
export default defineConfig({
  root: fileURLToPath(new URL('.', import.meta.url)),
  envPrefix: 'PUBLIC_',
  plugins: [svelte(), tailwindcss(), runnerWorkers(), compilerDir()],
  server: {
    // SharedArrayBuffer untuk input interaktif hanya tersedia saat crossOriginIsolated.
    headers: {
      'Cross-Origin-Opener-Policy': 'same-origin',
      'Cross-Origin-Embedder-Policy': 'credentialless',
    },
  },
  build: {
    outDir: 'dist',
    rolldownOptions: {
      output: {
        // Grup ikut menarik dependensinya, jadi grup node_modules diberi prioritas
        // lebih tinggi agar svelte dan vendor tidak tertelan chunk algohub-ui.
        codeSplitting: {
          groups: [
            { name: 'xterm', test: /[\\/]@xterm[\\/]/, priority: 4 },
            { name: 'monaco-loader', test: /[\\/]@monaco-editor[\\/]loader[\\/]/, priority: 4 },
            { name: 'svelte', test: /[\\/]node_modules[\\/]svelte[\\/]/, priority: 3 },
            { name: 'vendor', test: /[\\/]node_modules[\\/]/, priority: 2 },
            { name: 'algohub-ui', test: /[\\/]packages[\\/]ui[\\/]/, priority: 1 },
            { name: 'algohub-runner', test: /[\\/]packages[\\/]runner[\\/]src[\\/]/, priority: 1 },
          ],
        },
      },
    },
  },
});
