import { createHash } from 'node:crypto';
import { FAKE_FILES } from './fake-files';

const corrupt = process.env.FAKE_RELEASE_CORRUPT;
const sums = Object.entries(FAKE_FILES)
  .map(([name, body]) => `${createHash('sha256').update(body).digest('hex')}  ${name}`)
  .join('\n');

// Dimuat lewat `bun --preload` agar scripts/fetch-compiler.ts diuji tanpa jaringan dan tanpa diubah.
globalThis.fetch = (async (input: RequestInfo | URL) => {
  const name = String(input).split('/').pop() ?? '';
  if (name === 'SHA256SUMS') return new Response(sums);
  const body = FAKE_FILES[name];
  if (body === undefined) return new Response(null, { status: 404 });
  return new Response(name === corrupt ? `${body} rusak` : body);
}) as typeof fetch;
