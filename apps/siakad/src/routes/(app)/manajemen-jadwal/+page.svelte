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
    DialogFooter,
    DialogHeader,
    DialogTitle,
    Input,
    Label,
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
  import CalendarDays from '@lucide/svelte/icons/calendar-days';
  import Clock from '@lucide/svelte/icons/clock';
  import Download from '@lucide/svelte/icons/download';
  import FileSpreadsheet from '@lucide/svelte/icons/file-spreadsheet';
  import Filter from '@lucide/svelte/icons/filter';
  import Loader2 from '@lucide/svelte/icons/loader-2';
  import Pencil from '@lucide/svelte/icons/pencil';
  import Plus from '@lucide/svelte/icons/plus';
  import Save from '@lucide/svelte/icons/save';
  import Trash2 from '@lucide/svelte/icons/trash-2';
  import { createQuery } from '@tanstack/svelte-query';
  import { createRpcMutation, invalidateRpcs, rpcQueryOptions } from '#lib/queries/rpc.ts';
  import { confirm } from '#lib/stores/confirm.svelte.ts';
  import { toast } from '#lib/toast.ts';
  import type { PageProps } from './$types';

  interface Schedule {
    id: number;
    day_of_week: string;
    start_time: string;
    end_time: string;
    title: string;
    major?: string;
    class_code?: string;
  }

  const DAYS = ['Senin', 'Selasa', 'Rabu', 'Kamis', 'Jumat', 'Sabtu'];
  const MAJORS = [
    'Teknik Tenaga Listrik',
    'Teknik Elektro',
    'Teknik Sistem Energi',
    'Teknologi Listrik',
  ];
  const MAJOR_BADGE: Record<string, [string, string]> = {
    'Teknik Tenaga Listrik': ['TTL', 'bg-blue-50 text-blue-700 border-blue-200'],
    'Teknik Elektro': ['TE', 'bg-orange-50 text-orange-700 border-orange-200'],
    'Teknik Sistem Energi': ['TSE', 'bg-purple-50 text-purple-700 border-purple-200'],
    'Teknologi Listrik': ['TL', 'bg-green-50 text-green-700 border-green-200'],
  };

  let { data }: PageProps = $props();
  const user = $derived(data.auth.user);
  const isAdminOrSecretary = $derived(
    user.role === 'koordinator' || user.division?.toLowerCase() === 'sekretaris',
  );

  const schedulesQuery = createQuery(() =>
    rpcQueryOptions<Schedule[]>('get_schedules_secure', { p_viewer_id: user.id }),
  );
  const upsertSchedule = createRpcMutation('upsert_schedule_secure');
  const deleteSchedule = createRpcMutation('delete_schedule_secure', ['get_schedules_secure']);

  $effect(() => {
    if (schedulesQuery.error) console.error('Error fetch:', schedulesQuery.error);
  });

  let filterDay = $state('all');
  const schedules = $derived(
    (schedulesQuery.data || []).filter((s) => filterDay === 'all' || s.day_of_week === filterDay),
  );

  let loading = $state(false);
  let importing = $state(false);
  let fileInput = $state<HTMLInputElement | null>(null);
  let isDialogOpen = $state(false);
  let isEditMode = $state(false);
  let editId = $state<number | null>(null);
  let day = $state('');
  let startTime = $state('');
  let endTime = $state('');
  let title = $state('');
  let major = $state('');
  let classCode = $state('');

  async function handleDownloadTemplate() {
    const XLSX = await import('xlsx');
    const templateData = [
      {
        'Mata Kuliah': 'Praktikum Algoritma',
        Jurusan: 'Teknik Informatika',
        Kelas: 'A',
        Hari: 'Senin',
        'Jam Mulai': '08:00',
        'Jam Selesai': '10:00',
      },
      {
        'Mata Kuliah': 'Praktikum Basis Data',
        Jurusan: 'Sistem Informasi',
        Kelas: 'B',
        Hari: 'Selasa',
        'Jam Mulai': '13:00',
        'Jam Selesai': '15:00',
      },
    ];
    const ws = XLSX.utils.json_to_sheet(templateData);
    const wb = XLSX.utils.book_new();
    XLSX.utils.book_append_sheet(wb, ws, 'Template Jadwal');
    XLSX.writeFile(wb, 'Template_Import_Jadwal.xlsx');
  }

  function handleFileUpload(e: Event & { currentTarget: HTMLInputElement }) {
    const file = e.currentTarget.files?.[0];
    if (!file) return;
    const reader = new FileReader();
    reader.onload = async (evt) => {
      importing = true;
      try {
        if (!isAdminOrSecretary) {
          toast.error('Akses Ditolak', {
            description: 'Hanya Koordinator atau Sekretaris yang dapat mengimport jadwal.',
          });
          return;
        }
        const XLSX = await import('xlsx');
        const wb = XLSX.read(evt.target?.result, { type: 'binary' });
        const rows = XLSX.utils.sheet_to_json<Record<string, string>>(wb.Sheets[wb.SheetNames[0]]);
        if (rows.length === 0) {
          toast.error('File Excel kosong!');
          return;
        }
        const validData = rows
          .map((row) => ({
            title: row['Mata Kuliah'] || row['Title'],
            major: row['Jurusan'] || row['Major'],
            class_code: row['Kelas'] || row['Class'],
            day_of_week: row['Hari'] || row['Day'],
            start_time: row['Jam Mulai'] || row['Start'],
            end_time: row['Jam Selesai'] || row['End'],
          }))
          .filter((d) => d.title && d.day_of_week && d.start_time);
        if (validData.length === 0) {
          toast.error('Format Excel tidak sesuai. Gunakan template.');
          return;
        }
        let successCount = 0;
        for (const item of validData) {
          // Import dianggap langsung approved karena hanya koordinator/sekretaris yang lolos.
          const { error } = await upsertSchedule.mutateAsync({
            p_caller_id: user.id,
            p_id: 0,
            p_day_of_week: item.day_of_week,
            p_start_time: item.start_time,
            p_end_time: item.end_time,
            p_title: item.title,
            p_major: item.major,
            p_class_code: item.class_code,
            p_type: 'praktikum',
            p_status: 'approved',
          });
          if (!error) successCount++;
        }
        toast.success(`Berhasil mengimport ${successCount} jadwal!`);
        void invalidateRpcs(['get_schedules_secure']);
      } catch (error) {
        console.error(error);
        toast.error('Gagal import: ' + (error as Error).message);
      } finally {
        importing = false;
        if (fileInput) fileInput.value = '';
      }
    };
    reader.readAsBinaryString(file);
  }

  function openAddDialog() {
    isEditMode = false;
    day = '';
    startTime = '';
    endTime = '';
    title = '';
    major = '';
    classCode = '';
    isDialogOpen = true;
  }

  function openEditDialog(item: Schedule) {
    isEditMode = true;
    editId = item.id;
    day = item.day_of_week;
    startTime = item.start_time;
    endTime = item.end_time;
    title = item.title;
    major = item.major || '';
    classCode = item.class_code || '';
    isDialogOpen = true;
  }

  async function handleSave() {
    if (!day || !startTime || !endTime || !title || !major || !classCode) {
      return toast.error('Semua kolom wajib diisi!');
    }
    loading = true;
    try {
      const { error } = await upsertSchedule.mutateAsync({
        p_caller_id: user.id,
        p_id: isEditMode && editId ? editId : 0,
        p_day_of_week: day,
        p_start_time: startTime,
        p_end_time: endTime,
        p_title: title,
        p_major: major,
        p_class_code: classCode,
        p_type: 'praktikum',
        p_status: 'approved',
      });
      if (error) throw error;
      toast.success(isEditMode ? 'Jadwal diperbarui!' : 'Jadwal disimpan!');
      isDialogOpen = false;
      void invalidateRpcs(['get_schedules_secure']);
    } catch (error) {
      toast.error('Gagal: ' + (error as Error).message);
    } finally {
      loading = false;
    }
  }

  async function handleDelete(id: number) {
    if (!(await confirm('Hapus jadwal ini?'))) return;
    const { error } = await deleteSchedule.mutateAsync({ p_caller_id: user.id, p_id: id });
    if (error) toast.error('Gagal: ' + error.message);
    else toast.success('Jadwal dihapus');
  }
</script>

<div class="mx-auto max-w-6xl space-y-6">
  <div
    class="flex flex-col items-start justify-between gap-4 rounded-lg border bg-white p-4 shadow-sm xl:flex-row xl:items-center"
  >
    <div>
      <h1 class="text-primary flex items-center gap-2 text-2xl font-bold">
        <CalendarDays class="h-6 w-6" /> Manajemen Jadwal
      </h1>
      <p class="text-muted-foreground mt-1 text-sm">
        Atur slot waktu, jurusan, dan kelas praktikum.
      </p>
    </div>
    <div class="flex w-full flex-col items-start gap-2 sm:flex-row sm:items-center xl:w-auto">
      <div class="flex w-full items-center gap-2 rounded-md border bg-gray-50 px-3 py-1 sm:w-auto">
        <Filter class="text-muted-foreground h-4 w-4" />
        <Select type="single" bind:value={filterDay}>
          <SelectTrigger class="h-8 w-[130px] border-none bg-transparent focus:ring-0"
            >{filterDay === 'all' ? 'Semua Hari' : filterDay}</SelectTrigger
          >
          <SelectContent>
            <SelectItem value="all" label="Semua Hari" />
            {#each DAYS as d (d)}<SelectItem value={d} label={d} />{/each}
          </SelectContent>
        </Select>
      </div>
      {#if isAdminOrSecretary}
        <div class="flex w-full gap-2 sm:w-auto">
          <div class="flex gap-1">
            <input
              type="file"
              bind:this={fileInput}
              accept=".xlsx, .xls"
              class="hidden"
              onchange={handleFileUpload}
            />
            <Button
              variant="outline"
              class="border-green-600 text-green-700 hover:bg-green-50"
              onclick={() => fileInput?.click()}
              disabled={importing}
            >
              {#if importing}<Loader2 class="h-4 w-4 animate-spin" />{:else}<FileSpreadsheet
                  class="mr-2 h-4 w-4"
                />{/if}
              Import
            </Button>
            <Button
              variant="ghost"
              size="icon"
              class="text-muted-foreground"
              title="Download Template Excel"
              onclick={handleDownloadTemplate}
            >
              <Download class="h-4 w-4" />
            </Button>
          </div>
          <Button onclick={openAddDialog} class="bg-primary hover:bg-primary/90"
            ><Plus class="mr-2 h-4 w-4" /> Baru</Button
          >
        </div>
      {/if}
    </div>
  </div>

  <Card>
    <CardHeader><CardTitle>Daftar Jadwal Tersedia ({schedules.length})</CardTitle></CardHeader>
    <CardContent>
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Hari & Jam</TableHead>
            <TableHead>Mata Kuliah / Sesi</TableHead>
            <TableHead>Jurusan</TableHead>
            <TableHead>Kelas</TableHead>
            {#if isAdminOrSecretary}<TableHead class="text-right">Aksi</TableHead>{/if}
          </TableRow>
        </TableHeader>
        <TableBody>
          {#if schedules.length === 0}
            <TableRow
              ><TableCell colspan={5} class="text-muted-foreground py-10 text-center"
                >Belum ada jadwal.</TableCell
              ></TableRow
            >
          {:else}
            {#each schedules as s (s.id)}
              <TableRow>
                <TableCell>
                  <div class="font-bold">{s.day_of_week}</div>
                  <div class="text-muted-foreground mt-1 flex items-center gap-1 text-xs">
                    <Clock class="h-3 w-3" />
                    {s.start_time.slice(0, 5)} - {s.end_time.slice(0, 5)}
                  </div>
                </TableCell>
                <TableCell class="font-medium">{s.title}</TableCell>
                <TableCell>
                  {#if s.major && MAJOR_BADGE[s.major]}
                    <Badge variant="outline" class={MAJOR_BADGE[s.major][1]}
                      >{MAJOR_BADGE[s.major][0]}</Badge
                    >
                  {:else}
                    <span class="text-sm">{s.major}</span>
                  {/if}
                </TableCell>
                <TableCell
                  ><div
                    class="flex h-8 w-8 items-center justify-center rounded-full border bg-gray-100 text-sm font-bold"
                  >
                    {s.class_code}
                  </div></TableCell
                >
                {#if isAdminOrSecretary}
                  <TableCell class="text-right">
                    <div class="flex justify-end gap-2">
                      <Button
                        variant="outline"
                        size="icon"
                        aria-label="Edit jadwal"
                        class="h-8 w-8 text-blue-600 hover:bg-blue-50"
                        onclick={() => openEditDialog(s)}><Pencil class="h-4 w-4" /></Button
                      >
                      <Button
                        variant="destructive"
                        size="icon"
                        aria-label="Hapus jadwal"
                        class="h-8 w-8"
                        onclick={() => handleDelete(s.id)}><Trash2 class="h-4 w-4" /></Button
                      >
                    </div>
                  </TableCell>
                {/if}
              </TableRow>
            {/each}
          {/if}
        </TableBody>
      </Table>
    </CardContent>
  </Card>

  <Dialog bind:open={isDialogOpen}>
    <DialogContent class="sm:max-w-[500px]">
      <DialogHeader
        ><DialogTitle>{isEditMode ? 'Edit Jadwal' : 'Buat Jadwal Baru'}</DialogTitle></DialogHeader
      >
      <div class="grid gap-4 py-2">
        <div class="space-y-1">
          <Label for="js-title">Nama Mata Kuliah / Sesi</Label><Input
            id="js-title"
            placeholder="Contoh: Praktikum Algoritma"
            bind:value={title}
          />
        </div>
        <div class="grid grid-cols-2 gap-4">
          <div class="space-y-1">
            <Label>Jurusan</Label>
            <Select type="single" bind:value={major}>
              <SelectTrigger class="w-full"
                >{#if major}{major}{:else}<span class="text-muted-foreground">Pilih Jurusan</span
                  >{/if}</SelectTrigger
              >
              <SelectContent>
                {#each MAJORS as m (m)}<SelectItem value={m} label={m} />{/each}
              </SelectContent>
            </Select>
          </div>
          <div class="space-y-1">
            <Label for="js-class">Kelas / Kode</Label><Input
              id="js-class"
              placeholder="Contoh: A, B"
              bind:value={classCode}
            />
          </div>
        </div>
        <div class="space-y-1">
          <Label>Hari</Label>
          <Select type="single" bind:value={day}>
            <SelectTrigger class="w-full"
              >{#if day}{day}{:else}<span class="text-muted-foreground">Pilih Hari</span
                >{/if}</SelectTrigger
            >
            <SelectContent
              >{#each DAYS as d (d)}<SelectItem value={d} label={d} />{/each}</SelectContent
            >
          </Select>
        </div>
        <div class="grid grid-cols-2 gap-4">
          <div class="space-y-1">
            <Label for="js-start">Jam Mulai</Label><Input
              id="js-start"
              type="time"
              bind:value={startTime}
            />
          </div>
          <div class="space-y-1">
            <Label for="js-end">Jam Selesai</Label><Input
              id="js-end"
              type="time"
              bind:value={endTime}
            />
          </div>
        </div>
      </div>
      <DialogFooter>
        <Button onclick={handleSave} disabled={loading} class="w-full"
          >{#if loading}<Loader2 class="mr-2 animate-spin" />{:else}<Save
              class="mr-2 h-4 w-4"
            />{/if}
          Simpan</Button
        >
      </DialogFooter>
    </DialogContent>
  </Dialog>
</div>
