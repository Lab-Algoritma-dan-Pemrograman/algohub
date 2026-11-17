<script lang="ts" generics="TData">
  import { createVirtualizer } from '@tanstack/svelte-virtual';
  import { getCoreRowModel, type ColumnDef, type Row } from '@tanstack/table-core';
  import { untrack } from 'svelte';
  import { cn } from '../../utils.js';
  import { createSvelteTable, FlexRender } from '../ui/data-table/index.js';
  import * as Table from '../ui/table/index.js';
  import { shouldVirtualize, virtualPadding } from './virtual.js';
  let {
    data,
    columns,
    getRowId,
    estimateRowHeight = 48,
    maxHeight = '70vh',
    class: className,
  }: {
    data: TData[];
    // Tipe nilai tiap kolom berbeda; ColumnDef<TData, unknown> menolak kolom accessor bertipe.
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    columns: ColumnDef<TData, any>[];
    getRowId?: (row: TData, index: number) => string;
    estimateRowHeight?: number;
    maxHeight?: string;
    class?: string;
  } = $props();
  const table = createSvelteTable({
    get data() {
      return data;
    },
    get columns() {
      return columns;
    },
    get getRowId() {
      return getRowId;
    },
    getCoreRowModel: getCoreRowModel(),
    manualPagination: true,
    manualSorting: true,
    manualFiltering: true,
  });
  const rows = $derived(table.getRowModel().rows);
  const virtual = $derived(shouldVirtualize(rows.length));
  let scrollEl = $state<HTMLDivElement | null>(null);
  const virtualizer = createVirtualizer<HTMLDivElement, HTMLTableRowElement>({
    count: 0,
    getScrollElement: () => scrollEl,
    estimateSize: () => estimateRowHeight,
    overscan: 10,
  });
  $effect(() => {
    const count = virtual ? rows.length : 0;
    const size = estimateRowHeight;
    void scrollEl;
    // setOptions memperbarui store; membacanya tanpa untrack membuat effect ini berulang terus.
    untrack(() => $virtualizer).setOptions({ count, estimateSize: () => size });
  });
  const padding = $derived(
    virtualPadding($virtualizer.getVirtualItems(), $virtualizer.getTotalSize()),
  );
</script>

{#snippet header()}
  <Table.Header>
    {#each table.getHeaderGroups() as headerGroup (headerGroup.id)}
      <Table.Row>
        {#each headerGroup.headers as head (head.id)}
          <Table.Head colspan={head.colSpan}>
            {#if !head.isPlaceholder}
              <FlexRender content={head.column.columnDef.header} context={head.getContext()} />
            {/if}
          </Table.Head>
        {/each}
      </Table.Row>
    {/each}
  </Table.Header>
{/snippet}
{#snippet row(r: Row<TData>)}
  <Table.Row>
    {#each r.getVisibleCells() as cell (cell.id)}
      <Table.Cell>
        <FlexRender content={cell.column.columnDef.cell} context={cell.getContext()} />
      </Table.Cell>
    {/each}
  </Table.Row>
{/snippet}
{#if virtual}
  <div bind:this={scrollEl} class={cn('overflow-auto', className)} style:max-height={maxHeight}>
    <Table.Root>
      {@render header()}
      <Table.Body>
        {#if padding.top > 0}
          <tr style:height="{padding.top}px"></tr>
        {/if}
        {#each $virtualizer.getVirtualItems() as item (item.key)}
          {@const r = rows[item.index]}
          <!-- Baris bisa berkurang sebelum effect memperbarui count virtualizer. -->
          {#if r}
            {@render row(r)}
          {/if}
        {/each}
        {#if padding.bottom > 0}
          <tr style:height="{padding.bottom}px"></tr>
        {/if}
      </Table.Body>
    </Table.Root>
  </div>
{:else}
  <div class={className}>
    <Table.Root>
      {@render header()}
      <Table.Body>
        {#each rows as r (r.id)}
          {@render row(r)}
        {/each}
      </Table.Body>
    </Table.Root>
  </div>
{/if}
