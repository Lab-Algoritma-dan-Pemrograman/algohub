import { WORKER_VERSION } from './constants';
export interface RunnerConfig {
  compilerBaseUrl: string;
  workerBaseUrl?: string;
}
export interface CompilerUrls {
  clang: string;
  lld: string;
  memfs: string;
  sysroot: string;
}
export interface ResolvedConfig {
  compiler: CompilerUrls;
  workers: { c: string; python: string };
}
export const withTrailingSlash = (url: string) => (url.endsWith('/') ? url : `${url}/`);
export function resolveConfig(config: RunnerConfig): ResolvedConfig {
  const compilerBase = (config.compilerBaseUrl ?? '').trim();
  if (!compilerBase) throw new Error('PUBLIC_COMPILER_BASE_URL wajib diisi');
  const compiler = withTrailingSlash(compilerBase);
  const workers = withTrailingSlash(config.workerBaseUrl?.trim() || '/runner/');
  const version = `?v=${WORKER_VERSION}`;
  return {
    compiler: {
      clang: `${compiler}clang`,
      lld: `${compiler}lld`,
      memfs: `${compiler}memfs`,
      sysroot: `${compiler}sysroot.tar`,
    },
    workers: {
      c: `${workers}clang-worker.js${version}`,
      python: `${workers}pyodide-worker.js${version}`,
    },
  };
}
