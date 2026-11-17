import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { INIT_MAX_RETRIES, INIT_RETRY_DELAY_MS } from './constants';
import { messages } from './messages';
import { Runtime, type RuntimeOptions } from './runtime';
type Message = { type: string; [key: string]: unknown };
class FakeWorker {
  static instances: FakeWorker[] = [];
  static onInit: (worker: FakeWorker) => void = (worker) => worker.emit({ type: 'ready' });
  onmessage: ((event: { data: Message }) => void) | null = null;
  onerror: ((event: { message: string }) => void) | null = null;
  posted: Message[] = [];
  terminated = false;
  constructor(readonly url: string) {
    FakeWorker.instances.push(this);
  }
  postMessage(message: Message): void {
    this.posted.push(message);
    if (message.type === 'init') FakeWorker.onInit(this);
  }
  terminate(): void {
    this.terminated = true;
  }
  emit(data: Message): void {
    this.onmessage?.({ data });
  }
  lastRun(): Message & { id: number } {
    const run = this.posted.filter((m) => m.type === 'run').at(-1);
    if (!run) throw new Error('belum ada pesan run');
    return run as Message & { id: number };
  }
}
const flush = async () => {
  for (let i = 0; i < 20; i++) await Promise.resolve();
};
const worker = (index = -1) => FakeWorker.instances.at(index)!;
function createRuntime(options: Partial<RuntimeOptions> = {}) {
  const onStatus = vi.fn();
  const runtime = new Runtime({
    workerUrl: '/runner/test-worker.js',
    sab: new SharedArrayBuffer(4096),
    createInit: async () => ({ message: { type: 'init' } }),
    stopMode: 'terminate',
    unavailableMessage: 'tidak tersedia',
    onStatus,
    ...options,
  });
  return { runtime, onStatus };
}
describe('Runtime', () => {
  beforeEach(() => {
    vi.useFakeTimers();
    vi.stubGlobal('Worker', FakeWorker);
    FakeWorker.instances = [];
    FakeWorker.onInit = (w) => w.emit({ type: 'ready' });
  });
  afterEach(() => {
    vi.useRealTimers();
    vi.unstubAllGlobals();
    vi.restoreAllMocks();
  });
  it('menolak run kedua di tick yang sama', async () => {
    const { runtime } = createRuntime();
    const first = runtime.run('a');
    await expect(runtime.run('b')).rejects.toThrow('Program lain masih berjalan');
    await flush();
    const w = worker();
    expect(w.posted.filter((m) => m.type === 'run')).toHaveLength(1);
    w.emit({ type: 'exit', id: w.lastRun().id, error: null, waitingForInput: false });
    await expect(first).resolves.toMatchObject({ status: 'ok' });
    const next = runtime.run('c');
    await flush();
    w.emit({ type: 'exit', id: w.lastRun().id, error: null, waitingForInput: false });
    await expect(next).resolves.toMatchObject({ status: 'ok' });
  });
  it('mengumpulkan output dan meneruskan onOutput', async () => {
    const { runtime, onStatus } = createRuntime();
    const onOutput = vi.fn();
    const result = runtime.run('a', { onOutput });
    await flush();
    const w = worker();
    const { id } = w.lastRun();
    w.emit({ type: 'started', id });
    w.emit({ type: 'write', id, data: 'halo ' });
    w.emit({ type: 'write', id, data: 'dunia' });
    w.emit({ type: 'exit', id, error: null, waitingForInput: false, images: ['data:x'] });
    await expect(result).resolves.toEqual({
      status: 'ok',
      output: 'halo dunia',
      error: null,
      images: ['data:x'],
      waitingForInput: false,
    });
    expect(onOutput.mock.calls).toEqual([['halo '], ['dunia']]);
    expect(onStatus.mock.calls.map(([s]) => s)).toEqual(['loading', 'ready', 'busy', 'ready']);
  });
  it('mode batch mengirim stdin dan tidak interaktif', async () => {
    const { runtime } = createRuntime();
    void runtime.run('a', { stdin: '3\n4\n' });
    await flush();
    expect(worker().lastRun()).toMatchObject({ code: 'a', stdin: '3\n4\n' });
  });
  it('tanpa SAB run memakai stdin kosong', async () => {
    const { runtime } = createRuntime({ sab: null });
    expect(runtime.interactive).toBe(false);
    void runtime.run('a');
    await flush();
    expect(worker().lastRun()).toMatchObject({ stdin: '' });
  });
  it('mode interaktif mengirim input lewat SAB saat diminta', async () => {
    const sab = new SharedArrayBuffer(4096);
    const meta = new Int32Array(sab, 0, 2);
    const { runtime, onStatus } = createRuntime({ sab });
    void runtime.run('a');
    await flush();
    const w = worker();
    const { id, stdin } = w.lastRun();
    expect(stdin).toBeUndefined();
    expect(runtime.sendInput('3')).toBe(true);
    expect(Atomics.load(meta, 0)).toBe(0);
    w.emit({ type: 'need-input', id });
    expect(Atomics.load(meta, 0)).toBe(1);
    expect(new TextDecoder().decode(new Uint8Array(sab, 8, Atomics.load(meta, 1)))).toBe('3\n');
    expect(onStatus).toHaveBeenLastCalledWith('busy');
  });
  it('timeout dijeda saat menunggu input dan mulai ulang setelah input', async () => {
    const { runtime } = createRuntime({ timeoutMs: 1000 });
    const result = runtime.run('a');
    await flush();
    const first = worker();
    const { id } = first.lastRun();
    first.emit({ type: 'started', id });
    vi.advanceTimersByTime(900);
    first.emit({ type: 'need-input', id });
    vi.advanceTimersByTime(5000);
    expect(first.terminated).toBe(false);
    runtime.sendInput('1');
    vi.advanceTimersByTime(999);
    expect(first.terminated).toBe(false);
    vi.advanceTimersByTime(1);
    await expect(result).resolves.toMatchObject({ status: 'timeout', error: messages.timeout });
    expect(first.terminated).toBe(true);
    await flush();
    expect(FakeWorker.instances).toHaveLength(2);
  });
  it('stop mode terminate mengakhiri run dan membuat worker baru', async () => {
    const { runtime } = createRuntime();
    const result = runtime.run('a');
    await flush();
    const first = worker();
    first.emit({ type: 'write', id: first.lastRun().id, data: 'x' });
    runtime.stop();
    await expect(result).resolves.toMatchObject({ status: 'stopped', output: 'x' });
    expect(first.terminated).toBe(true);
    await flush();
    expect(FakeWorker.instances).toHaveLength(2);
    const next = runtime.run('b');
    await flush();
    const second = worker();
    second.emit({ type: 'exit', id: second.lastRun().id, error: null, waitingForInput: false });
    await expect(next).resolves.toMatchObject({ status: 'ok' });
  });
  it('stop mode cancel memberi flag batal dan menunggu exit dari worker', async () => {
    const sab = new SharedArrayBuffer(4096);
    const { runtime } = createRuntime({ sab, stopMode: 'cancel' });
    const result = runtime.run('a');
    await flush();
    const w = worker();
    runtime.stop();
    expect(Atomics.load(new Int32Array(sab, 0, 2), 0)).toBe(2);
    expect(w.terminated).toBe(false);
    w.emit({
      type: 'exit',
      id: w.lastRun().id,
      error: 'Program dibatalkan',
      waitingForInput: false,
    });
    await expect(result).resolves.toMatchObject({ status: 'stopped' });
  });
  it('worker crash setelah ready mengakhiri run dengan error dan membuat worker baru', async () => {
    const { runtime } = createRuntime();
    const result = runtime.run('a');
    await flush();
    const first = worker();
    first.onerror?.({ message: 'boom' });
    await expect(result).resolves.toMatchObject({ status: 'error', error: 'boom' });
    expect(first.terminated).toBe(true);
    await flush();
    expect(FakeWorker.instances).toHaveLength(2);
  });
  it('init dicoba ulang lalu berhasil', async () => {
    let attempts = 0;
    FakeWorker.onInit = (w) =>
      w.emit(
        ++attempts <= INIT_MAX_RETRIES ? { type: 'init-error', error: 'x' } : { type: 'ready' },
      );
    const { runtime } = createRuntime();
    const ready = runtime.ensure();
    for (let i = 0; i < INIT_MAX_RETRIES; i++) {
      await flush();
      await vi.advanceTimersByTimeAsync(INIT_RETRY_DELAY_MS);
    }
    await expect(ready).resolves.toBe(true);
    expect(FakeWorker.instances).toHaveLength(INIT_MAX_RETRIES + 1);
  });
  it('init gagal total menghasilkan unavailable dan bisa dicoba lagi', async () => {
    vi.spyOn(console, 'error').mockImplementation(() => {});
    FakeWorker.onInit = (w) => w.emit({ type: 'init-error', error: 'x' });
    const { runtime, onStatus } = createRuntime();
    const result = runtime.run('a');
    for (let i = 0; i < INIT_MAX_RETRIES; i++) {
      await flush();
      await vi.advanceTimersByTimeAsync(INIT_RETRY_DELAY_MS);
    }
    await expect(result).resolves.toMatchObject({ status: 'unavailable', error: 'tidak tersedia' });
    expect(onStatus).toHaveBeenLastCalledWith('failed');
    expect(FakeWorker.instances.every((w) => w.terminated)).toBe(true);
    FakeWorker.onInit = (w) => w.emit({ type: 'ready' });
    void runtime.run('b');
    await flush();
    expect(worker().lastRun()).toMatchObject({ code: 'b' });
  });
});
