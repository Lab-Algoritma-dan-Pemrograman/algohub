<script lang="ts">
  import type { ActivityLog } from '@algohub/api-client/elearning';
  import { Button, Input, VirtualList } from '@algohub/ui';
  import ChevronLeft from '@lucide/svelte/icons/chevron-left';
  import ChevronRight from '@lucide/svelte/icons/chevron-right';
  import Download from '@lucide/svelte/icons/download';
  import RefreshCw from '@lucide/svelte/icons/refresh-cw';
  import Search from '@lucide/svelte/icons/search';
  import Trash from '@lucide/svelte/icons/trash-2';
  import { createQuery, keepPreviousData } from '@tanstack/svelte-query';
  import { onMount } from 'svelte';
  import { SvelteSet } from 'svelte/reactivity';
  import { perform } from '#lib/admin/action.ts';
  import ConfirmDialog from '#lib/components/admin/ConfirmDialog.svelte';
  import {
    ACTIVITY_LOG_PAGE_SIZE,
    activityLogActions,
    activityLogsQuery,
  } from '#lib/queries/monitoring.ts';
  import { watchActivityLogs } from '#lib/realtime/monitoring.ts';
  import { normalizeRole } from '#lib/roles.ts';
  import type { PageProps } from './$types';

  const EVENTS: Record<string, { label: string; color: string }> = {
    login: { label: 'Login', color: 'bg-emerald-50 text-emerald-700 border-emerald-200' },
    logout: { label: 'Logout', color: 'bg-zinc-100 text-zinc-600 border-zinc-200' },
    access_modified: { label: 'Akses Diubah', color: 'bg-red-50 text-red-700 border-red-200' },
    curriculum_modified: {
      label: 'Ubah Kurikulum',
      color: 'bg-rose-50 text-rose-700 border-rose-200',
    },
    progress_reset: {
      label: 'Reset Progress',
      color: 'bg-amber-50 text-amber-700 border-amber-200',
    },
    xp_adjusted: { label: 'Ubah XP', color: 'bg-indigo-50 text-indigo-700 border-indigo-200' },
    user_deleted: { label: 'Hapus Peserta', color: 'bg-rose-50 text-rose-700 border-rose-200' },
  };
  const FALLBACK_COLOR = 'bg-zinc-100 text-zinc-600 border-zinc-200';

  let { data }: PageProps = $props();
  // Hanya admin yang boleh menghapus log; kordas hanya membaca, seperti AuditLogPage lama.
  const canDelete = $derived(normalizeRole(data.auth.user.role) === 'admin');

  let page = $state(1);
  let search = $state('');
  let eventFilter = $state('all');
  let dateFrom = $state('');
  let dateTo = $state('');
  let deleting = $state(false);
  let confirmOpen = $state(false);
  const selected = new SvelteSet<string>();

  const logs = createQuery(() => ({
    ...activityLogsQuery(page),
    placeholderData: keepPreviousData,
  }));
  const items = $derived(logs.data?.items ?? []);
  const total = $derived(logs.data?.total ?? 0);
  const pageCount = $derived(Math.max(1, Math.ceil(total / ACTIVITY_LOG_PAGE_SIZE)));

  const visible = $derived.by(() => {
    const term = search.toLowerCase();
    const from = dateFrom ? new Date(`${dateFrom}T00:00:00`).getTime() : null;
    const to = dateTo ? new Date(`${dateTo}T23:59:59.999`).getTime() : null;
    return items.filter((log) => {
      const at = log.timestamp ? new Date(log.timestamp).getTime() : 0;
      return (
        (term === '' ||
          (log.nama ?? '').toLowerCase().includes(term) ||
          (log.nim ?? '').includes(search) ||
          (log.details ?? '').toLowerCase().includes(term)) &&
        (eventFilter === 'all' || log.event_type === eventFilter) &&
        (from === null || at >= from) &&
        (to === null || at <= to)
      );
    });
  });

  const wib = (iso: string | null, options: Intl.DateTimeFormatOptions) =>
    iso
      ? new Intl.DateTimeFormat('id-ID', {
          timeZone: 'Asia/Jakarta',
          hour12: false,
          ...options,
        }).format(new Date(iso))
      : '-';
  const full = (iso: string | null) =>
    wib(iso, {
      day: '2-digit',
      month: 'short',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
    });

  onMount(() => watchActivityLogs());

  // Berpindah halaman atau mengubah filter membuang pilihan lama supaya tidak menghapus log yang tak terlihat.
  $effect(() => {
    void page;
    selected.clear();
  });

  function toggleAll() {
    const ids = visible.map((l) => l.id);
    if (ids.length > 0 && ids.every((id) => selected.has(id))) selected.clear();
    else for (const id of ids) selected.add(id);
  }

  function toggle(id: string) {
    if (selected.has(id)) selected.delete(id);
    else selected.add(id);
  }

  async function deleteSelected() {
    deleting = true;
    const ids = [...selected];
    await perform(
      () => activityLogActions.remove(ids),
      `${ids.length} log dihapus.`,
      'Gagal menghapus log',
    );
    selected.clear();
    deleting = false;
  }

  const csv = (value: string | null) => `"${(value ?? '').replace(/"/g, '""')}"`;

  function exportCsv() {
    const rows = visible.map((l: ActivityLog) =>
      [full(l.timestamp), l.nim, l.nama, l.event_type, l.details].map(csv).join(','),
    );
    const body = ['Timestamp (GMT+7),NIM,Nama,Event,Detail', ...rows].join('\n');
    const url = URL.createObjectURL(new Blob([body], { type: 'text/csv' }));
    const link = document.createElement('a');
    link.href = url;
    link.download = `audit-log-${new Date().toISOString().slice(0, 10)}.csv`;
    link.click();
    URL.revokeObjectURL(url);
  }

  function clearFilters() {
    search = '';
    eventFilter = 'all';
    dateFrom = '';
    dateTo = '';
  }
</script>

<div class="space-y-6">
  <div class="flex flex-wrap items-center justify-between gap-4">
    <div>
      <h1 class="text-2xl font-bold tracking-tight">Audit Log</h1>
      <p class="mt-1 text-sm text-zinc-500">{total} entri · Zona waktu GMT+7 (WIB)</p>
    </div>
    <div class="flex items-center gap-2">
      <Button variant="outline" class="rounded-xl font-bold" onclick={exportCsv}>
        <Download class="mr-2 h-4 w-4" /> Export CSV
      </Button>
      <Button
        variant="outline"
        class="rounded-xl font-bold"
        disabled={logs.isFetching}
        onclick={() => logs.refetch()}
      >
        <RefreshCw class="mr-2 h-4 w-4 {logs.isFetching ? 'animate-spin' : ''}" /> Refresh
      </Button>
    </div>
  </div>

  <div class="space-y-3 rounded-2xl border border-zinc-200 bg-white p-4">
    <div class="flex flex-wrap gap-3">
      <div class="relative min-w-[200px] flex-1">
        <Search size={16} class="absolute top-1/2 left-3 -translate-y-1/2 text-zinc-400" />
        <Input bind:value={search} placeholder="Cari nama, NIM, atau detail..." class="pl-9" />
      </div>
      <select
        class="rounded-xl border border-zinc-200 bg-zinc-50 px-4 py-2.5 text-sm font-bold"
        bind:value={eventFilter}
      >
        <option value="all">Semua Event</option>
        {#each Object.entries(EVENTS) as [key, value] (key)}<option value={key}
            >{value.label}</option
          >{/each}
      </select>
    </div>
    <div class="flex flex-wrap items-center gap-3">
      <input
        type="date"
        class="rounded-lg border border-zinc-200 bg-zinc-50 px-3 py-2 text-sm"
        bind:value={dateFrom}
        aria-label="Dari tanggal"
      />
      <span class="text-xs text-zinc-400">s/d</span>
      <input
        type="date"
        class="rounded-lg border border-zinc-200 bg-zinc-50 px-3 py-2 text-sm"
        bind:value={dateTo}
        aria-label="Sampai tanggal"
      />
      {#if search || eventFilter !== 'all' || dateFrom || dateTo}
        <Button variant="ghost" size="sm" class="text-xs font-bold" onclick={clearFilters}
          >Hapus Filter</Button
        >
      {/if}
    </div>
  </div>

  {#if canDelete && selected.size > 0}
    <div class="flex items-center gap-3 rounded-xl border border-rose-200 bg-rose-50 px-4 py-3">
      <span class="text-sm font-bold text-rose-700">{selected.size} dipilih</span>
      <div class="flex-1"></div>
      <Button
        size="sm"
        class="rounded-lg bg-red-600 text-xs font-bold text-white"
        disabled={deleting}
        onclick={() => (confirmOpen = true)}
      >
        <Trash class="mr-1 h-3 w-3" /> Hapus Terpilih
      </Button>
    </div>
  {/if}

  <div class="overflow-hidden rounded-2xl border border-zinc-200 bg-white">
    <div
      class="grid grid-cols-[2rem_9rem_1fr_9rem_1.5fr] items-center gap-3 border-b border-zinc-100 bg-zinc-50/50 p-3 text-[10px] font-bold tracking-widest text-zinc-400 uppercase"
    >
      <span>
        {#if canDelete}
          <input
            type="checkbox"
            aria-label="Pilih semua"
            checked={visible.length > 0 && visible.every((l) => selected.has(l.id))}
            onchange={toggleAll}
          />
        {/if}
      </span>
      <span>Waktu (GMT+7)</span>
      <span>Mahasiswa</span>
      <span>Event</span>
      <span>Detail</span>
    </div>
    {#if logs.isPending}
      <p class="p-12 text-center text-zinc-400">Memuat log...</p>
    {:else if visible.length === 0}
      <p class="p-12 text-center font-bold text-zinc-400">Tidak ada log ditemukan</p>
    {:else}
      <VirtualList
        items={visible}
        key={(l) => l.id}
        estimateSize={56}
        threshold={40}
        maxHeight="65vh"
      >
        {#snippet item(log)}
          {@const config = EVENTS[log.event_type]}
          <div
            class="grid grid-cols-[2rem_9rem_1fr_9rem_1.5fr] items-center gap-3 border-b border-zinc-50 p-3 hover:bg-zinc-50/50"
          >
            <span>
              {#if canDelete}
                <input
                  type="checkbox"
                  aria-label="Pilih log"
                  checked={selected.has(log.id)}
                  onchange={() => toggle(log.id)}
                />
              {/if}
            </span>
            <span>
              <span class="block font-mono text-sm text-zinc-700"
                >{wib(log.timestamp, {
                  hour: '2-digit',
                  minute: '2-digit',
                  second: '2-digit',
                })}</span
              >
              <span class="text-[10px] text-zinc-400"
                >{wib(log.timestamp, { day: '2-digit', month: 'short', year: 'numeric' })}</span
              >
            </span>
            <span class="min-w-0">
              <span class="block truncate text-sm font-bold">{log.nama}</span>
              <span class="text-[10px] text-zinc-400">{log.nim}</span>
            </span>
            <span>
              <span
                class="inline-flex rounded-lg border px-2 py-1 text-[10px] font-bold {config?.color ??
                  FALLBACK_COLOR}"
              >
                {config?.label ?? log.event_type}
              </span>
            </span>
            <span class="truncate text-sm text-zinc-600" title={log.details ?? ''}
              >{log.details}</span
            >
          </div>
        {/snippet}
      </VirtualList>
    {/if}

    <div class="flex items-center justify-between border-t border-zinc-100 p-4">
      <span class="text-xs text-zinc-500">Halaman {page} dari {pageCount}</span>
      <div class="flex gap-2">
        <Button
          variant="outline"
          size="sm"
          class="rounded-lg"
          disabled={page <= 1}
          onclick={() => (page -= 1)}
        >
          <ChevronLeft class="h-4 w-4" />
        </Button>
        <Button
          variant="outline"
          size="sm"
          class="rounded-lg"
          disabled={page >= pageCount}
          onclick={() => (page += 1)}
        >
          <ChevronRight class="h-4 w-4" />
        </Button>
      </div>
    </div>
  </div>
</div>

<ConfirmDialog
  bind:open={confirmOpen}
  title="Hapus Log Audit?"
  message="{selected.size} log terpilih akan dihapus. Tindakan ini tidak bisa dibatalkan."
  confirmLabel="Hapus"
  busy={deleting}
  onconfirm={deleteSelected}
/>
