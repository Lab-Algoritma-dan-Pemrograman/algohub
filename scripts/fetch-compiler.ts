import { mkdir, readFile, rename } from 'node:fs/promises';
import { join } from 'node:path';
const RELEASE_URL =
  'https://github.com/Lab-Algoritma-dan-Pemrograman/algohub/releases/download/compiler-v1/';
const FILES = ['clang', 'lld', 'memfs', 'sysroot.tar'];
function fail(message: string): never {
  console.error(message);
  process.exit(1);
}
// Streaming Bun (Bun.write dari Response, Bun.file().stream()) kadang macet tanpa error di
// Windows (Bun 1.3.14). File terbesar ±30 MB, jadi isi file diproses utuh di memori.
function sha256(data: ArrayBuffer | Uint8Array): string {
  return new Bun.CryptoHasher('sha256').update(data).digest('hex');
}
async function existingSha256(path: string): Promise<string | null> {
  try {
    return sha256(await readFile(path));
  } catch {
    return null;
  }
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
  if ((await existingSha256(target)) === expected) {
    process.stdout.write(`${name}: dilewati (hash cocok)\n`);
    continue;
  }
  const body = await (await download(name)).arrayBuffer();
  const actual = sha256(body);
  if (actual !== expected) {
    fail(`SHA-256 ${name} tidak cocok: diharapkan ${expected}, didapat ${actual}`);
  }
  const tmp = `${target}.tmp`;
  await Bun.write(tmp, body);
  await rename(tmp, target);
  process.stdout.write(`${name}: diunduh\n`);
}
