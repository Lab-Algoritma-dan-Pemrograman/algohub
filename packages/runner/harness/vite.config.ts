import { createReadStream } from 'node:fs';
import { join } from 'node:path';
import { fileURLToPath } from 'node:url';
import { svelte } from '@sveltejs/vite-plugin-svelte';
import tailwindcss from '@tailwindcss/vite';
import { defineConfig, type Plugin } from 'vite';
import { runnerWorkers } from '../src/vite';
const COMPILER_FILES = ['clang', 'lld', 'memfs', 'sysroot.tar'];
// Biner compiler tidak disimpan di repo; harness membacanya dari folder lokal.
function compilerDir(): Plugin {
  return {
    name: 'algohub-compiler-dir',
    apply: 'serve',
    configureServer(server) {
      server.middlewares.use('/compiler/v1/', (req, res, next) => {
        const file = req.url?.split('?')[0].replace(/^\//, '');
        if (!file || !COMPILER_FILES.includes(file)) return next();
        const dir = process.env.ALGOHUB_COMPILER_DIR;
        if (!dir) {
          res.statusCode = 500;
          res.end(
            'ALGOHUB_COMPILER_DIR belum diisi. Arahkan ke folder berisi clang, lld, memfs, dan sysroot.tar.',
          );
          return;
        }
        res.setHeader('Content-Type', 'application/octet-stream');
        res.setHeader('Cross-Origin-Resource-Policy', 'same-origin');
        createReadStream(join(dir, file)).on('error', next).pipe(res);
      });
    },
  };
}
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
