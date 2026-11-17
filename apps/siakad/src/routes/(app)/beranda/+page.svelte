<script lang="ts">
  import { Badge, Card, CardContent, CardHeader, CardTitle } from '@algohub/ui';
  import BookOpen from '@lucide/svelte/icons/book-open';
  import Calendar from '@lucide/svelte/icons/calendar';
  import CheckCircle2 from '@lucide/svelte/icons/circle-check';
  import ClipboardCheck from '@lucide/svelte/icons/clipboard-check';
  import Database from '@lucide/svelte/icons/database';
  import FileText from '@lucide/svelte/icons/file-text';
  import GraduationCap from '@lucide/svelte/icons/graduation-cap';
  import MessageSquare from '@lucide/svelte/icons/message-square';
  import QrCode from '@lucide/svelte/icons/qr-code';
  import Users from '@lucide/svelte/icons/users';
  import { createQuery } from '@tanstack/svelte-query';
  import UserContactCard from '#lib/components/beranda/UserContactCard.svelte';
  import { dashboardStatsQuery } from '#lib/queries/beranda.ts';
  import { rpcQueryOptions } from '#lib/queries/rpc.ts';
  import { canonRole } from '#lib/roles.ts';
  import { buildWaText, getGreeting, getHonorific, getWaLink } from '#lib/wa.ts';
  import type { PageProps } from './$types';

  interface AttendanceLog {
    check_in_time: string;
    status: string;
  }

  interface PersonalSchedule {
    student_id: number | string;
    schedule_title: string;
    schedule_day: string;
    schedule_start?: string;
    schedule_end?: string;
    schedule_class: string;
    schedule_major: string;
    assistant_name?: string;
    assistant_phone?: string;
    assistant_code?: string;
  }

  let { data }: PageProps = $props();
  const user = $derived(data.auth.user);
  // Role kanonik; fallback praktikan menjaga perilaku lama saat role belum terisi.
  const role = $derived(canonRole(user.role) || 'praktikan');

  const statsQuery = createQuery(() => ({
    ...dashboardStatsQuery(user.id),
    enabled: role !== 'penyewa',
  }));
  const s = $derived(statsQuery.data ?? null);
  const attendanceQuery = createQuery(() =>
    rpcQueryOptions<AttendanceLog[]>(
      'get_attendance_logs_secure',
      { p_viewer_id: user.id },
      { enabled: role === 'praktikan' },
    ),
  );
  // Route praktikan diizinkan get_personal_schedules_secure; get_group_members_secure wajib
  // p_schedule_id sehingga selalu kosong.
  const scheduleQuery = createQuery(() =>
    rpcQueryOptions<PersonalSchedule[]>(
      'get_personal_schedules_secure',
      { p_viewer_id: user.id },
      { enabled: role === 'praktikan' },
    ),
  );
  const absensi = $derived((attendanceQuery.data || []).slice(0, 4));
  const jadwal = $derived(
    (scheduleQuery.data || [])
      .filter((r) => String(r.student_id) === String(user.id))
      .map((r) => ({
        title: r.schedule_title,
        day_of_week: r.schedule_day,
        start_time: r.schedule_start,
        end_time: r.schedule_end,
        class_code: r.schedule_class,
        major: r.schedule_major,
        assistant: {
          full_name: r.assistant_name,
          phone_number: r.assistant_phone,
          assistant_code: undefined as string | undefined,
        },
      })),
  );
</script>

{#snippet statCard(
  border: string,
  title: string,
  icon: typeof Calendar,
  iconClass: string,
  value: number | string,
  note: string,
)}
  {@const Icon = icon}
  <Card class="card-hover rounded-2xl border-2 bg-card {border}">
    <CardHeader class="flex flex-row items-center justify-between pb-2">
      <CardTitle class="text-muted-foreground text-sm font-medium">{title}</CardTitle>
      <Icon class="h-4 w-4 {iconClass}" />
    </CardHeader>
    <CardContent>
      <div class="text-2xl font-bold">{value}</div>
      <p class="text-muted-foreground text-xs">{note}</p>
    </CardContent>
  </Card>
{/snippet}

{#snippet pendingCard(
  wrap: string,
  label: string,
  labelClass: string,
  icon: typeof Calendar,
  iconClass: string,
  value: number,
  valueClass: string,
  note: string,
  noteClass: string,
)}
  {@const Icon = icon}
  <div
    class="group cursor-pointer rounded-2xl border-2 p-5 shadow-sm transition-all hover:-translate-y-1 {wrap}"
  >
    <div class="mb-2 flex items-center justify-between">
      <p class="text-xs font-black {labelClass}">{label}</p>
      <Icon class="h-4 w-4 group-hover:animate-bounce {iconClass}" />
    </div>
    <h4 class="mb-1 text-3xl font-black {valueClass}">{value}</h4>
    <p class="text-[10px] font-bold {noteClass}">{note}</p>
  </div>
{/snippet}

{#if role === 'praktikan'}
  <div class="space-y-6">
    <div>
      <h1 class="text-foreground text-2xl font-bold">Selamat Datang, {user.full_name}!</h1>
      <p class="text-muted-foreground mt-1">
        Pantau jadwal, absensi, dan informasi praktikum Anda di sini.
      </p>
    </div>
    <div class="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
      <UserContactCard userId={user.id} />
      {@render statCard(
        'border-fun-blue shadow-bubbly-blue',
        'Kelas Diambil',
        BookOpen,
        'text-blue-600',
        s?.total_kelas || 0,
        'Kelompok Praktikum',
      )}
      {@render statCard(
        'border-fun-purple shadow-bubbly-purple',
        'Kehadiran',
        CheckCircle2,
        'text-purple-600',
        `${s?.attendance_rate || 0}%`,
        'Rasio Kehadiran',
      )}
      {@render statCard(
        'border-fun-yellow shadow-bubbly-yellow',
        'Kritik & Saran',
        MessageSquare,
        'text-orange-600',
        s?.total_feedback || 0,
        'Pesan dikirim',
      )}
    </div>
    <div class="grid gap-6 md:grid-cols-2">
      <Card class="shadow-sm">
        <CardHeader>
          <CardTitle class="flex items-center gap-2 text-lg"
            ><Calendar class="text-primary h-5 w-5" /> Jadwal Kelompok Saya</CardTitle
          >
        </CardHeader>
        <CardContent>
          <div class="space-y-3">
            {#if jadwal.length === 0}
              <p class="text-muted-foreground text-sm">Belum ada jadwal plotting.</p>
            {:else}
              {#each jadwal as schedule, i (i)}
                <div class="bg-muted/50 flex flex-col gap-3 rounded-lg p-3">
                  <div class="flex items-center justify-between">
                    <div>
                      <p class="text-foreground font-medium">{schedule.title}</p>
                      <p class="text-muted-foreground text-xs">
                        {schedule.day_of_week} • {schedule.start_time?.slice(0, 5)} - {schedule.end_time?.slice(
                          0,
                          5,
                        )}
                      </p>
                    </div>
                    <Badge variant="secondary">Kelas {schedule.class_code}</Badge>
                  </div>
                  {#if schedule.assistant.full_name}
                    <div class="flex items-center justify-between gap-2 border-t pt-2">
                      <p class="text-muted-foreground truncate text-xs">
                        Asisten: <span class="text-foreground font-semibold"
                          >{schedule.assistant.full_name}</span
                        >
                      </p>
                      {#if schedule.assistant.phone_number}
                        <a
                          href={getWaLink(
                            schedule.assistant.phone_number,
                            buildWaText(null, {
                              greeting: getGreeting(),
                              honorific: getHonorific(schedule.assistant.assistant_code) || 'Kak',
                              assistant: schedule.assistant.full_name,
                              student: user.full_name,
                              nim: user.username,
                              major: schedule.major,
                              kelas: schedule.class_code,
                            }),
                          )}
                          target="_blank"
                          rel="noreferrer external"
                          class="shrink-0 text-xs font-bold text-[#25D366]"
                        >
                          Chat WA
                        </a>
                      {/if}
                    </div>
                  {/if}
                </div>
              {/each}
            {/if}
          </div>
        </CardContent>
      </Card>
      <Card class="shadow-sm">
        <CardHeader>
          <CardTitle class="flex items-center gap-2 text-lg"
            ><QrCode class="text-primary h-5 w-5" /> Riwayat Absensi</CardTitle
          >
        </CardHeader>
        <CardContent>
          <div class="space-y-3">
            {#if absensi.length === 0}
              <p class="text-muted-foreground text-sm">Belum ada data absensi.</p>
            {:else}
              {#each absensi as log, i (i)}
                <div class="bg-muted/50 flex items-center justify-between rounded-lg p-3">
                  <div>
                    <p class="text-foreground text-sm font-medium">
                      {new Date(log.check_in_time).toLocaleDateString('id-ID', {
                        weekday: 'long',
                        day: 'numeric',
                        month: 'short',
                      })}
                    </p>
                    <p class="text-muted-foreground text-xs">
                      Pukul: {new Date(log.check_in_time).toLocaleTimeString('id-ID', {
                        hour: '2-digit',
                        minute: '2-digit',
                      })}
                    </p>
                  </div>
                  <span
                    class="rounded-full px-2 py-1 text-xs font-medium {log.status === 'Hadir'
                      ? 'bg-green-100 text-green-700'
                      : 'bg-orange-100 text-orange-700'}"
                  >
                    {log.status}
                  </span>
                </div>
              {/each}
            {/if}
          </div>
        </CardContent>
      </Card>
    </div>
  </div>
{:else if role === 'asisten'}
  <div class="space-y-6">
    <div>
      <h1 class="text-foreground text-2xl font-bold">Dashboard Asisten</h1>
      <p class="text-muted-foreground mt-1">
        Pantau total praktikan kelompok Anda dan jadwal bimbingan.
      </p>
    </div>
    <div class="grid gap-4 md:grid-cols-2 lg:grid-cols-5">
      <div class="lg:col-span-2"><UserContactCard userId={user.id} /></div>
      {@render statCard(
        'border-fun-purple shadow-bubbly-purple',
        'Jadwal Jaga',
        Calendar,
        'text-purple-600',
        s?.upcoming_shifts_count || 0,
        'Shift mendatang',
      )}
      {@render statCard(
        'border-fun-blue shadow-bubbly-blue',
        'Praktikan',
        Users,
        'text-blue-600',
        s?.total_students_under_me || 0,
        'Dalam bimbingan',
      )}
      {@render statCard(
        'border-fun-yellow shadow-bubbly-yellow',
        'Log Kehadiran',
        QrCode,
        'text-orange-600',
        s?.my_attendance_count || 0,
        'Absen divalidasi',
      )}
    </div>
  </div>
{:else if role === 'koordinator'}
  <div class="space-y-8">
    <div>
      <h1 class="text-dark mb-2 text-3xl font-black tracking-tight">Pusat Kendali Koordinator</h1>
      <p class="text-sm font-bold text-gray-500">
        Rincian data global laboratorium dan daftar tugas validasi Anda.
      </p>
    </div>
    <div class="grid grid-cols-1 gap-6 md:grid-cols-2 lg:grid-cols-4">
      <div
        class="shadow-bubbly-red card-bubbly group relative cursor-default overflow-hidden rounded-3xl border-2 border-red-100 bg-white p-6"
      >
        <div class="mb-4 flex items-start justify-between">
          <div>
            <p class="text-[10px] font-black tracking-widest text-gray-400 uppercase">
              Total Pengguna
            </p>
            <h3 class="text-dark mt-1 text-4xl font-black">{s?.total_users || 0}</h3>
          </div>
          <div
            class="text-maroon flex h-12 w-12 items-center justify-center rounded-2xl bg-red-50 text-xl shadow-inner transition-transform group-hover:scale-110 group-hover:rotate-6"
          >
            <Database class="text-maroon h-5 w-5" />
          </div>
        </div>
        <p class="text-xs font-bold text-gray-500">Semua akun di sistem</p>
        <Users class="absolute -right-4 -bottom-4 h-20 w-20 text-red-50 opacity-50" />
      </div>
      <div
        class="shadow-bubbly-blue card-bubbly group relative cursor-default overflow-hidden rounded-3xl border-2 border-blue-100 bg-white p-6"
      >
        <div class="mb-4 flex items-start justify-between">
          <div>
            <p class="text-[10px] font-black tracking-widest text-gray-400 uppercase">
              Total Asisten
            </p>
            <h3 class="text-dark mt-1 text-4xl font-black">{s?.total_assistants || 0}</h3>
          </div>
          <div
            class="text-fun-blue flex h-12 w-12 items-center justify-center rounded-2xl bg-blue-50 text-xl shadow-inner transition-transform group-hover:scale-110 group-hover:rotate-6"
          >
            <Users class="text-fun-blue h-5 w-5" />
          </div>
        </div>
        <p class="text-xs font-bold text-gray-500">Asisten lab terdaftar</p>
        <Users class="absolute -right-4 -bottom-4 h-20 w-20 text-blue-50 opacity-50" />
      </div>
      <div
        class="shadow-bubbly-orange card-bubbly group relative cursor-default overflow-hidden rounded-3xl border-2 border-orange-100 bg-white p-6"
      >
        <div class="mb-4 flex items-start justify-between">
          <div>
            <p class="text-[10px] font-black tracking-widest text-gray-400 uppercase">
              Total Praktikan
            </p>
            <h3 class="text-dark mt-1 text-4xl font-black">{s?.total_students || 0}</h3>
          </div>
          <div
            class="text-fun-orange flex h-12 w-12 items-center justify-center rounded-2xl bg-orange-50 text-xl shadow-inner transition-transform group-hover:scale-110 group-hover:rotate-6"
          >
            <Users class="text-fun-orange h-5 w-5" />
          </div>
        </div>
        <p class="text-xs font-bold text-gray-500">Praktikan terdaftar</p>
        <GraduationCap class="absolute -right-4 -bottom-4 h-20 w-20 text-orange-50 opacity-50" />
      </div>
      <div
        class="shadow-bubbly-purple card-bubbly group hover:border-fun-purple relative cursor-pointer overflow-hidden rounded-3xl border-2 border-purple-100 bg-white p-6"
      >
        <div class="mb-4 flex items-start justify-between">
          <div>
            <p class="text-[10px] font-black tracking-widest text-gray-400 uppercase">
              Kritik & Saran
            </p>
            <h3
              class="mt-1 text-4xl font-black {(s?.total_feedback || 0) > 0
                ? 'text-dark'
                : 'text-gray-300'}"
            >
              {s?.total_feedback || 0}
            </h3>
          </div>
          <div
            class="text-fun-purple flex h-12 w-12 items-center justify-center rounded-2xl bg-purple-50 text-xl shadow-inner transition-transform group-hover:scale-110 group-hover:rotate-6"
          >
            <MessageSquare class="text-fun-purple h-5 w-5" />
          </div>
        </div>
        <p class="flex items-center gap-1 text-xs font-bold text-gray-500">
          <CheckCircle2 class="text-fun-green inline h-4 w-4 shrink-0" /> Kotak masuk bersih
        </p>
      </div>
    </div>
    <div>
      <h3 class="text-dark mb-4 flex items-center gap-2 text-lg font-black">
        <CheckCircle2 class="text-maroon h-5 w-5 animate-pulse" /> Tugas & Validasi Pending
      </h3>
      <div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-5">
        {@render pendingCard(
          'bg-blue-50/50 border-blue-100 hover:shadow-[0_4px_0_#BFDBFE]',
          'Jadwal Jaga',
          'text-gray-600',
          Calendar,
          'text-fun-blue',
          s?.upcoming_shifts_count || 0,
          'text-dark',
          'Shift mendatang',
          'text-gray-400',
        )}
        {@render pendingCard(
          'bg-green-50/50 border-green-100 hover:shadow-[0_4px_0_#BBF7D0]',
          'Kelas Bimbingan',
          'text-gray-600',
          BookOpen,
          'text-fun-green',
          s?.total_classes_under_me || 0,
          'text-dark',
          'Jadwal kelompok Anda',
          'text-gray-400',
        )}
        {@render pendingCard(
          'bg-teal-50/50 border-teal-100 hover:shadow-[0_4px_0_#99F6E4]',
          'Praktikan Bimbingan',
          'text-gray-600',
          Users,
          'text-teal-400',
          s?.total_students_under_me || 0,
          'text-dark',
          'Di kelompok Anda',
          'text-gray-400',
        )}
        {@render pendingCard(
          'bg-red-50 border-red-200 hover:shadow-[0_4px_0_#FECACA] relative overflow-hidden',
          'Approval Absen',
          'text-maroon',
          FileText,
          'text-maroon',
          s?.pending_attendance || 0,
          'text-maroon',
          'Praktikan pending',
          'text-red-400',
        )}
        {@render pendingCard(
          'bg-purple-50 border-purple-200 hover:shadow-[0_4px_0_#E9D5FF]',
          'Approval Izin',
          'text-fun-purple',
          ClipboardCheck,
          'text-fun-purple',
          s?.pending_assistant_requests || 0,
          'text-fun-purple',
          'Asisten pending',
          'text-purple-400',
        )}
      </div>
    </div>
  </div>
{:else if role === 'penyewa'}
  <p class="text-muted-foreground text-sm">
    Gunakan menu Sewa &amp; Pinjam untuk mengajukan peminjaman barang laboratorium.
  </p>
{/if}
