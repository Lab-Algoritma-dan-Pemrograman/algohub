<script lang="ts">
  import Play from '@lucide/svelte/icons/play';
  import RotateCcw from '@lucide/svelte/icons/rotate-ccw';
  import Sparkles from '@lucide/svelte/icons/sparkles';
  import Square from '@lucide/svelte/icons/square';
  import { formatResult, messages, type Language } from '@algohub/runner';
  import { CodeEditor, Terminal } from '@algohub/ui';
  import { onDestroy, onMount } from 'svelte';
  import { DEFAULT_CODE, DEMO_SCRIPTS } from '#lib/demo-scripts.ts';
  import { elearningRunner } from '#lib/runner.ts';

  const runner = elearningRunner();
  const idle = `\x1b[2m${messages.idle}\x1b[0m`;

  let terminal: ReturnType<typeof Terminal> | undefined = $state();
  let language = $state<Language>('python');
  let code = $state(DEFAULT_CODE.python);
  let activeTemplate = $state<number | null>(null);
  let running = $state(false);
  let plotImages = $state<string[]>([]);
  // Reset, Stop, atau ganti bahasa membatalkan run yang berjalan; hasilnya tidak boleh ditulis
  // ke terminal yang sudah dibersihkan.
  let runToken = 0;

  const loading = $derived(
    runner.status[language] === 'idle' || runner.status[language] === 'loading',
  );

  onMount(() => {
    // Compiler diunduh begitu demo tampil supaya Jalankan pertama tidak menunggu lama.
    void runner.ensure(language);
    terminal?.writeln(idle);
  });

  onDestroy(() => {
    if (running) cancelRun();
  });

  function cancelRun() {
    runToken++;
    runner.stop();
    running = false;
  }

  function reset() {
    if (running) cancelRun();
    terminal?.clear();
    plotImages = [];
    terminal?.writeln(idle);
  }

  function selectLanguage(next: Language) {
    if (next === language) return;
    language = next;
    code = DEFAULT_CODE[next];
    activeTemplate = null;
    reset();
    void runner.ensure(next);
  }

  function selectTemplate(index: number) {
    activeTemplate = index;
    code = DEMO_SCRIPTS[language][index].code;
    reset();
    terminal?.writeln(`\x1b[1;36mTemplate: ${DEMO_SCRIPTS[language][index].name}\x1b[0m`);
  }

  function stop() {
    reset();
    terminal?.writeln(`\x1b[1;31m⚠ ${messages.stopped}\x1b[0m`);
  }

  function sendInput(line: string) {
    if (!runner.sendInput(line)) {
      terminal?.writeln(`\r\n\x1b[33m⚠ ${messages.noInteractive}\x1b[0m`);
    }
  }

  async function run() {
    if (running) return;
    const token = ++runToken;
    terminal?.clear();
    plotImages = [];
    running = true;
    terminal?.focus();
    try {
      const result = await runner.run(language, code, {
        onOutput: (data) => terminal?.write(data),
      });
      if (token !== runToken) return;
      plotImages = result.images;
      terminal?.write(formatResult(result));
    } catch (e) {
      terminal?.writeln(`\x1b[33m⚠ ${(e as Error).message}\x1b[0m`);
    } finally {
      if (token === runToken) running = false;
    }
  }
</script>

<div class="relative z-10 flex w-full flex-col gap-5">
  <div
    class="shadow-soft border-maroon/5 flex flex-wrap items-center justify-between gap-3 rounded-[2rem] border bg-white/95 p-4 backdrop-blur-md"
  >
    <div
      class="flex items-center rounded-2xl border border-zinc-200/50 bg-zinc-100 p-1 shadow-inner"
    >
      {#each [['python', 'Python'], ['c', 'Bahasa C']] as const as [value, label] (value)}
        <button
          onclick={() => selectLanguage(value)}
          class="cursor-pointer rounded-xl px-5 py-2 text-xs font-black transition-all {language ===
          value
            ? 'bg-maroon shadow-bubbly-maroon text-white'
            : 'text-zinc-500 hover:text-zinc-900'}"
        >
          {label}
        </button>
      {/each}
    </div>
    <div class="flex flex-wrap items-center gap-1.5">
      <span
        class="text-maroon/60 mr-1.5 flex items-center gap-1 text-[11px] font-black tracking-wider uppercase"
      >
        <Sparkles size={12} class="text-fun-yellow fill-fun-yellow" />
        Template:
      </span>
      {#each DEMO_SCRIPTS[language] as script, index (script.name)}
        <button
          onclick={() => selectTemplate(index)}
          class="cursor-pointer rounded-xl border px-3.5 py-2 text-xs font-bold transition-all hover:scale-105 active:scale-95 {activeTemplate ===
          index
            ? 'bg-maroon-bg border-maroon/30 text-maroon font-black'
            : 'border-zinc-200 bg-white text-zinc-600 hover:border-zinc-300 hover:bg-zinc-50'}"
        >
          {script.name}
        </button>
      {/each}
    </div>
  </div>

  <div
    class="relative grid grid-cols-1 gap-6 overflow-hidden rounded-[3rem] border border-white/10 bg-zinc-950 p-6 shadow-[0_50px_100px_-20px_rgba(138,21,56,0.35)] lg:grid-cols-3"
  >
    <div
      class="relative z-10 overflow-hidden rounded-2xl border border-white/5 shadow-2xl lg:col-span-2"
    >
      <div class="flex items-center gap-2 border-b border-white/5 bg-zinc-900/60 px-4 py-3">
        {#if running}
          <button
            onclick={stop}
            class="flex h-8 cursor-pointer items-center gap-1.5 rounded-lg bg-red-600 px-3 text-xs font-bold text-white hover:bg-red-700"
          >
            <Square size={12} /> Stop
          </button>
        {:else}
          <button
            onclick={run}
            class="bg-maroon hover:bg-maroon-light flex h-8 cursor-pointer items-center gap-1.5 rounded-lg px-3 text-xs font-bold text-white"
          >
            <Play size={12} /> Jalankan
            <span class="text-[9px] font-normal text-white/50">Ctrl+Enter</span>
          </button>
        {/if}
        <button
          onclick={reset}
          class="flex h-8 cursor-pointer items-center gap-1.5 rounded-lg border border-zinc-700 bg-zinc-800 px-3 text-xs font-bold text-zinc-300 hover:bg-zinc-700"
        >
          <RotateCcw size={11} /> Reset
        </button>
      </div>
      <CodeEditor bind:value={code} {language} height="320px" onrun={run} />
    </div>

    <div
      class="relative z-10 flex min-h-[320px] flex-col overflow-hidden rounded-2xl border border-white/5 bg-[#09090b] shadow-2xl"
    >
      <div class="flex items-center gap-2 border-b border-white/5 bg-zinc-900/60 px-4 py-3">
        <div
          class="h-2.5 w-2.5 rounded-full {running
            ? 'animate-ping bg-amber-500'
            : 'bg-emerald-500'}"
        ></div>
        <span class="font-mono text-[10px] font-bold tracking-wider text-zinc-400 uppercase"
          >Terminal Output</span
        >
      </div>
      <Terminal
        bind:this={terminal}
        class="flex-1"
        inputEnabled={running}
        oninputline={sendInput}
      />
      {#if plotImages.length > 0}
        <div class="max-h-[300px] overflow-y-auto border-t border-white/10 bg-zinc-950 p-3">
          {#each plotImages as src, i (i)}
            <img
              {src}
              alt={`Grafik matplotlib ${i + 1}`}
              class="mb-2 max-w-full rounded-lg bg-white"
            />
          {/each}
        </div>
      {/if}
    </div>

    <div
      class="relative z-10 flex items-center justify-between px-2 text-[10px] font-bold text-zinc-500 lg:col-span-3"
    >
      <span class="flex items-center gap-1.5">
        {#if loading}
          <span class="inline-block h-2.5 w-2.5 animate-ping rounded-full bg-amber-500"></span>
          {language === 'python' ? messages.loadingPython : messages.loadingC}
        {:else}
          <span class="inline-block h-2 w-2 rounded-full bg-emerald-500"></span>
          Interpreter WebAssembly siap
        {/if}
      </span>
      <span>Ketik langsung di terminal saat input() / scanf() dipanggil</span>
    </div>
  </div>
</div>
