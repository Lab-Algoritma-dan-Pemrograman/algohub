import { formatResult, messages, type Language } from '@algohub/runner';
import { elearningRunner } from './runner.ts';

export interface TerminalHandle {
  write(text: string): void;
  writeln(text: string): void;
  clear(): void;
  focus(): void;
}

export const IDLE_TEXT = `\x1b[2m${messages.idle}\x1b[0m`;

// Menjalankan kode di Terminal @algohub/ui untuk /playground dan /terminal-demo. Input scanf dan
// input() diketik di terminal lalu dikirim lewat saluran input runner.
export class TerminalRun {
  running = $state(false);
  images = $state<string[]>([]);
  readonly #runner = elearningRunner();
  readonly #terminal: () => TerminalHandle | undefined;
  // Reset, Stop, atau ganti bahasa membatalkan run yang berjalan; hasilnya tidak boleh ditulis
  // lagi ke terminal yang sudah dibersihkan.
  #token = 0;

  constructor(terminal: () => TerminalHandle | undefined) {
    this.#terminal = terminal;
  }

  status(language: Language) {
    return this.#runner.status[language];
  }

  prepare(language: Language): void {
    void this.#runner.ensure(language);
  }

  async run(language: Language, code: string, header = ''): Promise<void> {
    if (this.running) return;
    const token = ++this.#token;
    const term = this.#terminal();
    term?.clear();
    if (header) term?.write(header);
    this.images = [];
    this.running = true;
    term?.focus();
    try {
      const result = await this.#runner.run(language, code, {
        onOutput: (data) => {
          if (token === this.#token) term?.write(data);
        },
      });
      if (token !== this.#token) return;
      this.images = result.images;
      term?.write(formatResult(result));
    } catch (error) {
      if (token === this.#token) term?.writeln(`\x1b[33m⚠ ${(error as Error).message}\x1b[0m`);
    } finally {
      if (token === this.#token) this.running = false;
    }
  }

  sendInput(line: string): void {
    if (!this.#runner.sendInput(line)) {
      this.#terminal()?.writeln(`\r\n\x1b[33m⚠ ${messages.noInteractive}\x1b[0m`);
    }
  }

  // Membatalkan run dan membersihkan terminal; pemanggil menulis teks awal sendiri.
  reset(): void {
    this.cancel();
    this.#terminal()?.clear();
    this.images = [];
  }

  cancel(): void {
    if (!this.running) return;
    this.#token++;
    this.#runner.stop();
    this.running = false;
  }
}
