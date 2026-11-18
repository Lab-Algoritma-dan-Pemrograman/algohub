<script lang="ts">
  import Minus from '@lucide/svelte/icons/minus';
  import RotateCcw from '@lucide/svelte/icons/rotate-ccw';
  import Trophy from '@lucide/svelte/icons/trophy';
  import Medal from '@lucide/svelte/icons/medal';
  import Users from '@lucide/svelte/icons/users';
  import { createQuery } from '@tanstack/svelte-query';
  import { fade } from 'svelte/transition';
  import {
    DEFAULT_LEADERBOARD_LIMIT,
    jakartaWeekLabel,
    leaderboardFilter,
    leaderboardParams,
    studentWeeklyParams,
  } from '#lib/gamification.ts';
  import {
    leaderboardFiltersQuery,
    leaderboardListQuery,
    leaderboardRankQuery,
    weeklyLeaderboardQuery,
  } from '#lib/queries/leaderboard.ts';
  import { isStaff } from '#lib/roles.ts';
  import { appStore } from '#lib/stores/app.svelte.ts';
  import type { PageProps } from './$types';

  interface Row {
    nim: string;
    nama: string | null;
    level: number;
    xp: number;
    carry: boolean;
  }

  let { data }: PageProps = $props();
  const user = $derived(data.auth.user);
  const staff = $derived(isStaff(user.role));
  const student = $derived(!staff);

  // Mahasiswa selalu melihat papan mingguan; staf memilih Mingguan atau Total dan boleh memfilter.
  let period = $state<'weekly' | 'total'>('weekly');
  let kelas = $state('');
  let jurusan = $state('');
  let shift = $state('');
  let limit = $state(DEFAULT_LEADERBOARD_LIMIT);

  const weeklyMode = $derived(student || period === 'weekly');
  const view = $derived({ kelas, jurusan, shift, limit });
  const filter = $derived(leaderboardFilter(view));
  const hasActiveFilter = $derived(
    kelas !== '' || jurusan !== '' || shift !== '' || limit !== DEFAULT_LEADERBOARD_LIMIT,
  );
  const weekParams = $derived(student ? studentWeeklyParams(user) : leaderboardParams(view));
  const rankParams = $derived(
    student ? { ...studentWeeklyParams(user), limit: 0 } : { ...leaderboardParams(view), limit: 0 },
  );

  const options = createQuery(() => ({ ...leaderboardFiltersQuery(), enabled: staff }));
  const weekly = createQuery(() => ({
    ...weeklyLeaderboardQuery(weekParams),
    enabled: weeklyMode,
  }));
  const weeklyAll = createQuery(() => ({
    ...weeklyLeaderboardQuery(rankParams),
    enabled: weeklyMode,
  }));
  const total = createQuery(() => ({
    ...leaderboardListQuery(leaderboardParams(view)),
    enabled: !weeklyMode,
  }));
  const totalRank = createQuery(() => ({
    ...leaderboardRankQuery(appStore.profile?.xp ?? 0, filter),
    enabled: !weeklyMode,
  }));

  const rows = $derived<Row[]>(
    weeklyMode
      ? (weekly.data?.items ?? []).map((e) => ({
          nim: e.nim,
          nama: e.nama,
          level: e.level,
          xp: e.weeklyXp,
          carry: e.isCarry,
        }))
      : (total.data?.items ?? []).map((e) => ({
          nim: e.nim,
          nama: e.nama,
          level: e.level,
          xp: e.xp,
          carry: false,
        })),
  );
  const loading = $derived(weeklyMode ? weekly.isPending : total.isPending);

  const myWeekly = $derived((weeklyAll.data?.items ?? []).find((e) => e.nim === user.nim));
  const myWeeklyXp = $derived(myWeekly?.weeklyXp ?? 0);
  const userRank = $derived.by(() => {
    if (!weeklyMode) return totalRank.data?.rank ?? null;
    if (!myWeekly || myWeekly.weeklyXp <= 0) return null;
    return (weeklyAll.data?.items ?? []).filter((e) => e.weeklyXp > myWeekly.weeklyXp).length + 1;
  });
  const isUserInTop = $derived(rows.some((row) => row.nim === user.nim));
  const profile = $derived(appStore.profile);
  const week = jakartaWeekLabel();

  function resetFilters() {
    kelas = '';
    jurusan = '';
    shift = '';
    limit = DEFAULT_LEADERBOARD_LIMIT;
  }

  const podium = $derived(
    [
      {
        index: 1,
        rank: 2,
        height: 'h-40 sm:h-48',
        color: 'bg-zinc-100',
        border: 'border-zinc-300',
        shadow: '#71717a',
        medal: 'text-zinc-400',
      },
      {
        index: 0,
        rank: 1,
        height: 'h-56 sm:h-64',
        color: 'bg-amber-50',
        border: 'border-amber-200',
        shadow: '#D97706',
        medal: 'text-amber-500',
      },
      {
        index: 2,
        rank: 3,
        height: 'h-32 sm:h-40',
        color: 'bg-orange-50',
        border: 'border-orange-200',
        shadow: '#C2410C',
        medal: 'text-orange-500',
      },
    ].filter((slot) => rows[slot.index]),
  );
  const selectClass =
    'cursor-pointer rounded-xl border border-zinc-200 bg-zinc-50 px-4 py-2.5 text-sm font-black shadow-inner focus:ring-2 focus:ring-rose-700/20 focus:outline-none';
  const tabClass = (active: boolean) =>
    `cursor-pointer rounded-xl border-2 px-6 py-2.5 text-sm font-black transition-all ${
      active
        ? 'bg-maroon border-maroon text-white shadow-[0_3px_0_rgba(0,0,0,0.2)]'
        : 'hover:border-maroon/30 border-zinc-200 bg-white text-zinc-500'
    }`;
</script>

<div class="mx-auto max-w-4xl space-y-8 pb-12" in:fade={{ duration: 200 }}>
  <div class="space-y-3 text-center">
    <h1
      class="text-maroon mb-2 flex items-center justify-center text-4xl font-black tracking-tight select-none md:text-5xl"
    >
      {weeklyMode ? 'Papan Peringkat Mingguan' : 'Papan Peringkat Global'}
      <Trophy class="text-fun-yellow ml-4 animate-bounce" size={32} />
    </h1>
    <p class="text-zinc-550 mx-auto max-w-xl leading-relaxed font-bold">
      {#if student}
        Top 3 Penjelajah di Kelas {user.kelas}{user.shift ? ` - Shift ${user.shift}` : ''} - {user.jurusan ??
          ''}
      {:else if weeklyMode}
        Peringkat XP minggu berjalan. Reset tiap Senin 00:00 WIB.
      {:else}
        Lihat peringkat semua praktikan berdasarkan total XP.
      {/if}
    </p>
    {#if weeklyMode}
      <p
        class="inline-block rounded-full border border-amber-200 bg-amber-50 px-4 py-1.5 text-xs font-black tracking-wide text-amber-700"
      >
        Periode {week} — reset Senin 00:00 WIB
      </p>
    {/if}
    {#if student && myWeeklyXp > 0}
      <p class="text-sm font-black text-rose-700">
        XP minggumu: {myWeeklyXp.toLocaleString('id-ID')} XP
      </p>
    {/if}
  </div>

  {#if staff}
    <div class="flex justify-center gap-2">
      <button onclick={() => (period = 'weekly')} class={tabClass(period === 'weekly')}
        >Mingguan</button
      >
      <button onclick={() => (period = 'total')} class={tabClass(period === 'total')}>Total</button>
    </div>
    <div class="shadow-soft rounded-3xl border-2 border-gray-100 bg-white p-5">
      <div class="flex flex-wrap items-center gap-3">
        <select bind:value={kelas} class={selectClass} aria-label="Filter kelas">
          <option value="">Semua Kelas</option>
          {#each options.data?.kelas ?? [] as k (k)}<option value={k}>{k}</option>{/each}
        </select>
        <select bind:value={jurusan} class={selectClass} aria-label="Filter jurusan">
          <option value="">Semua Jurusan</option>
          {#each options.data?.jurusan ?? [] as j (j)}<option value={j}>{j}</option>{/each}
        </select>
        <select bind:value={shift} class={selectClass} aria-label="Filter shift">
          <option value="">Semua Shift</option>
          {#each options.data?.shift ?? [] as s (s)}<option value={s}>Shift {s}</option>{/each}
        </select>
        <select bind:value={limit} class={selectClass} aria-label="Jumlah baris">
          <option value={3}>Top 3</option>
          <option value={10}>Top 10</option>
          <option value={20}>Top 20</option>
          <option value={0}>Semua</option>
        </select>
        {#if hasActiveFilter}
          <button
            onclick={resetFilters}
            class="btn-bubbly ml-auto flex cursor-pointer items-center gap-2 rounded-xl border-2 border-rose-200 bg-rose-50 px-5 py-2.5 text-sm font-black text-rose-700 shadow-[0_3px_0_#8A1538] transition-all active:translate-y-[3px] active:shadow-none"
          >
            <RotateCcw size={14} />
            Reset Filter
          </button>
        {/if}
      </div>
    </div>
  {/if}

  {#if loading}
    <div class="flex h-64 items-center justify-center">
      <div
        class="h-12 w-12 animate-spin rounded-full border-4 border-rose-700 border-t-transparent"
      ></div>
    </div>
  {:else}
    {#if podium.length > 0}
      <div class="flex items-end justify-center gap-6 overflow-x-auto px-4 pt-12 pb-8">
        {#each podium as slot (slot.rank)}
          {@const entry = rows[slot.index]}
          <div class="group/podium flex min-w-[120px] flex-col items-center gap-4 select-none">
            <div class="relative">
              <div
                class="flex h-16 w-16 rotate-[-4deg] items-center justify-center rounded-full border-4 border-white bg-zinc-200 text-xl font-black text-zinc-400 shadow-xl transition-transform group-hover/podium:rotate-0 sm:h-20 sm:w-20 sm:text-2xl"
              >
                {(entry.nama || 'U').charAt(0)}
              </div>
              <div class="absolute -right-2 -bottom-2 rounded-full bg-white p-1.5 shadow-md">
                {#if slot.rank === 1}
                  <Trophy class={slot.medal} size={22} />
                {:else}
                  <Medal class={slot.medal} size={22} />
                {/if}
              </div>
            </div>
            <div class="text-center">
              <div
                class="max-w-[100px] truncate text-sm font-black text-zinc-800 sm:max-w-[140px] sm:text-base"
              >
                {entry.nama || 'Anonymous'}
              </div>
              <div class="text-xs font-bold text-zinc-500 sm:text-sm">
                {entry.xp.toLocaleString('id-ID')} XP
              </div>
              {#if entry.carry}
                <div class="text-[9px] font-black tracking-widest text-sky-600 uppercase">
                  +50% minggu lalu
                </div>
              {/if}
            </div>
            <div
              class="relative flex w-full items-center justify-center rounded-t-3xl border-x-2 border-t-2 {slot.height} {slot.color} {slot.border}"
              style:box-shadow="0 8px 0 {slot.shadow}"
            >
              <span class="text-3xl font-black text-white/25 select-none sm:text-5xl"
                >#{slot.rank}</span
              >
            </div>
          </div>
        {/each}
      </div>
    {/if}

    <div
      class="shadow-soft hover:border-maroon/20 overflow-hidden rounded-[2rem] border-2 border-gray-100 bg-white transition-all"
    >
      <div
        class="bg-maroon-bg text-maroon/70 grid grid-cols-12 border-b-2 border-gray-100 px-6 py-4 text-xs font-black tracking-widest uppercase"
      >
        <div class="col-span-2 sm:col-span-1">Rank</div>
        <div class="col-span-6">Penjelajah</div>
        <div class="col-span-2 hidden text-center sm:block">Level</div>
        <div class="col-span-4 text-right sm:col-span-3">
          {weeklyMode ? 'XP Minggu Ini' : 'Total XP'}
        </div>
      </div>
      <div class="divide-y divide-zinc-100">
        {#each rows as entry, idx (entry.nim)}
          <div
            class="group grid grid-cols-12 items-center px-6 py-4 transition-colors hover:bg-zinc-50/50 {user.nim ===
            entry.nim
              ? 'bg-rose-50/40'
              : ''}"
          >
            <div class="col-span-2 flex items-center gap-2 sm:col-span-1">
              <span class="text-base font-black {idx < 3 ? 'text-rose-700' : 'text-zinc-800'}"
                >{idx + 1}</span
              >
              <Minus size={14} class="text-zinc-300" />
            </div>
            <div class="col-span-6 flex items-center gap-3">
              <div
                class="flex h-10 w-10 items-center justify-center rounded-full border-2 border-zinc-200 bg-zinc-100 font-black text-zinc-400 select-none"
              >
                {(entry.nama || 'U').charAt(0)}
              </div>
              <div class="flex flex-col">
                <span
                  class="flex max-w-[150px] items-center gap-2 truncate font-black text-zinc-900 sm:max-w-none"
                >
                  {entry.nama || 'Anonymous'}
                  {#if user.nim === entry.nim}
                    <span
                      class="rounded-lg border border-rose-200/50 bg-rose-100 px-2 py-0.5 text-[9px] font-black tracking-widest text-rose-700 uppercase"
                      >Anda</span
                    >
                  {/if}
                  {#if entry.carry}
                    <span
                      class="rounded-lg border border-sky-200/50 bg-sky-100 px-2 py-0.5 text-[9px] font-black tracking-widest text-sky-700 uppercase"
                      title="Dapat bonus 50% XP minggu lalu">+50%</span
                    >
                  {/if}
                </span>
                <span
                  class="text-[10px] font-black tracking-wider text-zinc-400 uppercase sm:hidden"
                  >Level {entry.level || 1}</span
                >
              </div>
            </div>
            <div class="col-span-2 hidden text-center sm:block">
              <span
                class="rounded-xl border border-zinc-200/50 bg-zinc-100 px-3 py-1 text-xs font-black text-zinc-600"
                >Lvl {entry.level || 1}</span
              >
            </div>
            <div
              class="col-span-4 text-right font-mono text-sm font-black text-rose-700 sm:col-span-3 sm:text-base"
            >
              {entry.xp.toLocaleString('id-ID')} XP
            </div>
          </div>
        {/each}

        {#if !isUserInTop && userRank}
          <div class="flex justify-center bg-zinc-50/30 py-3">
            {#each [0, 1, 2] as dot (dot)}
              <div class="mx-1 h-1.5 w-1.5 animate-pulse rounded-full bg-zinc-300"></div>
            {/each}
          </div>
          <div
            class="grid grid-cols-12 items-center border-t border-zinc-100 bg-rose-50/35 px-6 py-4"
          >
            <div class="col-span-2 flex items-center gap-2 sm:col-span-1">
              <span class="text-base font-black text-zinc-900">{userRank}</span>
              <Minus size={14} class="text-zinc-300" />
            </div>
            <div class="col-span-6 flex items-center gap-3">
              <div
                class="flex h-10 w-10 items-center justify-center rounded-full border-2 border-rose-200 bg-rose-100 font-black text-rose-700 select-none"
              >
                {(profile?.nama || user.nama || 'U').charAt(0)}
              </div>
              <span class="flex items-center gap-2 font-black text-zinc-900">
                {profile?.nama || user.nama || 'Anonymous'}
                <span
                  class="rounded-lg border border-rose-200/50 bg-rose-100 px-2 py-0.5 text-[9px] font-black tracking-widest text-rose-700 uppercase"
                  >Anda</span
                >
              </span>
            </div>
            <div class="col-span-2 hidden text-center sm:block">
              <span
                class="rounded-xl border border-rose-200/50 bg-rose-100 px-3 py-1 text-xs font-black text-rose-700"
                >Lvl {profile?.level || 1}</span
              >
            </div>
            <div
              class="col-span-4 text-right font-mono text-sm font-black text-rose-700 sm:col-span-3 sm:text-base"
            >
              {(weeklyMode ? myWeeklyXp : (profile?.xp ?? 0)).toLocaleString('id-ID')} XP
            </div>
          </div>
        {/if}
      </div>

      {#if rows.length === 0}
        <div class="space-y-4 py-12 text-center">
          <div
            class="mx-auto flex h-16 w-16 items-center justify-center rounded-full border border-zinc-100 bg-zinc-50 shadow-inner"
          >
            <Users class="text-zinc-400" size={26} />
          </div>
          <div class="space-y-1">
            <p class="font-black text-zinc-900">Belum ada data</p>
            <p class="text-sm font-bold text-balance text-zinc-500">
              Jadilah yang pertama untuk mencapai puncak papan peringkat!
            </p>
          </div>
        </div>
      {/if}
    </div>
  {/if}
</div>
