<script lang="ts">
  import { onDestroy, onMount } from 'svelte';
  import { Play, RefreshCw, Square } from '@lucide/svelte';
  import type { RunResponse } from '@algohub/api-client/praktikum';
  import { formatResult, messages, type Language } from '@algohub/runner';
  import { CodeEditor, Terminal } from '@algohub/ui';
  import { createMutation } from '@tanstack/svelte-query';
  import { api } from '#lib/api.ts';
  import { praktikumRunner } from '#lib/runner.ts';
  import { session } from '#lib/session.ts';

  let {
    value = $bindable(''),
    language = 'c',
    readonly = false,
    height = '350px',
    runnable = false,
    oninput,
  }: {
    value?: string;
    language?: string;
    readonly?: boolean;
    height?: string;
    runnable?: boolean;
    oninput?: () => void;
  } = $props();

  const runner = praktikumRunner();
  const idle = `\x1b[2m${messages.idle}\x1b[0m`;

  let terminal: ReturnType<typeof Terminal> | undefined = $state();
  let runLang = $state<Language>('c');
  let running = $state(false);
  let plotImages = $state<string[]>([]);
  // Reset/Stop/ganti bahasa membatalkan run yang sedang berjalan; hasilnya tidak boleh
  // ditulis lagi ke terminal yang sudah dibersihkan.
  let runToken = 0;

  const runMutation = createMutation(() => ({
    mutationFn: (body: { language: Language; source: string; stdin: string }) =>
      api.post<RunResponse>('/run', body),
  }));

  let statusText = $derived.by(() => {
    if (!runnable) return '';
    const status = runner.status[runLang];
    if (status !== 'idle' && status !== 'loading') return '';
    return runLang === 'python' ? messages.loadingPython : messages.loadingC;
  });

  // Prop `language` dari halaman induk wajib menggerakkan bahasa editor dan toolbar, tetapi
  // hanya saat prop berubah: halaman praktikum mengirim language="c" tetap, jadi tanpa penjaga
  // ini pilihan Python dari select langsung dipaksa balik ke C.
  let lastLangProp: string | undefined;
  $effect(() => {
    if (language === lastLangProp) return;
    lastLangProp = language;
    const want: Language = language === 'python' ? 'python' : 'c';
    if (want !== runLang) {
      runLang = want;
      onLangChange();
    }
  });

  onMount(() => {
    // Compiler diunduh begitu editor tampil supaya Run pertama tidak menunggu lama.
    if (runnable) void runner.ensure(runLang);
    terminal?.writeln(idle);
  });

  onDestroy(() => {
    if (running) cancelRun();
  });

  function onLangChange() {
    clearTerminal();
    if (runnable) void runner.ensure(runLang);
  }

  function cancelRun() {
    runToken++;
    runner.stop();
    running = false;
  }

  function clearTerminal() {
    if (running) cancelRun();
    terminal?.clear();
    plotImages = [];
    terminal?.writeln(idle);
  }

  function stopCode() {
    clearTerminal();
    terminal?.writeln(`\x1b[1;31m⚠ ${messages.stopped}\x1b[0m`);
  }

  function sendInput(line: string) {
    if (!runner.sendInput(line))
      terminal?.writeln(`\r\n\x1b[33m⚠ ${messages.noInteractive}\x1b[0m`);
  }

  async function runCode() {
    if (!runnable || running) return;
    const token = ++runToken;
    const lang = runLang;
    terminal?.clear();
    plotImages = [];
    running = true;
    terminal?.focus();
    try {
      const result = await runner.run(lang, value, {
        onOutput: (data) => terminal?.write(data),
      });
      if (token !== runToken) return;
      if (result.status === 'unavailable') {
        await runServerFallback(lang);
      } else {
        plotImages = result.images;
        terminal?.write(formatResult(result));
      }
    } catch (e) {
      // Run yang dibatalkan saat compiler masih dimuat tetap memegang runner sampai selesai.
      terminal?.writeln(`\x1b[33m⚠ ${(e as Error).message}\x1b[0m`);
    } finally {
      if (token === runToken) running = false;
    }
  }

  async function runServerFallback(lang: Language) {
    // /run butuh Bearer token; tanpa token klien API menerima 401 lalu mengalihkan
    // pengunjung publik /info/praktikum ke halaman login.
    if (!session.hasToken) {
      terminal?.writeln(
        '\x1b[33m⚠ Compiler C masih diunduh/diinisialisasi (±60MB). Tunggu sampai status "Menyiapkan compiler C" hilang, lalu klik Run lagi.\x1b[0m',
      );
      return;
    }
    try {
      const result = await runMutation.mutateAsync({ language: lang, source: value, stdin: '' });
      if (result.stdout) terminal?.write(result.stdout);
      if (result.stderr) terminal?.write(`\x1b[31m${result.stderr}\x1b[0m`);
      if (result.error) terminal?.write(`\r\n\x1b[1;31m❌ Error: ${result.error}\x1b[0m`);
      terminal?.writeln(`\r\n\x1b[1;32m✓ ${messages.done}\x1b[0m`);
    } catch (e) {
      terminal?.writeln(`\r\n\x1b[1;31m❌ Connection Error: ${(e as Error).message}\x1b[0m`);
    }
  }
</script>

<div class="flex flex-col border border-zinc-800 rounded-xl overflow-hidden shadow-sm mb-6">
  <div
    class="flex flex-wrap items-center justify-between gap-3 p-3 bg-zinc-900 border-b border-zinc-800 select-none"
  >
    <div class="flex items-center gap-2">
      {#if runnable}
        <select
          bind:value={runLang}
          onchange={onLangChange}
          class="h-8 rounded-lg bg-zinc-800 border border-zinc-700 text-zinc-100 px-2.5 py-1 text-xs font-bold font-mono outline-none cursor-pointer focus:border-zinc-500"
        >
          <option value="c">main.c</option>
          <option value="python">main.py</option>
        </select>

        {#if !running}
          <button
            class="h-8 bg-primary hover:bg-primary/95 text-white px-3 py-1 rounded-lg text-xs font-bold transition-all flex items-center gap-1.5 shadow-sm"
            onclick={runCode}
          >
            <Play size={12} /> Run
            <span class="text-[9px] text-white/50 font-normal">Ctrl+Enter</span>
          </button>
        {:else}
          <button
            class="h-8 bg-red-600 hover:bg-red-700 text-white px-3 py-1 rounded-lg text-xs font-bold transition-all flex items-center gap-1.5 shadow-sm"
            onclick={stopCode}
          >
            <Square size={12} /> Stop
          </button>
        {/if}

        <button
          class="h-8 bg-zinc-800 hover:bg-zinc-700 text-zinc-300 px-3 py-1 rounded-lg text-xs font-bold transition-all flex items-center gap-1.5 border border-zinc-700"
          onclick={clearTerminal}
        >
          <RefreshCw size={11} /> Reset Terminal
        </button>
      {/if}
    </div>

    <div class="text-[11px] font-mono text-zinc-400">
      {#if statusText}
        {statusText}
      {/if}
    </div>
  </div>

  <CodeEditor
    bind:value
    language={runLang}
    {readonly}
    {height}
    {oninput}
    onrun={runnable ? runCode : undefined}
  />
</div>

{#if runnable}
  <Terminal bind:this={terminal} class="flex-grow" inputEnabled={running} oninputline={sendInput} />

  {#if plotImages.length > 0}
    <div
      class="mt-3 rounded-xl border border-zinc-800 bg-[#141414] p-3 flex flex-col gap-3 shadow-lg"
    >
      {#each plotImages as src, i (i)}
        <figure class="flex flex-col gap-2">
          <div class="flex items-center justify-between">
            <span class="text-[10px] font-mono font-bold tracking-wider text-zinc-500">
              GRAFIK {plotImages.length > 1 ? `#${i + 1}` : ''}
            </span>
            <a
              href={src}
              download={`grafik_${i + 1}.png`}
              class="text-[10px] font-mono font-bold text-zinc-400 hover:text-zinc-100 transition-colors duration-200"
              >UNDUH PNG</a
            >
          </div>
          <img {src} alt={`Grafik matplotlib ${i + 1}`} class="max-w-full rounded-lg bg-white" />
        </figure>
      {/each}
    </div>
  {/if}
{/if}
