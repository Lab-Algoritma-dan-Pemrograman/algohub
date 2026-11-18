<script lang="ts">
  import { goto } from '$app/navigation';
  import { resolve } from '$app/paths';
  import BookOpenReader from '@lucide/svelte/icons/book-open';
  import Box from '@lucide/svelte/icons/package';
  import Check from '@lucide/svelte/icons/check';
  import Clock from '@lucide/svelte/icons/clock';
  import LaptopCode from '@lucide/svelte/icons/laptop';
  import List from '@lucide/svelte/icons/list';
  import Lock from '@lucide/svelte/icons/lock';
  import Play from '@lucide/svelte/icons/play';
  import Shapes from '@lucide/svelte/icons/shapes';
  import type { Component } from 'svelte';
  import { fly } from 'svelte/transition';
  import { resumeLessonId, type TreeLevel } from '#lib/curriculum.ts';

  let {
    module,
    locked,
    progress,
    isLangC,
    cardIndex,
    completed,
  }: {
    module: TreeLevel['modules'][number];
    locked: boolean;
    progress: number;
    isLangC: boolean;
    cardIndex: number;
    completed: ReadonlySet<string>;
  } = $props();

  interface Theme {
    icon: Component<{ size?: number }>;
    bg: string;
    textColor: string;
  }

  const THEMES: Theme[] = [
    {
      icon: BookOpenReader,
      bg: 'bg-blue-50 text-fun-blue shadow-[0_4px_0_#DBEAFE] group-hover:bg-fun-blue group-hover:text-white group-hover:shadow-[0_4px_0_#0284C7]',
      textColor: 'group-hover:text-fun-blue',
    },
    {
      icon: LaptopCode,
      bg: 'bg-teal-50 text-fun-green shadow-[0_4px_0_#CCFBF1] group-hover:bg-fun-green group-hover:text-white group-hover:shadow-[0_4px_0_#047857]',
      textColor: 'group-hover:text-fun-green',
    },
    {
      icon: Box,
      bg: 'bg-purple-50 text-fun-purple shadow-[0_4px_0_#F3E8FF] group-hover:bg-fun-purple group-hover:text-white group-hover:shadow-[0_4px_0_#7E22CE]',
      textColor: 'group-hover:text-fun-purple',
    },
    {
      icon: Shapes,
      bg: 'bg-amber-50 text-fun-yellow shadow-[0_4px_0_#FEF3C7] group-hover:bg-fun-yellow group-hover:text-white group-hover:shadow-[0_4px_0_#D97706]',
      textColor: 'group-hover:text-fun-yellow',
    },
  ];

  const theme = $derived(THEMES[cardIndex % THEMES.length]);
  const complete = $derived(progress === 100);
  const lessonCount = $derived(module.lessons.length);
  const Icon = $derived(theme.icon);

  function open(lessonId: string | null) {
    if (locked || !lessonId) return;
    void goto(resolve('/(app)/lesson/[id]', { id: lessonId }));
  }
</script>

<div
  in:fly={{ y: 30, duration: 300, delay: cardIndex * 100 }}
  class="group relative flex h-full flex-col overflow-hidden rounded-[2rem] border-2 p-7 shadow-soft transition-colors duration-300 select-none {locked
    ? 'cursor-not-allowed border-zinc-200 bg-zinc-50 opacity-50 grayscale'
    : 'card-hover border-gray-100 bg-white hover:border-maroon/20'}"
>
  {#if locked}
    <div class="absolute top-5 right-5 z-20 text-zinc-400">
      <div
        class="flex h-9 w-9 items-center justify-center rounded-xl border border-zinc-200 bg-zinc-100 shadow-sm"
      >
        <Lock size={16} />
      </div>
    </div>
  {/if}

  <div class="mt-2 mb-5 flex items-start justify-between">
    <div
      class="flex h-14 w-14 items-center justify-center rounded-2xl text-2xl transition-all duration-300 group-hover:scale-110 {locked
        ? 'bg-zinc-100 text-zinc-400'
        : theme.bg} {cardIndex % 2 === 0 ? 'group-hover:-rotate-6' : 'group-hover:rotate-6'}"
    >
      {#if locked}<Lock size={24} />{:else}<Icon size={24} />{/if}
    </div>
  </div>

  <h3
    class="mb-3 text-xl leading-tight font-black text-dark transition-colors {locked
      ? ''
      : theme.textColor}"
  >
    {module.title}
  </h3>

  <div class="mb-6 flex items-center space-x-2 text-xs font-bold text-gray-500">
    <span
      class="flex items-center rounded-lg border border-gray-100 bg-gray-50 px-2.5 py-1 text-gray-600"
    >
      <List size={12} class="mr-1.5 text-fun-yellow" />
      {lessonCount} Pelajaran
    </span>
    <span
      class="flex items-center rounded-lg border border-gray-100 bg-gray-50 px-2.5 py-1 text-gray-600"
    >
      <Clock size={12} class="mr-1.5 text-fun-blue" />
      ~{lessonCount * 5} mnt
    </span>
  </div>

  <div class="mb-6 flex-1 space-y-4">
    {#each module.lessons.slice(0, 3) as lesson (lesson.id)}
      {@const done = completed.has(lesson.id)}
      <button
        type="button"
        disabled={locked}
        onclick={() => open(lesson.id)}
        class="group/item flex w-full items-start space-x-3 text-left {locked
          ? 'cursor-not-allowed'
          : 'cursor-pointer'}"
      >
        <span
          class="mt-0.5 flex h-6 w-6 shrink-0 items-center justify-center rounded-full border-2 bg-white shadow-sm transition-all group-hover/item:scale-110 {done
            ? 'border-fun-green bg-fun-green text-white'
            : 'border-gray-300 group-hover/item:border-maroon/50'}"
        >
          <Check size={14} class="text-white transition-all {done ? 'opacity-100' : 'opacity-0'}" />
        </span>
        <span
          class="truncate pt-0.5 text-sm font-bold transition-colors {locked
            ? 'text-zinc-400'
            : 'text-gray-600 group-hover/item:text-maroon group-hover/item:underline'}"
        >
          {lesson.title}
        </span>
      </button>
    {/each}
  </div>

  <div class="mt-auto">
    {#if lessonCount > 3}
      <div
        class="mb-5 w-full rounded-lg px-3 py-1.5 text-left text-xs font-black tracking-widest uppercase {locked
          ? 'bg-zinc-100 text-zinc-400'
          : isLangC
            ? 'bg-blue-50 text-fun-blue'
            : 'bg-rose-50 text-rose-700'}"
      >
        + {lessonCount - 3} PELAJARAN LAGI
      </div>
    {:else}
      <div class="mb-5 h-[34px]"></div>
    {/if}
    <div class="flex items-center justify-between border-t-2 border-dashed border-gray-100 pt-5">
      <span
        class="rounded-full bg-gray-50 px-3 py-1 text-xs font-black tracking-widest text-gray-400 uppercase"
      >
        {progress}% Selesai
      </span>
      <button
        type="button"
        disabled={locked}
        onclick={() => open(resumeLessonId(module, completed))}
        aria-label={complete ? `Tinjau ${module.title}` : `Lanjutkan ${module.title}`}
        class="flex h-10 w-10 items-center justify-center rounded-xl shadow-sm transition-all {locked
          ? 'cursor-not-allowed bg-gray-100 text-gray-400'
          : complete
            ? 'shadow-bubbly -translate-y-0 cursor-pointer bg-fun-green text-white hover:-translate-y-1 hover:bg-emerald-600'
            : 'hover:shadow-bubbly-maroon cursor-pointer bg-gray-100 text-gray-500 hover:-translate-y-1 hover:bg-maroon hover:text-white'}"
      >
        {#if complete}<Check size={18} />{:else}<Play size={14} class="ml-0.5" />{/if}
      </button>
    </div>
  </div>
</div>
