<script lang="ts">
  import type { Language } from '@algohub/runner';
  import { CodeEditor } from '@algohub/ui';
  import Loader from '@lucide/svelte/icons/loader-circle';
  import Play from '@lucide/svelte/icons/play';

  let {
    value = $bindable(''),
    language,
    running = false,
    onrun,
    onreset,
    oninput,
  }: {
    value?: string;
    language: Language;
    running?: boolean;
    onrun: () => void;
    onreset?: () => void;
    oninput?: () => void;
  } = $props();
</script>

<div class="flex h-full flex-col rounded-2xl border border-zinc-800 bg-zinc-900 shadow-2xl">
  <div
    class="flex items-center justify-between rounded-t-2xl border-b border-zinc-800 bg-zinc-800/50 px-4 py-2"
  >
    <div class="flex items-center gap-2">
      <div class="h-3 w-3 rounded-full bg-red-500/80"></div>
      <div class="h-3 w-3 rounded-full bg-amber-500/80"></div>
      <div class="h-3 w-3 rounded-full bg-rose-700/80"></div>
      <span class="ml-2 font-mono text-xs text-zinc-400">
        {language === 'c' ? 'main.c' : 'main.py'}
      </span>
      <span
        class="rounded bg-zinc-800 px-2 py-0.5 text-[10px] font-bold tracking-wider text-zinc-500 uppercase"
      >
        {language === 'c' ? 'C' : 'Python'}
      </span>
    </div>
    <div class="flex items-center gap-2">
      {#if onreset}
        <button
          type="button"
          onclick={onreset}
          class="cursor-pointer rounded-lg border border-zinc-700 bg-zinc-800 px-3 py-1.5 text-xs font-bold text-zinc-300 transition-all hover:border-zinc-600 hover:bg-zinc-700 hover:text-white active:scale-95"
        >
          Reset
        </button>
      {/if}
      <button
        type="button"
        onclick={onrun}
        disabled={running}
        class="flex cursor-pointer items-center gap-2 rounded-lg bg-rose-700 px-4 py-1.5 text-sm font-bold text-white shadow-lg shadow-rose-700/20 transition-all hover:bg-rose-600 active:scale-95 disabled:opacity-50"
      >
        {#if running}
          <Loader class="h-4 w-4 animate-spin" />
        {:else}
          <Play class="h-4 w-4 fill-current" />
        {/if}
        Jalankan
        <span class="hidden text-[10px] font-normal text-white/50 sm:inline">Ctrl+Enter</span>
      </button>
    </div>
  </div>
  <div class="relative min-h-[200px] flex-1">
    <div class="absolute inset-0">
      <CodeEditor bind:value {language} height="100%" {onrun} {oninput} />
    </div>
  </div>
</div>
