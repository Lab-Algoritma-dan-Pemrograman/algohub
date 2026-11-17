<script lang="ts">
  import { siakadKeys } from '@algohub/api-client/siakad';
  import {
    Badge,
    Button,
    Card,
    CardContent,
    CardDescription,
    CardHeader,
    CardTitle,
    Dialog,
    DialogContent,
    DialogDescription,
    DialogHeader,
    DialogTitle,
    Label,
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
  } from '@algohub/ui';
  import Briefcase from '@lucide/svelte/icons/briefcase';
  import Calendar from '@lucide/svelte/icons/calendar';
  import DownloadCloud from '@lucide/svelte/icons/cloud-download';
  import Filter from '@lucide/svelte/icons/filter';
  import RefreshCw from '@lucide/svelte/icons/refresh-cw';
  import RotateCcw from '@lucide/svelte/icons/rotate-ccw';
  import Trash2 from '@lucide/svelte/icons/trash-2';
  import Users from '@lucide/svelte/icons/users';
  import Wand2 from '@lucide/svelte/icons/wand-sparkles';
  import { createQuery } from '@tanstack/svelte-query';
  import { queryClient } from '#lib/query.ts';
  import {
    createRpcMutation,
    fetchRpc,
    invalidateRpcs,
    rpcQueryOptions,
  } from '#lib/queries/rpc.ts';
  import { confirm } from '#lib/stores/confirm.svelte.ts';
  import { toast } from '#lib/toast.ts';
  import type { PageProps } from './$types';

  interface Schedule {
    id: number;
    title: string;
    type?: string;
    status?: string;
    major?: string;
    class_code?: string;
    day_of_week?: string;
    start_time: string;
    end_time: string;
  }

  interface AssignedRow {
    id: number;
    assistant_id: number;
    assistant_name?: string;
  }

  interface MemberRow {
    id: number;
    student_id: number;
    assistant_id: number | string | null;
    student_name?: string;
    student_nim?: string;
    student_class_code?: string | null;
    assistant_name?: string;
  }

  interface AvailableRow {
    user_id: number;
    user_full_name?: string;
  }

  let { data }: PageProps = $props();
  const user = $derived(data.auth.user);

  const accessQuery = createQuery(() =>
    rpcQueryOptions<boolean>('check_menu_access_secure', {
      p_viewer_id: user.id,
      p_menu_key: '/manajemen-kelas',
    }),
  );
  const hasEditAccess = $derived(!accessQuery.error && !!accessQuery.data);
  const schedulesQuery = createQuery(() =>
    rpcQueryOptions<Schedule[]>('get_schedules_secure', { p_viewer_id: user.id }),
  );
  const schedules = $derived(
    (schedulesQuery.data || []).filter((s) => s.type === 'praktikum' && s.status === 'approved'),
  );

  let selectedGroup = $state<Schedule | null>(null);
  let isDialogOpen = $state(false);
  let loading = $state(false);
  let selectedAssistantId = $state('');
  let filterDay = $state('all');
  let filterMajor = $state('all');
  let distributionMode = $state('selang-seling');

  const groupId = $derived(selectedGroup?.id ?? 0);
  const membersParams = $derived({ p_viewer_id: user.id, p_schedule_id: groupId });
  const assignedQuery = createQuery(() =>
    rpcQueryOptions<AssignedRow[]>(
      'get_group_assistants_secure',
      { p_viewer_id: user.id, p_schedule_id: groupId },
      { enabled: !!selectedGroup },
    ),
  );
  const membersQuery = createQuery(() =>
    rpcQueryOptions<MemberRow[]>('get_group_members_secure', membersParams, {
      enabled: !!selectedGroup,
    }),
  );
  const availableQuery = createQuery(() =>
    rpcQueryOptions<AvailableRow[]>(
      'get_assistant_availability_secure',
      {
        p_viewer_id: user.id,
        p_day: selectedGroup?.day_of_week,
        p_start: selectedGroup?.start_time,
        p_end: selectedGroup?.end_time,
      },
      { enabled: !!selectedGroup },
    ),
  );

  $effect(() => {
    const error = assignedQuery.error ?? membersQuery.error ?? availableQuery.error;
    if (error) toast.error('Terjadi kesalahan sistem: ' + error.message);
  });

  const assignedAssistants = $derived(
    (assignedQuery.data || []).map((a) => ({ ...a, users: { full_name: a.assistant_name } })),
  );
  const students = $derived(
    (membersQuery.data || [])
      .filter(
        (s) =>
          !s.student_class_code ||
          s.student_class_code === selectedGroup?.class_code ||
          s.student_class_code.endsWith(' ' + selectedGroup?.class_code),
      )
      .map((s) => ({
        ...s,
        users: { full_name: s.student_name, nim: s.student_nim },
        assistant: { full_name: s.assistant_name },
      })),
  );
  const availableAssistants = $derived.by(() => {
    const assignedIds = new Set(assignedAssistants.map((a) => a.assistant_id));
    return (availableQuery.data || [])
      .filter((a) => !assignedIds.has(a.user_id))
      .map((a) => ({ ...a, users: { full_name: a.user_full_name } }));
  });

  const detailFns = [
    'get_group_assistants_secure',
    'get_group_members_secure',
    'get_assistant_availability_secure',
  ];
  const refreshDetails = () => invalidateRpcs(detailFns);
  const syncStudents = createRpcMutation('sync_students_to_group_secure', detailFns);
  const updateMembers = createRpcMutation('update_group_members_batch_secure');
  const resetPlotting = createRpcMutation('reset_group_plotting_secure', detailFns);
  const upsertAssistant = createRpcMutation('upsert_group_assistant_secure', detailFns);
  const deleteAssistant = createRpcMutation('delete_group_assistant_secure', detailFns);

  function openGroup(schedule: Schedule) {
    selectedGroup = schedule;
    isDialogOpen = true;
    void refreshDetails();
  }

  async function handleSyncStudents() {
    if (!selectedGroup) return;
    if (!hasEditAccess) return toast.error('Akses Ditolak');
    loading = true;
    try {
      const { error } = await syncStudents.mutateAsync({
        p_caller_id: user.id,
        p_schedule_id: selectedGroup.id,
        p_major: selectedGroup.major,
        p_class_code: selectedGroup.class_code,
      });
      if (error) throw error;
      toast.success(`Proses sinkronisasi praktikan selesai!`);
    } catch (err) {
      toast.error('Gagal sync: ' + (err as Error).message);
    } finally {
      loading = false;
    }
  }

  async function handleDistributeStudents() {
    if (assignedAssistants.length === 0) return toast.error('Masukkan tim asisten dulu!');
    if (students.length === 0) return toast.error('Sync data mahasiswa dulu!');
    if (!hasEditAccess) return toast.error('Akses Ditolak');
    loading = true;
    try {
      const assistantIds = assignedAssistants.map((a) => a.assistant_id);
      const groupSize = Math.ceil(students.length / assistantIds.length);
      const updates = students.map((s, index) => ({
        id: s.id,
        assistant_id:
          distributionMode === 'selang-seling'
            ? assistantIds[index % assistantIds.length]
            : assistantIds[Math.min(Math.floor(index / groupSize), assistantIds.length - 1)],
      }));
      const { error } = await updateMembers.mutateAsync({
        p_caller_id: user.id,
        p_updates: updates,
      });
      if (error) throw error;
      toast.success('Mahasiswa berhasil dibagi rata!');
      void refreshDetails();
    } catch (err) {
      toast.error((err as Error).message);
    } finally {
      loading = false;
    }
  }

  async function handleResetPlotting() {
    if (
      !(await confirm(
        'Yakin ingin menghapus semua pembagian asisten? Data mahasiswa tidak akan hilang.',
      ))
    )
      return;
    if (!hasEditAccess) return toast.error('Akses Ditolak');
    loading = true;
    try {
      const { error } = await resetPlotting.mutateAsync({
        p_caller_id: user.id,
        p_schedule_id: selectedGroup?.id,
      });
      if (error) throw error;
      toast.success('Plotting berhasil di-reset!');
    } catch (err) {
      toast.error((err as Error).message);
    } finally {
      loading = false;
    }
  }

  async function handleUpdateStudentAssistant(memberId: number, newAssistantId: string) {
    if (!hasEditAccess) return toast.error('Akses Ditolak');
    queryClient.setQueryData<MemberRow[]>(
      siakadKeys.rpc('get_group_members_secure', membersParams),
      (rows) => rows?.map((s) => (s.id === memberId ? { ...s, assistant_id: newAssistantId } : s)),
    );
    const { error } = await updateMembers.mutateAsync({
      p_caller_id: user.id,
      p_updates: [
        {
          id: memberId,
          assistant_id: newAssistantId === 'unassigned' ? null : parseInt(newAssistantId),
        },
      ],
    });
    if (error) {
      toast.error('Gagal update: ' + error.message);
      void refreshDetails();
    } else {
      toast.success('Asisten diperbarui');
    }
  }

  async function handleAddAssistant() {
    if (!selectedAssistantId) return;
    if (!hasEditAccess) return toast.error('Akses Ditolak');
    const { error } = await upsertAssistant.mutateAsync({
      p_caller_id: user.id,
      p_schedule_id: selectedGroup?.id,
      p_assistant_id: parseInt(selectedAssistantId),
    });
    if (error) toast.error(error.message);
    else selectedAssistantId = '';
  }

  async function handleDeleteAssistant(id: number) {
    const { error } = await deleteAssistant.mutateAsync({ p_caller_id: user.id, p_id: id });
    if (error) toast.error(error.message);
  }

  const uniq = (values: (string | undefined)[]) =>
    [...new Set(values.filter((v): v is string => !!v))].sort();
  const uniqueDays = $derived(uniq(schedules.map((s) => s.day_of_week)));
  const uniqueMajors = $derived(uniq(schedules.map((s) => s.major)));
  const filteredSchedules = $derived(
    schedules.filter(
      (item) =>
        (filterDay === 'all' || item.day_of_week === filterDay) &&
        (filterMajor === 'all' || item.major === filterMajor),
    ),
  );

  async function handleExportSemuaKelompok() {
    // T-30: pembungkus toast lama tidak meneruskan loading(), jadi di prod baris ini melempar
    // TypeError dan file tidak pernah diunduh. Dibiarkan sama sampai temuan diputuskan.
    const loadingToast = (toast as unknown as { loading: (m: string) => string }).loading(
      'Menyiapkan file Excel Keseluruhan...',
    );
    try {
      const rows = await fetchRpc<Record<string, string>[]>('get_all_group_members_secure', {
        p_viewer_id: user.id,
      });
      if (!rows || rows.length === 0) {
        toast.dismiss(loadingToast);
        return toast.error('Data plotting belum tersedia.');
      }
      const XLSX = await import('xlsx');
      const excelData = rows.map((row) => ({
        Jurusan: row.schedule_major || '-',
        Kelas: row.schedule_class_code || '-',
        Hari: row.schedule_day || '-',
        Jam: `${row.schedule_start?.slice(0, 5)} - ${row.schedule_end?.slice(0, 5)}`,
        'Modul Praktikum': row.schedule_title || '-',
        'Asisten Pembimbing': row.assistant_name || 'Belum ditentukan',
        'NIM Praktikan': row.student_nim || '-',
        'Nama Praktikan': row.student_name || '-',
      }));
      excelData.sort((a, b) => {
        if (a.Jurusan !== b.Jurusan) return a.Jurusan.localeCompare(b.Jurusan);
        if (a.Kelas !== b.Kelas) return a.Kelas.localeCompare(b.Kelas);
        return a['Asisten Pembimbing'].localeCompare(b['Asisten Pembimbing']);
      });
      const worksheet = XLSX.utils.json_to_sheet(
        excelData.map((row, idx) => ({ No: idx + 1, ...row })),
      );
      worksheet['!cols'] = [5, 25, 10, 10, 15, 25, 20, 15, 30].map((wch) => ({ wch }));
      const workbook = XLSX.utils.book_new();
      XLSX.utils.book_append_sheet(workbook, worksheet, 'Seluruh Plotting');
      XLSX.writeFile(
        workbook,
        `Data_Seluruh_Plotting_${new Date().toISOString().split('T')[0]}.xlsx`,
      );
      toast.dismiss(loadingToast);
      toast.success('Excel Keseluruhan diunduh.');
    } catch (err) {
      toast.dismiss(loadingToast);
      toast.error('Gagal Export: ' + (err as Error).message);
    }
  }

  async function handleExportPerKelas() {
    if (!selectedGroup || students.length === 0) {
      return toast.error('Data mahasiswa di kelas ini masih kosong.');
    }
    const XLSX = await import('xlsx');
    const cleanData = students.map((s, idx) => ({
      No: idx + 1,
      'Asisten Pembimbing': s.assistant.full_name || 'Belum ditentukan',
      'NIM Praktikan': s.users.nim || '-',
      'Nama Praktikan': s.users.full_name || '-',
    }));
    cleanData.sort((a, b) =>
      a['Asisten Pembimbing'] !== b['Asisten Pembimbing']
        ? a['Asisten Pembimbing'].localeCompare(b['Asisten Pembimbing'])
        : a['Nama Praktikan'].localeCompare(b['Nama Praktikan']),
    );
    const worksheet = XLSX.utils.json_to_sheet(cleanData);
    worksheet['!cols'] = [{ wch: 5 }, { wch: 25 }, { wch: 15 }, { wch: 30 }];
    const workbook = XLSX.utils.book_new();
    XLSX.utils.book_append_sheet(workbook, worksheet, `Kelas ${selectedGroup.class_code}`);
    XLSX.writeFile(
      workbook,
      `Plotting_Kelas_${selectedGroup.class_code}_${selectedGroup.major}.xlsx`,
    );
    toast.success(`Excel Kelas ${selectedGroup.class_code} diunduh.`);
  }

  const modeLabels: Record<string, string> = {
    'selang-seling': 'Selang-seling',
    berurutan: 'Satu Kelompok Urt',
  };
</script>

{#snippet filterSelect(
  value: string,
  set: (v: string) => void,
  width: string,
  allLabel: string,
  options: string[],
)}
  <Select type="single" {value} onValueChange={set}>
    <SelectTrigger class="h-9 {width}">{value === 'all' ? allLabel : value}</SelectTrigger>
    <SelectContent>
      <SelectItem value="all" label={allLabel} />
      {#each options as option (option)}<SelectItem value={option} label={option} />{/each}
    </SelectContent>
  </Select>
{/snippet}

<div class="space-y-6">
  <div class="flex flex-col items-start justify-between gap-4 md:flex-row md:items-center">
    <div>
      <h1 class="text-primary flex items-center gap-2 text-2xl font-bold">
        <Users class="h-6 w-6" /> Plotting Asisten & Praktikan
      </h1>
      <p class="text-muted-foreground mt-1 text-sm">
        Kelola pembagian mahasiswa berdasarkan jadwal yang aktif.
      </p>
    </div>
    <div class="flex flex-wrap gap-2">
      {#if hasEditAccess}
        <Button
          variant="outline"
          class="border-green-600 text-green-700 hover:bg-green-50"
          onclick={handleExportSemuaKelompok}
        >
          <DownloadCloud class="mr-2 h-4 w-4" /> Export Semua Plotting
        </Button>
      {/if}
    </div>
  </div>

  {#if schedules.length > 0}
    <div class="flex flex-wrap items-center gap-3 rounded-lg border bg-white p-3 shadow-sm">
      <div class="text-muted-foreground mr-2 ml-1 flex items-center text-sm font-medium">
        <Filter class="mr-2 h-4 w-4" /> Filter:
      </div>
      {@render filterSelect(
        filterDay,
        (v) => (filterDay = v),
        'w-[140px]',
        'Semua Hari',
        uniqueDays,
      )}
      {@render filterSelect(
        filterMajor,
        (v) => (filterMajor = v),
        'w-[200px]',
        'Semua Jurusan',
        uniqueMajors,
      )}
      {#if filterDay !== 'all' || filterMajor !== 'all'}
        <Button
          variant="ghost"
          size="sm"
          onclick={() => {
            filterDay = 'all';
            filterMajor = 'all';
          }}
          class="text-red-500 hover:bg-red-50">Reset</Button
        >
      {/if}
    </div>
  {/if}

  <div class="grid grid-cols-1 gap-4 lg:grid-cols-2 xl:grid-cols-3">
    {#if filteredSchedules.length === 0}
      <div
        class="text-muted-foreground col-span-3 rounded border border-dashed bg-white py-10 text-center"
      >
        <p>Tidak ada jadwal yang sesuai.</p>
      </div>
    {:else}
      {#each filteredSchedules as item (item.id)}
        <Card class="group border-l-primary relative border-l-4 transition-shadow hover:shadow-md">
          <CardHeader class="pb-2">
            <CardTitle class="line-clamp-1 text-base font-bold" title={item.title}>
              {item.title}
            </CardTitle>
            <CardDescription class="mt-2 space-y-1 text-xs">
              <div class="flex items-center justify-between">
                <Badge variant="outline" class="border-blue-200 bg-blue-50 text-blue-700"
                  >{item.major}</Badge
                >
                <span class="rounded bg-gray-100 px-2 py-0.5 font-bold text-gray-700"
                  >Kelas {item.class_code}</span
                >
              </div>
              <div class="flex items-center gap-1 pt-1 font-medium text-gray-600">
                <Calendar class="h-3 w-3" />
                {item.day_of_week}, {item.start_time.slice(0, 5)} - {item.end_time.slice(0, 5)}
              </div>
            </CardDescription>
          </CardHeader>
          <CardContent>
            <Button
              size="sm"
              class="text-primary border-primary hover:bg-primary w-full border bg-white transition-colors hover:text-white"
              onclick={() => openGroup(item)}
            >
              <Wand2 class="mr-2 h-4 w-4" /> Kelola Plotting
            </Button>
          </CardContent>
        </Card>
      {/each}
    {/if}
  </div>

  <Dialog
    bind:open={isDialogOpen}
    onOpenChange={(open) => {
      if (!open) selectedGroup = null;
    }}
  >
    {#if selectedGroup}
      <DialogContent class="flex h-[85vh] max-w-6xl flex-col overflow-hidden p-0 sm:max-w-6xl">
        <DialogHeader
          class="flex flex-row items-center justify-between border-b bg-gray-50 px-6 py-4"
        >
          <DialogTitle class="flex flex-col gap-1">
            <span>Plotting: {selectedGroup.title}</span>
            <span class="text-muted-foreground text-xs font-normal"
              >{selectedGroup.major} - Kelas {selectedGroup.class_code}</span
            >
          </DialogTitle>
          <DialogDescription class="sr-only">
            Kelola pembagian asisten dan praktikan untuk kelompok praktikum ini.
          </DialogDescription>
          {#if hasEditAccess && students.length > 0}
            <Button
              size="sm"
              variant="outline"
              class="h-8 border-green-600 text-green-700 hover:bg-green-50"
              onclick={handleExportPerKelas}
            >
              <DownloadCloud class="mr-2 h-4 w-4" /> Download Kelompok
            </Button>
          {/if}
        </DialogHeader>
        <div class="grid flex-1 grid-cols-1 overflow-hidden md:grid-cols-3">
          <div class="flex flex-col overflow-y-auto border-r bg-gray-50/50 p-4">
            <h3 class="mb-3 flex items-center gap-2 text-sm font-bold">
              <Briefcase class="h-4 w-4" /> Tim Asisten ({assignedAssistants.length})
            </h3>
            <div class="mb-4 flex-1 space-y-2">
              {#if assignedAssistants.length === 0}
                <p class="text-muted-foreground text-xs italic">Belum ada tim.</p>
              {:else}
                {#each assignedAssistants as a (a.id)}
                  <div
                    class="flex items-center justify-between rounded border bg-white p-2 text-sm shadow-sm"
                  >
                    <span>{a.users.full_name}</span>
                    {#if hasEditAccess}
                      <button
                        type="button"
                        aria-label="Hapus asisten"
                        onclick={() => handleDeleteAssistant(a.id)}
                        ><Trash2 class="h-3 w-3 cursor-pointer text-red-400" /></button
                      >
                    {/if}
                  </div>
                {/each}
              {/if}
            </div>
            {#if hasEditAccess}
              <div class="mt-auto space-y-2 border-t pt-4">
                <Label class="text-xs">Tambah Asisten (Hanya yg Free)</Label>
                <Select type="single" bind:value={selectedAssistantId}>
                  <SelectTrigger class="h-9 w-full text-xs">
                    {#if selectedAssistantId}
                      {availableAssistants.find((a) => String(a.user_id) === selectedAssistantId)
                        ?.users.full_name}
                    {:else}<span class="text-muted-foreground">Pilih Asisten</span>{/if}
                  </SelectTrigger>
                  <SelectContent>
                    {#if availableAssistants.length === 0}
                      <div class="p-2 text-center text-xs text-red-500">Tidak ada yg free :(</div>
                    {:else}
                      {#each availableAssistants as av (av.user_id)}
                        <SelectItem value={String(av.user_id)} label={av.users.full_name} />
                      {/each}
                    {/if}
                  </SelectContent>
                </Select>
                <Button
                  size="sm"
                  class="w-full"
                  onclick={handleAddAssistant}
                  disabled={!selectedAssistantId}>Tambah Tim</Button
                >
              </div>
            {/if}
          </div>
          <div class="flex flex-col overflow-hidden bg-white p-4 md:col-span-2">
            <div class="mb-4 flex items-center justify-between border-b pb-4">
              <div>
                <h3 class="flex items-center gap-2 text-sm font-bold">
                  <Users class="h-4 w-4" /> Mahasiswa ({students.length})
                </h3>
                <p class="text-muted-foreground text-xs">Plotting manual atau otomatis</p>
              </div>
              {#if hasEditAccess}
                <div class="flex flex-wrap items-center gap-2">
                  <Button
                    size="sm"
                    variant="destructive"
                    class="h-8 text-xs"
                    onclick={handleResetPlotting}
                    disabled={loading || students.length === 0}
                  >
                    <RotateCcw class="mr-1 h-3 w-3" /> Reset
                  </Button>
                  <Button
                    size="sm"
                    variant="outline"
                    class="h-8 text-xs"
                    onclick={handleSyncStudents}
                    disabled={loading}
                  >
                    <RefreshCw class="mr-1 h-3 w-3 {loading ? 'animate-spin' : ''}" /> Sync User
                  </Button>
                  <div class="flex rounded border bg-gray-100">
                    <Select type="single" bind:value={distributionMode}>
                      <SelectTrigger
                        class="h-8 w-[140px] border-none bg-transparent text-xs shadow-none focus:ring-0"
                        >{modeLabels[distributionMode]}</SelectTrigger
                      >
                      <SelectContent>
                        {#each Object.entries(modeLabels) as [value, label] (value)}
                          <SelectItem {value} {label} />
                        {/each}
                      </SelectContent>
                    </Select>
                    <Button
                      size="sm"
                      class="h-8 rounded-l-none bg-blue-600 text-xs hover:bg-blue-700"
                      onclick={handleDistributeStudents}
                      disabled={loading || students.length === 0}
                    >
                      <Wand2 class="mr-1 h-3 w-3" /> Auto Bagi
                    </Button>
                  </div>
                </div>
              {/if}
            </div>
            <div class="flex-1 overflow-x-auto overflow-y-auto rounded-md border">
              <table class="w-full text-left text-sm">
                <thead class="sticky top-0 z-10 bg-gray-100">
                  <tr
                    ><th class="p-3 text-xs font-semibold text-gray-500 uppercase"
                      >Nama Mahasiswa</th
                    ><th class="p-3 text-xs font-semibold text-gray-500 uppercase"
                      >Pembimbing (Edit)</th
                    ></tr
                  >
                </thead>
                <tbody class="divide-y">
                  {#if students.length === 0}
                    <tr
                      ><td colspan={2} class="p-8 text-center text-gray-400"
                        >Data masih kosong.<br />Klik tombol 'Sync User'.</td
                      ></tr
                    >
                  {:else}
                    {#each students as s (s.id)}
                      {@const asstIdStr = s.assistant_id ? String(s.assistant_id) : 'unassigned'}
                      {@const isAssistantInList = assignedAssistants.some(
                        (a) => String(a.assistant_id) === asstIdStr,
                      )}
                      <tr class="hover:bg-gray-50">
                        <td class="p-3">
                          <div class="font-medium">
                            {s.users.full_name || 'Nama tidak ditemukan'}
                          </div>
                          <div class="text-xs text-gray-400">{s.users.nim || '-'}</div>
                        </td>
                        <td class="w-[250px] p-3">
                          {#if hasEditAccess}
                            <Select
                              type="single"
                              value={asstIdStr}
                              onValueChange={(val) => handleUpdateStudentAssistant(s.id, val)}
                            >
                              <SelectTrigger class="h-8 w-full border-gray-200 bg-white text-xs">
                                {asstIdStr === 'unassigned'
                                  ? '-- Lepas Pembimbing --'
                                  : (assignedAssistants.find(
                                      (a) => String(a.assistant_id) === asstIdStr,
                                    )?.users.full_name ??
                                    (s.assistant.full_name || `Asisten ID #${s.assistant_id}`))}
                              </SelectTrigger>
                              <SelectContent>
                                <SelectItem
                                  value="unassigned"
                                  label="-- Lepas Pembimbing --"
                                  class="font-medium text-red-500"
                                />
                                {#each assignedAssistants as a (a.assistant_id)}
                                  <SelectItem
                                    value={String(a.assistant_id)}
                                    label={a.users.full_name}
                                  />
                                {/each}
                                {#if !isAssistantInList && s.assistant_id}
                                  <SelectItem
                                    value={asstIdStr}
                                    label={s.assistant.full_name || `Asisten ID #${s.assistant_id}`}
                                  />
                                {/if}
                              </SelectContent>
                            </Select>
                          {:else}
                            <Badge
                              variant={s.assistant_id ? 'secondary' : 'outline'}
                              class={!s.assistant_id ? 'text-red-400' : ''}
                            >
                              {s.assistant.full_name || 'Belum diplot'}
                            </Badge>
                          {/if}
                        </td>
                      </tr>
                    {/each}
                  {/if}
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </DialogContent>
    {/if}
  </Dialog>
</div>
