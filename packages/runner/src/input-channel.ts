import { SAB_SIZE } from './constants';
const FLAG_WAIT = 0;
const FLAG_DATA = 1;
const FLAG_CANCEL = 2;
const DATA_OFFSET = 8;
const encoder = new TextEncoder();
export function createSab(): SharedArrayBuffer | null {
  if (typeof SharedArrayBuffer === 'undefined' || !globalThis.crossOriginIsolated) return null;
  return new SharedArrayBuffer(SAB_SIZE);
}
export class InputChannel {
  readonly #meta: Int32Array;
  readonly #data: Uint8Array;
  constructor(sab: SharedArrayBuffer) {
    this.#meta = new Int32Array(sab, 0, 2);
    this.#data = new Uint8Array(sab, DATA_OFFSET);
  }
  fits(line: string): boolean {
    return encoder.encode(`${line}\n`).length <= this.#data.length;
  }
  send(line: string): boolean {
    const bytes = encoder.encode(`${line}\n`);
    if (bytes.length > this.#data.length) return false;
    this.#data.set(bytes);
    Atomics.store(this.#meta, 1, bytes.length);
    Atomics.store(this.#meta, 0, FLAG_DATA);
    Atomics.notify(this.#meta, 0);
    return true;
  }
  cancel(): void {
    Atomics.store(this.#meta, 0, FLAG_CANCEL);
    Atomics.notify(this.#meta, 0);
  }
  reset(): void {
    Atomics.store(this.#meta, 0, FLAG_WAIT);
    Atomics.store(this.#meta, 1, 0);
  }
}
