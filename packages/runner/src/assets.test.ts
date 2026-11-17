import { describe, expect, it, vi } from 'vitest';
import { createAssetLoader } from './assets';
import type { CompilerUrls } from './config';
const urls: CompilerUrls = {
  clang: '/compiler/v1/clang',
  lld: '/compiler/v1/lld',
  memfs: '/compiler/v1/memfs',
  sysroot: '/compiler/v1/sysroot.tar',
};
const setup = (status: (url: string, call: number) => number = () => 200) => {
  let call = 0;
  const fetchImpl = vi.fn(async (url: string, _init?: RequestInit) => {
    call += 1;
    const code = status(url, call);
    return new Response(code === 200 ? new Uint8Array([1, 2, 3]) : null, { status: code });
  });
  const compile = vi.fn(async () => ({}) as WebAssembly.Module);
  return { fetchImpl, compile, loader: createAssetLoader(fetchImpl, compile) };
};
describe('createAssetLoader', () => {
  it('load dua kali hanya mengunduh tiap file sekali', async () => {
    const { fetchImpl, compile, loader } = setup();
    const first = await loader.load(urls);
    const second = await loader.load(urls);
    expect(second).toEqual(first);
    expect(second.sysroot).toBe(first.sysroot);
    expect(fetchImpl).toHaveBeenCalledTimes(4);
    expect(compile).toHaveBeenCalledTimes(3);
    expect(first.sysroot.byteLength).toBe(3);
  });
  it('prefetch lalu load memakai unduhan yang sama', async () => {
    const { fetchImpl, loader } = setup();
    loader.prefetch(urls);
    await vi.waitFor(() => expect(fetchImpl).toHaveBeenCalledTimes(4));
    expect(fetchImpl.mock.calls.every(([, init]) => init?.priority === 'low')).toBe(true);
    await loader.load(urls);
    loader.prefetch(urls);
    await new Promise((resolve) => setTimeout(resolve, 0));
    expect(fetchImpl).toHaveBeenCalledTimes(4);
  });
  it('mengunduh ulang setelah load gagal', async () => {
    const { fetchImpl, loader } = setup((url, call) => (url === urls.lld && call <= 4 ? 500 : 200));
    await expect(loader.load(urls)).rejects.toThrow('HTTP 500');
    await loader.load(urls);
    const lldCalls = fetchImpl.mock.calls.filter(([url]) => url === urls.lld);
    expect(lldCalls).toHaveLength(2);
    expect(fetchImpl).toHaveBeenCalledTimes(5);
  });
  it('melempar error saat HTTP 404', async () => {
    const { loader } = setup((url) => (url === urls.sysroot ? 404 : 200));
    await expect(loader.load(urls)).rejects.toThrow(
      'Gagal mengunduh /compiler/v1/sysroot.tar: HTTP 404',
    );
  });
});
