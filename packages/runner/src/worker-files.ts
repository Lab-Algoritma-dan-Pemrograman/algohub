export const WORKER_FILES = ['clang-worker.js', 'pyodide-worker.js', 'vendor/shared.js'] as const;
export type WorkerFile = (typeof WORKER_FILES)[number];
const isWorkerFile = (name: string): name is WorkerFile =>
  (WORKER_FILES as readonly string[]).includes(name);
export function resolveWorkerFile(url: string, base: string): WorkerFile | null {
  const path = url.split(/[?#]/, 1)[0];
  if (!path.startsWith(base)) return null;
  const name = path.slice(base.length);
  return isWorkerFile(name) ? name : null;
}
