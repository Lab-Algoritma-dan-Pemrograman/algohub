<script lang="ts">
  import { ArrowUp } from '@lucide/svelte';

  let scrollY = $state(0);
  let show = $derived(scrollY > 300);

  function scrollToTop() {
    window.scrollTo({ top: 0, behavior: 'smooth' });
  }
</script>

<svelte:window bind:scrollY />

{#if show}
  <button
    onclick={scrollToTop}
    class="fixed bottom-6 right-6 z-50 flex h-10 w-10 items-center justify-center rounded-full bg-primary text-white shadow-lg transition-all hover:bg-primary-hover hover:-translate-y-0.5 focus:outline-none"
    aria-label="Kembali ke atas"
  >
    <ArrowUp size={18} />
  </button>
{/if}
