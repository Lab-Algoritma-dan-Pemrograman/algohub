// @vitest-environment node
import { readdirSync, readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { describe, expect, it } from 'vitest';

const srcDir = fileURLToPath(new URL('../src/', import.meta.url));
const files = readdirSync(srcDir, { recursive: true, encoding: 'utf8' })
  .map((file) => file.replaceAll('\\', '/'))
  .filter((file) => /\.(ts|svelte)$/.test(file))
  .map((file) => ({ file, code: readFileSync(`${srcDir}${file}`, 'utf8') }));

// 03-frontend.md § Realtime: supabase-js hanya untuk .channel(); data lewat backend Go.
const forbidden = [
  /supabase\s*\.\s*(from|rpc|auth|storage)\b/,
  /\.rpc\(/,
  /\.from\(\s*['"]/,
  /\.storage\b/,
];

describe('tidak ada akses Supabase langsung', () => {
  it('sumber dipindai', () => {
    expect(files.length).toBeGreaterThan(10);
  });

  it.each(forbidden.map((re) => [re.source, re] as const))('tidak ada %s', (_name, re) => {
    expect(files.filter(({ code }) => re.test(code)).map(({ file }) => file)).toEqual([]);
  });

  it('createClient hanya di realtime/client.ts', () => {
    expect(
      files.filter(({ code }) => /\bcreateClient\b/.test(code)).map(({ file }) => file),
    ).toEqual(['lib/realtime/client.ts']);
  });
});
