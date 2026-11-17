<script lang="ts">
  import AlertTriangle from '@lucide/svelte/icons/triangle-alert';
  import ChevronDown from '@lucide/svelte/icons/chevron-down';
  import ChevronUp from '@lucide/svelte/icons/chevron-up';
  import Home from '@lucide/svelte/icons/house';
  import RefreshCw from '@lucide/svelte/icons/refresh-cw';

  let { error }: { error: unknown } = $props();
  let showDetails = $state(false);
</script>

<div
  class="flex min-h-screen w-full items-center justify-center bg-slate-950 p-6 font-sans text-slate-100"
>
  <div
    class="pointer-events-none absolute top-1/4 left-1/4 h-96 w-96 rounded-full bg-blue-600/10 blur-[100px]"
  ></div>
  <div
    class="pointer-events-none absolute right-1/4 bottom-1/4 h-96 w-96 rounded-full bg-indigo-600/10 blur-[100px]"
  ></div>
  <div
    class="relative z-10 w-full max-w-xl overflow-hidden rounded-2xl border border-slate-800 bg-slate-900/60 p-8 text-center shadow-2xl backdrop-blur-xl"
  >
    <div
      class="mx-auto mb-6 flex h-16 w-16 animate-pulse items-center justify-center rounded-full border border-amber-500/20 bg-amber-500/10"
    >
      <AlertTriangle class="h-8 w-8 text-amber-500" />
    </div>
    <h1 class="mb-2 text-2xl font-bold tracking-tight text-white">Ups! Terjadi Kesalahan Sistem</h1>
    <p class="mx-auto mb-8 max-w-md text-sm text-slate-400">
      Aplikasi mengalami gangguan rendering yang tidak terduga. Jangan khawatir, data Anda tetap
      aman.
    </p>
    <div class="mb-8 flex flex-col items-center justify-center gap-3 sm:flex-row">
      <button
        onclick={() => window.location.reload()}
        class="flex w-full items-center justify-center gap-2 rounded-lg bg-blue-600 px-5 py-2.5 text-sm font-medium text-white shadow-lg shadow-blue-600/20 transition-all duration-200 hover:bg-blue-500 active:scale-95 sm:w-auto"
      >
        <RefreshCw class="h-4 w-4" />
        Segarkan Halaman
      </button>
      <button
        onclick={() => (window.location.href = '/')}
        class="flex w-full items-center justify-center gap-2 rounded-lg border border-slate-700 bg-slate-800 px-5 py-2.5 text-sm font-medium text-slate-300 transition-all duration-200 hover:bg-slate-700 active:scale-95 sm:w-auto"
      >
        <Home class="h-4 w-4" />
        Kembali ke Beranda
      </button>
    </div>
    {#if error}
      <div class="border-t border-slate-800/80 pt-6 text-left">
        <button
          onclick={() => (showDetails = !showDetails)}
          class="flex items-center gap-2 text-xs text-slate-500 transition-colors duration-150 hover:text-slate-300 focus:outline-none"
        >
          {#if showDetails}
            <ChevronUp class="h-3.5 w-3.5" />
          {:else}
            <ChevronDown class="h-3.5 w-3.5" />
          {/if}
          {showDetails ? 'Sembunyikan detail teknis' : 'Tampilkan detail teknis'}
        </button>
        {#if showDetails}
          <div
            class="mt-4 max-h-48 overflow-x-auto rounded-lg border border-slate-800 bg-slate-950/80 p-4 font-mono text-[11px] leading-relaxed text-slate-400"
          >
            <p class="mb-1 font-semibold text-amber-500/90">Error: {String(error)}</p>
            {#if error instanceof Error && error.stack}
              <pre class="mt-2 whitespace-pre-wrap">{error.stack}</pre>
            {/if}
          </div>
        {/if}
      </div>
    {/if}
  </div>
</div>
