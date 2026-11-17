import { describe, expect, it } from 'vitest';
import { resolveWorkerFile } from './worker-files';
describe('resolveWorkerFile', () => {
  it('mengembalikan nama file dari allowlist tanpa query', () => {
    expect(resolveWorkerFile('/runner/clang-worker.js?v=1', '/runner/')).toBe('clang-worker.js');
    expect(resolveWorkerFile('/runner/pyodide-worker.js', '/runner/')).toBe('pyodide-worker.js');
    expect(resolveWorkerFile('/runner/vendor/shared.js?v=1#x', '/runner/')).toBe(
      'vendor/shared.js',
    );
  });
  it('menolak file di luar allowlist, base lain, dan ..', () => {
    expect(resolveWorkerFile('/runner/index.ts', '/runner/')).toBeNull();
    expect(resolveWorkerFile('/other/clang-worker.js', '/runner/')).toBeNull();
    expect(resolveWorkerFile('/runner/../package.json', '/runner/')).toBeNull();
    expect(resolveWorkerFile('/runner/vendor/../clang-worker.js', '/runner/')).toBeNull();
  });
});
