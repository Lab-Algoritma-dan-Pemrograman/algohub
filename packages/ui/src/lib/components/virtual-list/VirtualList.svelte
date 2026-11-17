<script lang="ts" generics="T">
  import { createVirtualizer } from '@tanstack/svelte-virtual';
  import { untrack, type Snippet } from 'svelte';
  import { cn } from '../../utils.js';
  import { shouldVirtualize, VIRTUAL_THRESHOLD, virtualPadding } from '../data-table/virtual.js';
  let {
    items,
    key,
    estimateSize = 120,
    threshold = VIRTUAL_THRESHOLD,
    maxHeight = '70vh',
    class: className,
    item,
  }: {
    items: T[];
    key: (item: T, index: number) => string | number;
    estimateSize?: number;
    threshold?: number;
    maxHeight?: string;
    class?: string;
    item: Snippet<[T, number]>;
  } = $props();
  const virtual = $derived(shouldVirtualize(items.length, threshold));
  let scrollEl = $state<HTMLDivElement | null>(null);
  const virtualizer = createVirtualizer<HTMLDivElement, HTMLDivElement>({
    count: 0,
    getScrollElement: () => scrollEl,
    estimateSize: () => estimateSize,
    overscan: 5,
  });
  $effect(() => {
    const list = items;
    const count = virtual ? list.length : 0;
    const size = estimateSize;
    const toKey = key;
    void scrollEl;
    // setOptions memperbarui store; membacanya tanpa untrack membuat effect ini berulang terus.
    untrack(() => $virtualizer).setOptions({
      count,
      estimateSize: () => size,
      getItemKey: (index) => toKey(list[index], index),
    });
  });
  const padding = $derived(
    virtualPadding($virtualizer.getVirtualItems(), $virtualizer.getTotalSize()),
  );
  // Tinggi kartu berbeda-beda, jadi tiap elemen diukur setelah dirender.
  function measure(node: HTMLDivElement) {
    untrack(() => $virtualizer).measureElement(node);
  }
</script>

{#if virtual}
  <div bind:this={scrollEl} class="overflow-auto" style:max-height={maxHeight}>
    <div
      class={cn(className)}
      style:padding-top="{padding.top}px"
      style:padding-bottom="{padding.bottom}px"
    >
      {#each $virtualizer.getVirtualItems() as v (v.key)}
        <!-- Item bisa berkurang sebelum effect memperbarui count virtualizer. -->
        {#if v.index < items.length}
          <div data-index={v.index} {@attach measure}>
            {@render item(items[v.index], v.index)}
          </div>
        {/if}
      {/each}
    </div>
  </div>
{:else}
  <div class={cn(className)}>
    {#each items as it, i (key(it, i))}
      {@render item(it, i)}
    {/each}
  </div>
{/if}
