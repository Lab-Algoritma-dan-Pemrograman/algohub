import type { Language, RunResult } from '@algohub/runner';
import { praktikumRunner } from './runner.ts';

export type PlaygroundStatus = 'init' | 'loading' | 'ready' | 'running' | 'unsupported';

export class Playground {
  output = $state('');
  running = $state(false);
  #unsupported = $state(false);
  readonly #language: Language;
  readonly #runner = praktikumRunner();

  constructor(language: Language) {
    this.#language = language;
  }

  get status(): PlaygroundStatus {
    if (this.#unsupported) return 'unsupported';
    if (this.running) return 'running';
    switch (this.#runner.status[this.#language]) {
      case 'idle':
        return 'init';
      case 'loading':
        return 'loading';
      case 'failed':
        return 'unsupported';
      default:
        return 'ready';
    }
  }

  // Runner tidak mengabarkan kapan program menunggu input, jadi form input tampil selama
  // program berjalan.
  get needInput(): boolean {
    return this.running;
  }

  start(): void {
    // SharedArrayBuffer butuh cross-origin isolation (header COOP/COEP).
    if (typeof SharedArrayBuffer === 'undefined' || !self.crossOriginIsolated) {
      this.#unsupported = true;
      return;
    }
    void this.#runner.ensure(this.#language);
  }

  async run(code: string): Promise<void> {
    if (this.status !== 'ready') return;
    this.output = '';
    this.running = true;
    try {
      this.#report(
        await this.#runner.run(this.#language, code, { onOutput: (data) => this.#append(data) }),
      );
    } finally {
      this.running = false;
    }
  }

  submitInput(line: string): void {
    if (!this.running) return;
    this.#append(line + '\n');
    if (!this.#runner.sendInput(line)) {
      this.#append('\n[Mode input interaktif tidak tersedia (SAB mati).]\n');
    }
  }

  stop(): void {
    this.#runner.stop();
  }

  dispose(): void {
    if (this.running) this.#runner.stop();
  }

  #append(text: string): void {
    this.output += text;
  }

  // Format keluaran mengikuti playground lama per bahasa (Python dulu memakai worker sendiri).
  #report(result: RunResult): void {
    if (this.#language === 'python') {
      if (result.error && result.status !== 'stopped') this.#append(`${result.error}\n`);
      return;
    }
    switch (result.status) {
      case 'compile-error':
        this.#append(result.output || result.error || 'Kompilasi gagal.');
        break;
      case 'timeout':
        this.#append(
          '\n[Program tidak merespons dalam 40 detik (kemungkinan infinite loop) — compiler di-reset.]\n',
        );
        break;
      case 'ok':
      case 'error':
        if (result.error) this.#append(`\n⚠ ${result.error}\n`);
        if (result.output.length > 0 && !result.output.endsWith('\n')) this.#append('\n');
        this.#append('\n✓ Program selesai!\n');
        break;
    }
  }
}
