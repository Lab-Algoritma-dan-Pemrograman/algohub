import { createReadStream } from 'node:fs';
import { readFile } from 'node:fs/promises';
import { join } from 'node:path';
import { fileURLToPath } from 'node:url';
import type { Connect, Plugin } from 'vite';
import { withTrailingSlash } from './config';
import { resolveWorkerFile, WORKER_FILES } from './worker-files';
const STATIC_DIR = fileURLToPath(new URL('../static/', import.meta.url));
const COMPILER_FILES = ['clang', 'lld', 'memfs', 'sysroot.tar'];
export interface RunnerWorkersOptions {
  base?: string;
}
export function runnerWorkers({ base = '/runner/' }: RunnerWorkersOptions = {}): Plugin {
  const prefix = withTrailingSlash(base);
  return {
    name: 'algohub-runner-workers',
    configureServer(server) {
      server.middlewares.use((req, res, next) => {
        const file = req.url ? resolveWorkerFile(req.url, prefix) : null;
        if (!file) return next();
        readFile(STATIC_DIR + file).then((source) => {
          res.setHeader('Content-Type', 'text/javascript');
          // Revalidasi 304 tidak membawa COEP sehingga browser memblokir worker
          // (ERR_BLOCKED_BY_RESPONSE); no-store memaksa selalu 200.
          res.setHeader('Cache-Control', 'no-store');
          res.setHeader('Cross-Origin-Embedder-Policy', 'credentialless');
          res.setHeader('Cross-Origin-Resource-Policy', 'same-origin');
          res.end(source);
        }, next);
      });
    },
    async generateBundle() {
      const outDir = prefix.replace(/^\//, '');
      for (const file of WORKER_FILES) {
        this.emitFile({
          type: 'asset',
          fileName: `${outDir}${file}`,
          source: await readFile(STATIC_DIR + file),
        });
      }
    },
  };
}
const serveCompilerFile: Connect.NextHandleFunction = (req, res, next) => {
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
};
// Biner compiler tidak disimpan di repo; server dev/preview membacanya dari folder lokal.
export function compilerDir(): Plugin {
  return {
    name: 'algohub-compiler-dir',
    apply: 'serve',
    configureServer(server) {
      server.middlewares.use('/compiler/v1/', serveCompilerFile);
    },
    configurePreviewServer(server) {
      server.middlewares.use('/compiler/v1/', serveCompilerFile);
    },
  };
}
