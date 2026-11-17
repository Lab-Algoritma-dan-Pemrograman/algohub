import { describe, expect, it } from 'vitest';
import { resolveConfig } from './config';
describe('resolveConfig', () => {
  it('membentuk URL compiler dari base absolut', () => {
    expect(
      resolveConfig({ compilerBaseUrl: 'https://algohub.web.id/compiler/v1/' }).compiler,
    ).toEqual({
      clang: 'https://algohub.web.id/compiler/v1/clang',
      lld: 'https://algohub.web.id/compiler/v1/lld',
      memfs: 'https://algohub.web.id/compiler/v1/memfs',
      sysroot: 'https://algohub.web.id/compiler/v1/sysroot.tar',
    });
  });
  it('menerima base root-relative dan menambah slash akhir', () => {
    const { compiler } = resolveConfig({ compilerBaseUrl: '/compiler/v1' });
    expect(compiler.clang).toBe('/compiler/v1/clang');
    expect(compiler.sysroot).toBe('/compiler/v1/sysroot.tar');
  });
  it('melempar error saat base kosong', () => {
    expect(() => resolveConfig({ compilerBaseUrl: '' })).toThrow(
      'PUBLIC_COMPILER_BASE_URL wajib diisi',
    );
    expect(() => resolveConfig({ compilerBaseUrl: '   ' })).toThrow('PUBLIC_COMPILER_BASE_URL');
  });
  it('memakai /runner/ sebagai base worker bawaan', () => {
    expect(resolveConfig({ compilerBaseUrl: '/compiler/v1/' }).workers).toEqual({
      c: '/runner/clang-worker.js?v=1',
      python: '/runner/pyodide-worker.js?v=1',
    });
  });
  it('memakai workerBaseUrl bila diisi', () => {
    const { workers } = resolveConfig({ compilerBaseUrl: '/c/', workerBaseUrl: '/w' });
    expect(workers.c).toBe('/w/clang-worker.js?v=1');
  });
});
