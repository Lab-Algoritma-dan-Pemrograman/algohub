<script lang="ts">
  import {
    Badge,
    Button,
    Card,
    CardContent,
    CardHeader,
    CardTitle,
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
    Input,
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
  } from '@algohub/ui';
  import ChevronLeft from '@lucide/svelte/icons/chevron-left';
  import ChevronRight from '@lucide/svelte/icons/chevron-right';
  import ClipboardList from '@lucide/svelte/icons/clipboard-list';
  import Clock from '@lucide/svelte/icons/clock';
  import Database from '@lucide/svelte/icons/database';
  import Eye from '@lucide/svelte/icons/eye';
  import Filter from '@lucide/svelte/icons/filter';
  import Loader2 from '@lucide/svelte/icons/loader-2';
  import RefreshCw from '@lucide/svelte/icons/refresh-cw';
  import Search from '@lucide/svelte/icons/search';
  import User from '@lucide/svelte/icons/user';
  import { createQuery, keepPreviousData } from '@tanstack/svelte-query';
  import { rpcQueryOptions } from '#lib/queries/rpc.ts';
  import { toast } from '#lib/toast.ts';
  import type { PageProps } from './$types';

  interface AuditRow {
    id: number;
    created_at: string;
    actor_name?: string;
    actor_id?: number | null;
    action_type: string;
    description?: string;
    payload?: unknown;
  }

  const LIMIT = 25;
  const ACTIONS: Record<string, string> = {
    all: 'Semua Aksi',
    LOGIN: 'LOGIN',
    INSERT: 'INSERT',
    UPDATE: 'UPDATE',
    DELETE: 'DELETE',
    BATCH_REGISTER: 'BATCH REGISTER',
    RESET_PLOTTING: 'RESET PLOTTING',
  };

  let { data }: PageProps = $props();
  const user = $derived(data.auth.user);

  let searchTerm = $state('');
  let debouncedSearch = $state('');
  let actionType = $state('all');
  let page = $state(1);
  let selectedPayload = $state<unknown>(null);
  let isDialogOpen = $state(false);

  $effect(() => {
    const term = searchTerm;
    const timer = setTimeout(() => {
      debouncedSearch = term;
      page = 1;
    }, 500);
    return () => clearTimeout(timer);
  });

  // Satu baris ekstra diminta untuk mengetahui apakah masih ada halaman berikutnya.
  const logsQuery = createQuery(() =>
    rpcQueryOptions<AuditRow[]>(
      'get_audit_logs_secure',
      {
        p_viewer_id: user.id,
        p_search: debouncedSearch || null,
        p_action_type: actionType === 'all' ? null : actionType,
        p_limit: LIMIT + 1,
        p_offset: (page - 1) * LIMIT,
      },
      { placeholderData: keepPreviousData },
    ),
  );

  $effect(() => {
    if (logsQuery.error) {
      console.error('Gagal mengambil audit log:', logsQuery.error);
      toast.error('Gagal memuat log riwayat aktivitas: ' + logsQuery.error.message);
    }
  });

  const results = $derived(logsQuery.data || []);
  const hasMore = $derived(results.length > LIMIT);
  const logs = $derived(results.slice(0, LIMIT));
  const loading = $derived(logsQuery.isPending);

  function handleRefresh() {
    page = 1;
    void logsQuery.refetch();
    toast.success('Log riwayat aktivitas berhasil diperbarui! 🔄✨');
  }

  function actionBadgeColor(action: string) {
    const act = action.toUpperCase();
    if (act.includes('DELETE')) return 'bg-red-100 text-red-700 border-red-200 hover:bg-red-100';
    if (act.includes('UPDATE'))
      return 'bg-blue-100 text-blue-700 border-blue-200 hover:bg-blue-100';
    if (act.includes('INSERT') || act.includes('CREATE')) {
      return 'bg-green-100 text-green-700 border-green-200 hover:bg-green-100';
    }
    if (act.includes('LOGIN')) {
      return 'bg-orange-100 text-orange-700 border-orange-200 hover:bg-orange-100';
    }
    return 'bg-slate-100 text-slate-700 border-slate-200 hover:bg-slate-100';
  }

  const formatIndonesianDate = (dateString: string) =>
    new Date(dateString).toLocaleString('id-ID', {
      day: 'numeric',
      month: 'short',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
    });
</script>

<div class="mx-auto max-w-6xl space-y-6">
  <div
    class="flex flex-col items-start justify-between gap-4 rounded-2xl border border-gray-100 bg-white p-6 shadow-sm xl:flex-row xl:items-center"
  >
    <div>
      <h1 class="text-primary flex items-center gap-2 text-2xl font-black tracking-tight uppercase">
        <ClipboardList class="text-maroon h-6 w-6 -rotate-3 transform" /> Audit Log Aktivitas
      </h1>
      <p class="text-muted-foreground mt-1 text-sm font-medium">
        Pantau dan verifikasi setiap perubahan data yang terjadi dalam sistem Lab AP.
      </p>
    </div>
    <Button
      onclick={handleRefresh}
      variant="outline"
      class="border-maroon/20 text-maroon hover:bg-maroon-bg rounded-xl font-bold transition-all duration-200 hover:scale-105"
      disabled={logsQuery.isFetching}
    >
      <RefreshCw class="mr-2 h-4 w-4 {logsQuery.isFetching ? 'animate-spin' : ''}" />
      Segarkan
    </Button>
  </div>

  <div
    class="grid grid-cols-1 gap-4 rounded-2xl border border-gray-100 bg-white p-5 shadow-sm md:grid-cols-3"
  >
    <div class="relative md:col-span-2">
      <Search class="absolute top-3.5 left-3.5 h-4 w-4 text-gray-400" />
      <Input
        placeholder="Cari berdasarkan nama aktor atau deskripsi aktivitas..."
        bind:value={searchTerm}
        class="focus:border-maroon focus:ring-maroon h-11 rounded-xl border-gray-200 pl-10 font-medium"
      />
    </div>
    <div class="flex h-11 items-center gap-2 rounded-xl border border-gray-200 bg-white px-3 py-1">
      <Filter class="h-4 w-4 shrink-0 text-gray-400" />
      <Select
        type="single"
        value={actionType}
        onValueChange={(val) => {
          actionType = val;
          page = 1;
        }}
      >
        <SelectTrigger
          class="h-full w-full border-none bg-transparent font-bold text-gray-600 focus:ring-0"
          >{ACTIONS[actionType]}</SelectTrigger
        >
        <SelectContent class="rounded-xl">
          {#each Object.entries(ACTIONS) as [value, label] (value)}<SelectItem
              {value}
              {label}
            />{/each}
        </SelectContent>
      </Select>
    </div>
  </div>

  <Card class="overflow-hidden rounded-2xl border border-gray-100 shadow-sm">
    <CardHeader class="border-b border-gray-100 bg-gray-50/50 px-6 py-4">
      <CardTitle class="text-dark flex items-center gap-2 text-lg font-black">
        <Database class="h-5 w-5 text-gray-400" /> Riwayat Log Aktivitas ({logs.length})
      </CardTitle>
    </CardHeader>
    <CardContent class="p-0">
      <Table>
        <TableHeader class="bg-gray-50/50">
          <TableRow>
            <TableHead class="px-6 py-3 font-extrabold text-gray-700">Waktu & Tanggal</TableHead>
            <TableHead class="px-6 py-3 font-extrabold text-gray-700">Aktor</TableHead>
            <TableHead class="px-6 py-3 font-extrabold text-gray-700">Tipe Aksi</TableHead>
            <TableHead class="px-6 py-3 font-extrabold text-gray-700">Deskripsi</TableHead>
            <TableHead class="px-6 py-3 text-center font-extrabold text-gray-700">Payload</TableHead
            >
          </TableRow>
        </TableHeader>
        <TableBody>
          {#if loading}
            <TableRow>
              <TableCell colspan={5} class="py-16 text-center">
                <div class="flex flex-col items-center gap-2">
                  <Loader2 class="text-primary h-8 w-8 animate-spin" />
                  <p class="text-sm font-bold text-gray-500">Sedang memuat data log...</p>
                </div>
              </TableCell>
            </TableRow>
          {:else if logs.length === 0}
            <TableRow>
              <TableCell colspan={5} class="text-muted-foreground py-16 text-center font-medium">
                Tidak ada log riwayat aktivitas yang ditemukan.
              </TableCell>
            </TableRow>
          {:else}
            {#each logs as log (log.id)}
              <TableRow class="transition-colors hover:bg-gray-50/30">
                <TableCell class="px-6 py-4">
                  <div class="flex items-center gap-2 text-xs font-bold text-gray-500">
                    <Clock class="h-3.5 w-3.5 shrink-0 text-gray-400" />
                    {formatIndonesianDate(log.created_at)}
                  </div>
                </TableCell>
                <TableCell class="text-dark px-6 py-4 text-sm font-bold">
                  <div class="flex items-center gap-2">
                    <div
                      class="flex h-7 w-7 shrink-0 items-center justify-center rounded-lg border bg-gray-100 text-gray-600"
                    >
                      <User class="h-4 w-4" />
                    </div>
                    <div>
                      <p class="leading-tight">{log.actor_name}</p>
                      {#if log.actor_id}<p class="mt-0.5 text-[10px] font-bold text-gray-400">
                          ID: #{log.actor_id}
                        </p>{/if}
                    </div>
                  </div>
                </TableCell>
                <TableCell class="px-6 py-4">
                  <Badge
                    variant="outline"
                    class="rounded-lg border px-2 py-0.5 text-[10px] font-black tracking-wider {actionBadgeColor(
                      log.action_type,
                    )}">{log.action_type}</Badge
                  >
                </TableCell>
                <TableCell
                  class="max-w-xs truncate px-6 py-4 text-sm font-semibold text-gray-600 xl:max-w-md"
                  title={log.description}
                >
                  {log.description || '-'}
                </TableCell>
                <TableCell class="px-6 py-4 text-center">
                  {#if log.payload}
                    <Button
                      variant="ghost"
                      size="icon"
                      class="text-maroon h-8 w-8 rounded-lg transition-colors hover:bg-red-50 hover:text-red-700"
                      onclick={() => {
                        selectedPayload = log.payload;
                        isDialogOpen = true;
                      }}
                      title="Lihat Detail JSON Payload"
                    >
                      <Eye class="h-4 w-4" />
                    </Button>
                  {:else}
                    <span class="text-xs font-bold text-gray-400">-</span>
                  {/if}
                </TableCell>
              </TableRow>
            {/each}
          {/if}
        </TableBody>
      </Table>
    </CardContent>
  </Card>

  {#if !loading && logs.length > 0}
    <div
      class="flex items-center justify-between rounded-2xl border border-gray-100 bg-white p-4 shadow-sm"
    >
      <p class="text-xs font-bold text-gray-400">Halaman {page}</p>
      <div class="flex gap-2">
        <Button
          variant="outline"
          size="sm"
          class="rounded-xl border-gray-200 font-bold"
          onclick={() => (page = Math.max(page - 1, 1))}
          disabled={page === 1}
        >
          <ChevronLeft class="mr-1 h-4 w-4" /> Sebelumnya
        </Button>
        <Button
          variant="outline"
          size="sm"
          class="rounded-xl border-gray-200 font-bold"
          onclick={() => (page += 1)}
          disabled={!hasMore}
        >
          Selanjutnya <ChevronRight class="ml-1 h-4 w-4" />
        </Button>
      </div>
    </div>
  {/if}

  <Dialog bind:open={isDialogOpen}>
    <DialogContent
      class="shadow-bubbly-dark rounded-2xl border-2 border-gray-100 bg-[#FDFBF7] sm:max-w-[600px]"
    >
      <DialogHeader>
        <DialogTitle
          class="text-dark flex items-center gap-2 text-xl font-black tracking-tight uppercase"
        >
          📂 Detail Payload Data
        </DialogTitle>
        <DialogDescription class="text-sm font-semibold text-gray-500">
          Data JSON payload dari aksi yang terekam dalam log aktivitas.
        </DialogDescription>
      </DialogHeader>
      <div
        class="max-h-[350px] overflow-y-auto rounded-xl border border-gray-200 bg-white/80 p-4 font-mono text-xs leading-relaxed text-gray-700 shadow-inner"
      >
        <pre class="break-all whitespace-pre-wrap">{selectedPayload
            ? JSON.stringify(selectedPayload, null, 2)
            : 'Tidak ada data.'}</pre>
      </div>
      <DialogFooter>
        <Button
          onclick={() => (isDialogOpen = false)}
          class="w-full rounded-xl border-2 border-slate-600 bg-slate-500 font-bold text-white duration-200 hover:-translate-y-0.5 hover:bg-slate-600"
        >
          Tutup
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</div>
