<script lang="ts">
  import { onDestroy, onMount } from 'svelte';
  import { Play, Square } from '@lucide/svelte';
  import { Playground } from '#lib/playground.svelte.ts';

  const playground = new Playground('python');

  let status = $derived(playground.status);
  let output = $derived(playground.output);
  let needInput = $derived(playground.needInput);
  let inputValue = $state('');
  let inputEl: HTMLInputElement | undefined = $state();

  let code = $state(`nama = input("Siapa namamu? ")
print("Halo,", nama, "!")
n = int(input("Masukkan sebuah angka: "))
print("Kuadratnya adalah", n * n)
`);

  $effect(() => {
    if (needInput) queueMicrotask(() => inputEl?.focus());
  });

  function run() {
    void playground.run(code);
  }

  function submitInput() {
    if (!needInput) return;
    playground.submitInput(inputValue);
    inputValue = '';
  }

  function stop() {
    playground.stop();
  }

  onMount(() => playground.start());
  onDestroy(() => playground.dispose());
</script>

<div class="mx-auto max-w-4xl">
  <h1 class="mb-1 text-2xl font-bold text-ink-heading">Playground Python (Live)</h1>
  <p class="mb-5 text-sm text-ink-caption">
    Jalankan kode Python langsung di browser — termasuk <code>input()</code> yang interaktif. (Eksperimental
    / Fase 1)
  </p>

  {#if status === 'unsupported'}
    <div class="rounded-lg bg-state-warning-bg p-4 text-sm text-state-warning">
      Browser ini belum mendukung mode interaktif (butuh <em>cross-origin isolation</em> / SharedArrayBuffer).
      Coba Chrome/Edge/Firefox versi terbaru, atau pastikan halaman dimuat lewat HTTPS dengan header COOP/COEP
      aktif.
    </div>
  {:else}
    <div class="grid gap-4">
      <div>
        <label class="label" for="code">Kode Python</label>
        <textarea
          id="code"
          bind:value={code}
          class="input min-h-48 font-mono text-sm"
          spellcheck="false"></textarea>
      </div>

      <div class="flex items-center gap-2">
        <button class="btn-primary" onclick={run} disabled={status !== 'ready'}>
          <Play size={16} />
          {status === 'loading' ? 'Memuat Python…' : status === 'running' ? 'Berjalan…' : 'Run'}
        </button>
        <button
          class="btn-outline"
          onclick={stop}
          disabled={status === 'loading' || status === 'init'}
        >
          <Square size={14} /> Stop / Reset
        </button>
        {#if status === 'loading'}
          <span class="text-xs text-ink-caption">Mengunduh runtime (~sekali, lalu ter-cache)…</span>
        {/if}
      </div>

      <div>
        <p class="mb-1 text-xs font-medium text-ink-caption">Output / Terminal:</p>
        <pre
          class="min-h-40 overflow-auto whitespace-pre-wrap rounded-lg bg-gray-900 p-3 text-sm text-gray-100">{output ||
            '(output akan muncul di sini)'}</pre>
        {#if needInput}
          <form
            class="mt-2 flex items-center gap-2"
            onsubmit={(e) => {
              e.preventDefault();
              submitInput();
            }}
          >
            <span class="font-mono text-sm text-ink-caption">&gt;</span>
            <input
              bind:this={inputEl}
              bind:value={inputValue}
              class="input flex-1 font-mono text-sm"
              placeholder="ketik input lalu Enter…"
              autocomplete="off"
            />
            <button class="btn-primary py-1.5" type="submit">Kirim</button>
          </form>
        {/if}
      </div>
    </div>
  {/if}
</div>
