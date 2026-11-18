<script lang="ts">
  import Lock from '@lucide/svelte/icons/lock';
  import Sparkles from '@lucide/svelte/icons/sparkles';
  import Star from '@lucide/svelte/icons/star';
  import { createQuery } from '@tanstack/svelte-query';
  import { fade } from 'svelte/transition';
  import ModuleCard from '#lib/components/ModuleCard.svelte';
  import {
    buildCurriculum,
    isModuleLocked,
    levelCompletionPercent,
    moduleProgressPercent,
    overridesOf,
  } from '#lib/curriculum.ts';
  import { getLevelLanguage } from '#lib/levelLanguage.ts';
  import { levelsQuery, lessonsQuery, modulesQuery } from '#lib/queries/curriculum.ts';
  import { completedLessonsQuery } from '#lib/queries/progress.ts';
  import { appStore } from '#lib/stores/app.svelte.ts';

  const levels = createQuery(levelsQuery);
  const modules = createQuery(modulesQuery);
  const lessons = createQuery(lessonsQuery);
  const completedQuery = createQuery(completedLessonsQuery);

  const curriculum = $derived(
    buildCurriculum(levels.data ?? [], modules.data ?? [], lessons.data ?? []),
  );
  const completed = $derived(new Set(completedQuery.data ?? []));
  const overrides = $derived(overridesOf(appStore.profile?.levelAccessOverrides));
</script>

<div class="space-y-12" in:fade={{ duration: 200 }}>
  <div class="mb-10">
    <h1
      class="mb-4 flex items-center text-4xl font-black tracking-tight text-maroon select-none md:text-5xl"
    >
      Eksplorasi Kurikulum
      <Sparkles class="ml-4 animate-pulse text-fun-yellow" size={30} />
    </h1>
    <p class="max-w-3xl text-lg leading-relaxed font-bold text-gray-600">
      Yuk mulai petualangan ngoding kamu! Pilih
      <span class="rounded-lg border border-blue-100 bg-blue-50 px-2 py-1 font-black text-python">
        Python
      </span>
      untuk kemudahan dan dunia data, atau
      <span class="rounded-lg border border-blue-100 bg-blue-50 px-2 py-1 font-black text-c">
        Bahasa C
      </span>
      untuk ngulik jeroan sistem lebih dalam.
    </p>
  </div>

  {#if levels.isPending}
    <p class="text-center font-bold text-gray-400" role="status">Memuat kurikulum...</p>
  {:else if curriculum.length === 0}
    <p class="text-center font-bold text-gray-400">Kurikulum belum tersedia.</p>
  {/if}

  <div class="space-y-16">
    {#each curriculum as level, lIdx (level.id)}
      {@const isLangC = getLevelLanguage(level, lIdx) === 'c'}
      {@const levelLocked = isModuleLocked(curriculum, lIdx, 0, completed, overrides)}
      {@const levelProgress = levelCompletionPercent(level, completed)}
      <section class="space-y-6">
        <div
          class="group relative mb-8 flex flex-col items-start gap-6 overflow-hidden rounded-[2rem] border-2 border-gray-100 bg-white p-6 shadow-soft transition-colors hover:border-maroon/30 md:flex-row md:items-center"
        >
          <div
            class="pointer-events-none absolute top-0 right-0 h-32 w-32 rounded-bl-full transition-all {isLangC
              ? 'bg-fun-blue/5'
              : 'bg-rose-500/5'}"
          ></div>
          <div class="relative shrink-0">
            <div
              class="relative z-10 flex h-20 w-20 rotate-[-5deg] items-center justify-center rounded-3xl text-4xl font-black text-white transition-all select-none group-hover:rotate-0 {levelLocked
                ? 'bg-zinc-300 shadow-[0_8px_0_#71717a]'
                : isLangC
                  ? 'bg-fun-blue shadow-[0_8px_0_#023E8A]'
                  : 'bg-rose-500 shadow-[0_8px_0_#9F1239]'}"
            >
              {#if levelLocked}<Lock size={30} />{:else}{lIdx + 1}{/if}
            </div>
            {#if !levelLocked}
              <Star
                class="absolute -top-2 -right-2 z-20 animate-pulse fill-fun-yellow text-fun-yellow"
                size={20}
              />
            {/if}
          </div>

          <div class="flex-1">
            <div class="mb-2 flex flex-wrap items-center gap-3">
              <h2 class="text-2xl leading-snug font-black tracking-tight text-dark uppercase">
                {level.title}
              </h2>
              {#if levelLocked}
                <span
                  class="rounded-xl border border-zinc-200 bg-zinc-100 px-2.5 py-1 text-[10px] font-black tracking-widest text-zinc-500 uppercase"
                >
                  Terkunci
                </span>
              {/if}
              <span
                class="rounded-xl border-2 px-3 py-1.5 text-xs font-black tracking-wide uppercase {isLangC
                  ? 'border-blue-200 bg-blue-100 text-blue-700'
                  : 'border-rose-200 bg-rose-100 text-rose-700'}"
              >
                {isLangC ? 'Bahasa C' : 'Python'}
              </span>
            </div>
            <p
              class="inline-block rounded-lg border border-gray-100 bg-gray-50 px-3 py-1 text-sm font-bold text-gray-500"
            >
              {level.description}
            </p>
          </div>

          {#if !levelLocked}
            <div
              class="relative flex w-full flex-col items-center justify-center overflow-hidden rounded-2xl border-2 border-maroon/10 bg-maroon-bg p-4 select-none md:w-56"
            >
              <div class="z-10 mb-1 text-xs font-black tracking-widest text-maroon/70 uppercase">
                Progres Level
              </div>
              <div class="z-10 flex items-baseline space-x-1">
                <span class="text-3xl font-black text-maroon">{levelProgress}%</span>
                <span class="text-sm font-bold text-maroon/70">Selesai</span>
              </div>
              <div
                class="z-10 mt-3 h-3 w-full overflow-hidden rounded-full border border-maroon/20 bg-white shadow-inner"
                role="progressbar"
                aria-valuenow={levelProgress}
                aria-valuemin={0}
                aria-valuemax={100}
                aria-label="Progres level"
              >
                <div
                  class="h-full rounded-full bg-maroon transition-all duration-500"
                  style="width: {levelProgress}%"
                ></div>
              </div>
            </div>
          {/if}
        </div>

        <div class="grid grid-cols-1 gap-6 md:grid-cols-2 lg:grid-cols-3">
          {#each level.modules as module, mIdx (module.id)}
            <ModuleCard
              {module}
              locked={isModuleLocked(curriculum, lIdx, mIdx, completed, overrides)}
              progress={moduleProgressPercent(module, completed)}
              {isLangC}
              cardIndex={mIdx}
              {completed}
            />
          {/each}
        </div>
      </section>
    {/each}
  </div>
</div>
