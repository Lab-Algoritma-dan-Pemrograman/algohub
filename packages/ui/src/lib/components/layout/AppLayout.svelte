<script lang="ts">
  import type { Snippet } from 'svelte';

  let { header, sidebar, children }: { header?: Snippet; sidebar?: Snippet; children: Snippet } =
    $props();
</script>

<div class="flex min-h-screen flex-col">
  {#if header}
    <header class="bg-navbar text-white">{@render header()}</header>
  {/if}
  <div class="flex flex-1">
    {#if sidebar}
      <aside class="bg-sidebar text-white">{@render sidebar()}</aside>
    {/if}
    <main class="bg-surface-base flex-1">{@render children()}</main>
  </div>
</div>
