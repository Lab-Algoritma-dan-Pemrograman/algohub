<script lang="ts">
  import { onDestroy, onMount } from 'svelte';
  import type { HTMLAttributes } from 'svelte/elements';
  import type { Terminal as XTerm } from '@xterm/xterm';
  import { cn } from '../../utils.js';
  import { createLineEditor } from './line-editor.js';

  let {
    inputEnabled = false,
    oninputline,
    class: className,
    ...restProps
  }: HTMLAttributes<HTMLDivElement> & {
    inputEnabled?: boolean;
    oninputline?: (line: string) => void;
  } = $props();

  let termEl: HTMLDivElement;
  let term: XTerm | null = null;
  let resizeObserver: ResizeObserver | null = null;
  let pending = '';
  const lineEditor = createLineEditor();

  // xterm menggambar ke canvas yang tidak terbaca assistive tech, jadi output juga
  // ditulis ke live-region agar bisa dibaca dan diverifikasi.
  // Dibatasi ekornya saja agar program yang mencetak tanpa henti tidak membuat
  // teks ini tumbuh tanpa batas (xterm sendiri hanya menyimpan 5000 baris).
  const MIRROR_LIMIT = 100_000;
  let outputText = $state('');
  let destroyed = false;

  function stripAnsi(str: string): string {
    // eslint-disable-next-line no-control-regex
    return str.replace(/\x1b\[[0-9;?]*[A-Za-z]/g, '');
  }

  export function write(str: string) {
    if (term) term.write(str);
    else pending += str;
    outputText = (outputText + stripAnsi(str)).slice(-MIRROR_LIMIT);
  }

  export function writeln(str: string) {
    write(str + '\n');
  }

  export function clear() {
    term?.clear();
    pending = '';
    outputText = '';
  }

  export function focus() {
    term?.focus();
  }

  $effect(() => {
    if (!inputEnabled) lineEditor.reset();
  });

  onMount(async () => {
    try {
      const [{ Terminal }, { FitAddon }] = await Promise.all([
        import('@xterm/xterm'),
        import('@xterm/addon-fit'),
        import('@xterm/xterm/css/xterm.css'),
      ]);
      if (destroyed) return;
      const instance = new Terminal({
        theme: {
          background: '#18181b',
          foreground: '#e4e4e7',
          cursor: '#a1a1aa',
          cursorAccent: '#18181b',
          selectionBackground: '#3f3f4680',
          green: '#4ade80',
          red: '#f87171',
          yellow: '#facc15',
          blue: '#60a5fa',
          magenta: '#c084fc',
          cyan: '#22d3ee',
        },
        fontSize: 13,
        fontFamily: '"JetBrains Mono", "Fira Code", monospace',
        cursorBlink: true,
        cursorStyle: 'bar',
        convertEol: true,
        scrollback: 5000,
      });
      const fitAddon = new FitAddon();
      instance.loadAddon(fitAddon);
      instance.open(termEl);
      fitAddon.fit();
      term = instance;
      if (pending) {
        instance.write(pending);
        pending = '';
      }

      instance.onKey(({ key, domEvent }) => {
        if (!inputEnabled) return;
        const { echo, line } = lineEditor.handle(key, domEvent);
        if (echo) write(echo);
        if (line !== undefined) oninputline?.(line);
      });

      resizeObserver = new ResizeObserver(() => fitAddon.fit());
      resizeObserver.observe(termEl);
    } catch (e) {
      console.error('Terminal gagal diinisialisasi:', e);
    }
  });

  onDestroy(() => {
    destroyed = true;
    resizeObserver?.disconnect();
    term?.dispose();
  });
</script>

<div
  class={cn(
    'flex min-h-0 flex-col overflow-hidden rounded-xl border border-zinc-800 bg-[#18181b] shadow-lg',
    className,
  )}
  {...restProps}
>
  <div
    class="flex flex-shrink-0 items-center justify-between border-b border-zinc-800 bg-[#141414] px-4 py-2 select-none"
  >
    <div class="flex items-center gap-1.5">
      <span class="h-2.5 w-2.5 rounded-full border border-[#e0443e] bg-[#ff5f56]"></span>
      <span class="h-2.5 w-2.5 rounded-full border border-[#dfa123] bg-[#ffbd2e]"></span>
      <span class="h-2.5 w-2.5 rounded-full border border-[#1aab29] bg-[#27c93f]"></span>
    </div>
    <span class="font-mono text-[10px] font-bold tracking-wider text-zinc-500">TERMINAL</span>
    <div class="w-12"></div>
  </div>
  <div class="min-h-[220px] flex-grow bg-[#18181b] p-3">
    <div bind:this={termEl} class="h-full min-h-[200px] w-full"></div>
    <div class="sr-only whitespace-pre-wrap" aria-live="polite" data-terminal-output>
      {outputText}
    </div>
  </div>
</div>
