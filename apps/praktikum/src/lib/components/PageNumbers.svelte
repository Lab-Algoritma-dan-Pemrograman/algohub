<script lang="ts">
  import { ChevronLeft, ChevronRight } from '@lucide/svelte';

  let {
    page,
    totalPages,
    onchange,
    activeClass = 'bg-primary text-white border-primary shadow-sm',
    prevLabel,
    nextLabel,
  }: {
    page: number;
    totalPages: number;
    onchange: (page: number) => void;
    activeClass?: string;
    prevLabel?: string;
    nextLabel?: string;
  } = $props();

  let pages = $derived(Array.from({ length: totalPages }, (_, i) => i + 1));
</script>

<div class="flex items-center gap-1">
  <button
    onclick={() => onchange(Math.max(1, page - 1))}
    disabled={page === 1}
    class="w-8 h-8 rounded-lg bg-white border border-slate-200 text-slate-600 flex items-center justify-center hover:bg-slate-50 hover:border-slate-300 transition-all disabled:opacity-40"
    aria-label={prevLabel}
  >
    <ChevronLeft size={16} />
  </button>

  {#each pages as p (p)}
    {#if p === 1 || p === totalPages || (p >= page - 1 && p <= page + 1)}
      <button
        onclick={() => onchange(p)}
        class="w-8 h-8 rounded-lg text-xs font-bold border transition-all {page === p
          ? activeClass
          : 'bg-white text-slate-600 border-slate-200 hover:bg-slate-50 hover:border-slate-300'}"
      >
        {p}
      </button>
    {:else if p === 2 || p === totalPages - 1}
      <span class="w-8 text-center text-xs font-bold text-slate-400">...</span>
    {/if}
  {/each}

  <button
    onclick={() => onchange(Math.min(totalPages, page + 1))}
    disabled={page === totalPages}
    class="w-8 h-8 rounded-lg bg-white border border-slate-200 text-slate-600 flex items-center justify-center hover:bg-slate-50 hover:border-slate-300 transition-all disabled:opacity-40"
    aria-label={nextLabel}
  >
    <ChevronRight size={16} />
  </button>
</div>
