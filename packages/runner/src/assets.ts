import type { CompilerUrls } from './config';
export interface CompilerAssets {
  clang: WebAssembly.Module;
  lld: WebAssembly.Module;
  memfs: WebAssembly.Module;
  sysroot: ArrayBuffer;
}
export interface AssetLoader {
  prefetch(urls: CompilerUrls): void;
  load(urls: CompilerUrls): Promise<CompilerAssets>;
}
type FetchLike = (url: string, init?: RequestInit) => Promise<Response>;
type CompileLike = (bytes: ArrayBuffer) => Promise<WebAssembly.Module>;
const whenIdle = (callback: () => void) => {
  if (typeof globalThis.requestIdleCallback === 'function') requestIdleCallback(callback);
  else setTimeout(callback, 0);
};
function memo<T>(cache: Map<string, Promise<T>>, key: string, create: () => Promise<T>) {
  const existing = cache.get(key);
  if (existing) return existing;
  const promise = create();
  cache.set(key, promise);
  promise.catch(() => {
    if (cache.get(key) === promise) cache.delete(key);
  });
  return promise;
}
// File compiler diunduh sekali per halaman lalu Module-nya dikirim ke setiap
// worker baru, jadi worker yang dibuat ulang setelah timeout tidak mengunduh lagi.
export function createAssetLoader(
  fetchImpl: FetchLike = (url, init) => fetch(url, init),
  compile: CompileLike = (bytes) => WebAssembly.compile(bytes),
): AssetLoader {
  const downloads = new Map<string, Promise<ArrayBuffer>>();
  const modules = new Map<string, Promise<WebAssembly.Module>>();
  const download = (url: string, init?: RequestInit) =>
    memo(downloads, url, async () => {
      const response = await fetchImpl(url, init);
      if (!response.ok) throw new Error(`Gagal mengunduh ${url}: HTTP ${response.status}`);
      return response.arrayBuffer();
    });
  const compileFrom = (url: string) =>
    memo(modules, url, async () => {
      const bytes = await download(url);
      downloads.delete(url);
      return compile(bytes);
    });
  return {
    prefetch(urls) {
      whenIdle(() => {
        for (const url of [urls.clang, urls.lld, urls.memfs, urls.sysroot]) {
          if (!modules.has(url)) download(url, { priority: 'low' }).catch(() => {});
        }
      });
    },
    async load(urls) {
      const [clang, lld, memfs, sysroot] = await Promise.all([
        compileFrom(urls.clang),
        compileFrom(urls.lld),
        compileFrom(urls.memfs),
        download(urls.sysroot),
      ]);
      return { clang, lld, memfs, sysroot };
    },
  };
}
export const assetLoader = createAssetLoader();
