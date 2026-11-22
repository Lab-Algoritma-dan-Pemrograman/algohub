<script lang="ts">
  import { messages, type Language } from '@algohub/runner';
  import { Terminal, toast } from '@algohub/ui';
  import BookOpen from '@lucide/svelte/icons/book-open';
  import Code from '@lucide/svelte/icons/code';
  import Copy from '@lucide/svelte/icons/copy';
  import TerminalIcon from '@lucide/svelte/icons/terminal';
  import Trash from '@lucide/svelte/icons/trash-2';
  import X from '@lucide/svelte/icons/x';
  import { createQuery } from '@tanstack/svelte-query';
  import { onDestroy, onMount } from 'svelte';
  import { fade, scale } from 'svelte/transition';
  import LessonEditor from '#lib/components/LessonEditor.svelte';
  import PlotDisplay from '#lib/components/PlotDisplay.svelte';
  import { playgroundExamplesQuery } from '#lib/queries/playground.ts';
  import { TerminalRun } from '#lib/terminal-run.svelte.ts';

  const DEFAULT_CODE: Record<Language, string> = {
    python:
      '# Tulis kode Python Anda di sini\n# Gunakan input() untuk menerima input dari terminal\n\nnama = input("Siapa nama kamu? ")\nprint(f"Halo, {nama}! Selamat datang!")',
    c: '#include <stdio.h>\n\nint main() {\n    char nama[100];\n    printf("Siapa nama kamu? ");\n    scanf("%[^\\n]", nama);\n    printf("Halo, %s! Selamat datang!\\n", nama);\n    return 0;\n}',
  };
  const IDLE = '\x1b[2mJalankan kode Anda untuk melihat hasil...\x1b[0m';

  let language = $state<Language>('python');
  let code = $state(DEFAULT_CODE.python);
  let showExamples = $state(false);
  let terminal = $state<ReturnType<typeof Terminal>>();

  const run = new TerminalRun(() => terminal);
  const languageName = $derived(language === 'python' ? 'Python' : 'Bahasa C');
  const runnerStatus = $derived(run.status(language));
  const preparing = $derived(runnerStatus === 'idle' || runnerStatus === 'loading');

  const examples = createQuery(() => ({
    ...playgroundExamplesQuery(language),
    enabled: showExamples,
  }));

  onMount(() => {
    terminal?.writeln(IDLE);
    run.prepare(language);
  });
  onDestroy(() => run.cancel());

  function runCode() {
    const header = `\x1b[1;36m$ ${language === 'python' ? 'python' : 'clang'} run\x1b[0m\r\n\r\n`;
    void run.run(language, code, header);
  }

  function clearOutput() {
    run.reset();
    terminal?.writeln(IDLE);
  }

  function changeLanguage(next: Language) {
    if (next === language) return;
    language = next;
    code = DEFAULT_CODE[next];
    clearOutput();
    run.prepare(next);
  }

  async function copyCode() {
    try {
      await navigator.clipboard.writeText(code);
      toast.success('Kode disalin.');
    } catch {
      toast.error('Kode gagal disalin.');
    }
  }

  function useExample(exampleCode: string) {
    code = exampleCode;
    showExamples = false;
  }
</script>

<svelte:window onkeydown={(e) => showExamples && e.key === 'Escape' && (showExamples = false)} />

<div class="flex h-auto flex-col gap-6 pb-6 lg:h-[calc(100dvh-160px)]" in:fade={{ duration: 200 }}>
  <div class="flex flex-col justify-between gap-4 select-none sm:flex-row sm:items-center">
    <div>
      <h1 class="mb-1 flex items-center text-3xl font-black tracking-tight text-maroon">
        Playground Kode
        <TerminalIcon class="ml-3 animate-pulse text-fun-yellow" size={26} />
      </h1>
      <p class="text-sm font-bold text-zinc-500">
        Bereksperimen dengan kode di lingkungan sandbox.
      </p>
    </div>
    <div class="flex items-center gap-3">
      <div
        class="flex items-center rounded-2xl border border-zinc-200/60 bg-zinc-100 p-1.5 shadow-inner"
        role="group"
        aria-label="Bahasa"
      >
        {#each [['python', 'Python'], ['c', 'C']] as const as [id, label] (id)}
          <button
            type="button"
            onclick={() => changeLanguage(id)}
            aria-pressed={language === id}
            class="btn-bubbly cursor-pointer rounded-xl px-5 py-2 text-sm font-black transition-all {language ===
            id
              ? 'bg-rose-700 text-white shadow-[0_3px_0_#5C0E25] active:translate-y-[3px] active:shadow-none'
              : 'text-zinc-500 hover:text-zinc-900'}"
          >
            {label}
          </button>
        {/each}
      </div>
      <div class="h-8 w-px bg-zinc-200"></div>
      <button
        type="button"
        onclick={copyCode}
        title="Salin Kode"
        aria-label="Salin Kode"
        class="btn-bubbly cursor-pointer rounded-xl border-2 border-rose-200 bg-rose-50 p-3 text-rose-700 shadow-[0_3px_0_#8A1538] transition-all hover:bg-rose-100 active:translate-y-[3px] active:shadow-none"
      >
        <Copy size={16} />
      </button>
      <div class="h-8 w-px bg-zinc-200"></div>
      <button
        type="button"
        onclick={() => (showExamples = true)}
        class="btn-bubbly flex cursor-pointer items-center gap-2 rounded-xl bg-rose-700 px-4 py-2.5 text-sm font-bold text-white shadow-[0_3px_0_#5C0E25] transition-all hover:bg-rose-600 active:translate-y-[3px] active:shadow-none"
      >
        <BookOpen size={16} />
        Contoh Kode
      </button>
    </div>
  </div>

  <div class="grid h-auto min-h-0 flex-1 grid-cols-1 gap-6 lg:h-full lg:grid-cols-3">
    <div
      class="flex h-[450px] flex-col justify-between rounded-[2rem] border-2 border-gray-100 bg-white p-4 shadow-soft lg:col-span-2 lg:h-full"
    >
      <LessonEditor bind:value={code} {language} running={run.running} onrun={runCode} />
    </div>
    <div
      class="flex h-[300px] flex-col overflow-hidden rounded-[2rem] border-2 border-zinc-950 bg-zinc-900 shadow-[0_8px_0_#000] lg:h-full"
    >
      <div class="flex items-center gap-2 border-b-2 border-zinc-950/40 bg-zinc-800/80 px-4 py-3">
        <span class="flex-1 font-mono text-xs font-bold text-zinc-400">
          {language === 'python' ? 'python' : 'clang'} â€” playground
        </span>
        <button
          type="button"
          onclick={clearOutput}
          title="Bersihkan Output"
          class="flex cursor-pointer items-center gap-1.5 rounded-lg border border-zinc-600 bg-zinc-700/60 px-2.5 py-1 text-[11px] font-bold text-zinc-300 transition-colors hover:bg-zinc-600 hover:text-white"
        >
          <Trash size={10} />
          <span>Bersihkan</span>
        </button>
      </div>
      <Terminal
        bind:this={terminal}
        class="min-h-0 flex-1 rounded-none border-0"
        inputEnabled={run.running}
        oninputline={(line) => run.sendInput(line)}
      />
      {#if run.images.length > 0}
        <div
          class="custom-scrollbar max-h-[350px] overflow-y-auto border-t border-zinc-800 bg-zinc-950 p-3"
        >
          <PlotDisplay images={run.images} />
        </div>
      {/if}
    </div>
  </div>

  <div class="flex items-center gap-2 text-xs font-bold text-zinc-500 select-none">
    {#if preparing}
      <span role="status">{language === 'c' ? messages.loadingC : messages.loadingPython}</span>
    {:else if language === 'python'}
      <span>
        Ketik langsung di terminal saat program meminta
        <code class="rounded border border-zinc-200 bg-zinc-100 px-1.5 py-0.5 text-zinc-600">
          input()
        </code>
      </span>
    {:else}
      <span>
        Ketik langsung di terminal saat program meminta
        <code class="rounded border border-zinc-200 bg-zinc-100 px-1.5 py-0.5 text-zinc-600">
          scanf
        </code>
      </span>
    {/if}
  </div>
</div>

{#if showExamples}
  <div
    transition:fade={{ duration: 150 }}
    class="fixed inset-0 z-[100] flex items-center justify-center bg-zinc-950/60 p-4 backdrop-blur-sm"
  >
    <button
      type="button"
      onclick={() => (showExamples = false)}
      class="absolute inset-0 cursor-default"
      aria-label="Tutup contoh kode"
      tabindex="-1"
    ></button>
    <div
      in:scale={{ start: 0.95, duration: 200 }}
      role="dialog"
      aria-modal="true"
      aria-labelledby="examples-title"
      class="relative flex max-h-[80vh] w-full max-w-2xl flex-col overflow-hidden rounded-[2.5rem] border border-zinc-100 bg-white shadow-2xl"
    >
      <div class="flex items-center justify-between border-b border-zinc-100 bg-zinc-50/50 p-6">
        <div>
          <h3
            id="examples-title"
            class="flex items-center gap-2 text-xl font-black tracking-tight text-zinc-900"
          >
            <BookOpen class="text-rose-700" size={20} />
            Contoh Kode
          </h3>
          <p class="mt-0.5 text-xs font-semibold text-zinc-500">
            Pilih contoh kode {languageName} untuk dipelajari.
          </p>
        </div>
        <button
          type="button"
          onclick={() => (showExamples = false)}
          aria-label="Tutup"
          class="rounded-full border border-zinc-200 bg-white p-2 text-zinc-400 transition-colors hover:text-zinc-600"
        >
          <X size={18} />
        </button>
      </div>
      <div class="custom-scrollbar flex-1 space-y-3 overflow-y-auto p-6">
        {#if examples.isPending}
          <div class="flex flex-col items-center justify-center gap-3 py-16" role="status">
            <div
              class="h-10 w-10 animate-spin rounded-full border-4 border-rose-700 border-t-transparent"
            ></div>
            <p class="text-sm font-bold text-zinc-400">Memuat contoh kode...</p>
          </div>
        {:else if (examples.data ?? []).length === 0}
          <div class="flex flex-col items-center justify-center gap-3 py-16 text-center">
            <div class="flex h-16 w-16 items-center justify-center rounded-2xl bg-zinc-100">
              <Code class="text-zinc-300" size={26} />
            </div>
            <p class="text-sm font-bold text-zinc-400">
              {examples.isError
                ? 'Contoh kode gagal dimuat.'
                : `Belum ada contoh kode untuk ${languageName}.`}
            </p>
            <p class="text-xs text-zinc-300">Hubungi admin untuk menambahkan contoh.</p>
          </div>
        {:else}
          {#each examples.data ?? [] as example (example.id)}
            <div
              class="group overflow-hidden rounded-2xl border border-zinc-200 transition-all hover:border-rose-200 hover:shadow-lg hover:shadow-rose-500/5"
            >
              <div class="flex items-start justify-between gap-3 p-4">
                <div class="min-w-0 flex-1">
                  <div class="mb-1 flex items-center gap-2">
                    <span
                      class="rounded-full px-2 py-0.5 text-[10px] font-black tracking-wider uppercase {example.language ===
                      'python'
                        ? 'bg-rose-100 text-rose-700'
                        : 'bg-blue-100 text-blue-700'}"
                    >
                      {example.language === 'python' ? 'Python' : 'C'}
                    </span>
                    <h4 class="truncate text-sm font-bold text-zinc-900">{example.title}</h4>
                  </div>
                  {#if example.description}
                    <p class="mt-1 line-clamp-2 text-xs text-zinc-500">{example.description}</p>
                  {/if}
                </div>
                <button
                  type="button"
                  onclick={() => useExample(example.code)}
                  class="shrink-0 rounded-xl bg-rose-700 px-4 py-2 text-xs font-bold text-white shadow-lg shadow-rose-700/20 transition-all hover:bg-rose-600 active:scale-95"
                >
                  Gunakan
                </button>
              </div>
              <div
                class="relative max-h-24 overflow-hidden border-t border-zinc-200 bg-zinc-900 px-4 py-3"
              >
                <pre
                  class="font-mono text-[11px] leading-relaxed whitespace-pre text-emerald-400">{example.code
                    .split('\n')
                    .slice(0, 5)
                    .join('\n')}</pre>
                {#if example.code.split('\n').length > 5}
                  <div
                    class="absolute right-0 bottom-0 left-0 h-8 bg-gradient-to-t from-zinc-900 to-transparent"
                  ></div>
                {/if}
              </div>
            </div>
          {/each}
        {/if}
      </div>
    </div>
  </div>
{/if}
