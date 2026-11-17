import { INIT_MAX_RETRIES, INIT_RETRY_DELAY_MS } from './constants';
import { InputChannel } from './input-channel';
import { InputQueue } from './input-queue';
import { messages } from './messages';
import { createRunTimer, type RunTimer } from './run-timer';
import { transition, type RuntimeEvent, type RuntimeStatus } from './state';
export type RunStatus = 'ok' | 'error' | 'compile-error' | 'timeout' | 'stopped' | 'unavailable';
export interface RunResult {
  status: RunStatus;
  output: string;
  error: string | null;
  images: string[];
  waitingForInput: boolean;
}
export interface RunOptions {
  stdin?: string;
  onOutput?: (data: string) => void;
}
export interface RuntimeOptions {
  workerUrl: string;
  sab: SharedArrayBuffer | null;
  createInit: () => Promise<{ message: object; transfer?: Transferable[] }>;
  timeoutMs?: number;
  stopMode: 'terminate' | 'cancel';
  unavailableMessage: string;
  onStatus: (status: RuntimeStatus) => void;
}
type WorkerMessage =
  | { type: 'ready' }
  | { type: 'init-error'; error: string }
  | { type: 'compile-error'; id: number; output: string; error: string }
  | { type: 'started'; id: number }
  | { type: 'write'; id: number; data: string }
  | { type: 'need-input'; id: number }
  | { type: 'exit'; id: number; error: string | null; waitingForInput: boolean; images?: string[] };
interface ActiveRun {
  id: number;
  output: string;
  stopping: boolean;
  onOutput?: (data: string) => void;
  resolve: (result: RunResult) => void;
}
const delay = (ms: number) => new Promise((resolve) => setTimeout(resolve, ms));
export class Runtime {
  readonly #options: RuntimeOptions;
  readonly #channel: InputChannel | null;
  readonly #queue: InputQueue | null;
  readonly #timer: RunTimer | null;
  #status: RuntimeStatus = 'idle';
  #worker: Worker | null = null;
  #ready: Promise<boolean> | null = null;
  #current: ActiveRun | null = null;
  #nextId = 0;
  constructor(options: RuntimeOptions) {
    this.#options = options;
    this.#channel = options.sab ? new InputChannel(options.sab) : null;
    this.#queue = this.#channel ? new InputQueue(this.#channel) : null;
    this.#timer = options.timeoutMs
      ? createRunTimer(options.timeoutMs, () => this.#timeout())
      : null;
  }
  get interactive(): boolean {
    return this.#channel !== null;
  }
  ensure(): Promise<boolean> {
    this.#ready ??= this.#start();
    return this.#ready;
  }
  async run(code: string, { stdin, onOutput }: RunOptions = {}): Promise<RunResult> {
    if (this.#current) throw new Error('Program lain masih berjalan.');
    if (!(await this.ensure()) || !this.#worker) {
      return { ...emptyResult('unavailable'), error: this.#options.unavailableMessage };
    }
    const worker = this.#worker;
    const interactive = stdin === undefined && this.interactive;
    this.#channel?.reset();
    this.#queue?.clear();
    const id = ++this.#nextId;
    return new Promise((resolve) => {
      this.#current = { id, output: '', stopping: false, onOutput, resolve };
      this.#emit({ type: 'start' });
      worker.postMessage({ type: 'run', id, code, stdin: interactive ? undefined : (stdin ?? '') });
    });
  }
  sendInput(line: string): boolean {
    if (!this.#current || !this.#queue) return false;
    const wasWaiting = this.#queue.waiting;
    if (!this.#queue.push(line)) return false;
    if (wasWaiting) this.#inputSent();
    return true;
  }
  stop(): void {
    const run = this.#current;
    if (!run) return;
    if (this.#options.stopMode === 'cancel') {
      run.stopping = true;
      this.#channel?.cancel();
      return;
    }
    // WASM yang berputar sinkron tidak bisa diinterupsi; satu-satunya cara
    // menghentikannya adalah terminate worker lalu membuat yang baru.
    this.#restart();
    run.resolve({ ...emptyResult('stopped'), output: run.output });
  }
  async #start(): Promise<boolean> {
    this.#emit({ type: 'init' });
    for (let attempt = 0; ; attempt++) {
      try {
        await this.#spawn();
        this.#emit({ type: 'ready' });
        return true;
      } catch (error) {
        const retry = attempt < INIT_MAX_RETRIES;
        this.#emit({ type: 'init-failed', retry });
        if (!retry) {
          console.error(error);
          this.#ready = null;
          return false;
        }
        await delay(INIT_RETRY_DELAY_MS);
      }
    }
  }
  #spawn(): Promise<void> {
    return new Promise((resolve, reject) => {
      const worker = new Worker(this.#options.workerUrl);
      const fail = (error: unknown) => {
        worker.terminate();
        reject(error);
      };
      worker.onerror = (event) => fail(new Error(event.message || 'Worker gagal dimuat'));
      worker.onmessage = (event: MessageEvent<WorkerMessage>) => {
        if (event.data.type === 'init-error') fail(new Error(event.data.error));
        if (event.data.type !== 'ready') return;
        this.#worker = worker;
        worker.onmessage = (next: MessageEvent<WorkerMessage>) => this.#handle(next.data);
        worker.onerror = (next) => this.#crash(next.message);
        resolve();
      };
      this.#options
        .createInit()
        .then(({ message, transfer = [] }) => worker.postMessage(message, transfer), fail);
    });
  }
  #handle(message: WorkerMessage): void {
    const run = this.#current;
    if (!run || !('id' in message) || message.id !== run.id) return;
    switch (message.type) {
      case 'started':
        this.#timer?.start();
        break;
      case 'write':
        run.output += message.data;
        run.onOutput?.(message.data);
        break;
      case 'need-input':
        this.#timer?.pause();
        this.#emit({ type: 'need-input' });
        this.#queue?.needInput();
        if (this.#queue && !this.#queue.waiting) this.#inputSent();
        break;
      case 'compile-error':
        this.#finish({
          ...emptyResult('compile-error'),
          output: message.output,
          error: message.error,
        });
        break;
      case 'exit':
        this.#finish({
          status: run.stopping ? 'stopped' : message.error ? 'error' : 'ok',
          output: run.output,
          error: message.error,
          images: message.images ?? [],
          waitingForInput: message.waitingForInput,
        });
        break;
    }
  }
  #inputSent(): void {
    this.#emit({ type: 'input' });
    this.#timer?.start();
  }
  #finish(result: RunResult): void {
    const run = this.#current;
    this.#timer?.clear();
    this.#queue?.clear();
    this.#current = null;
    this.#emit({ type: 'finish' });
    run?.resolve(result);
  }
  #timeout(): void {
    const run = this.#current;
    if (!run) return;
    this.#restart();
    run.resolve({ ...emptyResult('timeout'), output: run.output, error: messages.timeout });
  }
  #crash(error: string): void {
    const run = this.#current;
    this.#restart();
    run?.resolve({
      ...emptyResult('error'),
      output: run.output,
      error: error || 'Worker berhenti.',
    });
  }
  #restart(): void {
    this.#worker?.terminate();
    this.#worker = null;
    this.#ready = null;
    this.#current = null;
    this.#timer?.clear();
    this.#queue?.clear();
    this.#channel?.reset();
    this.#emit({ type: 'reset' });
    void this.ensure();
  }
  #emit(event: RuntimeEvent): void {
    const next = transition(this.#status, event);
    if (next === this.#status) return;
    this.#status = next;
    this.#options.onStatus(next);
  }
}
function emptyResult(status: RunStatus): RunResult {
  return { status, output: '', error: null, images: [], waitingForInput: false };
}
