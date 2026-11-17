import { assetLoader } from './assets';
import { prepareCSource } from './c-source';
import { resolveConfig, type CompilerUrls, type RunnerConfig } from './config';
import { PYODIDE_INDEX_URLS, RUN_TIMEOUT_MS } from './constants';
import { createSab } from './input-channel';
import { messages } from './messages';
import { Runtime, type RunOptions, type RunResult } from './runtime';
import type { RuntimeStatus } from './state';
export type Language = 'c' | 'python';
export class Runner {
  status = $state<Record<Language, RuntimeStatus>>({ c: 'idle', python: 'idle' });
  readonly #compiler: CompilerUrls;
  readonly #runtimes: Record<Language, Runtime>;
  constructor(config: RunnerConfig) {
    const resolved = resolveConfig(config);
    const cSab = createSab();
    const pythonSab = createSab();
    this.#compiler = resolved.compiler;
    this.#runtimes = {
      c: new Runtime({
        workerUrl: resolved.workers.c,
        sab: cSab,
        timeoutMs: RUN_TIMEOUT_MS,
        stopMode: 'terminate',
        unavailableMessage: messages.unavailableC,
        createInit: async () => ({
          message: { type: 'init', sab: cSab, assets: await assetLoader.load(resolved.compiler) },
        }),
        onStatus: (status) => (this.status.c = status),
      }),
      python: new Runtime({
        workerUrl: resolved.workers.python,
        sab: pythonSab,
        stopMode: 'cancel',
        unavailableMessage: messages.unavailablePython,
        createInit: async () => ({
          message: { type: 'init', sab: pythonSab, indexUrls: [...PYODIDE_INDEX_URLS] },
        }),
        onStatus: (status) => (this.status.python = status),
      }),
    };
  }
  get interactive(): boolean {
    return this.#runtimes.c.interactive;
  }
  prefetch(): void {
    assetLoader.prefetch(this.#compiler);
  }
  ensure(language: Language): Promise<boolean> {
    return this.#runtimes[language].ensure();
  }
  run(language: Language, code: string, options?: RunOptions): Promise<RunResult> {
    const source = language === 'c' ? prepareCSource(code) : code;
    return this.#runtimes[language].run(source, options);
  }
  sendInput(line: string): boolean {
    return this.#runtimes.c.sendInput(line) || this.#runtimes.python.sendInput(line);
  }
  stop(): void {
    this.#runtimes.c.stop();
    this.#runtimes.python.stop();
  }
}
let instance: { runner: Runner; key: string } | null = null;
// Dipanggil dari banyak komponen. Tanpa satu instance per halaman, setiap
// panggilan membuat set Worker baru sehingga compile dan run bisa mendarat di
// worker berbeda ("Please compile first") dan file compiler diunduh ulang.
export function getRunner(config: RunnerConfig): Runner {
  const key = JSON.stringify(resolveConfig(config));
  if (!instance) instance = { runner: new Runner(config), key };
  else if (instance.key !== key)
    throw new Error('getRunner sudah dipakai dengan konfigurasi lain.');
  return instance.runner;
}
