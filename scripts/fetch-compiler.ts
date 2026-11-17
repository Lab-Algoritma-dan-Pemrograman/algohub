import { mkdir, rename, rm } from 'node:fs/promises';
import { join } from 'node:path';
const RELEASE_URL =
  'https://github.com/Lab-Algoritma-dan-Pemrograman/algohub/releases/download/compiler-v1/';
const FILES = ['clang', 'lld', 'memfs', 'sysroot.tar'];
function fail(message: string): never {
  console.error(message);
  process.exit(1);
}
async function sha256(path: string): Promise<string | null> {
  const file = Bun.file(path);
  if (!(await file.exists())) return null;
  const hasher = new Bun.CryptoHasher('sha256');
  for await (const chunk of file.stream()) hasher.update(chunk);
  return hasher.digest('hex');
}
async function download(name: string): Promise<Response> {
  const res = await fetch(RELEASE_URL + name);
  if (!res.ok) fail(`Gagal mengunduh ${name}: HTTP ${res.status}`);
  return res;
}
function parseSums(text: string): Map<string, string> {
  const sums = new Map<string, string>();
  for (const line of text.split(/\r?\n/)) {
    if (!line.trim()) continue;
    const match = /^([0-9a-f]{64})\s+\*?(\S+)$/i.exec(line.trim());
    if (!match) fail(`Baris SHA256SUMS tidak dikenali: ${line}`);
    sums.set(match[2], match[1].toLowerCase());
  }
  const names = [...sums.keys()].sort();
  if (names.join() !== [...FILES].sort().join()) {
    fail(`SHA256SUMS harus berisi tepat ${FILES.join(', ')}; ditemukan ${names.join(', ')}`);
  }
  return sums;
}
const dest = process.argv[2];
if (!dest) fail('Pemakaian: bun scripts/fetch-compiler.ts <folder-tujuan>');
await mkdir(dest, { recursive: true });
const sums = parseSums(await (await download('SHA256SUMS')).text());
for (const name of FILES) {
  const expected = sums.get(name);
  const target = join(dest, name);
  if ((await sha256(target)) === expected) {
    process.stdout.write(`${name}: dilewati (hash cocok)\n`);
    continue;
  }
  const tmp = `${target}.tmp`;
  await Bun.write(tmp, await download(name));
  const actual = await sha256(tmp);
  if (actual !== expected) {
    await rm(tmp, { force: true });
    fail(`SHA-256 ${name} tidak cocok: diharapkan ${expected}, didapat ${actual}`);
  }
  await rename(tmp, target);
  process.stdout.write(`${name}: diunduh\n`);
}
