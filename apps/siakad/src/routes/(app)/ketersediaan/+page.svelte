<script lang="ts">
  import {
    Badge,
    Button,
    Card,
    CardContent,
    CardDescription,
    CardHeader,
    CardTitle,
    Checkbox,
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
    Tabs,
    TabsContent,
    TabsList,
    TabsTrigger,
  } from '@algohub/ui';
  import CalendarCheck from '@lucide/svelte/icons/calendar-check';
  import Clock from '@lucide/svelte/icons/clock';
  import FileText from '@lucide/svelte/icons/file-text';
  import ImageIcon from '@lucide/svelte/icons/image';
  import Loader2 from '@lucide/svelte/icons/loader-2';
  import Pencil from '@lucide/svelte/icons/pencil';
  import Plus from '@lucide/svelte/icons/plus';
  import Save from '@lucide/svelte/icons/save';
  import Sparkles from '@lucide/svelte/icons/sparkles';
  import Trash2 from '@lucide/svelte/icons/trash-2';
  import UploadCloud from '@lucide/svelte/icons/cloud-upload';
  import UserCog from '@lucide/svelte/icons/user-cog';
  import X from '@lucide/svelte/icons/x';
  import { createQuery } from '@tanstack/svelte-query';
  import { analyzeSchedule } from '#lib/api.ts';
  import { createRpcMutation, invalidateRpcs, rpcQueryOptions } from '#lib/queries/rpc.ts';
  import { confirm } from '#lib/stores/confirm.svelte.ts';
  import { toast } from '#lib/toast.ts';
  import type { PageProps } from './$types';

  interface Slot {
    id: number;
    day_of_week?: string;
    start_time?: string;
    end_time?: string;
  }

  const DAYS = ['Senin', 'Selasa', 'Rabu', 'Kamis', 'Jumat', 'Sabtu'];
  const DAY_ORDER: Record<string, number> = {
    Senin: 1,
    Selasa: 2,
    Rabu: 3,
    Kamis: 4,
    Jumat: 5,
    Sabtu: 6,
  };
  const AI_MODELS: Record<string, string> = {
    'gemini-3-flash': 'Gemini 3 Flash (High)',
    'gemini-2.5-flash': 'Gemini 2.5 Flash',
    'stepfun-3.7-flash': 'Stepfun 3.7 Flash (Bynara Router)',
  };

  let { data }: PageProps = $props();
  const user = $derived(data.auth.user);
  const isCoordinator = $derived(
    user.role === 'koordinator' || user.division?.toLowerCase() === 'sekretaris',
  );
  const hasEditAccess = $derived(isCoordinator || data.auth.allowedPaths.includes('/ketersediaan'));

  let targetUserId = $state<string | null>(null);
  $effect(() => {
    if (!targetUserId) targetUserId = user.id.toString();
  });

  const usersQuery = createQuery(() =>
    rpcQueryOptions<{ id: number; full_name: string; role: string }[]>(
      'get_users_secure',
      { p_viewer_id: user.id },
      { enabled: isCoordinator },
    ),
  );
  const assistantsList = $derived(
    (usersQuery.data || []).filter((u) => ['asisten', 'koordinator'].includes(u.role)),
  );
  const slotsQuery = createQuery(() =>
    rpcQueryOptions<Slot[]>(
      'get_all_availability_for_user_secure',
      { p_viewer_id: user.id, p_target_user_id: parseInt(targetUserId ?? '0') },
      { enabled: !!targetUserId },
    ),
  );
  $effect(() => {
    if (slotsQuery.error) console.error('Error Fetch DB:', slotsQuery.error);
  });
  const formatTime = (timeVal: unknown) => (timeVal ? String(timeVal).substring(0, 5) : '--:--');
  const getSafeString = (val: unknown, fallback = '-') => (val ? String(val) : fallback);
  const mySlots = $derived(
    [...(slotsQuery.data || [])].sort(
      (a, b) =>
        (DAY_ORDER[getSafeString(a?.day_of_week)] || 7) -
        (DAY_ORDER[getSafeString(b?.day_of_week)] || 7),
    ),
  );

  const saveSlot = createRpcMutation('save_assistant_availability_secure');
  const deleteSlot = createRpcMutation('delete_assistant_availability_secure');
  const refreshSlots = () => invalidateRpcs(['get_all_availability_for_user_secure']);

  let loading = $state(false);
  let activeTab = $state('manual');
  let day = $state('');
  let start = $state('');
  let end = $state('');
  let isEditMode = $state(false);
  let editingId = $state<number | null>(null);
  let uploadFile = $state<File | null>(null);
  let previewUrl = $state<string | null>(null);
  let analyzing = $state(false);
  let detectedSlots = $state<Slot[]>([]);
  let selectedAiModel = $state('gemini-3-flash');
  let selectedSlots = $state<number[]>([]);

  $effect(() => {
    void targetUserId;
    selectedSlots = [];
  });

  function handleCancelEdit() {
    day = '';
    start = '';
    end = '';
    isEditMode = false;
    editingId = null;
  }

  async function handleSaveSlot() {
    if (!day || !start || !end) return toast.error('Lengkapi data!');
    if (!targetUserId) return toast.error('Pilih Asisten terlebih dahulu!');
    loading = true;
    try {
      const { error } = await saveSlot.mutateAsync({
        p_caller_id: user.id,
        p_target_user_id: parseInt(targetUserId),
        p_id: isEditMode && editingId ? editingId : 0,
        p_day: day,
        p_start: start,
        p_end: end,
      });
      loading = false;
      if (error) {
        toast.error(error.message);
      } else {
        toast.success(isEditMode ? 'Jadwal berhasil diperbarui!' : 'Jadwal manual ditambahkan!');
        if (isEditMode) handleCancelEdit();
        else {
          day = '';
          start = '';
          end = '';
        }
        await refreshSlots();
      }
    } catch (err) {
      loading = false;
      toast.error((err as Error).message);
    }
  }

  function handleEditClick(slot: Slot) {
    day = getSafeString(slot?.day_of_week, 'Senin');
    start = formatTime(slot?.start_time);
    end = formatTime(slot?.end_time);
    isEditMode = true;
    editingId = slot?.id;
    activeTab = 'manual';
  }

  async function handleDelete(id: number) {
    if (!(await confirm('Hapus slot waktu ini?'))) return;
    const { error } = await deleteSlot.mutateAsync({ p_caller_id: user.id, p_id: id });
    if (!error) {
      toast.success('Berhasil menghapus jadwal');
      await refreshSlots();
      selectedSlots = selectedSlots.filter((sid) => sid !== id);
    } else {
      toast.error('Gagal menghapus data: ' + error.message);
    }
  }

  async function handleMultipleDelete() {
    if (selectedSlots.length === 0) return;
    if (!(await confirm(`Hapus ${selectedSlots.length} slot waktu yang dipilih?`))) return;
    loading = true;
    let successCount = 0;
    for (const id of selectedSlots) {
      const { error } = await deleteSlot.mutateAsync({ p_caller_id: user.id, p_id: id });
      if (!error) successCount++;
    }
    loading = false;
    toast.success(`${successCount} jadwal berhasil dihapus!`);
    selectedSlots = [];
    await refreshSlots();
  }

  function toggleSelectAll() {
    selectedSlots = selectedSlots.length === mySlots.length ? [] : mySlots.map((s) => s.id);
  }

  function toggleSelectSlot(id: number) {
    selectedSlots = selectedSlots.includes(id)
      ? selectedSlots.filter((sid) => sid !== id)
      : [...selectedSlots, id];
  }

  function handleFileUpload(e: Event & { currentTarget: HTMLInputElement }) {
    const file = e.currentTarget.files?.[0];
    if (!file) return;
    uploadFile = file;
    previewUrl = file.type.startsWith('image/') ? URL.createObjectURL(file) : null;
    detectedSlots = [];
  }

  function clearFile() {
    uploadFile = null;
    previewUrl = null;
    detectedSlots = [];
  }

  // API key AI hanya ada di backend; browser mengirim file sebagai base64.
  async function handleAnalyzeFile() {
    if (!uploadFile) return toast.error('Pilih file terlebih dahulu!');
    analyzing = true;
    try {
      const file = uploadFile;
      const fileBase64 = await new Promise<string>((resolve, reject) => {
        const reader = new FileReader();
        reader.onloadend = () => resolve((reader.result as string).split(',')[1]);
        reader.onerror = reject;
        reader.readAsDataURL(file);
      });
      const res = await analyzeSchedule<{ slots?: Slot[]; error?: string }>({
        fileBase64,
        mimeType: file.type,
        model: selectedAiModel,
      });
      if (!res.ok) {
        const errData = res.body ?? { error: 'Server error' };
        throw new Error(errData.error || `Server error: ${res.status}`);
      }
      const detectedData = res.body?.slots;
      if (!Array.isArray(detectedData) || detectedData.length === 0) {
        toast.error('AI tidak menemukan jadwal kosong di dokumen ini.');
        detectedSlots = [];
      } else {
        detectedSlots = detectedData;
        toast.success(`Berhasil mengekstrak ${detectedData.length} jadwal free!`);
      }
    } catch (err) {
      console.error('AI Error:', err);
      toast.error(
        'Gagal menganalisis file. ' +
          ((err as Error).message || 'Pastikan dokumen terbaca dengan jelas.'),
      );
    } finally {
      analyzing = false;
    }
  }

  async function handleSaveDetectedSlots() {
    if (detectedSlots.length === 0) return;
    if (!targetUserId) return toast.error('Pilih Asisten terlebih dahulu!');
    loading = true;
    try {
      let successCount = 0;
      for (const slot of detectedSlots) {
        const { error } = await saveSlot.mutateAsync({
          p_caller_id: user.id,
          p_target_user_id: parseInt(targetUserId),
          p_id: 0,
          p_day: getSafeString(slot?.day_of_week, 'Senin'),
          p_start: formatTime(slot?.start_time),
          p_end: formatTime(slot?.end_time),
        });
        if (!error) successCount++;
      }
      toast.success(`${successCount} jadwal free berhasil disimpan!`);
      clearFile();
      await refreshSlots();
    } catch (err) {
      toast.error((err as Error).message);
    } finally {
      loading = false;
    }
  }

  const targetLabel = $derived(
    targetUserId === user.id.toString()
      ? 'Diri Sendiri (Anda)'
      : (assistantsList.find((a) => a.id.toString() === targetUserId)?.full_name ?? ''),
  );
</script>

<div class="mx-auto max-w-6xl space-y-6">
  <div
    class="flex flex-col items-start justify-between gap-4 rounded-lg border bg-white p-5 shadow-sm md:flex-row md:items-center"
  >
    <div>
      <h1 class="text-primary flex items-center gap-2 text-2xl font-bold">
        <CalendarCheck class="h-7 w-7" /> Ketersediaan Asisten
      </h1>
      <p class="text-muted-foreground mt-1 text-sm">
        {isCoordinator
          ? 'Pilih asisten di samping untuk menginput atau mengedit jadwal mereka.'
          : 'Masukkan jadwal di mana Anda BISA/FREE untuk mengajar praktikum.'}
      </p>
    </div>
    {#if isCoordinator && targetUserId}
      <div
        class="flex w-full items-center gap-3 rounded-md border border-blue-200 bg-blue-50 p-2 px-3 md:w-auto"
      >
        <div class="flex items-center text-sm font-semibold text-blue-800">
          <UserCog class="mr-2 h-4 w-4" /> Target Asisten:
        </div>
        <Select type="single" bind:value={targetUserId}>
          <SelectTrigger class="h-8 w-[200px] bg-white">{targetLabel}</SelectTrigger>
          <SelectContent>
            <SelectItem value={user.id.toString()} label="Diri Sendiri (Anda)" />
            {#each assistantsList.filter((a) => a.id !== user.id) as a (a.id)}
              <SelectItem value={a.id.toString()} label={a.full_name} />
            {/each}
          </SelectContent>
        </Select>
      </div>
    {/if}
  </div>

  <div class="grid gap-6 lg:grid-cols-5">
    <div class="lg:col-span-2">
      <Tabs bind:value={activeTab} class="w-full">
        <TabsList class="mb-4 grid w-full grid-cols-2">
          <TabsTrigger value="manual">Input Manual</TabsTrigger>
          <TabsTrigger value="auto" class="flex items-center gap-1">
            <Sparkles class="h-3 w-3" /> Auto Deteksi
          </TabsTrigger>
        </TabsList>
        <TabsContent value="manual">
          <Card class={isEditMode ? 'border-blue-300 shadow-blue-100' : ''}>
            <CardHeader class="pb-3">
              <CardTitle
                class="flex items-center gap-2 text-base {isEditMode ? 'text-blue-700' : ''}"
              >
                {#if isEditMode}<Pencil class="h-4 w-4" /> Edit Slot Jadwal{:else}Tambah Slot Manual{/if}
              </CardTitle>
            </CardHeader>
            <CardContent class="space-y-4">
              <div class="space-y-2">
                <Label>Hari</Label>
                <Select type="single" bind:value={day}>
                  <SelectTrigger class="w-full"
                    >{#if day}{day}{:else}<span class="text-muted-foreground">Pilih Hari</span
                      >{/if}</SelectTrigger
                  >
                  <SelectContent>
                    {#each DAYS as d (d)}<SelectItem value={d} label={d} />{/each}
                  </SelectContent>
                </Select>
              </div>
              <div class="grid grid-cols-2 gap-4">
                <div class="space-y-2">
                  <Label for="slot-start">Mulai Free</Label>
                  <Input id="slot-start" type="time" bind:value={start} />
                </div>
                <div class="space-y-2">
                  <Label for="slot-end">Selesai Free</Label>
                  <Input id="slot-end" type="time" bind:value={end} />
                </div>
              </div>
              <div class="mt-2 flex gap-2">
                <Button
                  onclick={handleSaveSlot}
                  disabled={loading}
                  class="flex-1 {isEditMode ? 'bg-blue-600 hover:bg-blue-700' : ''}"
                >
                  {#if isEditMode}<Save class="mr-2 h-4 w-4" /> Simpan{:else}<Plus
                      class="mr-2 h-4 w-4"
                    /> Tambah{/if}
                </Button>
                {#if isEditMode}
                  <Button variant="outline" onclick={handleCancelEdit} disabled={loading}
                    >Batal</Button
                  >
                {/if}
              </div>
            </CardContent>
          </Card>
        </TabsContent>
        <TabsContent value="auto">
          <Card class="border-blue-100 bg-blue-50/30">
            <CardHeader class="pb-3">
              <CardTitle class="flex items-center gap-2 text-base text-blue-800">
                <UploadCloud class="h-4 w-4" /> Upload Jadwal Kuliah
              </CardTitle>
              <CardDescription class="text-xs">
                AI akan mengekstrak jam kosong dari KRS/Jadwal (rentang jam 06:00 - 22:00).
              </CardDescription>
            </CardHeader>
            <CardContent class="space-y-4">
              {#if !uploadFile}
                <div
                  class="relative cursor-pointer rounded-lg border-2 border-dashed border-blue-200 bg-white p-6 text-center transition-colors hover:bg-blue-50"
                >
                  <input
                    type="file"
                    accept="image/png, image/jpeg, application/pdf"
                    aria-label="Upload jadwal kuliah"
                    class="absolute inset-0 h-full w-full cursor-pointer opacity-0"
                    onchange={handleFileUpload}
                  />
                  <div class="mb-2 flex justify-center gap-2 text-blue-400">
                    <ImageIcon class="h-7 w-7" />
                    <FileText class="h-7 w-7" />
                  </div>
                  <p class="text-sm font-medium text-blue-700">Klik / Tarik file ke sini</p>
                  <p class="text-muted-foreground mt-1 text-xs">Format: JPG, PNG, PDF</p>
                </div>
              {:else}
                <div class="space-y-3">
                  <div class="relative overflow-hidden rounded-md border bg-white">
                    {#if previewUrl}
                      <img
                        src={previewUrl}
                        alt="Preview"
                        class="h-32 w-full object-cover opacity-80 transition-opacity hover:opacity-100"
                      />
                    {:else}
                      <div
                        class="flex h-32 w-full flex-col items-center justify-center bg-gray-50 p-4 text-center"
                      >
                        <FileText class="mb-2 h-10 w-10 text-red-500" />
                        <span class="line-clamp-2 text-sm font-medium text-gray-700"
                          >{uploadFile.name || 'Dokumen'}</span
                        >
                        <span class="text-muted-foreground mt-1 text-xs"
                          >{(uploadFile.size / 1024 / 1024).toFixed(2)} MB</span
                        >
                      </div>
                    {/if}
                    <Button
                      variant="destructive"
                      size="icon"
                      aria-label="Hapus file"
                      class="absolute top-2 right-2 h-6 w-6 rounded-full"
                      onclick={clearFile}
                    >
                      <X class="h-3 w-3" />
                    </Button>
                  </div>
                  {#if detectedSlots.length === 0}
                    <div class="space-y-3">
                      {#if hasEditAccess}
                        <div class="flex items-center gap-2">
                          <Label class="text-muted-foreground w-[120px] text-xs"
                            >Pilih Model AI:</Label
                          >
                          <Select type="single" bind:value={selectedAiModel}>
                            <SelectTrigger class="h-8 bg-white text-xs"
                              >{AI_MODELS[selectedAiModel]}</SelectTrigger
                            >
                            <SelectContent>
                              {#each Object.entries(AI_MODELS) as [value, label] (value)}
                                <SelectItem {value} {label} />
                              {/each}
                            </SelectContent>
                          </Select>
                        </div>
                      {/if}
                      <Button
                        onclick={handleAnalyzeFile}
                        disabled={analyzing}
                        class="w-full bg-blue-600 hover:bg-blue-700"
                      >
                        {#if analyzing}<Loader2 class="mr-2 h-4 w-4 animate-spin" /> Sedang Membaca Dokumen...{:else}<Sparkles
                            class="mr-2 h-4 w-4"
                          /> Cari Jam Kosong{/if}
                      </Button>
                    </div>
                  {:else}
                    <div class="space-y-3 rounded border border-green-200 bg-white p-3 shadow-sm">
                      <p class="text-xs font-bold text-green-700">
                        Berhasil Ditemukan ({detectedSlots.length} slot):
                      </p>
                      <ul class="space-y-1 text-xs">
                        {#each detectedSlots as slot, i (i)}
                          <li class="flex justify-between border-b pb-1">
                            <span>{getSafeString(slot?.day_of_week, '?')}</span>
                            <span class="font-mono font-medium text-green-700">
                              {formatTime(slot?.start_time)} - {formatTime(slot?.end_time)}
                            </span>
                          </li>
                        {/each}
                      </ul>
                      <Button
                        onclick={handleSaveDetectedSlots}
                        disabled={loading}
                        size="sm"
                        class="w-full bg-green-600 hover:bg-green-700"
                      >
                        <Save class="mr-2 h-4 w-4" /> Simpan Semua
                      </Button>
                    </div>
                  {/if}
                </div>
              {/if}
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>

    <div class="lg:col-span-3">
      <Card class="h-full">
        <CardHeader class="mb-2 border-b pb-3">
          <CardTitle class="flex items-center justify-between text-lg">
            <div class="flex items-center gap-2">
              <span>Jadwal Free Tersimpan</span>
              {#if selectedSlots.length > 0}
                <Button
                  variant="destructive"
                  size="sm"
                  class="animate-in fade-in slide-in-from-left-2 flex h-7 items-center gap-1 px-2 text-xs"
                  onclick={handleMultipleDelete}
                  disabled={loading}
                >
                  <Trash2 class="h-3 w-3" /> Hapus {selectedSlots.length}
                </Button>
              {/if}
            </div>
            {#if isCoordinator && targetUserId !== user.id.toString()}
              <Badge class="border-none bg-blue-100 text-xs text-blue-800"
                >Lihat Mode Koordinator</Badge
              >
            {/if}
          </CardTitle>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead class="w-[40px]">
                  <Checkbox
                    aria-label="Pilih semua"
                    checked={mySlots.length > 0 && selectedSlots.length === mySlots.length}
                    onCheckedChange={toggleSelectAll}
                  />
                </TableHead>
                <TableHead>Hari</TableHead>
                <TableHead>Jam Free</TableHead>
                <TableHead class="text-right">Aksi</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {#if mySlots.length === 0}
                <TableRow
                  ><TableCell colspan={3} class="text-muted-foreground py-10 text-center"
                    >Belum ada jadwal yang diinput.</TableCell
                  ></TableRow
                >
              {:else}
                {#each mySlots as slot, index (slot.id ?? `fallback-${index}`)}
                  {@const isChecked = selectedSlots.includes(slot.id)}
                  <TableRow class={isChecked ? 'bg-blue-50/50' : ''}>
                    <TableCell>
                      <Checkbox
                        aria-label="Pilih slot"
                        checked={isChecked}
                        onCheckedChange={() => toggleSelectSlot(slot.id)}
                      />
                    </TableCell>
                    <TableCell class="font-bold"
                      >{getSafeString(slot?.day_of_week, 'Tidak diketahui')}</TableCell
                    >
                    <TableCell>
                      <Badge
                        variant="outline"
                        class="flex w-fit items-center gap-1 bg-gray-50 font-mono text-sm"
                      >
                        <Clock class="h-3 w-3 text-green-600" />
                        {formatTime(slot?.start_time)} - {formatTime(slot?.end_time)}
                      </Badge>
                    </TableCell>
                    <TableCell class="text-right">
                      <div class="flex justify-end gap-1">
                        <Button
                          variant="ghost"
                          size="icon"
                          aria-label="Edit slot"
                          class="h-8 w-8 text-blue-500 hover:bg-blue-50"
                          onclick={() => handleEditClick(slot)}
                        >
                          <Pencil class="h-4 w-4" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="icon"
                          aria-label="Hapus slot"
                          class="h-8 w-8 text-red-500 hover:bg-red-50"
                          onclick={() => handleDelete(slot.id)}
                        >
                          <Trash2 class="h-4 w-4" />
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                {/each}
              {/if}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </div>
  </div>
</div>
