<script lang="ts">
  import { messages } from '@algohub/runner';
  import { Terminal } from '@algohub/ui';
  import Loader from '@lucide/svelte/icons/loader-circle';
  import Play from '@lucide/svelte/icons/play';
  import RotateCcw from '@lucide/svelte/icons/rotate-ccw';
  import { onDestroy, onMount } from 'svelte';
  import { fade } from 'svelte/transition';
  import PlotDisplay from '#lib/components/PlotDisplay.svelte';
  import { TERMINAL_DEMO_SCRIPTS } from '#lib/terminal-demo-scripts.ts';
  import { TerminalRun } from '#lib/terminal-run.svelte.ts';

  const BANNER = [
    '\x1b[1;36mâ•”â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•—\x1b[0m',
    '\x1b[1;36mâ•‘\x1b[0m  \x1b[1mTerminal Interaktif Demo\x1b[0m                \x1b[1;36mâ•‘\x1b[0m',
    '\x1b[1;36mâ•‘\x1b[0m  Pilih script lalu klik \x1b[1;32mâ–¶ Jalankan\x1b[0m       \x1b[1;36mâ•‘\x1b[0m',
    '\x1b[1;36mâ•šâ•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•\x1b[0m',
    '',
  ];

  let terminal = $state<ReturnType<typeof Terminal>>();
  let selected = $state(0);

  const run = new TerminalRun(() => terminal);
  const script = $derived(TERMINAL_DEMO_SCRIPTS[selected]);
  const preparing = $derived(run.status('python') === 'idle' || run.status('python') === 'loading');
  const unavailable = $derived(run.status('python') === 'failed');

  function showBanner() {
    for (const line of BANNER) terminal?.writeln(line);
  }

  onMount(() => {
    showBanner();
    run.prepare('python');
  });
  onDestroy(() => run.cancel());

  function runScript() {
    const header = `\x1b[1;36mScript: ${script.name}\x1b[0m\r\n\x1b[2m${'â”€'.repeat(44)}\x1b[0m\r\n\r\n`;
    void run.run('python', script.code, header);
  }

  function reset() {
    run.reset();
    showBanner();
  }
</script>

<div class="flex h-[calc(100dvh-160px)] flex-col gap-6" in:fade={{ duration: 200 }}>
  <div class="flex flex-wrap items-center justify-between gap-4">
    <div>
      <h1 class="text-3xl font-bold tracking-tight">Terminal Demo</h1>
      <p class="text-zinc-500">
        Demo terminal interaktif dengan Python
        <code class="rounded bg-zinc-100 px-1.5 py-0.5 text-xs">input()</code> langsung di terminal
      </p>
    </div>
    <div class="flex items-center gap-2">
      <div class="flex items-center rounded-xl bg-zinc-100 p-1" role="group" aria-label="Script">
        {#each TERMINAL_DEMO_SCRIPTS as item, i (item.name)}
          <button
            type="button"
            onclick={() => (selected = i)}
            aria-pressed={selected === i}
            class="rounded-lg px-3 py-1.5 text-sm font-medium transition-all {selected === i
              ? 'bg-white text-zinc-900 shadow-sm'
              : 'text-zinc-500 hover:text-zinc-700'}"
          >
            {item.name}
          </button>
        {/each}
      </div>
      <button
        type="button"
        onclick={runScript}
        disabled={run.running || unavailable}
        class="flex items-center gap-2 rounded-xl bg-green-600 px-5 py-2.5 font-medium text-white transition-colors hover:bg-green-700 disabled:cursor-not-allowed disabled:opacity-50"
      >
        {#if run.running}
          <Loader class="h-4 w-4 animate-spin" />
          Berjalan...
        {:else}
          <Play class="h-4 w-4" />
          Jalankan
        {/if}
      </button>
      <button
        type="button"
        onclick={reset}
        title="Reset"
        aria-label="Reset"
        class="rounded-xl bg-zinc-100 p-2.5 text-zinc-600 transition-colors hover:bg-zinc-200"
      >
        <RotateCcw class="h-4 w-4" />
      </button>
    </div>
  </div>

  <Terminal
    bind:this={terminal}
    class="min-h-0 flex-1 shadow-2xl"
    inputEnabled={run.running}
    oninputline={(line) => run.sendInput(line)}
  />

  {#if run.images.length > 0}
    <div class="mt-2"><PlotDisplay images={run.images} /></div>
  {/if}

  <div class="flex items-center gap-4 text-xs text-zinc-400">
    {#if preparing}
      <span role="status">{messages.loadingPython}</span>
    {:else if unavailable}
      <span role="alert">{messages.unavailablePython}</span>
    {:else}
      <span>Klik <b>Jalankan</b> lalu ketik langsung di terminal saat diminta input</span>
      <span class="text-zinc-300">â€¢</span>
      <span>
        Tekan
        <kbd class="rounded bg-zinc-200 px-1.5 py-0.5 font-mono text-xs text-zinc-600">Enter</kbd>
        untuk mengirim input
      </span>
    {/if}
  </div>
</div>
