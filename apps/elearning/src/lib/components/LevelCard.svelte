<script lang="ts">
  import { resolve } from '$app/paths';
  import type { RouteId } from '$app/types';
  import Check from '@lucide/svelte/icons/check';
  import Lock from '@lucide/svelte/icons/lock';
  import { lessonIds, type TreeLevel } from '#lib/curriculum.ts';

  let {
    level,
    index,
    locked,
    completed,
    color,
  }: {
    level: TreeLevel;
    index: number;
    locked: boolean;
    completed: ReadonlySet<string>;
    color: 'blue' | 'rose';
  } = $props();

  const ids = $derived(lessonIds(level));
  const total = $derived(ids.length);
  const done = $derived(ids.filter((id) => completed.has(id)).length);
  const percent = $derived(total > 0 ? Math.round((done / total) * 100) : 0);
  const complete = $derived(percent === 100 && total > 0);
  const blue = $derived(color === 'blue');
  const radius = 28;
  const circumference = 2 * Math.PI * radius;
</script>

<div
  class="relative flex flex-col justify-between overflow-hidden rounded-[2rem] border-2 p-6 transition-all duration-300 {locked
    ? 'cursor-not-allowed border-zinc-200 bg-zinc-50 opacity-50 shadow-sm grayscale'
    : `card-hover hover:border-maroon/20 shadow-soft group bg-gradient-to-br ${blue ? 'border-blue-100/80 from-blue-50/70 via-white to-indigo-50/10' : 'border-rose-100/80 from-rose-50/70 via-white to-pink-50/10'}`}"
>
  {#if locked}
    <div
      class="absolute inset-0 z-10 flex items-center justify-center rounded-[1.8rem] bg-zinc-100/60 backdrop-blur-[2px]"
    >
      <div class="flex flex-col items-center gap-2 text-zinc-400">
        <div
          class="flex h-12 w-12 items-center justify-center rounded-2xl bg-zinc-200 shadow-inner"
        >
          <Lock size={16} />
        </div>
        <span class="text-[10px] font-black tracking-[0.2em] uppercase">Terkunci</span>
      </div>
    </div>
  {/if}

  <div class="relative z-[1] flex flex-1 flex-col justify-between">
    <div>
      <div class="mb-4 flex items-start justify-between">
        <div class="flex items-center gap-3">
          <div
            class="flex h-11 w-11 items-center justify-center rounded-xl text-sm font-black transition-all select-none {locked
              ? 'bg-zinc-200 text-zinc-400 shadow-sm'
              : complete
                ? 'bg-fun-green rotate-[-5deg] text-white shadow-[0_3px_0_#047857] group-hover:rotate-0'
                : blue
                  ? 'bg-fun-blue rotate-[-5deg] text-white shadow-[0_3px_0_#023E8A] group-hover:rotate-0'
                  : 'rotate-[-5deg] bg-rose-500 text-white shadow-[0_3px_0_#9F1239] group-hover:rotate-0'}"
          >
            {#if locked}
              <Lock size={12} />
            {:else if complete}
              <Check size={14} />
            {:else}
              {index + 1}
            {/if}
          </div>
          <div>
            <div
              class="text-[9px] font-black tracking-[0.2em] uppercase {complete
                ? 'text-fun-green'
                : blue
                  ? 'text-fun-blue'
                  : 'text-rose-600'}"
            >
              Level {index + 1}
            </div>
            <div class="mt-0.5 text-[10px] font-black tracking-wide text-zinc-500">
              {level.modules.length} Modul • {total} Pelajaran
            </div>
          </div>
        </div>
        {#if !locked}
          <div class="relative flex h-16 w-16 shrink-0 items-center justify-center">
            <svg class="h-16 w-16 -rotate-90" viewBox="0 0 64 64" aria-hidden="true">
              <circle
                cx="32"
                cy="32"
                r={radius}
                fill="none"
                stroke={blue ? '#eff6ff' : '#fff1f2'}
                stroke-width="5"
              />
              <circle
                cx="32"
                cy="32"
                r={radius}
                fill="none"
                stroke={blue ? '#2563eb' : '#e11d48'}
                stroke-width="5"
                stroke-linecap="round"
                stroke-dasharray={circumference}
                stroke-dashoffset={circumference - (percent / 100) * circumference}
                style="transition: stroke-dashoffset 1s ease-out"
              />
            </svg>
            <span
              class="absolute inset-0 flex items-center justify-center text-xs font-black {complete
                ? 'text-fun-green'
                : 'text-zinc-700'}">{percent}%</span
            >
          </div>
        {/if}
      </div>

      <h4 class="mb-3 line-clamp-2 text-base leading-tight font-black text-zinc-900">
        {#if !locked && ids.length > 0}
          <a href={resolve(`/lesson/${ids[0]}` as RouteId)} class="after:absolute after:inset-0"
            >{level.title}</a
          >
        {:else}
          {level.title}
        {/if}
      </h4>

      <div class="mb-4 space-y-2">
        {#each level.modules.slice(0, 3) as mod (mod.id)}
          {@const modDone = mod.lessons.filter((l) => completed.has(l.id)).length}
          {@const modComplete = modDone === mod.lessons.length && mod.lessons.length > 0}
          <div class="flex items-center gap-2.5 text-xs">
            <div
              class="flex h-5 w-5 shrink-0 items-center justify-center rounded-full border-2 bg-white shadow-sm {modComplete
                ? 'bg-fun-green border-fun-green text-white'
                : 'border-gray-300'}"
            >
              {#if modComplete}<Check size={12} />{/if}
            </div>
            <span
              class="flex-1 truncate text-xs {modComplete
                ? 'font-normal text-zinc-400 line-through'
                : 'font-bold text-zinc-700'}">{mod.title}</span
            >
            <span class="shrink-0 text-[10px] font-black text-zinc-400"
              >{modDone}/{mod.lessons.length}</span
            >
          </div>
        {/each}
        {#if level.modules.length > 3}
          <div class="ml-7 pt-0.5 text-[10px] font-black tracking-wide text-zinc-500">
            +{level.modules.length - 3} modul lagi
          </div>
        {/if}
      </div>
    </div>

    <div
      class="mt-3 flex flex-wrap items-center gap-1.5 border-t-2 border-dashed border-gray-100/60 pt-3"
    >
      {#each level.modules.flatMap((m) => m.lessons) as lesson (lesson.id)}
        {@const isDone = completed.has(lesson.id)}
        {#if locked}
          <span class="h-3.5 w-3.5 rounded-full bg-zinc-200/85"></span>
        {:else}
          <a
            href={resolve(`/lesson/${lesson.id}` as RouteId)}
            title={lesson.title}
            aria-label={lesson.title}
            class="relative z-10 flex h-3.5 w-3.5 items-center justify-center rounded-full shadow-sm transition-all hover:scale-125 {isDone
              ? 'bg-fun-green text-white shadow-[0_1.5px_0_#047857]'
              : 'bg-zinc-200/85 hover:bg-zinc-300'}"
          >
            {#if isDone}<Check size={8} />{/if}
          </a>
        {/if}
      {/each}
    </div>
  </div>
</div>
