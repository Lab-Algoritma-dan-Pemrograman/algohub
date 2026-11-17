import { spawnSync } from 'node:child_process';
import { existsSync, mkdtempSync, readdirSync, readFileSync, rmSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { fileURLToPath } from 'node:url';
import { afterEach, describe, expect, it } from 'vitest';
import { FAKE_FILES } from './fixtures/fake-files';

const appDir = fileURLToPath(new URL('..', import.meta.url));
const script = fileURLToPath(new URL('../../../scripts/fetch-compiler.ts', import.meta.url));
const fixture = fileURLToPath(new URL('./fixtures/fake-release.ts', import.meta.url));
// Vitest berjalan di Node; `bun run` mengisi npm_execpath dengan path bun yang sedang dipakai.
const bun = process.env.npm_execpath || 'bun';

let dest = '';

function fetchCompiler(corrupt?: string) {
  dest = mkdtempSync(join(tmpdir(), 'portal-compiler-'));
  return spawnSync(bun, ['--preload', fixture, script, dest], {
    encoding: 'utf8',
    env: { ...process.env, FAKE_RELEASE_CORRUPT: corrupt ?? '' },
  });
}

afterEach(() => rmSync(dest, { recursive: true, force: true }));

describe('fetch-compiler', () => {
  it('menyimpan keempat file kalau SHA-256 cocok', () => {
    const result = fetchCompiler();
    expect(result.status, result.stderr).toBe(0);
    for (const [name, body] of Object.entries(FAKE_FILES)) {
      expect(readFileSync(join(dest, name), 'utf8')).toBe(body);
    }
  });

  it('gagal dan tidak menyisakan file kalau SHA-256 tidak cocok', () => {
    const result = fetchCompiler('lld');
    expect(result.status).toBe(1);
    expect(result.stderr).toContain('SHA-256 lld tidak cocok');
    expect(readdirSync(dest).filter((file) => file.startsWith('lld'))).toEqual([]);
  });

  it('build portal berhenti kalau fetch-compiler gagal', () => {
    const { scripts } = JSON.parse(readFileSync(join(appDir, 'package.json'), 'utf8'));
    expect(scripts['fetch-compiler']).toBe(
      'bun ../../scripts/fetch-compiler.ts static/compiler/v1',
    );
    expect(scripts.build).toBe('bun run fetch-compiler && vite build');
    expect(existsSync(join(appDir, '../../scripts/fetch-compiler.ts'))).toBe(true);
  });
});
