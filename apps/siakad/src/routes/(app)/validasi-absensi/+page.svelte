<script lang="ts">
  import { siakadKeys } from '@algohub/api-client/siakad';
  import {
    Badge,
    Button,
    Card,
    CardContent,
    CardHeader,
    CardTitle,
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
  } from '@algohub/ui';
  import CheckCircle from '@lucide/svelte/icons/circle-check-big';
  import Clock from '@lucide/svelte/icons/clock';
  import History from '@lucide/svelte/icons/history';
  import Loader2 from '@lucide/svelte/icons/loader-2';
  import MessageCircle from '@lucide/svelte/icons/message-circle';
  import RotateCcw from '@lucide/svelte/icons/rotate-ccw';
  import XCircle from '@lucide/svelte/icons/circle-x';
  import { createQuery } from '@tanstack/svelte-query';
  import { queryClient } from '#lib/query.ts';
  import { createRpcMutation, invalidateRpcs, rpcQueryOptions } from '#lib/queries/rpc.ts';
  import { getDailyQuote } from '#lib/quotes.ts';
  import { subscribe } from '#lib/realtime/client.ts';
  import { canonRole } from '#lib/roles.ts';
  import { confirm } from '#lib/stores/confirm.svelte.ts';
  import { toast } from '#lib/toast.ts';
  import type { PageProps } from './$types';

  interface LogRow {
    id: number;
    status: string;
    notes?: string;
    check_in_time: string;
    verification_status?: string;
    is_verified?: boolean;
    reschedule_status?: string | null;
    reschedule_schedule_id?: number | null;
    schedule_title?: string;
    schedule_day?: string;
    schedule_time?: string;
    user_full_name?: string;
    user_role?: string;
    user_phone_number?: string | null;
  }

  let { data }: PageProps = $props();
  const user = $derived(data.auth.user);
  const params = $derived({ p_viewer_id: user.id });

  const logsQuery = createQuery(() =>
    rpcQueryOptions<LogRow[]>('get_attendance_logs_secure', params),
  );
  const verify = createRpcMutation('admin_verify_attendance_secure');

  $effect(() =>
    subscribe('validasi_live_updates', { table: 'attendance_logs' }, () => {
      void invalidateRpcs(['get_attendance_logs_secure']);
    }),
  );

  const formatted = $derived(
    (logsQuery.data || []).map((log) => ({
      ...log,
      users: {
        full_name: log.user_full_name,
        role: log.user_role,
        phone_number: log.user_phone_number,
      },
      schedules: log.reschedule_schedule_id
        ? {
            id: log.reschedule_schedule_id,
            title: log.schedule_title,
            day_of_week: log.schedule_day,
            start_time: log.schedule_time,
          }
        : null,
    })),
  );
  // Pending hanya permintaan izin (bukan Hadir, belum memilih jadwal pengganti); riwayat juga
  // memuat log dengan reschedule yang masih menunggu.
  const pendingLogs = $derived(
    formatted.filter(
      (l) =>
        l.verification_status === 'pending' &&
        canonRole(l.user_role) === 'praktikan' &&
        !l.reschedule_schedule_id &&
        l.status !== 'Hadir',
    ),
  );
  const historyLogs = $derived(
    formatted
      .filter(
        (l) =>
          canonRole(l.user_role) === 'praktikan' &&
          (['approved', 'rejected'].includes(l.verification_status ?? '') ||
            l.reschedule_status === 'pending'),
      )
      .slice(0, 50),
  );
  const loading = $derived(logsQuery.isFetching);
  const isStaff = $derived(['koordinator', 'asisten'].includes(user.role || ''));

  function getWaLink(phone: string | null | undefined, targetName?: string, isApproved?: boolean) {
    if (!phone) return '#';
    const cleanPhone = phone.replace(/[^0-9]/g, '').replace(/^0/, '62');
    const quote = getDailyQuote('praktikan');
    const result =
      isApproved === undefined
        ? 'telah diproses'
        : `telah ${isApproved ? 'DISETUJUI ✅' : 'DITOLAK ❌'}`;
    const msg = `Halo *${targetName || 'Praktikan'}*,\n\nPermohonan reschedule praktikum kamu ${result}.\n\n🌐 *Buka Web Lab AP:* https://www.lab-ap.web.id/absensi\n\n✨ "${quote}"`;
    return `https://wa.me/${cleanPhone}?text=${encodeURIComponent(msg)}`;
  }

  async function handleVerifyLicense(id: number, isApproved: boolean) {
    if (!isStaff) {
      toast.error('Akses Ditolak', {
        description: 'Anda tidak memiliki izin untuk memvalidasi izin.',
      });
      return;
    }
    const newStatus = isApproved ? 'approved' : 'rejected';
    // Optimistic supaya klik terasa responsif; dibatalkan dengan refetch bila gagal.
    queryClient.setQueryData<LogRow[]>(
      siakadKeys.rpc('get_attendance_logs_secure', params),
      (rows) =>
        rows?.map((l) =>
          l.id === id ? { ...l, verification_status: newStatus, is_verified: isApproved } : l,
        ),
    );
    const { error } = await verify.mutateAsync({
      p_caller_id: user.id,
      p_log_id: id,
      p_is_approved: isApproved,
      p_type: 'license',
    });
    if (error) {
      toast.error('Gagal memvalidasi: ' + error.message);
      void invalidateRpcs(['get_attendance_logs_secure']);
    } else {
      toast.success(isApproved ? 'Izin Disetujui. Praktikan bisa pilih jadwal.' : 'Izin Ditolak.');
    }
  }

  async function handleVerifyReschedule(id: number, isApproved: boolean) {
    if (!isStaff) {
      toast.error('Akses Ditolak');
      return;
    }
    const { error } = await verify.mutateAsync({
      p_caller_id: user.id,
      p_log_id: id,
      p_is_approved: isApproved,
      p_type: 'reschedule',
    });
    if (error) {
      toast.error('Gagal memproses jadwal: ' + error.message);
    } else {
      toast.success(isApproved ? 'Jadwal Disetujui & Terkunci.' : 'Jadwal Ditolak.');
      void invalidateRpcs(['get_attendance_logs_secure']);
    }
  }

  async function handleResetReschedule(id: number) {
    if (!(await confirm('Buka kunci jadwal ini?'))) return;
    if (!isStaff) {
      toast.error('Akses Ditolak');
      return;
    }
    // Reset diperlakukan sebagai menolak reschedule supaya jadwalnya kosong lagi.
    const { error } = await verify.mutateAsync({
      p_caller_id: user.id,
      p_log_id: id,
      p_is_approved: false,
      p_type: 'reschedule',
    });
    if (error) {
      toast.error('Gagal me-reset: ' + error.message);
    } else {
      toast.success('Jadwal di-reset.');
      void invalidateRpcs(['get_attendance_logs_secure']);
    }
  }
</script>

<div class="space-y-8">
  <div class="flex flex-col justify-between gap-4 md:flex-row md:items-center">
    <h1 class="flex items-center gap-2 text-2xl font-bold">
      <CheckCircle class="text-primary" /> Validasi Absensi
    </h1>
    <Badge
      variant="outline"
      class="animate-pulse border-green-200 bg-green-50 py-1 text-sm text-green-600"
    >
      <span class="mr-2 inline-block h-2 w-2 rounded-full bg-green-500"></span> Live Update
    </Badge>
  </div>

  <Card class="border-l-4 border-l-yellow-500 shadow-md">
    <CardHeader
      ><CardTitle class="flex items-center gap-2 text-yellow-700"
        ><Clock class="h-5 w-5" /> Permintaan Izin ({pendingLogs.length})</CardTitle
      ></CardHeader
    >
    <CardContent>
      {#if loading && pendingLogs.length === 0}
        <div class="flex justify-center p-4">
          <Loader2 class="text-muted-foreground h-6 w-6 animate-spin" />
        </div>
      {:else}
        <div class="overflow-x-auto">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead class="w-[1%] pr-6 whitespace-nowrap">Nama</TableHead>
                <TableHead class="w-[1%] pr-6 whitespace-nowrap">Alasan</TableHead>
                <TableHead class="w-[1%] pr-6 whitespace-nowrap">Kontak</TableHead>
                <TableHead class="w-[160px] text-center">Aksi</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {#if pendingLogs.length === 0}
                <TableRow
                  ><TableCell colspan={4} class="text-muted-foreground py-6 text-center italic"
                    >Tidak ada permintaan izin baru.</TableCell
                  ></TableRow
                >
              {:else}
                {#each pendingLogs as log (log.id)}
                  <TableRow data-testid="pending-row">
                    <TableCell>
                      <div class="font-bold">{log.users.full_name}</div>
                      <div class="text-muted-foreground text-xs">
                        {new Date(log.check_in_time).toLocaleString('id-ID', {
                          timeZone: 'Asia/Jakarta',
                        })}
                      </div>
                    </TableCell>
                    <TableCell>
                      <Badge variant="outline" class="mr-2 mb-1 bg-yellow-50">{log.status}</Badge>
                      <span class="text-sm">{log.notes}</span>
                    </TableCell>
                    <TableCell>
                      {#if log.users.phone_number}
                        <a
                          href={getWaLink(log.users.phone_number) || '#'}
                          target="_blank"
                          rel="noreferrer external"
                          ><Button
                            size="sm"
                            variant="outline"
                            class="h-8 border-green-200 bg-green-50 text-green-600 hover:bg-green-100"
                            ><MessageCircle class="mr-2 h-3 w-3" />Chat WA</Button
                          ></a
                        >
                      {:else}
                        <span class="text-muted-foreground text-xs">-</span>
                      {/if}
                    </TableCell>
                    <TableCell class="text-right">
                      <Button
                        size="sm"
                        class="mr-2 bg-green-600 hover:bg-green-700"
                        onclick={() => handleVerifyLicense(log.id, true)}>Setuju</Button
                      >
                      <Button
                        size="sm"
                        variant="destructive"
                        onclick={() => handleVerifyLicense(log.id, false)}>Tolak</Button
                      >
                    </TableCell>
                  </TableRow>
                {/each}
              {/if}
            </TableBody>
          </Table>
        </div>
      {/if}
    </CardContent>
  </Card>

  <Card class="border-l-4 border-l-blue-500 opacity-95 shadow-sm">
    <CardHeader
      ><CardTitle class="flex items-center gap-2 text-blue-700"
        ><History class="h-5 w-5" /> Status Jadwal Pengganti</CardTitle
      ></CardHeader
    >
    <CardContent>
      <div class="overflow-x-auto">
        <Table>
          <TableHeader>
            <TableRow
              ><TableHead>Nama</TableHead><TableHead>Status Izin</TableHead><TableHead
                >Jadwal Dipilih</TableHead
              ><TableHead class="text-right">Aksi Jadwal</TableHead></TableRow
            >
          </TableHeader>
          <TableBody>
            {#if historyLogs.length === 0}
              <TableRow
                ><TableCell colspan={4} class="text-muted-foreground py-6 text-center italic"
                  >Belum ada data riwayat.</TableCell
                ></TableRow
              >
            {:else}
              {#each historyLogs as log (log.id)}
                <TableRow class="bg-gray-50/50" data-testid="history-row">
                  <TableCell>
                    <div class="font-medium">{log.users.full_name}</div>
                    <div class="text-muted-foreground text-xs">
                      {new Date(log.check_in_time).toLocaleDateString('id-ID', {
                        timeZone: 'Asia/Jakarta',
                      })}
                    </div>
                  </TableCell>
                  <TableCell>
                    {#if log.verification_status === 'approved'}
                      <Badge class="border-transparent bg-green-100 text-green-800">Izin OK</Badge>
                    {:else}
                      <Badge class="border-transparent bg-red-100 text-red-800">Ditolak</Badge>
                    {/if}
                  </TableCell>
                  <TableCell>
                    {#if log.verification_status === 'rejected'}
                      <span class="text-muted-foreground">-</span>
                    {:else if !log.reschedule_schedule_id}
                      <span class="text-xs text-orange-600 italic">Belum pilih jadwal</span>
                    {:else}
                      <div class="flex flex-col items-start gap-1">
                        <span class="text-sm font-bold text-blue-700">
                          {log.schedules?.day_of_week}, {log.schedules?.start_time?.slice(0, 5)}
                        </span>
                        <span class="text-muted-foreground text-xs">({log.schedules?.title})</span>
                        {#if log.reschedule_status === 'pending'}<Badge
                            variant="outline"
                            class="w-fit border-yellow-200 bg-yellow-50 text-yellow-700"
                            >Menunggu Acc</Badge
                          >{/if}
                        {#if log.reschedule_status === 'approved'}<Badge
                            variant="outline"
                            class="w-fit border-green-200 bg-green-50 text-green-700"
                            >Disetujui & Fix</Badge
                          >{/if}
                        {#if log.reschedule_status === 'rejected'}<Badge
                            variant="outline"
                            class="w-fit border-red-200 bg-red-50 text-red-700">Ditolak</Badge
                          >{/if}
                      </div>
                    {/if}
                  </TableCell>
                  <TableCell class="text-right">
                    {#if log.verification_status === 'approved' && log.reschedule_schedule_id}
                      {#if log.reschedule_status === 'pending'}
                        <div class="flex justify-end gap-2">
                          <Button
                            size="sm"
                            class="h-8 bg-blue-600 px-3 hover:bg-blue-700"
                            onclick={() => handleVerifyReschedule(log.id, true)}
                            ><CheckCircle class="mr-1 h-4 w-4" /> Acc</Button
                          >
                          <Button
                            size="sm"
                            variant="destructive"
                            class="h-8 px-3"
                            onclick={() => handleVerifyReschedule(log.id, false)}
                            ><XCircle class="mr-1 h-4 w-4" /> Tolak</Button
                          >
                        </div>
                      {/if}
                      {#if log.reschedule_status === 'approved'}
                        <Button
                          size="sm"
                          variant="outline"
                          class="h-8 px-3 text-gray-500"
                          onclick={() => handleResetReschedule(log.id)}
                          ><RotateCcw class="mr-1 h-4 w-4" /> Buka Kunci</Button
                        >
                      {/if}
                    {/if}
                  </TableCell>
                </TableRow>
              {/each}
            {/if}
          </TableBody>
        </Table>
      </div>
    </CardContent>
  </Card>
</div>
