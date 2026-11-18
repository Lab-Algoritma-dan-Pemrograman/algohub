<script lang="ts">
  import { Button, Input, VirtualList } from '@algohub/ui';
  import Activity from '@lucide/svelte/icons/activity';
  import Clock from '@lucide/svelte/icons/clock';
  import RefreshCw from '@lucide/svelte/icons/refresh-cw';
  import Search from '@lucide/svelte/icons/search';
  import Trophy from '@lucide/svelte/icons/trophy';
  import Users from '@lucide/svelte/icons/users';
  import { createQuery } from '@tanstack/svelte-query';
  import { onMount } from 'svelte';
  import { perform } from '#lib/admin/action.ts';
  import ConfirmDialog from '#lib/components/admin/ConfirmDialog.svelte';
  import StudentCard from '#lib/components/admin/StudentCard.svelte';
  import { buildCurriculum, lessonIds } from '#lib/curriculum.ts';
  import { levelsQuery, lessonsQuery, modulesQuery } from '#lib/queries/curriculum.ts';
  import {
    STALE_SESSION_MINUTES,
    cleanupStaleSessions,
    monitoringActions,
    monitoringProgressQuery,
    monitoringSessionsQuery,
    monitoringStudentsQuery,
  } from '#lib/queries/monitoring.ts';
  import { watchMonitoring } from '#lib/realtime/monitoring.ts';
  import { isAdminOrKordas } from '#lib/roles.ts';
  import type { PageProps } from './$types';

  type Status = 'all' | 'online' | 'offline';
  type Sort = 'nama' | 'xp' | 'progress' | 'last_active';

  const CLEANUP_INTERVAL_MS = 45_000;
  const STUDENT_ROLES = ['praktikan', 'mahasiswa'];

  let { data }: PageProps = $props();
  // Asisten hanya melihat; admin dan kordas boleh mengubah XP, reset, dan hapus.
  const canManage = $derived(isAdminOrKordas(data.auth.user.role));

  const students = createQuery(monitoringStudentsQuery);
  const sessions = createQuery(monitoringSessionsQuery);
  const progress = createQuery(monitoringProgressQuery);
  const levels = createQuery(levelsQuery);
  const modules = createQuery(modulesQuery);
  const lessons = createQuery(lessonsQuery);

  let search = $state('');
  let kelas = $state('all');
  let status = $state<Status>('all');
  let sort = $state<Sort>('nama');
  let expandedNim = $state<string | null>(null);
  let refreshing = $state(false);
  let busy = $state(false);
  let confirmOpen = $state(false);
  let confirm = $state<{ title: string; message: string; run: () => Promise<void> }>({
    title: '',
    message: '',
    run: async () => {},
  });

  const curriculum = $derived(
    buildCurriculum(levels.data ?? [], modules.data ?? [], lessons.data ?? []),
  );
  const totalLessons = $derived(curriculum.reduce((sum, l) => sum + lessonIds(l).length, 0));
  const list = $derived(students.data ?? []);

  // Hanya sesi milik mahasiswa yang dihitung online, seperti query lama.
  const heartbeats = $derived(
    new Map(
      (sessions.data ?? [])
        .filter((s) => s.last_heartbeat && STUDENT_ROLES.includes(s.users?.role ?? ''))
        .map((s) => [s.nim, s.last_heartbeat as string]),
    ),
  );
  const completedBy = $derived(new Map((progress.data ?? []).map((p) => [p.nim, p.lessonIds])));

  const percentOf = (nim: string) =>
    totalLessons > 0 ? Math.round(((completedBy.get(nim)?.length ?? 0) / totalLessons) * 100) : 0;
  const lastActive = (nim: string, fallback: string | null) =>
    new Date(heartbeats.get(nim) ?? fallback ?? 0).getTime();

  const kelasList = $derived([...new Set(list.map((s) => s.kelas))].sort());

  const visible = $derived.by(() => {
    const term = search.toLowerCase();
    const filtered = list.filter((s) => {
      const online = heartbeats.has(s.nim);
      return (
        (s.nama.toLowerCase().includes(term) || s.nim.includes(search)) &&
        (kelas === 'all' || s.kelas === kelas) &&
        (status === 'all' || (status === 'online' ? online : !online))
      );
    });
    return filtered.sort((a, b) => {
      if (sort === 'xp') return b.xp - a.xp;
      if (sort === 'progress') return percentOf(b.nim) - percentOf(a.nim);
      if (sort === 'last_active')
        return lastActive(b.nim, b.last_active) - lastActive(a.nim, a.last_active);
      return a.nama.localeCompare(b.nama);
    });
  });

  const onlineCount = $derived(list.filter((s) => heartbeats.has(s.nim)).length);
  const averageXp = $derived(
    list.length > 0 ? Math.round(list.reduce((sum, s) => sum + s.xp, 0) / list.length) : 0,
  );
  const studySeconds = $derived(list.reduce((sum, s) => sum + (s.study_time ?? 0), 0));
  const studyText = $derived.by(() => {
    if (studySeconds <= 0) return '0 menit';
    const hours = Math.floor(studySeconds / 3600);
    const minutes = Math.floor((studySeconds % 3600) / 60);
    return hours > 0 ? `${hours} jam ${minutes} menit` : `${minutes} menit`;
  });

  async function refresh() {
    refreshing = true;
    try {
      await cleanupStaleSessions().catch(() => undefined);
      await Promise.all([students.refetch(), sessions.refetch(), progress.refetch()]);
    } finally {
      refreshing = false;
    }
  }

  onMount(() => {
    void cleanupStaleSessions().catch(() => undefined);
    const stopRealtime = watchMonitoring();
    const timer = setInterval(
      () => void cleanupStaleSessions().catch(() => undefined),
      CLEANUP_INTERVAL_MS,
    );
    return () => {
      stopRealtime();
      clearInterval(timer);
    };
  });

  function ask(
    title: string,
    message: string,
    run: () => Promise<unknown>,
    ok: string,
    fail: string,
  ) {
    confirm = {
      title,
      message,
      run: async () => {
        busy = true;
        try {
          await perform(run, ok, fail);
        } finally {
          busy = false;
        }
      },
    };
    confirmOpen = true;
  }
</script>

<div class="space-y-8 pb-12">
  <div
    class="flex flex-col justify-between gap-4 rounded-3xl border border-zinc-200/80 bg-white p-6 shadow-sm sm:flex-row sm:items-center"
  >
    <div>
      <h1 class="text-3xl font-black tracking-tight text-zinc-900">Monitoring Mahasiswa</h1>
      <p class="mt-1 text-sm font-medium text-zinc-500">
        Pantau keaktifan, progres belajar, dan nilai XP praktikan secara realtime.
      </p>
      <p class="mt-1 text-[11px] text-zinc-400">
        Sesi tanpa heartbeat {STALE_SESSION_MINUTES} menit dibersihkan otomatis.
      </p>
    </div>
    <Button class="rounded-2xl bg-zinc-900 font-black" disabled={refreshing} onclick={refresh}>
      <RefreshCw class="mr-2 h-4 w-4 {refreshing ? 'animate-spin' : ''}" /> Refresh Data
    </Button>
  </div>

  <div class="grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-4">
    {#each [{ icon: Users, label: 'Total Praktikan', value: `${list.length} Orang` }, { icon: Activity, label: 'Sedang Online', value: `${onlineCount} Aktif` }, { icon: Trophy, label: 'Rata-rata XP', value: `${averageXp.toLocaleString('id-ID')} XP` }, { icon: Clock, label: 'Total Waktu Belajar', value: studyText }] as card (card.label)}
      <div
        class="flex items-center gap-5 rounded-3xl border border-zinc-200/80 bg-white p-6 shadow-sm"
      >
        <div
          class="flex h-14 w-14 shrink-0 items-center justify-center rounded-2xl bg-rose-700 text-white"
        >
          <card.icon size={24} />
        </div>
        <div>
          <div
            class="mb-1.5 text-[10px] leading-none font-bold tracking-widest text-zinc-400 uppercase"
          >
            {card.label}
          </div>
          <div class="text-xl font-black text-zinc-900">{card.value}</div>
        </div>
      </div>
    {/each}
  </div>

  <div
    class="flex flex-col items-center justify-between gap-4 rounded-3xl border border-zinc-200/80 bg-white p-5 shadow-sm md:flex-row"
  >
    <div class="relative w-full md:max-w-xs">
      <Search size={16} class="absolute top-1/2 left-3.5 -translate-y-1/2 text-zinc-400" />
      <Input bind:value={search} placeholder="Cari nama atau NIM..." class="rounded-2xl pl-10" />
    </div>
    <div class="flex w-full flex-wrap items-center gap-3 md:w-auto">
      <div class="flex gap-1 rounded-2xl border border-zinc-200 bg-zinc-50 p-1">
        {#each [['all', 'Semua'], ['online', 'Online'], ['offline', 'Offline']] as const as [key, label] (key)}
          <button
            type="button"
            class="rounded-xl px-3 py-1.5 text-xs font-bold {status === key
              ? 'bg-white text-zinc-900 shadow-sm'
              : 'text-zinc-500'}"
            onclick={() => (status = key)}
          >
            {label}
          </button>
        {/each}
      </div>
      <select
        class="rounded-2xl border border-zinc-200 bg-zinc-50 px-4 py-2.5 text-xs font-bold"
        bind:value={kelas}
      >
        <option value="all">Semua Kelas</option>
        {#each kelasList as k (k)}<option value={k}>{k}</option>{/each}
      </select>
      <select
        class="rounded-2xl border border-zinc-200 bg-zinc-50 px-4 py-2.5 text-xs font-bold"
        bind:value={sort}
      >
        <option value="nama">Urut: Nama (A-Z)</option>
        <option value="xp">Urut: XP Terbanyak</option>
        <option value="progress">Urut: Progres Tertinggi</option>
        <option value="last_active">Urut: Terakhir Aktif</option>
      </select>
    </div>
  </div>

  {#if students.isPending}
    <p class="py-16 text-center text-sm font-semibold text-zinc-500">Memuat data monitoring...</p>
  {:else if students.isError}
    <p class="py-16 text-center text-sm font-semibold text-red-600">
      Gagal memuat data monitoring.
    </p>
  {:else if visible.length === 0}
    <div class="rounded-[2rem] border border-zinc-200/80 bg-white py-16 text-center shadow-sm">
      <Users size={40} class="mx-auto mb-4 text-zinc-300" />
      <p class="text-lg font-extrabold text-zinc-800">Tidak ada mahasiswa ditemukan</p>
      <p class="mx-auto mt-1 max-w-xs text-xs text-zinc-400">
        Silakan ubah kata kunci pencarian atau bersihkan filter yang aktif.
      </p>
    </div>
  {:else}
    <VirtualList items={visible} key={(s) => s.nim} estimateSize={96} maxHeight="75vh">
      {#snippet item(student)}
        <StudentCard
          {student}
          online={heartbeats.has(student.nim)}
          heartbeat={heartbeats.get(student.nim) ?? null}
          completed={completedBy.get(student.nim) ?? []}
          {totalLessons}
          {curriculum}
          {canManage}
          expanded={expandedNim === student.nim}
          ontoggle={() => (expandedNim = expandedNim === student.nim ? null : student.nim)}
          onadjustxp={(xp) =>
            ask(
              'Ubah XP',
              `Ubah XP ${student.nama} dari ${student.xp} menjadi ${xp}?`,
              () => monitoringActions.adjustXp(student.nim, xp),
              `XP berhasil diubah menjadi ${xp}.`,
              'Gagal mengubah XP',
            )}
          onreset={() =>
            ask(
              'Reset Progres?',
              `Reset progres belajar ${student.nama} (${student.nim})? Seluruh riwayat pelajaran, XP, dan streak kembali ke 0. Tindakan ini permanen.`,
              () => monitoringActions.resetUser(student.nim),
              'Progres berhasil direset.',
              'Gagal mereset progres',
            )}
          ondelete={() =>
            ask(
              'Hapus Akun?',
              `Hapus akun ${student.nama} (${student.nim})? Seluruh data progres dan pencapaian terhapus selamanya.`,
              () => monitoringActions.deleteUser(student.nim),
              'Akun berhasil dihapus.',
              'Gagal menghapus akun',
            )}
        />
      {/snippet}
    </VirtualList>
  {/if}
</div>

<ConfirmDialog
  bind:open={confirmOpen}
  title={confirm.title}
  message={confirm.message}
  {busy}
  onconfirm={confirm.run}
/>
