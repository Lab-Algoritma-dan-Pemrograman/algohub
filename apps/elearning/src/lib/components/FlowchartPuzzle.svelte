<script lang="ts">
  import ArrowDown from '@lucide/svelte/icons/arrow-down';
  import CheckCircle from '@lucide/svelte/icons/circle-check';
  import GripVertical from '@lucide/svelte/icons/grip-vertical';
  import RotateCcw from '@lucide/svelte/icons/rotate-ccw';
  import Trash from '@lucide/svelte/icons/trash-2';
  import { untrack } from 'svelte';
  import {
    SHAPE_LABEL,
    branchGroupOf,
    buildBank,
    buildSlots,
    parseFlow,
    sameCard,
    type BankItem,
    type FlowchartShape,
    type FlowchartSymbol,
    type SlotRef,
  } from '#lib/flowchart.ts';

  // Latihan susun simbol flowchart dari kode sumber. Penilaian murni di klien: kartu di setiap
  // slot harus sama (bentuk dan label) dengan solusi, termasuk cabang Ya/Tidak tiap decision.
  let {
    code,
    solutionJson,
    distractors = [],
    onresult,
    disabled = false,
  }: {
    code: string;
    solutionJson: string;
    distractors?: FlowchartSymbol[];
    onresult: (correct: boolean) => void;
    disabled?: boolean;
  } = $props();

  const flow = $derived(parseFlow(solutionJson));
  const slotRefs = $derived(buildSlots(flow));
  const slotById = $derived(new Map(slotRefs.map((r) => [r.id, r])));

  // Urutan acak ditentukan sekali saat komponen dibuat; halaman induk membuat ulang komponen
  // per lesson, jadi perubahan prop setelahnya tidak perlu mengacak ulang bank.
  const initialBank = untrack(() => buildBank(buildSlots(parseFlow(solutionJson)), distractors));

  let bank = $state<BankItem[]>([...initialBank]);
  let placed = $state<Record<string, BankItem>>({});
  let checked = $state<boolean | null>(null);
  let dragKey = $state<string | null>(null);
  let tapped = $state<string | null>(null);

  const filledCount = $derived(slotRefs.filter((r) => placed[r.id]).length);

  const findItem = (key: string): BankItem | null =>
    bank.find((b) => b.key === key) ?? Object.values(placed).find((p) => p.key === key) ?? null;
  const slotOfKey = (key: string): string | null =>
    Object.entries(placed).find(([, v]) => v.key === key)?.[0] ?? null;

  function placeAt(key: string, slotId: string) {
    if (disabled || !slotById.has(slotId)) return;
    const item = findItem(key);
    if (!item) return;
    const next = { ...placed };
    for (const [sid, it] of Object.entries(next)) if (it.key === key) delete next[sid];
    const displaced = next[slotId];
    next[slotId] = item;
    placed = next;
    bank = [...bank.filter((b) => b.key !== key), ...(displaced ? [displaced] : [])];
    checked = null;
  }

  function removeFromSlot(slotId: string) {
    if (disabled) return;
    const item = placed[slotId];
    if (!item) return;
    const next = { ...placed };
    delete next[slotId];
    placed = next;
    bank = [...bank, item];
    checked = null;
  }

  function reset() {
    bank = [...initialBank];
    placed = {};
    checked = null;
    tapped = null;
    onresult(false);
  }

  function check() {
    const correct =
      filledCount === slotRefs.length && slotRefs.every((r) => sameCard(placed[r.id], r.expected));
    checked = correct;
    onresult(correct);
  }

  function onDragStart(event: DragEvent, key: string) {
    event.dataTransfer?.setData('text/plain', key);
    dragKey = key;
  }

  function onDropSlot(event: DragEvent, slotId: string) {
    event.preventDefault();
    event.stopPropagation();
    const key = event.dataTransfer?.getData('text/plain');
    if (key) placeAt(key, slotId);
    dragKey = null;
  }

  function onDropBank(event: DragEvent) {
    event.preventDefault();
    event.stopPropagation();
    const key = event.dataTransfer?.getData('text/plain');
    if (!key) return;
    const slotId = slotOfKey(key);
    if (slotId) removeFromSlot(slotId);
    dragKey = null;
  }

  function allowDrop(event: DragEvent) {
    event.preventDefault();
    event.stopPropagation();
  }

  function onSlotClick(slot: SlotRef) {
    if (placed[slot.id]) removeFromSlot(slot.id);
    else if (tapped) {
      placeAt(tapped, slot.id);
      tapped = null;
    }
  }

  const SHAPE_CLASS: Record<FlowchartShape, string> = {
    terminator: 'min-w-[120px] rounded-full border-sky-600 bg-sky-50 text-sky-900',
    process: 'rounded-lg border-blue-600 bg-blue-50 text-blue-900',
    io: 'skew-x-[-12deg] rounded-md border-indigo-600 bg-indigo-50 text-indigo-900',
    decision:
      'my-2 h-[110px] min-h-0 w-[110px] rotate-45 rounded-md border-rose-600 bg-rose-50 text-rose-900',
    connector:
      'h-[56px] min-h-0 w-[56px] rounded-full border-orange-500 bg-orange-50 px-1 text-orange-900',
    predefined:
      'rounded-md border-teal-600 bg-teal-50 text-teal-900 [clip-path:polygon(12%_0,88%_0,100%_50%,88%_100%,12%_100%,0_50%)]',
  };

  const LABEL_CLASS: Partial<Record<FlowchartShape, string>> = {
    decision: '-rotate-45',
    io: 'skew-x-[12deg]',
  };
</script>

{#snippet symbol(shape: FlowchartShape, label: string)}
  <div
    class="flex min-h-[44px] items-center justify-center border-2 px-3 py-1 text-center text-[11px] font-bold select-none {SHAPE_CLASS[
      shape
    ]}"
  >
    <span class="max-w-[160px] leading-tight break-words {LABEL_CLASS[shape] ?? ''}">{label}</span>
  </div>
{/snippet}

{#snippet slotView(ref: SlotRef, compact: boolean)}
  {@const item = placed[ref.id]}
  {@const ok = sameCard(item, ref.expected)}
  <button
    type="button"
    ondragover={allowDrop}
    ondrop={(e) => onDropSlot(e, ref.id)}
    onclick={() => onSlotClick(ref)}
    aria-label={item ? `Lepas ${item.label}` : 'Slot kosong'}
    class="group relative flex w-full cursor-pointer items-center justify-center rounded-xl border-2 border-dashed transition-all {compact
      ? 'min-h-[56px]'
      : 'min-h-[64px]'} {item
      ? 'border-transparent'
      : 'border-zinc-300 hover:border-rose-400 hover:bg-rose-50/40'} {!item && tapped
      ? 'animate-pulse border-rose-400 bg-rose-50/60'
      : ''}"
  >
    {#if item}
      <div class="relative">
        {@render symbol(item.shape, item.label)}
        <div
          class="absolute -top-2 -right-2 rounded-full bg-red-500 p-0.5 text-white opacity-0 shadow transition-opacity group-hover:opacity-100"
        >
          <Trash size={10} />
        </div>
      </div>
    {:else}
      <span class="text-[10px] font-bold tracking-widest text-zinc-300 uppercase">Slot</span>
    {/if}
    {#if checked !== null && item}
      <div
        class="absolute top-1/2 -left-2 -translate-y-1/2 rounded-full p-0.5 {ok
          ? 'text-emerald-600'
          : 'text-red-500'}"
      >
        {#if ok}
          <CheckCircle size={16} />
        {:else}
          <span class="text-xs font-black">✕</span>
        {/if}
      </div>
    {/if}
  </button>
{/snippet}

<div class="flex h-full flex-col gap-5">
  <div class="shrink-0 overflow-hidden rounded-2xl border border-zinc-800 bg-zinc-950">
    <div class="flex items-center gap-1.5 border-b border-zinc-800 px-4 py-2">
      <span class="h-2.5 w-2.5 rounded-full bg-[#ff5f56]"></span>
      <span class="h-2.5 w-2.5 rounded-full bg-[#ffbd2e]"></span>
      <span class="h-2.5 w-2.5 rounded-full bg-[#27c93f]"></span>
      <span class="ml-2 text-[10px] font-bold tracking-widest text-zinc-500 uppercase">
        kode sumber — terjemahkan ke flowchart
      </span>
    </div>
    <pre
      class="max-h-56 overflow-x-auto overflow-y-auto p-4 font-mono text-xs whitespace-pre-wrap text-zinc-100">{code}</pre>
  </div>

  <div class="grid min-h-0 flex-1 grid-cols-1 gap-5 md:grid-cols-2">
    <div
      role="group"
      aria-label="Bank simbol"
      class="custom-scrollbar overflow-y-auto rounded-2xl border-2 border-dashed border-zinc-200 bg-zinc-50 p-4"
      ondragover={allowDrop}
      ondrop={onDropBank}
    >
      <div
        class="mb-3 flex items-center gap-1.5 text-[10px] font-black tracking-widest text-zinc-400 uppercase"
      >
        <GripVertical size={12} /> Bank Simbol — seret atau ketuk kartu, lalu ketuk slot
      </div>
      <div class="flex flex-wrap gap-2">
        {#each bank as item (item.key)}
          <button
            type="button"
            draggable={!disabled}
            ondragstart={(e) => onDragStart(e, item.key)}
            ondragend={() => (dragKey = null)}
            onclick={() => (tapped = tapped === item.key ? null : item.key)}
            {disabled}
            title={SHAPE_LABEL[item.shape]}
            class="cursor-grab rounded-xl transition-all hover:scale-[1.03] hover:shadow-md active:cursor-grabbing {tapped ===
            item.key
              ? 'scale-[1.03] ring-2 ring-rose-500 ring-offset-2'
              : ''} {dragKey === item.key ? 'opacity-40' : ''}"
          >
            {@render symbol(item.shape, item.label)}
          </button>
        {/each}
        {#if bank.length === 0}
          <p class="text-xs text-zinc-400 italic">
            Semua kartu sudah dipasang. Ketuk kartu di slot untuk melepasnya.
          </p>
        {/if}
      </div>
    </div>

    <div class="custom-scrollbar overflow-y-auto rounded-2xl border border-zinc-200 bg-white p-4">
      <div class="mb-3 text-[10px] font-black tracking-widest text-zinc-400 uppercase">
        Susunan Flowchart ({filledCount}/{slotRefs.length})
      </div>
      <div class="flex flex-col items-center gap-0">
        {#each flow.main as step, i (i)}
          {@const k = branchGroupOf(flow.main, i)}
          {@const yesRefs =
            step.shape === 'decision'
              ? slotRefs.filter((r) => r.branch === 'yes' && r.owner === k)
              : []}
          {@const noRefs =
            step.shape === 'decision'
              ? slotRefs.filter((r) => r.branch === 'no' && r.owner === k)
              : []}
          <div class="w-full">
            {@render slotView(slotById.get(`m-${i}`)!, false)}
          </div>
          {#if step.shape === 'decision' && (yesRefs.length > 0 || noRefs.length > 0)}
            <div class="my-2 grid w-full grid-cols-2 gap-3">
              <div class="flex flex-col items-center gap-1">
                <span class="text-[10px] font-black tracking-widest text-emerald-600 uppercase">
                  Ya
                </span>
                {#each yesRefs as ref (ref.id)}
                  {@render slotView(ref, true)}
                {:else}
                  <span class="text-[10px] text-zinc-300 italic">(tanpa langkah)</span>
                {/each}
              </div>
              <div class="flex flex-col items-center gap-1">
                <span class="text-[10px] font-black tracking-widest text-red-500 uppercase">
                  Tidak
                </span>
                {#each noRefs as ref (ref.id)}
                  {@render slotView(ref, true)}
                {:else}
                  <span class="text-[10px] text-zinc-300 italic">(tanpa langkah)</span>
                {/each}
              </div>
            </div>
          {/if}
          {#if i < flow.main.length - 1}
            <ArrowDown size={14} class="my-0.5 shrink-0 text-zinc-300" />
          {/if}
        {/each}
      </div>
    </div>
  </div>

  <div class="flex shrink-0 gap-3">
    <button
      type="button"
      onclick={reset}
      {disabled}
      class="flex items-center gap-2 rounded-2xl border border-zinc-200 bg-zinc-100 px-5 py-3 text-sm font-bold text-zinc-600 transition-all hover:bg-zinc-200 active:scale-95"
    >
      <RotateCcw size={15} /> Acak Ulang
    </button>
    <button
      type="button"
      onclick={check}
      disabled={disabled || filledCount < slotRefs.length}
      class="flex-1 rounded-2xl bg-zinc-900 py-3 text-sm font-bold text-white transition-all hover:bg-zinc-800 active:scale-95 disabled:opacity-40"
    >
      Periksa Susunan
    </button>
  </div>

  {#if checked !== null}
    <div
      role="status"
      class="rounded-2xl border p-4 text-center text-sm font-bold {checked
        ? 'border-emerald-200 bg-emerald-50 text-emerald-700'
        : 'border-red-200 bg-red-50 text-red-600'}"
    >
      {checked
        ? 'Susunan flowchart tepat! Alur kode sudah benar dipetakan.'
        : 'Belum tepat — perhatikan tanda ✕ pada slot yang salah, lalu perbaiki urutannya.'}
    </div>
  {/if}
</div>
