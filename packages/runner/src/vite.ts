import { readFile } from 'node:fs/promises';
import { fileURLToPath } from 'node:url';
import type { Plugin } from 'vite';
import { withTrailingSlash } from './config';
import { resolveWorkerFile, WORKER_FILES } from './worker-files';
const STATIC_DIR = fileURLToPath(new URL('../static/', import.meta.url));
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
