<script lang="ts">
  import { elearningKeys } from '@algohub/api-client/elearning';
  import Flame from '@lucide/svelte/icons/flame';
  import Gem from '@lucide/svelte/icons/gem';
  import RefreshCw from '@lucide/svelte/icons/refresh-cw';
  import Star from '@lucide/svelte/icons/star';
  import Trophy from '@lucide/svelte/icons/trophy';
  import { createMutation, createQuery } from '@tanstack/svelte-query';
  import { onMount } from 'svelte';
  import { fade } from 'svelte/transition';
  import { themeFor } from '#lib/achievement-theme.ts';
  import {
    lessonsToNextLevel,
    prestigeErrorMessage,
    semesterLabelJakarta,
    XP_PER_LEVEL,
    xpProgress,
  } from '#lib/gamification.ts';
  import { loadProfile } from '#lib/profile.ts';
  import {
    achievementCatalogQuery,
    myAchievementsQuery,
    unlockedIds,
  } from '#lib/queries/achievements.ts';
  import { prestigeEligibilityQuery, prestigeResetOptions } from '#lib/queries/prestige.ts';
  import { completedLessonsQuery } from '#lib/queries/progress.ts';
  import { queryClient } from '#lib/query.ts';
  import { subscribe } from '#lib/realtime/client.ts';
  import { appStore } from '#lib/stores/app.svelte.ts';
  import type { PageProps } from './$types';

  let { data }: PageProps = $props();
  const user = $derived(data.auth.user);
  const profile = $derived(appStore.profile);

  const catalog = createQuery(achievementCatalogQuery);
  const mine = createQuery(myAchievementsQuery);
  const eligibility = createQuery(() => prestigeEligibilityQuery());
  const completedQuery = createQuery(completedLessonsQuery);
  const reset = createMutation(() => prestigeResetOptions());

  const unlocked = $derived(unlockedIds(mine.data));
  const completedCount = $derived(completedQuery.data?.length ?? 0);
  const xp = $derived(profile?.xp ?? 0);
  const level = $derived(profile?.level ?? 1);
  const progress = $derived(xpProgress(xp));
  const prestige = $derived(eligibility.data);
  const joined = $derived(
    profile?.createdAt
      ? new Date(profile.createdAt).toLocaleDateString('id-ID', { month: 'long', year: 'numeric' })
      : '',
  );

  let confirming = $state(false);
  let message = $state('');
  let syncing = $state(false);

  // Perubahan di tabel milik user ini (achievement terbuka, progres) diambil ulang tanpa reload.
  onMount(() => {
    const filter = `nim=eq.${user.nim}`;
    const stopAchievements = subscribe(
      `profile-achievements-${user.nim}`,
      {
        table: 'unlocked_achievements',
        filter,
      },
      () => {
        void queryClient.invalidateQueries({ queryKey: elearningKeys.achievements.mine() });
      },
    );
    const stopProgress = subscribe(
      `profile-progress-${user.nim}`,
      {
        table: 'student_progress',
        filter,
      },
      () => {
        void queryClient.invalidateQueries({ queryKey: elearningKeys.progress.lessons() });
        void queryClient.invalidateQueries({ queryKey: elearningKeys.prestige.eligibility() });
      },
    );
    return () => {
      stopAchievements();
      stopProgress();
    };
  });

  function doReset() {
    message = '';
    reset.mutate(undefined, {
      onSuccess: (res) => {
        message = `Prestige #${res.prestigeNo} aktif: tier ${res.tier} x${res.mult.toFixed(2)} untuk lesson. Progres lama diarsip, papan mingguan tetap base.`;
        confirming = false;
      },
      onError: (error) => {
        message = prestigeErrorMessage(error);
        confirming = false;
      },
    });
  }

  async function sync() {
    syncing = true;
    try {
      await loadProfile();
    } catch {
      message = 'Gagal menyinkronkan profil.';
    } finally {
      syncing = false;
    }
  }
</script>

<div class="mx-auto max-w-5xl space-y-12 pb-12" in:fade={{ duration: 200 }}>
  <div
    class="shadow-soft hover:border-maroon/20 flex flex-col items-center gap-8 rounded-[3rem] border-2 border-gray-100 bg-white p-8 transition-all md:flex-row"
  >
    <div
      class="flex h-32 w-32 rotate-[-4deg] items-center justify-center rounded-full border-4 border-white bg-zinc-100 text-4xl font-black text-zinc-300 shadow-xl select-none"
    >
      {(profile?.nama || user.nama || 'U').charAt(0)}
    </div>
    <div class="flex-1 space-y-2 text-center md:text-left">
      <h1 class="text-3xl font-black tracking-tight text-zinc-900 select-none">
        {profile?.nama || user.nama || 'User'}
      </h1>
      <p class="font-bold text-zinc-500">
        Penjelajah{joined ? ` • Bergabung ${joined}` : ''}
      </p>
      <div class="flex flex-wrap justify-center gap-4 pt-4 md:justify-start">
        <div
          class="flex items-center gap-2 rounded-full border-2 border-amber-200 bg-amber-50 px-4 py-2 text-xs font-black tracking-wider text-amber-700 uppercase shadow-[0_3px_0_#D97706]"
        >
          <Flame size={14} class="text-amber-600" />
          {profile?.streak ?? 0} Hari Beruntun
        </div>
        <div
          class="flex items-center gap-2 rounded-full border-2 border-rose-200 bg-rose-50 px-4 py-2 text-xs font-black tracking-wider text-rose-700 uppercase shadow-[0_3px_0_#8A1538]"
        >
          <Trophy size={14} />
          Level {level}
        </div>
        <div
          class="flex items-center gap-2 rounded-full border-2 border-blue-200 bg-blue-50 px-4 py-2 text-xs font-black tracking-wider text-blue-700 uppercase shadow-[0_3px_0_#023E8A]"
        >
          <Star size={14} class="text-blue-600" />
          {xp.toLocaleString('id-ID')} XP
        </div>
        {#if profile?.prestigeTier}
          <div
            class="flex items-center gap-2 rounded-full border-2 border-purple-200 bg-purple-50 px-4 py-2 text-xs font-black tracking-wider text-purple-700 uppercase shadow-[0_3px_0_#7C3AED]"
          >
            <Gem size={14} class="text-purple-600" />
            {profile.prestigeTier} x{Number(profile.prestigeMult || 1).toFixed(2)}
          </div>
        {/if}
      </div>
    </div>
    <button
      onclick={sync}
      disabled={syncing}
      class="hover:text-maroon flex cursor-pointer items-center gap-2 self-start rounded-2xl p-3.5 text-sm font-black text-zinc-400 transition-all hover:bg-rose-50 disabled:opacity-50"
      aria-label="Sinkronkan profil"
    >
      <RefreshCw size={18} class={syncing ? 'animate-spin' : ''} />
      <span class="hidden sm:inline">Sinkronkan</span>
    </button>
  </div>

  <div class="grid grid-cols-1 gap-8 lg:grid-cols-3">
    <div class="space-y-6 lg:col-span-2">
      <div class="flex flex-wrap items-center justify-between gap-2 select-none">
        <h3 class="flex items-center gap-3 text-xl font-black">
          <div
            class="flex h-10 w-10 items-center justify-center rounded-xl bg-gradient-to-br from-rose-500 to-pink-600 text-white shadow-[0_3px_0_#5C0E25]"
          >
            <Trophy size={16} />
          </div>
          Pencapaian
        </h3>
        <div
          class="text-maroon rounded-full border border-rose-200 bg-rose-50 px-3.5 py-1.5 text-xs font-black shadow-sm"
        >
          {unlocked.size} / {catalog.data?.length ?? 0} Terbuka
        </div>
      </div>
      <div class="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4">
        {#each catalog.data ?? [] as achievement (achievement.id)}
          {@const theme = themeFor(achievement.icon)}
          {@const Icon = theme.icon}
          {@const open = unlocked.has(achievement.id)}
          <div
            class="flex flex-col items-center justify-center space-y-3 rounded-[2rem] border-2 p-5 text-center transition-all select-none {open
              ? 'shadow-soft hover:border-maroon/20 border-gray-100 bg-white'
              : 'border-zinc-200 bg-zinc-50 opacity-40 grayscale'}"
          >
            <div
              class="mb-1 flex h-14 w-14 items-center justify-center rounded-2xl {open
                ? `bg-gradient-to-br ${theme.bg} rotate-[-5deg] text-white shadow-[0_4px_0_#5C0E25]`
                : 'border border-zinc-300 bg-zinc-200 text-zinc-400 shadow-inner'}"
            >
              <Icon size={22} />
            </div>
            <div>
              <div class="mb-1 text-[11px] leading-tight font-black text-zinc-900">
                {achievement.title}
              </div>
              <div class="line-clamp-2 text-[9px] leading-tight font-bold text-zinc-500">
                {achievement.description}
              </div>
            </div>
          </div>
        {/each}
      </div>
    </div>

    <div class="space-y-8">
      <div
        class="shadow-soft hover:border-maroon/20 rounded-[2rem] border-2 border-gray-100 bg-white p-6 transition-all"
      >
        <h3 class="mb-4 flex items-center gap-3 text-lg font-black select-none">
          <div
            class="flex h-10 w-10 rotate-[-4deg] items-center justify-center rounded-xl bg-gradient-to-br from-purple-500 to-indigo-600 text-white shadow-[0_3px_0_#4C1D95]"
          >
            <Gem size={16} />
          </div>
          Prestige Sukarela
        </h3>
        {#if prestige}
          <div class="space-y-3 text-sm">
            <p class="leading-relaxed font-bold text-zinc-600">{prestige.reason}</p>
            <div class="space-y-1 text-xs font-black tracking-wider text-zinc-500 uppercase">
              <p>
                Tamat {prestige.levelsDone}/{prestige.totalLevels} level • {prestige.lessonsDone}/{prestige.totalLessons}
                lesson
              </p>
              <p>
                Semester {semesterLabelJakarta()} • dipakai {prestige.semesterUsed}/2 • cooldown 14
                hari
              </p>
              {#if prestige.eligible}
                <p>Multiplier tunggal Master x1.50 lesson-only • papan mingguan tetap base</p>
              {:else}
                <p>Multiplier x1.50 hanya aktif setelah tamat penuh 6/6</p>
              {/if}
            </div>
            {#if message}
              <p
                class="rounded-xl border border-zinc-200 bg-zinc-50 p-3 text-xs font-bold text-zinc-700"
                role="status"
              >
                {message}
              </p>
            {/if}
            {#if !confirming}
              <button
                onclick={() => {
                  if (prestige.eligible) confirming = true;
                }}
                disabled={!prestige.eligible || reset.isPending}
                class="w-full rounded-2xl border-2 px-4 py-3 text-sm font-black transition-all {prestige.eligible
                  ? 'cursor-pointer border-purple-700 bg-purple-600 text-white shadow-[0_3px_0_#4C1D95] active:translate-y-[3px] active:shadow-none'
                  : 'cursor-not-allowed border-zinc-200 bg-zinc-100 text-zinc-400'}"
              >
                {reset.isPending
                  ? 'Memproses...'
                  : prestige.eligible
                    ? `Reset & Ambil x${prestige.mult.toFixed(2)}`
                    : 'Belum Layak Prestige'}
              </button>
            {:else}
              <div class="space-y-2 rounded-2xl border-2 border-red-200 bg-red-50 p-4">
                <p class="text-xs leading-relaxed font-black text-red-700">
                  Yakin? Progres aktif diarsip lalu dihapus, XP kembali 0, multiplier lesson x{prestige.mult.toFixed(
                    2,
                  )}. Papan mingguan tetap base. Tidak bisa dibatalkan.
                </p>
                <div class="flex gap-2">
                  <button
                    onclick={doReset}
                    disabled={reset.isPending}
                    class="flex-1 cursor-pointer rounded-xl border-2 border-red-700 bg-red-600 px-4 py-2.5 text-sm font-black text-white disabled:opacity-50"
                  >
                    {reset.isPending ? '...' : 'Ya, Reset'}
                  </button>
                  <button
                    onclick={() => (confirming = false)}
                    class="flex-1 cursor-pointer rounded-xl border-2 border-zinc-200 bg-white px-4 py-2.5 text-sm font-black text-zinc-600"
                  >
                    Batal
                  </button>
                </div>
              </div>
            {/if}
          </div>
        {:else if eligibility.isError}
          <p class="text-sm font-bold text-zinc-400">Gagal memuat status prestige.</p>
        {:else}
          <p class="text-sm font-bold text-zinc-400">Memuat status prestige...</p>
        {/if}
      </div>

      <div
        class="shadow-soft hover:border-maroon/20 rounded-[2rem] border-2 border-gray-100 bg-white p-6 transition-all"
      >
        <h3 class="mb-6 flex items-center gap-3 text-lg font-black select-none">
          <div
            class="flex h-10 w-10 rotate-[-4deg] items-center justify-center rounded-xl bg-gradient-to-br from-amber-400 to-orange-500 text-white shadow-[0_3px_0_#D97706]"
          >
            <Star size={16} />
          </div>
          Statistik Belajar
        </h3>
        <div class="space-y-4">
          {#each [{ label: 'Pelajaran Selesai', value: String(completedCount) }, { label: 'Kuis Lulus', value: String(completedCount) }, { label: 'Latihan Terpecahkan', value: String(completedCount) }, { label: 'Total Waktu', value: `${Math.round((completedCount * 15) / 60)} jam` }] as stat (stat.label)}
            <div
              class="flex items-center justify-between border-b-2 border-dashed border-gray-100 py-3.5 last:border-0"
            >
              <span class="text-sm font-bold text-zinc-500">{stat.label}</span>
              <span class="font-mono text-base font-black text-zinc-900">{stat.value}</span>
            </div>
          {/each}
        </div>
      </div>

      <div
        class="relative overflow-hidden rounded-[2rem] border-2 border-zinc-950 bg-zinc-950 p-6 text-white shadow-[0_6px_0_#000]"
      >
        <h3 class="mb-4 flex items-center gap-2 text-lg font-black select-none">
          🎯 Target Berikutnya
        </h3>
        <div class="space-y-4">
          <div class="mb-1 flex justify-between text-sm font-bold">
            <span class="text-zinc-400">Level {level + 1}</span>
            <span class="font-mono font-black">{progress.current} / {XP_PER_LEVEL} XP</span>
          </div>
          <div
            class="h-2.5 w-full overflow-hidden rounded-full border border-zinc-900 bg-zinc-800 shadow-inner"
          >
            <div
              class="h-full rounded-full bg-rose-600 transition-all duration-500"
              style:width="{progress.percent}%"
            ></div>
          </div>
          <p class="text-xs leading-relaxed font-bold text-zinc-400">
            Selesaikan {lessonsToNextLevel(xp)} pelajaran lagi untuk mencapai Level {level + 1} dan membuka
            modul baru!
          </p>
        </div>
      </div>
    </div>
  </div>
</div>
