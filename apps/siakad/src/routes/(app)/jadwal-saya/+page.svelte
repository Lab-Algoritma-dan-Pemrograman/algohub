<script lang="ts">
  import {
    Badge,
    Button,
    Card,
    CardContent,
    CardHeader,
    CardTitle,
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
  import BookOpen from '@lucide/svelte/icons/book-open';
  import Calendar from '@lucide/svelte/icons/calendar';
  import Clock from '@lucide/svelte/icons/clock';
  import Filter from '@lucide/svelte/icons/filter';
  import Loader2 from '@lucide/svelte/icons/loader-2';
  import MapPin from '@lucide/svelte/icons/map-pin';
  import MessageCircle from '@lucide/svelte/icons/message-circle';
  import User from '@lucide/svelte/icons/user';
  import Users from '@lucide/svelte/icons/users';
  import { createQueries, createQuery } from '@tanstack/svelte-query';
  import { createRpcMutation, rpcQueryOptions } from '#lib/queries/rpc.ts';
  import { canonRole } from '#lib/roles.ts';
  import { toast } from '#lib/toast.ts';
  import { buildWaText, getGreeting, getHonorific, getWaLink } from '#lib/wa.ts';
  import type { PageProps } from './$types';

  interface PersonalRow {
    schedule_id: number;
    schedule_title?: string;
    schedule_day?: string;
    schedule_start?: string;
    schedule_end?: string;
    schedule_major?: string;
    schedule_class?: string;
    schedule_room?: string;
    assistant_id?: number | null;
    assistant_name?: string;
    assistant_phone?: string;
    assistant_code?: string;
    student_id?: number | null;
    student_name?: string;
    student_nim?: string;
    student_shift?: string;
    student_phone?: string;
  }

  interface GroupMemberRow {
    id: number;
    student_id: number;
    assistant_id: number | null;
    student_name?: string;
    student_nim?: string;
    student_shift?: string;
    student_class_code?: string;
  }

  interface ScheduleInfo {
    id: number;
    title?: string;
    day_of_week?: string;
    start_time?: string;
    end_time?: string;
    major?: string;
    class_code?: string;
    room?: string;
  }

  interface StudentEntry {
    student_id: number;
    student: {
      id: number;
      full_name?: string;
      username?: string;
      shift?: string;
      phone_number?: string;
    };
  }

  interface ScheduleItem {
    schedule_id: number;
    schedule: ScheduleInfo;
    student_shift?: string;
    assistant?: {
      id?: number | null;
      full_name?: string;
      phone_number?: string;
      assistant_code?: string;
    };
    students?: StudentEntry[];
  }

  let { data }: PageProps = $props();
  const user = $derived(data.auth.user);
  // Role sudah kanonik dari auth; 'mahasiswa' hanya jaring pengaman untuk data lama.
  const isPraktikan = $derived(canonRole(user.role) === 'praktikan' || user.role === 'mahasiswa');
  const isStaff = $derived(user.role === 'asisten' || user.role === 'koordinator');

  const schedulesQuery = createQuery(() =>
    rpcQueryOptions<PersonalRow[]>('get_personal_schedules_secure', { p_viewer_id: user.id }),
  );
  // Pengaturan lengkap (template WA) hanya untuk staff; RPC-nya menolak praktikan dengan 400.
  const settingsQuery = createQuery(() =>
    rpcQueryOptions<{ wa_templates?: Record<string, unknown> }[]>(
      'get_system_settings_full_secure',
      { p_viewer_id: user.id },
      { enabled: !isPraktikan },
    ),
  );
  const syncRosters = createRpcMutation('sync_class_rosters_secure');

  const scheduleInfo = (row: PersonalRow): ScheduleInfo => ({
    id: row.schedule_id,
    title: row.schedule_title,
    day_of_week: row.schedule_day,
    start_time: row.schedule_start,
    end_time: row.schedule_end,
    major: row.schedule_major,
    class_code: row.schedule_class,
  });

  const schedulesData = $derived.by((): ScheduleItem[] => {
    const rows = schedulesQuery.data || [];
    if (isPraktikan) {
      return rows.map((row) => ({
        ...row,
        schedule: scheduleInfo(row),
        assistant: {
          id: row.assistant_id,
          full_name: row.assistant_name,
          phone_number: row.assistant_phone,
          assistant_code: row.assistant_code,
        },
      }));
    }
    const grouped: ScheduleItem[] = [];
    const byId: Record<number, ScheduleItem> = {};
    for (const current of rows) {
      let schedule = byId[current.schedule_id];
      if (!schedule) {
        schedule = {
          schedule_id: current.schedule_id,
          schedule: scheduleInfo(current),
          students: [],
        };
        byId[current.schedule_id] = schedule;
        grouped.push(schedule);
      }
      if (current.student_id) {
        schedule.students!.push({
          student_id: current.student_id,
          student: {
            id: current.student_id,
            full_name: current.student_name,
            username: current.student_nim,
            shift: current.student_shift,
          },
        });
      }
    }
    return grouped;
  });

  // Jalur .from() lama memakai anon key dan RLS group_members selalu mengosongkan hasilnya;
  // RPC ini tidak mengembalikan kolom telepon, jadi nomor HP teman sekelompok tidak tampil.
  const kelompokQueries = createQueries(() => ({
    queries: isPraktikan
      ? schedulesData.map((s) =>
          rpcQueryOptions<GroupMemberRow[]>('get_group_members_secure', {
            p_viewer_id: user.id,
            p_schedule_id: s.schedule_id,
          }),
        )
      : [],
  }));

  const kelompok = $derived.by(() => {
    const map: Record<number, (GroupMemberRow & { student: Record<string, unknown> })[]> = {};
    schedulesData.forEach((s, i) => {
      map[s.schedule_id] = (kelompokQueries[i]?.data || []).map((m) => ({
        ...m,
        student: {
          id: m.student_id,
          full_name: m.student_name,
          username: m.student_nim,
          shift: m.student_shift,
          class_code: m.student_class_code,
        },
      }));
    });
    return map;
  });

  const loading = $derived(
    schedulesQuery.isPending || (isPraktikan && kelompokQueries.some((q) => q.isPending)),
  );

  $effect(() => {
    const error = schedulesQuery.error ?? kelompokQueries.find((q) => q.error)?.error;
    if (error) toast.error('Gagal memuat jadwal: ' + error.message);
  });

  const waTemplates = $derived(
    settingsQuery.data && settingsQuery.data.length > 0 ? settingsQuery.data[0].wa_templates : null,
  );

  // class_rosters hanya dipakai fitur koordinator dan RPC-nya khusus koordinator.
  let syncedAt = 0;
  $effect(() => {
    const updatedAt = schedulesQuery.dataUpdatedAt;
    if (user.role !== 'koordinator' || !schedulesQuery.data || updatedAt === syncedAt) return;
    syncedAt = updatedAt;
    const payload = schedulesData.flatMap((item) =>
      (item.students || []).map((s) => ({
        schedule_id: item.schedule_id,
        assistant_id: undefined,
        student_id: s.student_id,
      })),
    );
    if (payload.length === 0) return;
    syncRosters
      .mutateAsync({ p_caller_id: user.id, p_payload: payload })
      .then(({ error }) => {
        if (error) console.error('Sync Error:', error);
      })
      .catch((err) => console.error('Failed to sync to class_rosters:', err));
  });

  let filterDay = $state('all');
  let filterMajor = $state('all');
  let filterClass = $state('all');

  const uniq = (values: (string | undefined)[]) =>
    [...new Set(values.filter((v): v is string => !!v))].sort();
  const uniqueDays = $derived(uniq(schedulesData.map((i) => i.schedule.day_of_week)));
  const uniqueMajors = $derived(uniq(schedulesData.map((i) => i.schedule.major)));
  const uniqueClasses = $derived(uniq(schedulesData.map((i) => i.schedule.class_code)));

  const filteredSchedules = $derived(
    schedulesData.filter(
      (item) =>
        (filterDay === 'all' || item.schedule.day_of_week === filterDay) &&
        (filterMajor === 'all' || item.schedule.major === filterMajor) &&
        (filterClass === 'all' || item.schedule.class_code === filterClass),
    ),
  );

  function teman(item: ScheduleItem) {
    // Praktikan yang belum diplotting ke asisten melihat seluruh anggota jadwalnya.
    const anggota = kelompok[item.schedule_id] || [];
    return item.assistant?.id
      ? anggota.filter((m) => m.assistant_id === item.assistant!.id)
      : anggota;
  }

  function waTextFor(item: ScheduleItem) {
    return buildWaText(waTemplates, {
      greeting: getGreeting(),
      honorific: getHonorific(item.assistant?.assistant_code) || 'Kak',
      assistant: item.assistant?.full_name || 'Asisten',
      student: user.full_name,
      nim: user.username,
      major: item.schedule.major,
      kelas: item.schedule.class_code,
    });
  }
</script>

{#snippet emptyCard(text: string)}
  <Card class="border-dashed bg-gray-50"
    ><CardContent class="text-muted-foreground py-12 text-center">{text}</CardContent></Card
  >
{/snippet}

{#snippet filterSelect(
  value: string,
  set: (v: string) => void,
  width: string,
  allLabel: string,
  options: string[],
  prefix: string,
)}
  <Select type="single" {value} onValueChange={set}>
    <SelectTrigger class="h-9 {width}"
      >{value === 'all' ? allLabel : `${prefix}${value}`}</SelectTrigger
    >
    <SelectContent>
      <SelectItem value="all" label={allLabel} />
      {#each options as option (option)}
        <SelectItem value={option} label={`${prefix}${option}`} />
      {/each}
    </SelectContent>
  </Select>
{/snippet}

<div class="space-y-6">
  <div class="flex flex-col items-start justify-between gap-4 md:flex-row md:items-center">
    <div>
      <h1 class="flex items-center gap-2 text-2xl font-bold">
        <BookOpen class="text-primary" /> Jadwal Praktikum Saya
      </h1>
      <p class="text-muted-foreground">
        {isPraktikan
          ? 'Jadwal praktikum Anda, kontak asisten pembimbing, dan daftar kelompok bimbingan Anda.'
          : 'Berikut adalah daftar kelas dan praktikan yang Anda bimbing.'}
      </p>
    </div>
  </div>

  {#if !loading && schedulesData.length > 0 && isStaff}
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
        '',
      )}
      {@render filterSelect(
        filterMajor,
        (v) => (filterMajor = v),
        'w-[200px]',
        'Semua Jurusan',
        uniqueMajors,
        '',
      )}
      {@render filterSelect(
        filterClass,
        (v) => (filterClass = v),
        'w-[140px]',
        'Semua Kelas',
        uniqueClasses,
        'Kelas ',
      )}
      {#if filterDay !== 'all' || filterMajor !== 'all' || filterClass !== 'all'}
        <Button
          variant="ghost"
          size="sm"
          onclick={() => {
            filterDay = 'all';
            filterMajor = 'all';
            filterClass = 'all';
          }}
          class="h-9 px-3 text-red-500 hover:bg-red-50 hover:text-red-600"
        >
          Reset Filter
        </Button>
      {/if}
    </div>
  {/if}

  {#if loading}
    <div class="flex justify-center py-12">
      <Loader2 class="text-primary h-8 w-8 animate-spin" />
    </div>
  {:else if isPraktikan}
    {#if schedulesData.length === 0}
      {@render emptyCard('Anda belum dimasukkan ke dalam kelompok praktikum mana pun.')}
    {:else if filteredSchedules.length === 0}
      {@render emptyCard('Tidak ada jadwal yang sesuai dengan filter pencarian Anda.')}
    {:else}
      <div class="grid grid-cols-1 gap-6 lg:grid-cols-2">
        {#each filteredSchedules as item, idx (idx)}
          {@const temanKelompok = teman(item)}
          <Card class="border-l-4 border-l-blue-600 shadow-md transition-shadow hover:shadow-lg">
            <CardHeader class="pb-3">
              <div class="flex items-start justify-between">
                <Badge variant="outline" class="border-blue-200 bg-blue-50 text-blue-700">
                  Kelas {item.schedule.class_code || '-'}
                </Badge>
                <div class="flex flex-col items-end gap-1">
                  <Badge variant="secondary" class="max-w-[160px] truncate"
                    >{item.schedule.major}</Badge
                  >
                  {#if item.student_shift}<Badge
                      variant="outline"
                      class="border-orange-200 bg-orange-50 text-orange-700"
                      >Shift {item.student_shift}</Badge
                    >{/if}
                </div>
              </div>
              <CardTitle class="mt-2 text-xl">{item.schedule.title || 'Praktikum'}</CardTitle>
            </CardHeader>
            <CardContent class="space-y-4">
              <div class="flex flex-wrap gap-x-5 gap-y-2 text-sm text-gray-700">
                <span class="flex items-center"
                  ><Calendar class="text-muted-foreground mr-2 h-4 w-4" /><span class="font-medium"
                    >{item.schedule.day_of_week}</span
                  ></span
                >
                <span class="flex items-center"
                  ><Clock
                    class="text-muted-foreground mr-2 h-4 w-4"
                  />{item.schedule.start_time?.slice(0, 5)} - {item.schedule.end_time?.slice(
                    0,
                    5,
                  )}</span
                >
                {#if item.schedule.room}<span class="flex items-center"
                    ><MapPin class="text-muted-foreground mr-2 h-4 w-4" />{item.schedule.room}</span
                  >{/if}
              </div>
              <div class="mt-2 border-t pt-4">
                <div
                  class="text-muted-foreground mb-2 text-xs font-semibold tracking-wider uppercase"
                >
                  Asisten Pembimbing:
                </div>
                <div class="flex items-center justify-between gap-3">
                  <div class="flex min-w-0 items-center gap-2">
                    <User class="h-5 w-5 shrink-0 text-blue-600" />
                    <div class="min-w-0">
                      <p class="truncate font-bold text-gray-900">
                        {item.assistant?.full_name || 'Belum ditentukan'}
                      </p>
                      <p class="text-muted-foreground truncate text-xs">
                        {item.assistant?.phone_number
                          ? `+${getWaLink(item.assistant.phone_number).match(/wa\.me\/(\d+)/)?.[1] || ''}`
                          : 'Nomor WA belum tersedia'}
                      </p>
                    </div>
                  </div>
                  {#if item.assistant?.phone_number}
                    <a
                      href={getWaLink(item.assistant.phone_number, waTextFor(item))}
                      target="_blank"
                      rel="noreferrer external"
                    >
                      <Button
                        size="sm"
                        class="bg-[#25D366] text-white shadow-sm hover:bg-[#1ebd5c]"
                      >
                        <MessageCircle class="mr-2 h-4 w-4" /> Chat Asisten
                      </Button>
                    </a>
                  {/if}
                </div>
              </div>
              <div class="border-t pt-4">
                <div class="mb-3 flex items-center justify-between">
                  <div
                    class="text-muted-foreground flex items-center gap-2 text-xs font-semibold tracking-wider uppercase"
                  >
                    <Users class="h-4 w-4" /> Kelompok Saya
                  </div>
                  <Badge variant="outline" class="bg-white">{temanKelompok.length} orang</Badge>
                </div>
                {#if temanKelompok.length === 0}
                  <p class="text-muted-foreground text-sm italic">
                    Belum ada rekan satu kelompok yang diplotting.
                  </p>
                {:else}
                  <ol class="space-y-1.5">
                    {#each temanKelompok as m, i (m.id)}
                      {@const isMe = m.student_id === user.id}
                      <li
                        class="flex items-center justify-between gap-3 rounded-lg px-3 py-2 text-sm {isMe
                          ? 'border border-blue-200 bg-blue-50'
                          : 'bg-muted/50'}"
                      >
                        <span class="flex min-w-0 items-center gap-2">
                          <span class="text-muted-foreground w-4 shrink-0 text-xs">{i + 1}.</span>
                          <span
                            class="truncate {isMe ? 'font-black text-blue-800' : 'font-medium'}"
                          >
                            {m.student_name}{isMe ? ' (Saya)' : ''}
                          </span>
                        </span>
                        <span class="flex shrink-0 items-center gap-2">
                          <span class="text-muted-foreground font-mono text-xs"
                            >{m.student_nim || '-'}</span
                          >
                          {#if m.student_shift}<Badge
                              variant="outline"
                              class="border-orange-200 bg-orange-50 text-[9px] text-orange-700"
                              >Shift {m.student_shift}</Badge
                            >{/if}
                        </span>
                      </li>
                    {/each}
                  </ol>
                {/if}
              </div>
            </CardContent>
          </Card>
        {/each}
      </div>
    {/if}
  {:else if schedulesData.length === 0}
    {@render emptyCard('Anda belum ditugaskan sebagai pembimbing di kelas mana pun.')}
  {:else if filteredSchedules.length === 0}
    {@render emptyCard('Tidak ada jadwal yang sesuai dengan filter pencarian Anda.')}
  {:else}
    <div class="space-y-6">
      {#each filteredSchedules as item, idx (idx)}
        <Card class="overflow-hidden border-l-4 border-l-green-500 shadow-sm">
          <CardHeader class="border-b border-green-100 bg-green-50/50 pb-4">
            <div class="flex flex-col justify-between gap-4 sm:flex-row sm:items-center">
              <div>
                <div class="mb-2 flex items-center gap-2">
                  <Badge class="border-none bg-green-100 text-green-800 hover:bg-green-100"
                    >Kelas {item.schedule.class_code}</Badge
                  >
                  <span class="text-sm font-medium text-green-700">{item.schedule.major}</span>
                </div>
                <CardTitle class="text-xl">{item.schedule.title || 'Praktikum'}</CardTitle>
              </div>
              <div
                class="flex flex-col items-end rounded border bg-white p-2 px-3 text-sm font-medium text-gray-700"
              >
                <div class="flex items-center">
                  <Calendar class="text-muted-foreground mr-2 h-4 w-4" />
                  {item.schedule.day_of_week}
                </div>
                <div class="mt-1 flex items-center">
                  <Clock class="text-muted-foreground mr-2 h-4 w-4" />
                  {item.schedule.start_time?.slice(0, 5)} - {item.schedule.end_time?.slice(0, 5)}
                </div>
              </div>
            </div>
          </CardHeader>
          <CardContent class="overflow-x-auto p-0 pt-4">
            <div class="flex items-center justify-between border-b bg-gray-50 px-6 py-3">
              <span class="flex items-center gap-2 font-semibold text-gray-700"
                ><Users class="h-4 w-4" /> Daftar Praktikan Bimbingan</span
              >
              <Badge variant="outline" class="bg-white"
                >{item.students?.length || 0} Mahasiswa</Badge
              >
            </div>
            <Table>
              <TableHeader>
                <TableRow class="bg-white">
                  <TableHead class="w-[80px] text-center">No</TableHead>
                  <TableHead>Nama Mahasiswa</TableHead>
                  <TableHead>NIM & Shift</TableHead>
                  <TableHead class="text-right">No. WhatsApp</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {#if !item.students || item.students.length === 0}
                  <TableRow
                    ><TableCell colspan={4} class="text-muted-foreground py-6 text-center"
                      >Belum ada praktikan yang dimasukkan ke kelompok Anda.</TableCell
                    ></TableRow
                  >
                {:else}
                  {#each item.students as student, sIdx (sIdx)}
                    <TableRow>
                      <TableCell class="text-center font-medium text-gray-500">{sIdx + 1}</TableCell
                      >
                      <TableCell class="font-bold"
                        >{student.student.full_name || 'Unknown'}</TableCell
                      >
                      <TableCell>
                        <div class="flex flex-col gap-1">
                          <Badge variant="secondary" class="w-fit font-mono"
                            >{student.student.username || '-'}</Badge
                          >
                          {#if student.student.shift}<Badge
                              variant="outline"
                              class="w-fit border-orange-200 bg-orange-50 text-[9px] text-orange-700"
                              >Shift {student.student.shift}</Badge
                            >{/if}
                        </div>
                      </TableCell>
                      <TableCell class="text-right">
                        {#if student.student.phone_number}
                          <a
                            href={getWaLink(
                              student.student.phone_number,
                              `Halo ${student.student.full_name}, ini Asisten Praktikum kamu.`,
                            )}
                            target="_blank"
                            rel="noreferrer external"
                          >
                            <Button
                              size="sm"
                              variant="outline"
                              class="h-8 border-green-200 bg-green-50 text-green-600 hover:bg-green-100"
                            >
                              <MessageCircle class="mr-2 h-4 w-4" /> Hubungi
                            </Button>
                          </a>
                        {:else}
                          <span class="text-muted-foreground text-xs italic">-</span>
                        {/if}
                      </TableCell>
                    </TableRow>
                  {/each}
                {/if}
              </TableBody>
            </Table>
          </CardContent>
        </Card>
      {/each}
    </div>
  {/if}
</div>
