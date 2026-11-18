<script lang="ts">
  import { resolve } from '$app/paths';
  import ArrowLeft from '@lucide/svelte/icons/arrow-left';
  import BookOpen from '@lucide/svelte/icons/book-open';
  import CheckCircle from '@lucide/svelte/icons/circle-check';
  import Lock from '@lucide/svelte/icons/lock';
  import X from '@lucide/svelte/icons/x';
  import { onMount } from 'svelte';
  import { fade, fly } from 'svelte/transition';
  import { isLessonLevelLocked, type AccessOverrides, type TreeLevel } from '#lib/curriculum.ts';

  let {
    curriculum,
    currentLessonId,
    currentLevelIdx,
    completed,
    overrides,
    onselect,
    onclose,
  }: {
    curriculum: TreeLevel[];
    currentLessonId: string;
    currentLevelIdx: number;
    completed: ReadonlySet<string>;
    overrides: AccessOverrides;
    onselect: (lessonId: string) => void;
    onclose: () => void;
  } = $props();

  let container = $state<HTMLDivElement>();

  // Daftar panjang dibuka di posisi lesson aktif supaya tidak perlu menggulir dari atas.
  onMount(() => {
    const timer = setTimeout(() => {
      const active = container?.querySelector<HTMLElement>('[aria-current="true"]');
      if (!container || !active) return;
      const top = active.offsetTop - container.clientHeight / 2 + active.offsetHeight / 2;
      container.scrollTo({ top, behavior: 'smooth' });
    }, 150);
    return () => clearTimeout(timer);
  });
</script>

<svelte:window onkeydown={(e) => e.key === 'Escape' && onclose()} />

<button
  type="button"
  transition:fade={{ duration: 150 }}
  onclick={onclose}
  class="fixed inset-0 z-[60] cursor-pointer bg-zinc-950/40 backdrop-blur-sm"
  aria-label="Tutup daftar pelajaran"
></button>
<aside
  transition:fly={{ x: 340, duration: 250 }}
  aria-label="Kurikulum belajar"
  class="fixed top-0 right-0 bottom-0 z-[70] flex w-[21rem] max-w-full flex-col overflow-hidden rounded-l-[2.5rem] border-l border-zinc-200 bg-white shadow-2xl"
>
  <div class="flex items-center justify-between border-b border-zinc-100 bg-zinc-50/50 p-6">
    <div class="flex items-center gap-2">
      <div class="flex h-8 w-8 items-center justify-center rounded-lg bg-rose-700/10 text-rose-700">
        <BookOpen size={18} />
      </div>
      <span class="text-base font-black tracking-tight text-zinc-800">Kurikulum Belajar</span>
    </div>
    <button
      type="button"
      onclick={onclose}
      aria-label="Tutup"
      class="rounded-xl border border-transparent p-2 text-zinc-500 transition-colors hover:border-zinc-200 hover:bg-zinc-200/50 hover:text-zinc-800"
    >
      <X size={18} />
    </button>
  </div>

  <div bind:this={container} class="custom-scrollbar flex-1 space-y-6 overflow-y-auto bg-white p-6">
    {#each curriculum as level, lIdx (level.id)}
      {@const levelLocked = isLessonLevelLocked(curriculum, lIdx, completed, overrides)}
      {@const levelActive = lIdx === currentLevelIdx}
      <div
        class="relative overflow-hidden rounded-[1.75rem] border bg-zinc-50/40 p-5 transition-all duration-300 {levelActive
          ? 'border-rose-200 bg-rose-50/10 shadow-md shadow-rose-500/5'
          : 'border-zinc-200 hover:border-zinc-300'}"
      >
        <div class="mb-4 flex items-center gap-3">
          <div
            class="flex h-8 w-8 items-center justify-center rounded-xl border text-xs font-black shadow-sm {levelActive
              ? 'border-rose-600 bg-rose-700 text-white shadow-rose-700/20'
              : 'border-zinc-200 bg-white text-zinc-500'}"
          >
            {#if levelLocked}<Lock size={12} />{:else}{lIdx + 1}{/if}
          </div>
          <div class="min-w-0 flex-1">
            <div
              class="mb-1 text-[9px] leading-none font-black tracking-widest text-zinc-400 uppercase"
            >
              Level {lIdx + 1}
            </div>
            <h4 class="truncate text-xs font-bold text-zinc-800">{level.title}</h4>
          </div>
        </div>

        {#if !levelLocked}
          <div class="space-y-4 pt-1">
            {#each level.modules as mod (mod.id)}
              <div class="relative pl-3">
                <div
                  class="mb-2 flex items-center gap-1.5 text-[9px] font-black tracking-widest text-zinc-400 uppercase"
                >
                  <div class="h-1.5 w-1.5 shrink-0 rounded-full bg-zinc-300"></div>
                  <span class="truncate">{mod.title}</span>
                </div>
                <div class="relative ml-0.5 space-y-1.5 border-l border-zinc-200 pl-3">
                  {#each mod.lessons as les (les.id)}
                    {@const active = les.id === currentLessonId}
                    {@const done = completed.has(les.id)}
                    <button
                      type="button"
                      aria-current={active ? 'true' : undefined}
                      onclick={() => onselect(les.id)}
                      class="group relative flex w-full items-center gap-2.5 rounded-xl border border-transparent px-3 py-2 text-left text-xs transition-all {active
                        ? '-translate-y-px scale-[1.02] border-b-2 border-rose-800 bg-gradient-to-r from-rose-700 to-rose-600 font-bold text-white shadow-md shadow-rose-700/25'
                        : done
                          ? 'text-zinc-500 hover:translate-x-0.5 hover:bg-zinc-100/70 hover:text-zinc-800'
                          : 'text-zinc-700 hover:translate-x-0.5 hover:bg-zinc-100/70 hover:text-zinc-950'}"
                    >
                      {#if done}
                        <CheckCircle
                          size={13}
                          class="shrink-0 {active ? 'text-white' : 'text-rose-700'}"
                        />
                      {:else if active}
                        <span class="relative my-0.5 flex h-2 w-2 shrink-0">
                          <span
                            class="absolute inline-flex h-full w-full animate-ping rounded-full bg-rose-200 opacity-75"
                          ></span>
                          <span class="relative inline-flex h-2 w-2 rounded-full bg-white"></span>
                        </span>
                      {:else}
                        <div
                          class="h-2 w-2 shrink-0 rounded-full border border-zinc-300 bg-white group-hover:border-zinc-400"
                        ></div>
                      {/if}
                      <span class="flex-1 truncate">{les.title}</span>
                    </button>
                  {/each}
                </div>
              </div>
            {/each}
          </div>
        {:else}
          <div class="flex items-center gap-1.5 pt-1 pl-1 text-[10px] text-zinc-400 italic">
            <Lock size={10} />
            Selesaikan level sebelumnya untuk membuka
          </div>
        {/if}
      </div>
    {/each}
  </div>

  <div class="border-t border-zinc-100 bg-zinc-50/50 p-5">
    <a
      href={resolve('/(app)/courses')}
      class="flex w-full items-center justify-center gap-2 rounded-2xl bg-zinc-100 py-3.5 text-xs font-bold tracking-wider text-zinc-700 uppercase transition-all hover:bg-zinc-200/80"
    >
      <ArrowLeft size={14} />
      Kembali ke Belajar
    </a>
  </div>
</aside>
