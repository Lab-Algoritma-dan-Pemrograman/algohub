<script lang="ts">
  import { createQuery } from '@tanstack/svelte-query';
  import { Users, UserCheck, GraduationCap } from '@lucide/svelte';
  import { dashboardQuery } from '#lib/queries/admin/dashboard.ts';
  import { labelJenis } from '#lib/utils.ts';
  import { Skeleton } from '@algohub/ui';

  const statQuery = createQuery(dashboardQuery);
  let stat = $derived(statQuery.data ?? null);
  let err = $derived(statQuery.error?.message ?? '');

  const numFmt = new Intl.NumberFormat('id-ID');
  const fmt = (n: number | null | undefined) => (n == null ? '-' : numFmt.format(n));
</script>

<h1 class="mb-1 text-2xl font-bold text-ink-heading">Dashboard Admin</h1>
<p class="mb-5 text-sm text-ink-caption">Ringkasan aktivitas praktikum.</p>

{#if err}<p class="rounded-lg bg-state-error-bg p-3 text-state-error">{err}</p>{/if}

{#if stat}
  <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
    <div class="card flex items-start justify-between">
      <div>
        <p class="text-sm text-ink-caption">Total Mahasiswa</p>
        <p class="mt-1 text-3xl font-bold tabular-nums text-primary">{fmt(stat.total_mahasiswa)}</p>
      </div>
      <span
        class="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-xl bg-primary/10 text-primary"
      >
        <Users size={20} />
      </span>
    </div>
    <div class="card flex items-start justify-between">
      <div>
        <p class="text-sm text-ink-caption">Sudah Register</p>
        <p class="mt-1 text-3xl font-bold tabular-nums text-state-info">{fmt(stat.sudah_register)}</p>
        <p class="text-xs tabular-nums text-ink-caption">Belum: {fmt(stat.belum_register)}</p>
      </div>
      <span
        class="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-xl bg-state-info-bg text-state-info"
      >
        <UserCheck size={20} />
      </span>
    </div>
    <div class="card flex items-start justify-between">
      <div>
        <p class="text-sm text-ink-caption">Total Asisten</p>
        <p class="mt-1 text-3xl font-bold tabular-nums text-ink-heading">{fmt(stat.total_asisten)}</p>
      </div>
      <span
        class="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-xl bg-surface-soft text-ink-heading"
      >
        <GraduationCap size={20} />
      </span>
    </div>
  </div>

  <h2 class="mb-3 mt-8 text-xl font-bold text-ink-heading">Mahasiswa per Kelas &amp; Shift</h2>
  <div class="table-wrap max-w-xl">
    <table class="tbl">
      <thead><tr><th>Kelas</th><th>Shift</th><th>Jumlah</th></tr></thead>
      <tbody>
        {#each stat.per_kelas_shift ?? [] as r, i (i)}
          <tr><td>{r.nama_kelas}</td><td>{r.shift}</td><td class="tabular-nums">{fmt(r.jumlah)}</td></tr>
        {:else}
          <tr><td colspan="3" class="text-center text-ink-caption">Belum ada data kelas.</td></tr>
        {/each}
      </tbody>
    </table>
  </div>

  <h2 class="mb-3 mt-8 text-xl font-bold text-ink-heading">Sesi Aktif &amp; Progress</h2>
  {#if (stat.sesi_aktif ?? []).length === 0}
    <p class="text-ink-caption">Tidak ada sesi aktif.</p>
  {:else}
    <div class="grid gap-4 md:grid-cols-2">
      {#each stat.sesi_aktif ?? [] as s (s.aktivasi_sesi_id)}
        <div class="card">
          <h3 class="text-lg">{s.judul_sesi}</h3>
          <p class="text-sm text-ink-caption">{s.nama_kelas} · Shift {s.shift}</p>
          <div class="mt-3 space-y-2">
            {#each s.courses ?? [] as c (c.course_id)}
              <div class="rounded-lg border border-gray-100 p-2 text-sm">
                <div class="flex justify-between">
                  <span>{labelJenis(c.jenis)}</span>
                  <span
                    class="badge {c.is_open
                      ? 'bg-state-success-bg text-state-success'
                      : 'bg-gray-100 text-ink-caption'}">{c.is_open ? 'Terbuka' : 'Tertutup'}</span
                  >
                </div>
                <div class="mt-1 text-xs text-ink-caption">
                  Selesai {c.selesai} · Sedang {c.sedang} · Belum {c.belum}
                </div>
              </div>
            {/each}
          </div>
        </div>
      {/each}
    </div>
  {/if}
{:else if !err}
  <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3" aria-label="Memuat dashboard admin">
    {#each Array(3) as _, i (i)}
      <div class="card flex items-start justify-between">
        <div class="flex-1">
          <Skeleton class="h-4 w-28" />
          <Skeleton class="mt-2 h-8 w-16" />
          {#if i === 1}<Skeleton class="mt-1 h-3 w-24" />{/if}
        </div>
        <Skeleton class="h-10 w-10 rounded-xl" />
      </div>
    {/each}
  </div>

  <Skeleton class="mb-3 mt-8 h-6 w-48" />
  <div class="table-wrap max-w-xl">
    <div class="space-y-3 p-4">
      <Skeleton class="h-4 w-full" />
      <Skeleton class="h-4 w-full" />
      <Skeleton class="h-4 w-full" />
    </div>
  </div>

  <Skeleton class="mb-3 mt-8 h-6 w-44" />
  <div class="grid gap-4 md:grid-cols-2">
    <Skeleton class="h-40 rounded-2xl" />
    <Skeleton class="h-40 rounded-2xl" />
  </div>
{/if}
