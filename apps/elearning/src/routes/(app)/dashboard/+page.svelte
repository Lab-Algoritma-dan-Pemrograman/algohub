<script lang="ts">
  import { goto } from '$app/navigation';
  import { resolve } from '$app/paths';
  import type { RouteId } from '$app/types';
  import BookOpenCheck from '@lucide/svelte/icons/book-open-check';
  import Clock from '@lucide/svelte/icons/clock';
  import Flame from '@lucide/svelte/icons/flame';
  import Play from '@lucide/svelte/icons/play';
  import Trophy from '@lucide/svelte/icons/trophy';
  import Zap from '@lucide/svelte/icons/zap';
  import { createQuery } from '@tanstack/svelte-query';
  import { fade } from 'svelte/transition';
  import { api } from '#lib/api.ts';
  import LevelCard from '#lib/components/LevelCard.svelte';
  import {
    buildCurriculum,
    isLevelLocked,
    levelProgressPercent,
    lessonIds,
    overridesOf,
    pickContinueLesson,
    type Checkpoint,
  } from '#lib/curriculum.ts';
  import { getLevelLanguage } from '#lib/levelLanguage.ts';
  import { formatStudyTime } from '#lib/profile.ts';
  import { levelsQuery, lessonsQuery, modulesQuery } from '#lib/queries/curriculum.ts';
  import { completedLessonsQuery, recentProgressQuery } from '#lib/queries/progress.ts';
  import { isAdminOrKordas } from '#lib/roles.ts';
  import { appStore } from '#lib/stores/app.svelte.ts';
  import type { PageProps } from './$types';

  let { data }: PageProps = $props();
  const user = $derived(data.auth.user);

  const levels = createQuery(levelsQuery);
  const modules = createQuery(modulesQuery);
  const lessons = createQuery(lessonsQuery);
  const completedQuery = createQuery(completedLessonsQuery);
  const recentQuery = createQuery(() => recentProgressQuery(5));

  const profile = $derived(appStore.profile);
  const curriculum = $derived(
    buildCurriculum(levels.data ?? [], modules.data ?? [], lessons.data ?? []),
  );
  const completed = $derived(new Set(completedQuery.data ?? []));
  const overrides = $derived(overridesOf(profile?.levelAccessOverrides));
  const locked = $derived(
    curriculum.map((_, i) => isLevelLocked(curriculum, i, completed, overrides)),
  );
  const totalLessons = $derived(
    curriculum.reduce((sum, level, i) => sum + (locked[i] ? 0 : lessonIds(level).length), 0),
  );
  const completedCount = $derived(completed.size);
  const percent = $derived(
    totalLessons > 0 ? Math.round((completedCount / totalLessons) * 100) : 0,
  );
  // Bahasa ditentukan dari judul level; indeks asli kurikulum dipertahankan untuk fallback dan kunci.
  const entries = $derived(
    curriculum.map((level, index) => ({ level, index, language: getLevelLanguage(level, index) })),
  );
  const cLevels = $derived(entries.filter((e) => e.language === 'c'));
  const pyLevels = $derived(entries.filter((e) => e.language === 'python'));
  const cProgress = $derived(
    levelProgressPercent(
      cLevels.map((e) => e.level),
      curriculum,
      completed,
      overrides,
    ),
  );
  const pyProgress = $derived(
    levelProgressPercent(
      pyLevels.map((e) => e.level),
      curriculum,
      completed,
      overrides,
    ),
  );
  const lessonTitle = $derived(
    new Map((lessons.data ?? []).map((lesson) => [lesson.id, lesson.title])),
  );
  const emptyForAdmin = $derived(
    !levels.isPending && curriculum.length === 0 && isAdminOrKordas(user.role),
  );

  // Menangkap achievement XP yang terlewat (mis. dari Bug Hunt) saat halaman dibuka.
  let xpChecked = false;
  $effect(() => {
    const xp = profile?.xp;
    if (xpChecked || !xp) return;
    xpChecked = true;
    api.achievements
      .checkXp(xp)
      .then((res) => {
        if (res.achievement) appStore.pushAchievement(res.achievement);
      })
      .catch(() => undefined);
  });

  function readCheckpoint(): Checkpoint | null {
    try {
      const raw = localStorage.getItem('last-lesson-checkpoint');
      return raw ? (JSON.parse(raw) as Checkpoint) : null;
    } catch {
      return null;
    }
  }

  function continueLearning() {
    const target = pickContinueLesson(curriculum, completed, overrides, readCheckpoint());
    if (!target) return;
    if (target.resetStep) localStorage.setItem(`lesson-step:${target.lessonId}`, 'learn');
    void goto(resolve(`/lesson/${target.lessonId}` as RouteId));
  }

  const formatDate = (iso: string) =>
    new Date(iso).toLocaleDateString('id-ID', {
      weekday: 'long',
      day: 'numeric',
      month: 'long',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
</script>

<div class="space-y-8" in:fade={{ duration: 200 }}>
  {#if emptyForAdmin}
    <div
      class="space-y-4 rounded-3xl border-2 border-dashed border-rose-200 bg-rose-50 p-8 text-center"
    >
      <div
        class="mx-auto flex h-16 w-16 items-center justify-center rounded-2xl bg-rose-100 text-rose-700"
      >
        <Zap size={32} />
      </div>
      <div class="mx-auto max-w-md">
        <h2 class="text-xl font-bold text-rose-900">Kurikulum masih kosong</h2>
        <p class="mt-2 text-sm text-rose-700">
          Isi kurikulum E-Learning lewat Admin Panel supaya mahasiswa bisa mulai belajar.
        </p>
      </div>
      <a
        href={resolve('/admin' as RouteId)}
        class="inline-block rounded-xl bg-rose-700 px-8 py-3 font-bold text-white shadow-lg shadow-rose-700/20 transition-all hover:bg-rose-800 active:scale-95"
        >Buka Admin Panel</a
      >
    </div>
  {/if}

  <div class="flex flex-col justify-between gap-6 md:flex-row md:items-center">
    <div>
      <h1
        class="text-dark text-3xl leading-tight font-black tracking-tight select-none md:text-4xl"
      >
        Selamat datang kembali, {(profile?.nama ?? user.nama).split(' ')[0] || 'Penjelajah'}!
      </h1>
      <p class="mt-1 font-bold text-zinc-500">Kamu membuat kemajuan besar. Pertahankan!</p>
    </div>
    <div
      class="flex w-full flex-wrap items-center justify-start gap-3 md:ml-auto md:w-auto md:flex-nowrap md:justify-end"
    >
      <div
        class="flex items-center gap-3 rounded-2xl border-2 border-amber-200 bg-white px-5 py-3 shadow-sm select-none"
      >
        <div
          class="bg-fun-yellow text-maroon flex h-10 w-10 items-center justify-center rounded-xl shadow-sm"
        >
          <Flame size={18} />
        </div>
        <div>
          <div class="text-[10px] leading-none font-black tracking-widest text-amber-600 uppercase">
            Beruntun
          </div>
          <div class="mt-1 text-lg font-black text-amber-950">
            {profile?.streak ?? 0} <span class="text-xs font-bold text-amber-600">hari</span>
          </div>
        </div>
      </div>
      <div
        class="flex items-center gap-3 rounded-2xl border-2 border-rose-200 bg-white px-5 py-3 shadow-sm select-none"
      >
        <div
          class="bg-maroon flex h-10 w-10 items-center justify-center rounded-xl text-white shadow-sm"
        >
          <Trophy size={16} />
        </div>
        <div>
          <div class="text-[10px] leading-none font-black tracking-widest text-rose-600 uppercase">
            Total XP
          </div>
          <div class="mt-1 text-lg font-black text-rose-950">
            {(profile?.xp ?? 0).toLocaleString('id-ID')}
          </div>
        </div>
      </div>
      <div
        class="flex items-center gap-3 rounded-2xl border-2 border-blue-200 bg-white px-5 py-3 shadow-sm select-none"
      >
        <div
          class="bg-fun-blue flex h-10 w-10 items-center justify-center rounded-xl text-white shadow-sm"
        >
          <Clock size={16} />
        </div>
        <div>
          <div class="text-[10px] leading-none font-black tracking-widest text-blue-600 uppercase">
            Waktu Belajar
          </div>
          <div class="mt-1 text-lg font-black text-blue-950">
            {formatStudyTime(profile?.studyTime ?? 0)}
          </div>
        </div>
      </div>
    </div>
  </div>

  <div class="grid grid-cols-1 gap-8 lg:grid-cols-3">
    <div class="space-y-8 lg:col-span-2">
      <div
        class="shadow-soft card-hover hover:border-maroon/20 group relative flex flex-col overflow-hidden rounded-[2rem] border-2 border-gray-100 bg-white p-8 select-none md:p-10"
      >
        <div class="relative z-10">
          <div
            class="mb-4 inline-flex items-center gap-2 rounded-full border border-rose-200/60 bg-rose-50 px-3 py-1.5 text-xs font-black tracking-[0.2em] text-rose-700 uppercase"
          >
            <Play size={9} /> Lanjutkan Belajar
          </div>
          <h2 class="text-dark mb-3 text-3xl leading-tight font-black md:text-4xl">
            {curriculum[0]?.title || 'Belajar kamu'}
          </h2>
          <p class="mb-8 max-w-lg leading-relaxed font-bold text-zinc-500">
            Kuasai materi yang telah disusun. Kamu telah menyelesaikan
            <span class="text-maroon font-black">{completedCount}</span> dari
            <span class="text-dark font-black">{totalLessons}</span> pelajaran.
          </p>
          <div class="flex flex-col items-start gap-6 sm:flex-row sm:items-center">
            <button
              onclick={continueLearning}
              disabled={curriculum.length === 0}
              class="bg-maroon hover:bg-maroon-light shadow-bubbly-maroon btn-bubbly flex cursor-pointer items-center gap-3 rounded-2xl px-8 py-4 font-black text-white transition-all disabled:cursor-not-allowed disabled:opacity-60"
            >
              <Play size={14} />
              Mulai Belajar
            </button>
            <div class="flex flex-col gap-4">
              {#each [{ label: 'Bahasa C', value: cProgress, show: cLevels.length > 0, bar: 'from-blue-500 to-indigo-500', dot: 'bg-fun-blue' }, { label: 'Python', value: pyProgress, show: pyLevels.length > 0, bar: 'from-rose-500 to-pink-500', dot: 'bg-rose-500' }] as lang (lang.label)}
                {#if lang.show}
                  <div class="flex flex-col gap-1.5">
                    <div
                      class="flex items-center gap-1.5 text-[10px] font-black tracking-wider text-zinc-500 uppercase"
                    >
                      <div class="h-2.5 w-2.5 rounded-full {lang.dot}"></div>
                      {lang.label}
                    </div>
                    <div class="flex items-center gap-3">
                      <div
                        class="h-3 w-36 overflow-hidden rounded-full border border-zinc-200 bg-zinc-100 shadow-inner"
                      >
                        <div
                          class="h-full rounded-full bg-gradient-to-r {lang.bar}"
                          style:width="{lang.value}%"
                        ></div>
                      </div>
                      <span class="font-mono text-xs font-black text-zinc-600">{lang.value}%</span>
                    </div>
                  </div>
                {/if}
              {/each}
            </div>
          </div>
        </div>
      </div>

      <div class="space-y-12">
        {#each [{ title: 'Kurikulum Bahasa C', badge: 'C', list: cLevels, color: 'blue', badgeClass: 'bg-fun-blue shadow-[0_4px_0_#023E8A] rotate-[-6deg]' }, { title: 'Kurikulum Python', badge: 'Py', list: pyLevels, color: 'rose', badgeClass: 'bg-rose-500 shadow-[0_4px_0_#9F1239] rotate-[6deg]' }] as section (section.title)}
          {#if section.list.length > 0}
            <div class="space-y-4">
              <h3 class="flex items-center gap-3 text-xl font-black select-none">
                <div
                  class="flex h-10 w-10 items-center justify-center rounded-xl font-black text-white {section.badgeClass}"
                >
                  {section.badge}
                </div>
                <div>
                  <span class="text-dark">{section.title}</span>
                  <div class="mt-0.5 text-xs font-bold text-zinc-400">
                    {section.list.length} Level
                  </div>
                </div>
              </h3>
              <div class="grid grid-cols-1 gap-6 md:grid-cols-2">
                {#each section.list as entry (entry.level.id)}
                  <LevelCard
                    level={entry.level}
                    index={entry.index}
                    locked={locked[entry.index]}
                    {completed}
                    color={section.color as 'blue' | 'rose'}
                  />
                {/each}
              </div>
            </div>
          {/if}
        {/each}
      </div>
    </div>

    <div class="space-y-8">
      <div
        class="shadow-soft card-hover hover:border-maroon/20 rounded-[2rem] border-2 border-gray-100 bg-white p-6"
      >
        <h3 class="mb-5 flex items-center gap-3 text-lg font-black select-none">
          <div
            class="flex h-10 w-10 rotate-[-4deg] items-center justify-center rounded-xl bg-gradient-to-br from-rose-500 to-pink-600 text-white shadow-[0_3px_0_#5C0E25]"
          >
            <Trophy size={14} />
          </div>
          Progres Keseluruhan
        </h3>
        <div class="space-y-4">
          <div class="flex items-end justify-between">
            <div>
              <div class="text-4xl font-black text-zinc-900">{percent}%</div>
              <div class="mt-1 text-xs font-bold text-zinc-500">
                {completedCount} / {totalLessons} pelajaran
              </div>
            </div>
            <div class="text-right">
              <div class="text-xs font-bold text-zinc-400">Total XP</div>
              <div class="text-xl font-black text-rose-700">
                {(profile?.xp ?? 0).toLocaleString('id-ID')}
              </div>
            </div>
          </div>
          <div
            class="h-3 w-full overflow-hidden rounded-full border border-zinc-200 bg-zinc-100 shadow-inner"
          >
            <div
              class="h-full rounded-full bg-gradient-to-r from-rose-600 to-pink-500 transition-[width] duration-1000 ease-out"
              style:width="{percent}%"
            ></div>
          </div>
        </div>
      </div>

      <div
        class="shadow-soft card-hover hover:border-maroon/20 rounded-[2rem] border-2 border-gray-100 bg-white p-6"
      >
        <h3 class="mb-6 flex items-center gap-3 text-lg font-black select-none">
          <div
            class="flex h-10 w-10 rotate-[4deg] items-center justify-center rounded-xl bg-gradient-to-br from-amber-400 to-orange-500 text-white shadow-[0_3px_0_#D97706]"
          >
            <Clock size={14} />
          </div>
          Aktivitas Terbaru
        </h3>
        <div class="space-y-5">
          {#each recentQuery.data ?? [] as activity, i (activity.id)}
            <div class="relative flex items-start gap-4">
              <div class="flex flex-col items-center">
                <div
                  class="mt-1 h-3 w-3 shrink-0 rounded-full bg-gradient-to-br from-rose-500 to-pink-600 shadow-sm"
                ></div>
                {#if i < (recentQuery.data?.length ?? 0) - 1}
                  <div class="mt-1 w-px flex-1 bg-zinc-200"></div>
                {/if}
              </div>
              <div class="min-w-0 flex-1 pb-4">
                <div class="truncate text-sm font-bold">
                  {lessonTitle.has(activity.lesson_id)
                    ? `${lessonTitle.get(activity.lesson_id)} - Latihan`
                    : activity.lesson_id}
                </div>
                {#if activity.completed_at}
                  <div class="mt-0.5 text-xs font-medium text-zinc-400">
                    {formatDate(activity.completed_at)}
                  </div>
                {/if}
              </div>
            </div>
          {:else}
            <div class="py-8 text-center">
              <div
                class="mx-auto mb-3 flex h-16 w-16 items-center justify-center rounded-2xl border border-zinc-100 bg-zinc-50 shadow-inner"
              >
                <BookOpenCheck size={22} class="text-zinc-400" />
              </div>
              <p class="text-sm font-bold text-zinc-500">Belum ada aktivitas</p>
              <p class="mt-0.5 text-xs text-zinc-400">Mulai pelajaran pertama kamu!</p>
            </div>
          {/each}
        </div>
      </div>

      <div
        class="relative overflow-hidden rounded-[2rem] border-2 border-rose-800 bg-gradient-to-br from-rose-700 via-rose-800 to-pink-900 p-6 text-white shadow-[0_6px_0_#5C0E25]"
      >
        <div class="relative z-10">
          <h3 class="mb-2 text-lg font-black">Tips Pro!</h3>
          <p class="text-sm leading-relaxed font-bold text-rose-100">
            Konsistensi adalah kunci. Cobalah untuk memprogram setidaknya 15 menit setiap hari untuk
            membangun memori otot.
          </p>
        </div>
      </div>
    </div>
  </div>
</div>
