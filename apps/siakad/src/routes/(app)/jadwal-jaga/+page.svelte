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
    DialogHeader,
    DialogTitle,
    Input,
    Label,
    Select,
    SelectContent,
    SelectGroup,
    SelectGroupHeading,
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
  import ArrowRight from '@lucide/svelte/icons/arrow-right';
  import Bell from '@lucide/svelte/icons/bell';
  import BellRing from '@lucide/svelte/icons/bell-ring';
  import CalendarDays from '@lucide/svelte/icons/calendar-days';
  import CheckCircle from '@lucide/svelte/icons/circle-check-big';
  import Clock from '@lucide/svelte/icons/clock';
  import FileDown from '@lucide/svelte/icons/file-down';
  import FileUp from '@lucide/svelte/icons/file-up';
  import Filter from '@lucide/svelte/icons/filter';
  import HelpCircle from '@lucide/svelte/icons/circle-question-mark';
  import History from '@lucide/svelte/icons/history';
  import Info from '@lucide/svelte/icons/info';
  import Loader2 from '@lucide/svelte/icons/loader-2';
  import Pencil from '@lucide/svelte/icons/pencil';
  import RefreshCw from '@lucide/svelte/icons/refresh-cw';
  import RotateCcw from '@lucide/svelte/icons/rotate-ccw';
  import Smartphone from '@lucide/svelte/icons/smartphone';
  import Trash2 from '@lucide/svelte/icons/trash-2';
  import UserPlus from '@lucide/svelte/icons/user-plus';
  import Users from '@lucide/svelte/icons/users';
  import XCircle from '@lucide/svelte/icons/circle-x';
  import { createQuery } from '@tanstack/svelte-query';
  import { dashboardStatsQuery } from '#lib/queries/beranda.ts';
  import { createRpcMutation, invalidateRpcs, rpcQueryOptions } from '#lib/queries/rpc.ts';
  import {
    checkAndNotifyUpcomingShifts,
    clearShiftNotifiedHistory,
    getNotificationPermissionState,
    requestNotificationPermission,
    sendBrowserNotification,
    subscribeToWebPush,
    triggerServerWebPush,
    triggerServerWhatsApp,
  } from '#lib/notifications.ts';
  import { getDailyQuote } from '#lib/quotes.ts';
  import { subscribe } from '#lib/realtime/client.ts';
  import {
    assignmentsQuery,
    mapAssignment,
    type Assignment,
  } from '#lib/shift-notifications.svelte.ts';
  import { filterByPeriod, splitActiveHistory } from '#lib/jadwal-jaga.ts';
  import { confirm } from '#lib/stores/confirm.svelte.ts';
  import { toast } from '#lib/toast.ts';
  import type { PageProps } from './$types';

  interface Schedule {
    id: number;
    type?: string;
    day_of_week?: string;
    start_time: string;
    end_time?: string;
    title?: string;
    major?: string;
    class_code?: string;
  }

  interface UserRow {
    id: number;
    full_name: string;
    role: string;
    phone?: string;
    phone_number?: string;
  }

  type Row = Assignment & {
    status?: string | null;
    task_role?: string;
    activity_name?: string;
    activity_date?: string | null;
    original_user_id?: number | null;
    substitute_user_id?: number | null;
  };

  const taskOptions = ['Operator', 'Pemateri', 'Pre/Post Test', 'Program Keterampilan', 'Absen'];
  const activityOptions = [
    'Pengarahan',
    'Praktikum Modul 1',
    'Modul 2',
    'Modul 3&4',
    'Modul 5',
    'Modul 6',
    'Ujian Praktik',
  ];
  const filterLabels: Record<string, string> = {
    all: 'Semua Waktu',
    this_week: 'Minggu Ini',
    this_month: 'Bulan Ini',
    custom: 'Range Tanggal...',
  };

  let { data }: PageProps = $props();
  const user = $derived(data.auth.user);

  const accessQuery = createQuery(() =>
    rpcQueryOptions<boolean>('check_menu_access_secure', {
      p_viewer_id: user.id,
      p_menu_key: '/jadwal-jaga',
    }),
  );
  const hasEditAccess = $derived(!accessQuery.error && !!accessQuery.data);
  createQuery(() => dashboardStatsQuery(user.id));
  const settingsQuery = createQuery(() =>
    rpcQueryOptions<{ wa_templates?: Record<string, string> }[]>(
      'get_system_settings_full_secure',
      {
        p_viewer_id: user.id,
      },
    ),
  );
  const assignQuery = createQuery(() => assignmentsQuery(user.id));
  const schedulesQuery = createQuery(() =>
    rpcQueryOptions<Schedule[]>('get_schedules_secure', { p_viewer_id: user.id }),
  );
  const usersQuery = createQuery(() =>
    rpcQueryOptions<UserRow[]>('get_users_secure', { p_viewer_id: user.id }),
  );

  const waTemplates = $derived(
    settingsQuery.data && settingsQuery.data.length > 0 ? settingsQuery.data[0].wa_templates : null,
  );
  const assignments = $derived((assignQuery.data || []).map(mapAssignment) as Row[]);
  const availableSchedules = $derived(
    (schedulesQuery.data || []).filter((s) => s.type === 'praktikum'),
  );
  const allUsers = $derived(usersQuery.data || []);
  const coordinatorUser = $derived(allUsers.find((u) => u.role === 'koordinator'));
  const coordPhone = $derived(coordinatorUser?.phone || coordinatorUser?.phone_number || null);
  const assistants = $derived(allUsers.filter((u) => ['asisten', 'koordinator'].includes(u.role)));

  const refetchAll = () =>
    invalidateRpcs([
      'get_dashboard_stats_secure',
      'get_system_settings_full_secure',
      'get_schedule_assignments_secure',
      'get_schedules_secure',
      'get_users_secure',
    ]);
  const upsertAssignment = createRpcMutation('upsert_schedule_assignment_secure');
  const deleteAssignment = createRpcMutation('delete_schedule_assignment_secure');
  const updateSwap = createRpcMutation('update_swap_status_secure');
  const approveSwap = createRpcMutation('approve_swap_secure');

  let loading = $state(false);
  let isDialogOpen = $state(false);
  let isImportDialogOpen = $state(false);
  let isEditMode = $state(false);
  let editingId = $state<number | null>(null);
  let filterType = $state('all');
  let customStartDate = $state('');
  let customEndDate = $state('');
  let selectedScheduleId = $state('');
  let selectedAssistantId = $state('');
  let selectedRole = $state('');
  let activityName = $state('');
  let activityDate = $state('');
  let leaveReason = $state('');
  let submittingLeave = $state(false);
  let leaveAssignmentId = $state('');
  let notifPermission = $state<string>(getNotificationPermissionState());

  $effect(() => {
    if (assignQuery.data) checkAndNotifyUpcomingShifts(user, assignments);
  });

  // Pengingat 2 jam, 1 jam, dan 30 menit dicek ulang tiap menit.
  $effect(() => {
    const current = assignments;
    const viewer = user;
    const timer = setInterval(() => {
      if (current.length > 0) checkAndNotifyUpcomingShifts(viewer, current);
    }, 60_000);
    return () => clearInterval(timer);
  });

  $effect(() => {
    const viewer = user;
    return subscribe('jadwal_jaga_live_updates', { table: 'schedule_assignments' }, (payload) => {
      void refetchAll();
      const next = payload.new as Record<string, unknown> | undefined;
      if (payload.eventType === 'INSERT' && next?.user_id === viewer.id) {
        void sendBrowserNotification('🚀 Penugasan Jadwal Jaga Baru!', {
          body: `Kamu ditugaskan sebagai ${next.task_role || 'Petugas'} untuk ${next.activity_name || 'Praktikum'}.`,
          onClickUrl: '/jadwal-jaga',
        });
      } else if (next?.status === 'mencari_pengganti' && next?.user_id !== viewer.id) {
        void sendBrowserNotification('🔄 Permintaan Tukar Jadwal Jaga!', {
          body: `Ada asisten yang sedang mencari pengganti jadwal jaga. Klik untuk melihat & membantu.`,
          onClickUrl: '/jadwal-jaga',
        });
      }
    });
  });

  async function handleEnableNotif() {
    const granted = await requestNotificationPermission();
    notifPermission = getNotificationPermissionState();
    if (granted) {
      toast.success('Notifikasi Aktif!', {
        description: 'Kamu akan menerima notifikasi browser saat ada jadwal jaga.',
      });
      void sendBrowserNotification('🔔 Notifikasi Jadwal Jaga Aktif!', {
        body: 'Kamu akan menerima notifikasi browser otomatis untuk jadwal jaga hari ini & H-1.',
        onClickUrl: '/jadwal-jaga',
      });
      if (assignments.length > 0) checkAndNotifyUpcomingShifts(user, assignments);
    } else {
      toast.warning('Notifikasi Diblokir Browser', {
        description:
          "Klik ikon gembok / setelan situs (🔒/🎛️) di address bar browser -> Ubah Notifikasi menjadi 'Allow / Izinkan'.",
      });
    }
  }

  async function handleTestNotif() {
    await subscribeToWebPush(user);
    void sendBrowserNotification('🔔 Uji Notifikasi Browser', {
      body: 'Notifikasi browser aktif! Device token terdaftar di database.',
      onClickUrl: '/jadwal-jaga',
    });
    toast.success('Device Token VAPID berhasil terdaftar & disimpan di database!');
  }

  const split = $derived(
    splitActiveHistory(filterByPeriod(assignments, filterType, customStartDate, customEndDate)),
  );
  const activeAssignments = $derived(split.active);
  const historyAssignments = $derived(split.history);

  // Ketersediaan asisten tidak pernah diambil di kode lama, jadi semua asisten selalu masuk
  // kelompok "Bentrok / Tidak Ada Info" begitu kelas dipilih (dicatat sebagai temuan).
  const assistantAvailabilities: {
    user_id: number;
    day_of_week: string;
    start_time: string;
    end_time: string;
  }[] = [];
  const smartAssistantList = $derived.by(() => {
    if (!selectedScheduleId) return { available: assistants, unavailable: [] as UserRow[] };
    const targetSched = availableSchedules.find((s) => s.id.toString() === selectedScheduleId);
    if (!targetSched) return { available: assistants, unavailable: [] as UserRow[] };
    const available: UserRow[] = [];
    const unavailable: UserRow[] = [];
    for (const ast of assistants) {
      const isFree = assistantAvailabilities.some(
        (avail) =>
          avail.user_id === ast.id &&
          avail.day_of_week === targetSched.day_of_week &&
          avail.start_time <= targetSched.start_time &&
          avail.end_time >= (targetSched.end_time ?? ''),
      );
      (isFree ? available : unavailable).push(ast);
    }
    return { available, unavailable };
  });

  function openAddDialog() {
    isEditMode = false;
    editingId = null;
    selectedScheduleId = '';
    selectedAssistantId = '';
    selectedRole = '';
    activityName = '';
    activityDate = '';
    isDialogOpen = true;
  }

  function openEditDialog(item: Row) {
    isEditMode = true;
    editingId = item.id;
    selectedScheduleId = String(item.schedule.id);
    selectedAssistantId = String(item.user.id);
    selectedRole = item.task_role ?? '';
    activityName = item.activity_name || '';
    activityDate = item.activity_date || '';
    isDialogOpen = true;
  }

  async function handleSaveAssignment() {
    if (
      !selectedScheduleId ||
      !selectedAssistantId ||
      !selectedRole ||
      !activityDate ||
      !activityName
    ) {
      return toast.error('Lengkapi semua data!');
    }
    if (!hasEditAccess) {
      toast.error('Akses Ditolak', {
        description: 'Anda tidak memiliki izin untuk mengedit jadwal jaga.',
      });
      return;
    }
    loading = true;
    try {
      const { error } = await upsertAssignment.mutateAsync({
        p_caller_id: user.id,
        p_id: isEditMode && editingId ? editingId : 0,
        p_schedule_id: parseInt(selectedScheduleId),
        p_user_id: parseInt(selectedAssistantId),
        p_task_role: selectedRole,
        p_activity_name: activityName,
        p_activity_date: activityDate,
      });
      if (error) throw error;
      clearShiftNotifiedHistory(editingId || undefined);
      if (selectedAssistantId) {
        void triggerServerWebPush(
          parseInt(selectedAssistantId),
          isEditMode ? '✏️ Perubahan Jadwal Jaga!' : '🚀 Penugasan Jadwal Jaga Baru!',
          `Kamu ditugaskan sebagai ${selectedRole} untuk ${activityName} (${activityDate}).\n\n✨ "${getDailyQuote('asisten')}"`,
          '/jadwal-jaga',
        );
      }
      void sendBrowserNotification(
        isEditMode ? '✏️ Jadwal Jaga Diperbarui!' : '🚀 Jadwal Jaga Ditambahkan!',
        {
          body: `Jadwal jaga ${activityName} (${activityDate}) berhasil disimpan.`,
          onClickUrl: '/jadwal-jaga',
        },
      );
      toast.success(isEditMode ? 'Petugas berhasil diperbarui!' : 'Petugas ditambahkan!');
      isDialogOpen = false;
      void refetchAll();
    } catch (err) {
      toast.error((err as Error).message);
    } finally {
      loading = false;
    }
  }

  async function handleDelete(id: number) {
    if (!(await confirm('Hapus petugas ini?'))) return;
    const { error } = await deleteAssignment.mutateAsync({ p_caller_id: user.id, p_id: id });
    if (error) toast.error(error.message);
    else {
      toast.success('Tugas dihapus');
      void refetchAll();
    }
  }

  async function handleDownloadTemplateExcel() {
    const XLSX = await import('xlsx');
    const ws = XLSX.utils.json_to_sheet([
      {
        'Tanggal Kegiatan': '2026-03-20',
        'Nama Kegiatan': 'Praktikum Modul 1',
        Jurusan: 'Teknik Elektro',
        Kelas: 'A',
        'Nama Asisten': 'Budi Asisten',
        Tugas: 'Operator',
      },
    ]);
    const wb = XLSX.utils.book_new();
    XLSX.utils.book_append_sheet(wb, ws, 'Template Plotting');
    XLSX.writeFile(wb, 'Template_Plotting_Asisten.xlsx');
  }

  async function handleExportExcel() {
    const XLSX = await import('xlsx');
    const format = (list: Row[]) =>
      list.map((item) => ({
        Tanggal: item.activity_date
          ? new Date(item.activity_date).toLocaleDateString('id-ID')
          : '-',
        Kegiatan: item.activity_name || '-',
        Hari: item.schedule?.day_of_week || '-',
        Jam: item.schedule?.start_time?.slice(0, 5) || '-',
        Jurusan: item.schedule?.major || '-',
        Kelas: item.schedule?.class_code || '-',
        Asisten: item.user?.full_name || '-',
        Tugas: item.task_role || '-',
        'Status Jaga': item.status || 'aktif',
      }));
    const activeData = format(activeAssignments);
    const historyData = format(historyAssignments);
    const wb = XLSX.utils.book_new();
    XLSX.utils.book_append_sheet(
      wb,
      XLSX.utils.json_to_sheet(
        activeData.length > 0 ? activeData : [{ Message: 'Tidak ada jadwal aktif' }],
      ),
      'Jadwal Aktif',
    );
    XLSX.utils.book_append_sheet(
      wb,
      XLSX.utils.json_to_sheet(
        historyData.length > 0 ? historyData : [{ Message: 'Tidak ada riwayat jaga' }],
      ),
      'Riwayat Jaga',
    );
    XLSX.writeFile(wb, `Rekap_Jadwal_Jaga_${new Date().toISOString().split('T')[0]}.xlsx`);
    toast.success('File Excel berhasil di-download!');
  }

  function handleImportExcel(e: Event & { currentTarget: HTMLInputElement }) {
    const input = e.currentTarget;
    const file = input.files?.[0];
    if (!file) return;
    if (!hasEditAccess) {
      toast.error('Akses Ditolak', {
        description: 'Anda tidak memiliki izin untuk mengimport jadwal.',
      });
      return;
    }
    loading = true;
    const reader = new FileReader();
    reader.onload = async (evt) => {
      try {
        const XLSX = await import('xlsx');
        const wb = XLSX.read(evt.target?.result, { type: 'binary' });
        const rows = XLSX.utils.sheet_to_json<Record<string, string>>(wb.Sheets[wb.SheetNames[0]]);
        const days = ['Minggu', 'Senin', 'Selasa', 'Rabu', 'Kamis', 'Jumat', 'Sabtu'];
        const norm = (v: unknown) =>
          String(v ?? '')
            .toLowerCase()
            .trim();
        const errors: string[] = [];
        const payload = rows
          .map((row, index) => {
            const assistantName = row['Nama Asisten'] || row['nama_asisten'];
            const major = row['Jurusan'] || row['jurusan'];
            const classCode = row['Kelas'] || row['kelas'];
            const rowDate = row['Tanggal Kegiatan'] || row['tanggal_kegiatan'];
            if (!assistantName || !classCode || !major || !rowDate) return null;
            const dayName = days[new Date(rowDate).getDay()];
            const assistant = assistants.find((a) => norm(a.full_name) === norm(assistantName));
            const schedule = availableSchedules.find(
              (s) =>
                norm(s.major) === norm(major) &&
                norm(s.class_code) === norm(classCode) &&
                norm(s.day_of_week) === dayName.toLowerCase(),
            );
            if (!assistant) {
              errors.push(`Baris ${index + 2}: Asisten "${assistantName}" tidak ditemukan.`);
              return null;
            }
            if (!schedule) {
              errors.push(
                `Baris ${index + 2}: Jadwal "${major} Kelas ${classCode}" tidak ditemukan di hari ${dayName}.`,
              );
              return null;
            }
            return {
              activity_date: rowDate,
              activity_name: row['Nama Kegiatan'] || row['nama_kegiatan'],
              schedule_id: schedule.id,
              user_id: assistant.id,
              task_role: row['Tugas'] || row['tugas'],
            };
          })
          .filter((p) => p !== null);
        if (errors.length > 0) {
          const shown = errors.slice(0, 3);
          if (errors.length > 3) shown.push(`...dan ${errors.length - 3} lainnya.`);
          toast.error('Gagal Import', {
            description: shown.join('\n'),
            classes: { description: 'whitespace-pre-line max-h-[150px] overflow-auto' },
          });
        }
        if (payload.length > 0) {
          let success = 0;
          for (const p of payload) {
            const { error } = await upsertAssignment.mutateAsync({
              p_caller_id: user.id,
              p_id: 0,
              p_schedule_id: parseInt(String(p.schedule_id)),
              p_user_id: p.user_id,
              p_task_role: p.task_role,
              p_activity_name: p.activity_name,
              p_activity_date: p.activity_date,
            });
            if (!error) success++;
          }
          toast.success(`${success} jadwal berhasil diimpor!`);
          isImportDialogOpen = false;
          void refetchAll();
        } else if (errors.length === 0) {
          toast.error('Gagal', { description: 'Format Excel salah atau data tidak lengkap.' });
        }
      } catch (err) {
        toast.error('Gagal mengimpor Excel: ' + (err as Error).message);
      } finally {
        loading = false;
      }
    };
    reader.readAsBinaryString(file);
    input.value = '';
  }

  function getGreeting() {
    const hour = new Date().getHours();
    if (hour < 11) return 'Pagi';
    if (hour < 15) return 'Siang';
    if (hour < 19) return 'Sore';
    return 'Malam';
  }

  async function handleRequestLeave() {
    if (!leaveAssignmentId) return toast.error('Pilih jadwal yang mau di-swap!');
    if (!leaveReason) return toast.error('Alasan harus diisi untuk Koordinator!');
    submittingLeave = true;
    try {
      const { error: updateError } = await updateSwap.mutateAsync({
        p_caller_id: user.id,
        p_id: parseInt(leaveAssignmentId),
        p_status: 'mencari_pengganti',
        p_substitute_id: null,
      });
      if (updateError) throw updateError;
      const targetAssignment = assignments.find((a) => a.id.toString() === leaveAssignmentId);
      const scheduleInfo = targetAssignment
        ? `${targetAssignment.schedule?.day_of_week}, ${targetAssignment.schedule?.start_time?.slice(0, 5)} WIB (${targetAssignment.activity_name})`
        : 'Jadwal Jaga';
      const quote = getDailyQuote('asisten');
      const coordinator = assistants.find((a) => a.role === 'koordinator');
      if (coordinator) {
        await triggerServerWebPush(
          coordinator.id,
          '🔄 Permohonan Swap Jadwal Jaga Baru',
          `Asisten ${user.full_name} mengajukan swap jadwal: ${scheduleInfo}.\nAlasan: ${leaveReason}\n\n✨ "${quote}"`,
          '/jadwal-jaga',
        );
      }
      const targetPhone = coordinator?.phone || coordinator?.phone_number || coordPhone;
      if (targetPhone) {
        let cleanPhone = targetPhone.replace(/\D/g, '');
        if (cleanPhone.startsWith('0')) cleanPhone = '62' + cleanPhone.slice(1);
        let text = `Selamat ${getGreeting()} mas, saya *${user.full_name}* izin tidak dapat jaga dan sedang mencari pengganti (Swap).\nAlasan: ${leaveReason}`;
        if (waTemplates?.asisten_swap) {
          text = waTemplates.asisten_swap
            .replace(/{{waktu}}/g, getGreeting())
            .replace(/{{nama}}/g, user.full_name || '')
            .replace(/{{jadwal}}/g, scheduleInfo)
            .replace(/{{alasan}}/g, leaveReason);
        }
        await triggerServerWhatsApp(0, text, cleanPhone);
        window.open(`https://wa.me/${cleanPhone}?text=${encodeURIComponent(text)}`, '_blank');
        toast.success('Jadwal dilempar ke Bursa! Mengalihkan ke WhatsApp Koordinator...');
      } else {
        toast.success('Berhasil! Permohonan swap berhasil dikirim ke Koordinator.');
      }
      leaveReason = '';
      leaveAssignmentId = '';
    } catch (err) {
      toast.error((err as Error).message);
    } finally {
      submittingLeave = false;
    }
  }

  async function handleOfferSubstitute(assignmentId: number) {
    if (!(await confirm('Anda yakin bersedia menggantikan jadwal ini?'))) return;
    try {
      const { error } = await updateSwap.mutateAsync({
        p_caller_id: user.id,
        p_id: assignmentId,
        p_status: 'menunggu_persetujuan',
        p_substitute_id: user.id,
      });
      if (error) throw error;
      const coordinator = assistants.find((a) => a.role === 'koordinator');
      if (coordinator) {
        await triggerServerWebPush(
          coordinator.id,
          '🙋 Penawaran Pengganti Shift (Swap)',
          `Asisten ${user.full_name} menawarkan diri untuk menggantikan jadwal ID #${assignmentId}. Mohon persetujuan Koordinator.`,
          '/jadwal-jaga',
        );
      }
      toast.success('Berhasil menawarkan diri! Menunggu persetujuan Koordinator.');
    } catch (err) {
      toast.error('Gagal menawarkan diri: ' + (err as Error).message);
    }
  }

  async function handleApproveSwap(
    assignmentId: number,
    substituteId: number | null | undefined,
    originalUserId: number | null | undefined,
  ) {
    if (user.role !== 'koordinator' && user.role !== 'asisten') {
      toast.error('Akses Ditolak');
      return;
    }
    if (!(await confirm('Setujui pertukaran jadwal ini?'))) return;
    try {
      const { error } = await approveSwap.mutateAsync({ p_caller_id: user.id, p_id: assignmentId });
      if (error) throw error;
      if (originalUserId) {
        await triggerServerWebPush(
          originalUserId,
          '✅ Swap Jadwal Disetujui Koordinator',
          `Permohonan swap jadwal jaga kamu telah disetujui Koordinator!`,
          '/jadwal-jaga',
        );
      }
      if (substituteId) {
        await triggerServerWebPush(
          substituteId,
          '✅ Kamu Resmi Menjadi Pengganti Shift',
          `Selamat! Penawaran kamu untuk menggantikan shift telah disetujui Koordinator.`,
          '/jadwal-jaga',
        );
      }
      toast.success('Pertukaran jadwal berhasil disetujui!');
    } catch (err) {
      toast.error('Gagal menyetujui: ' + (err as Error).message);
    }
  }

  async function handleRejectSwap(assignmentId: number) {
    if (!(await confirm('Tolak penawaran ini dan kembalikan jadwal awal?'))) return;
    try {
      const { error } = await updateSwap.mutateAsync({
        p_caller_id: user.id,
        p_id: assignmentId,
        p_status: 'mencari_pengganti',
        p_substitute_id: null,
      });
      if (error) throw error;
      toast.info('Penawaran ditolak. Jadwal dikembalikan ke sebelumnya.');
    } catch (err) {
      toast.error('Gagal menolak: ' + (err as Error).message);
    }
  }

  async function handleCancelSwapKoordinator(assignmentId: number) {
    if (!(await confirm('Batalkan pencarian dan paksa jadwal kembali aktif ke asisten awal?')))
      return;
    try {
      const { error } = await updateSwap.mutateAsync({
        p_caller_id: user.id,
        p_id: assignmentId,
        p_status: 'aktif',
        p_substitute_id: null,
      });
      if (error) throw error;
      toast.success('Pencarian dibatalkan. Jadwal kembali aktif.');
      void refetchAll();
    } catch (err) {
      toast.error('Gagal membatalkan: ' + (err as Error).message);
    }
  }

  const myActiveAssignments = $derived(
    activeAssignments.filter(
      (item) => item.user?.id === user.id && (!item.status || item.status === 'aktif'),
    ),
  );
  const swapRequests = $derived(
    activeAssignments.filter((item) => item.status === 'mencari_pengganti'),
  );
  const mySwapRequestsStatus = $derived(
    activeAssignments.filter(
      (item) =>
        item.user?.id === user.id &&
        (item.status === 'mencari_pengganti' || item.status === 'menunggu_persetujuan'),
    ),
  );
  const pendingApprovals = $derived(
    activeAssignments.filter((item) => item.status === 'menunggu_persetujuan'),
  );
  const swappedHistory = $derived(
    assignments.filter(
      (item) => item.original_user_id !== null && item.original_user_id !== undefined,
    ),
  );

  const nameOf = (id: number | null | undefined) => assistants.find((a) => a.id === id)?.full_name;
  const shortDate = (value: string | null | undefined, month: 'short' | 'long' = 'short') =>
    value ? new Date(value).toLocaleDateString('id-ID', { day: 'numeric', month }) : '';
  const fullDate = (value: string | null | undefined) =>
    value
      ? new Date(value).toLocaleDateString('id-ID', {
          day: 'numeric',
          month: 'short',
          year: 'numeric',
        })
      : '-';
  const scheduleLabel = (s: Schedule) =>
    `${s.day_of_week}, ${s.start_time.slice(0, 5)} - ${s.title} (${s.major} Kelas ${s.class_code})`;
  const myAssignmentLabel = (item: Row) =>
    `${shortDate(item.activity_date)} - ${item.activity_name} (${item.schedule?.start_time?.slice(0, 5)})`;
</script>

{#snippet assignmentTable(dataList: Row[], isHistory: boolean)}
  <Table>
    <TableHeader>
      <TableRow>
        <TableHead>Tanggal</TableHead>
        <TableHead>Kegiatan & Waktu</TableHead>
        <TableHead>Jurusan & Kelas</TableHead>
        <TableHead>Asisten</TableHead>
        <TableHead>Status / Tugas</TableHead>
        {#if hasEditAccess}<TableHead class="text-right">Aksi</TableHead>{/if}
      </TableRow>
    </TableHeader>
    <TableBody>
      {#if dataList.length === 0}
        <TableRow
          ><TableCell colspan={6} class="text-muted-foreground py-8 text-center"
            >Tidak ada data.</TableCell
          ></TableRow
        >
      {:else}
        {#each dataList as item (item.id)}
          <TableRow
            data-testid="assignment-row"
            class={isHistory
              ? 'bg-gray-50/50 opacity-75'
              : item.status === 'mencari_pengganti' || item.status === 'menunggu_persetujuan'
                ? 'bg-yellow-50/50'
                : ''}
          >
            <TableCell class="font-medium">{fullDate(item.activity_date)}</TableCell>
            <TableCell>
              <div class="text-sm font-bold">{item.activity_name || 'Praktikum'}</div>
              <div class="text-muted-foreground mt-1 flex items-center gap-1 text-xs">
                <Clock class="h-3 w-3" />
                {item.schedule?.day_of_week}, {item.schedule?.start_time?.slice(0, 5)}
              </div>
            </TableCell>
            <TableCell>
              <div class="text-sm font-medium">{item.schedule?.major || '-'}</div>
              <div class="text-muted-foreground mt-1 text-xs">
                Kelas <span class="rounded border bg-gray-100 px-1 py-0.5 font-bold"
                  >{item.schedule?.class_code || '-'}</span
                >
              </div>
            </TableCell>
            <TableCell class={item.status !== 'aktif' ? 'font-medium text-yellow-700' : ''}>
              {item.user?.full_name}
            </TableCell>
            <TableCell>
              <div class="flex flex-col items-start gap-1">
                <Badge variant={isHistory ? 'outline' : 'secondary'}>{item.task_role}</Badge>
                {#if item.status === 'mencari_pengganti'}<Badge
                    class="bg-yellow-500 text-[10px] hover:bg-yellow-600">Cari Pengganti</Badge
                  >{/if}
                {#if item.status === 'menunggu_persetujuan'}<Badge
                    class="bg-blue-500 text-[10px] hover:bg-blue-600">Menunggu Approval</Badge
                  >{/if}
              </div>
            </TableCell>
            {#if hasEditAccess}
              <TableCell class="text-right">
                <div class="flex justify-end gap-1">
                  <Button
                    variant="ghost"
                    size="icon"
                    aria-label="Edit penugasan"
                    class="h-8 w-8 text-blue-500 hover:bg-blue-50"
                    onclick={() => openEditDialog(item)}><Pencil class="h-4 w-4" /></Button
                  >
                  <Button
                    variant="ghost"
                    size="icon"
                    aria-label="Hapus penugasan"
                    class="h-8 w-8 text-red-500 hover:bg-red-50"
                    onclick={() => handleDelete(item.id)}><Trash2 class="h-4 w-4" /></Button
                  >
                </div>
              </TableCell>
            {/if}
          </TableRow>
        {/each}
      {/if}
    </TableBody>
  </Table>
{/snippet}

{#snippet optionSelect(label: string, value: string, set: (v: string) => void, options: string[])}
  <div class="space-y-2">
    <Label>{label}</Label>
    <Select type="single" {value} onValueChange={set}>
      <SelectTrigger class="w-full" aria-label={label}
        >{#if value}{value}{:else}<span class="text-muted-foreground">Pilih...</span
          >{/if}</SelectTrigger
      >
      <SelectContent>
        {#each options as opt (opt)}<SelectItem value={opt} label={opt} />{/each}
        {#if value && !options.includes(value)}<SelectItem {value} label={value} />{/if}
      </SelectContent>
    </Select>
  </div>
{/snippet}

<div class="space-y-6">
  <div class="flex flex-col items-start justify-between gap-4 sm:flex-row sm:items-center">
    <div class="flex flex-wrap items-center gap-3">
      <h1 class="flex items-center gap-2 text-2xl font-bold">
        <Users class="text-primary" /> Plotting Jadwal Jaga
      </h1>
      <Badge
        variant="outline"
        class="animate-pulse border-green-200 bg-green-50 text-[10px] text-green-600 shadow-sm"
      >
        <span class="mr-1.5 inline-block h-1.5 w-1.5 rounded-full bg-green-500"></span> Live Update
      </Badge>
      {#if notifPermission === 'granted'}
        <Button
          variant="outline"
          size="sm"
          class="h-7 border-emerald-200 bg-emerald-50 text-[11px] text-emerald-700 shadow-sm hover:bg-emerald-100"
          onclick={handleTestNotif}
        >
          <Bell class="mr-1 h-3.5 w-3.5 text-emerald-600" />
          Notifikasi Browser Aktif
        </Button>
      {:else}
        <Button
          variant="outline"
          size="sm"
          class="h-7 border-amber-200 bg-amber-50 text-[11px] text-amber-700 shadow-sm hover:bg-amber-100"
          onclick={handleEnableNotif}
        >
          <BellRing class="mr-1 h-3.5 w-3.5 animate-bounce text-amber-600" />
          Aktifkan Notifikasi
        </Button>
      {/if}
    </div>

    {#if hasEditAccess}
      <div class="flex flex-wrap gap-2">
        <Button
          variant="outline"
          class="border-blue-600 text-blue-700 hover:bg-blue-50"
          onclick={handleExportExcel}
        >
          <FileDown class="mr-2 h-4 w-4" /> Export Excel
        </Button>
        <Button
          variant="outline"
          class="border-green-600 text-green-700 hover:bg-green-50"
          onclick={() => (isImportDialogOpen = true)}
          ><FileUp class="mr-2 h-4 w-4" /> Import Excel</Button
        >
        <Dialog bind:open={isImportDialogOpen}>
          <DialogContent class="max-w-3xl sm:max-w-3xl">
            <DialogHeader>
              <DialogTitle>Import Data Plotting Asisten (Excel)</DialogTitle>
              <DialogDescription>
                Gunakan fitur ini untuk mengunggah jadwal jaga asisten secara massal menggunakan
                file Excel.
              </DialogDescription>
            </DialogHeader>
            <div class="space-y-4 py-2">
              <div class="flex items-start gap-3 rounded-lg border border-blue-100 bg-blue-50 p-4">
                <Info class="mt-0.5 h-5 w-5 text-blue-600" />
                <div class="text-sm text-blue-800">
                  <p class="mb-1 font-semibold">Cara Menggunakan Fitur Import Excel:</p>
                  <ol class="ml-1 list-inside list-decimal space-y-1">
                    <li>Download template Excel menggunakan tombol di bawah.</li>
                    <li>
                      Sistem baru: Gunakan kolom <strong>Jurusan</strong> dan
                      <strong>Kelas</strong>.
                    </li>
                    <li>
                      Sistem akan otomatis mendeteksi Hari berdasarkan <strong
                        >Tanggal Kegiatan</strong
                      >.
                    </li>
                    <li>Pastikan Nama Asisten sesuai dengan yang terdaftar.</li>
                  </ol>
                  <Button
                    size="sm"
                    onclick={handleDownloadTemplateExcel}
                    class="mt-3 bg-blue-600 hover:bg-blue-700"
                    ><FileDown class="mr-2 h-4 w-4" /> Download Template Excel</Button
                  >
                </div>
              </div>
              <div class="pt-2">
                <Label for="import-excel" class="mb-2 block font-semibold"
                  >Upload File Excel (.xlsx) yang sudah diisi:</Label
                >
                <Input
                  id="import-excel"
                  type="file"
                  accept=".xlsx, .xls"
                  onchange={handleImportExcel}
                  disabled={loading}
                />
              </div>
            </div>
          </DialogContent>
        </Dialog>

        <Button onclick={openAddDialog}><UserPlus class="mr-2 h-4 w-4" /> Tambah Petugas</Button>
        <Dialog bind:open={isDialogOpen}>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>{isEditMode ? 'Edit Penugasan Asisten' : 'Tugaskan Asisten'}</DialogTitle
              >
              <DialogDescription>
                {isEditMode
                  ? 'Perbarui detail penugasan asisten untuk kegiatan praktikum.'
                  : 'Tentukan asisten dan peran mereka untuk jadwal praktikum tertentu.'}
              </DialogDescription>
            </DialogHeader>
            <div class="space-y-4 py-4">
              <div class="grid grid-cols-2 gap-4">
                <div class="space-y-2">
                  <Label for="activity-date">Tanggal Bertugas</Label><Input
                    id="activity-date"
                    type="date"
                    bind:value={activityDate}
                  />
                </div>
                {@render optionSelect(
                  'Kegiatan',
                  activityName,
                  (v) => (activityName = v),
                  activityOptions,
                )}
              </div>
              <div class="space-y-2">
                <Label>Target Kelas</Label>
                <Select type="single" bind:value={selectedScheduleId}>
                  <SelectTrigger class="w-full" aria-label="Target Kelas">
                    {#if selectedScheduleId}
                      {@const s = availableSchedules.find(
                        (x) => x.id.toString() === selectedScheduleId,
                      )}
                      {s ? scheduleLabel(s) : ''}
                    {:else}<span class="text-muted-foreground">Pilih Kelas...</span>{/if}
                  </SelectTrigger>
                  <SelectContent>
                    {#each availableSchedules as s (s.id)}
                      <SelectItem value={s.id.toString()} label={scheduleLabel(s)} />
                    {/each}
                  </SelectContent>
                </Select>
              </div>
              <div class="grid grid-cols-2 gap-4">
                <div class="space-y-2">
                  <Label>Asisten</Label>
                  <Select
                    type="single"
                    bind:value={selectedAssistantId}
                    disabled={!selectedScheduleId}
                  >
                    <SelectTrigger class="w-full" aria-label="Asisten">
                      {#if selectedAssistantId}{nameOf(parseInt(selectedAssistantId))}{:else}<span
                          class="text-muted-foreground">Pilih...</span
                        >{/if}
                    </SelectTrigger>
                    <SelectContent>
                      <SelectGroup>
                        <SelectGroupHeading class="bg-green-100 text-green-800"
                          >✅ Bisa Hadir (Jadwal Kosong)</SelectGroupHeading
                        >
                        {#each smartAssistantList.available as a (a.id)}
                          <SelectItem value={a.id.toString()} label={a.full_name} />
                        {/each}
                      </SelectGroup>
                      <SelectGroup class="mt-2">
                        <SelectGroupHeading class="bg-red-50 text-red-700"
                          >❌ Bentrok / Tidak Ada Info</SelectGroupHeading
                        >
                        {#each smartAssistantList.unavailable as a (a.id)}
                          <SelectItem value={a.id.toString()} label={a.full_name} />
                        {/each}
                      </SelectGroup>
                    </SelectContent>
                  </Select>
                </div>
                {@render optionSelect(
                  'Tugas',
                  selectedRole,
                  (v) => (selectedRole = v),
                  taskOptions,
                )}
              </div>
              <Button onclick={handleSaveAssignment} disabled={loading} class="mt-2 w-full"
                >{#if loading}<Loader2 class="mr-2 h-4 w-4 animate-spin" />{:else}Simpan Perubahan{/if}</Button
              >
            </div>
          </DialogContent>
        </Dialog>
      </div>
    {/if}
  </div>

  <div class="flex w-fit flex-wrap items-center gap-2 rounded border bg-white p-2 shadow-sm">
    <Filter class="text-muted-foreground ml-2 h-4 w-4" />
    <Select type="single" bind:value={filterType}>
      <SelectTrigger
        class="h-8 w-[160px] border-none bg-transparent font-medium shadow-none focus:ring-0"
        >{filterLabels[filterType]}</SelectTrigger
      >
      <SelectContent>
        {#each Object.entries(filterLabels) as [value, label] (value)}
          <SelectItem {value} {label} />
        {/each}
      </SelectContent>
    </Select>
    {#if filterType === 'custom'}
      <div class="ml-1 flex items-center gap-2 border-l pl-3">
        <Input
          type="date"
          aria-label="Tanggal mulai"
          class="h-8 w-[120px] text-xs"
          bind:value={customStartDate}
        />
        <span class="text-muted-foreground text-xs font-medium">s/d</span>
        <Input
          type="date"
          aria-label="Tanggal selesai"
          class="h-8 w-[120px] text-xs"
          bind:value={customEndDate}
        />
      </div>
    {/if}
  </div>

  <Tabs value="active">
    <TabsList
      class="mb-4 grid w-full {hasEditAccess
        ? 'grid-cols-3 lg:w-[500px]'
        : 'grid-cols-2 lg:w-[350px]'}"
    >
      <TabsTrigger value="active">Jadwal Aktif</TabsTrigger>
      {#if hasEditAccess}<TabsTrigger value="history">Riwayat Jaga</TabsTrigger>{/if}
      <TabsTrigger value="izin">Perizinan Jaga</TabsTrigger>
    </TabsList>

    <TabsContent value="active">
      <Card
        ><CardContent class="pt-6">{@render assignmentTable(activeAssignments, false)}</CardContent
        ></Card
      >
    </TabsContent>

    {#if hasEditAccess}
      <TabsContent value="history" class="space-y-6">
        <Card class="border-l-4 border-l-blue-500 opacity-95 shadow-sm">
          <CardHeader>
            <CardTitle class="flex items-center gap-2 text-blue-700">
              <History class="h-5 w-5" /> Riwayat Pertukaran (Swap) Asisten
            </CardTitle>
          </CardHeader>
          <CardContent>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Pertukaran Asisten</TableHead>
                  <TableHead>Tanggal Pertukaran</TableHead>
                  <TableHead>Jadwal & Modul</TableHead>
                  <TableHead class="text-right">Status</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {#if swappedHistory.length === 0}
                  <TableRow
                    ><TableCell colspan={4} class="text-muted-foreground py-6 text-center"
                      >Belum ada riwayat pertukaran asisten.</TableCell
                    ></TableRow
                  >
                {:else}
                  {#each swappedHistory as item (item.id)}
                    <TableRow class="bg-gray-50/50">
                      <TableCell>
                        <div class="text-sm font-bold text-gray-800">
                          {item.original_user?.full_name}
                        </div>
                        <div class="text-muted-foreground mt-1 flex items-center gap-1 text-xs">
                          <ArrowRight class="h-3 w-3 text-blue-500" /> Digantikan oleh:
                          <span class="font-semibold text-green-700">{item.user?.full_name}</span>
                        </div>
                      </TableCell>
                      <TableCell
                        ><div class="text-sm font-medium">
                          {fullDate(item.activity_date)}
                        </div></TableCell
                      >
                      <TableCell>
                        <div class="text-sm font-bold text-blue-700">
                          {item.schedule?.day_of_week}, {item.schedule?.start_time?.slice(0, 5)}
                        </div>
                        <div class="text-muted-foreground mt-1 text-xs">
                          {item.activity_name} ({item.schedule?.major} Kls {item.schedule
                            ?.class_code})
                        </div>
                      </TableCell>
                      <TableCell class="text-right">
                        <Badge
                          class="border-none bg-green-100 text-green-800 shadow-none hover:bg-green-100"
                          >Disetujui & Fix</Badge
                        >
                      </TableCell>
                    </TableRow>
                  {/each}
                {/if}
              </TableBody>
            </Table>
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle class="text-muted-foreground flex items-center gap-2 text-base">
              <CalendarDays class="h-4 w-4" /> Data Penugasan Masa Lalu (Selesai)
            </CardTitle>
          </CardHeader>
          <CardContent>{@render assignmentTable(historyAssignments, true)}</CardContent>
        </Card>
      </TabsContent>
    {/if}

    <TabsContent value="izin" class="space-y-6">
      {#if hasEditAccess && pendingApprovals.length > 0}
        <Card class="border-blue-200 shadow-md">
          <CardHeader class="border-b border-blue-100 bg-blue-50/70 pb-3">
            <CardTitle class="flex items-center gap-2 text-base text-blue-800">
              <CheckCircle class="h-5 w-5" /> Menunggu Persetujuan Pertukaran (Swap) Anda
            </CardTitle>
          </CardHeader>
          <CardContent class="pt-4">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Pertukaran Asisten</TableHead>
                  <TableHead>Jadwal Kelas & Tugas</TableHead>
                  <TableHead class="text-right">Aksi Koordinator</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {#each pendingApprovals as item (item.id)}
                  <TableRow class="bg-blue-50/20">
                    <TableCell>
                      <div class="flex items-center gap-2 text-sm">
                        <span class="font-medium text-red-600 line-through decoration-red-300"
                          >{item.user?.full_name}</span
                        >
                        <ArrowRight class="text-muted-foreground h-4 w-4" />
                        <span class="font-bold text-green-700"
                          >{nameOf(item.substitute_user_id) || 'Asisten Baru'}</span
                        >
                      </div>
                    </TableCell>
                    <TableCell>
                      <div class="font-medium">
                        {item.activity_name} ({item.schedule?.major} Kls {item.schedule
                          ?.class_code})
                      </div>
                      <div class="text-muted-foreground mt-1 text-xs">
                        {shortDate(item.activity_date)} • Jam {item.schedule?.start_time?.slice(
                          0,
                          5,
                        )} •
                        <span class="font-bold">{item.task_role}</span>
                      </div>
                    </TableCell>
                    <TableCell class="text-right">
                      <div class="flex justify-end gap-2">
                        <Button
                          size="sm"
                          variant="outline"
                          class="border-red-200 text-red-600 hover:bg-red-50"
                          onclick={() => handleRejectSwap(item.id)}
                        >
                          <XCircle class="mr-1 h-4 w-4" /> Tolak
                        </Button>
                        <Button
                          size="sm"
                          class="bg-blue-600 hover:bg-blue-700"
                          onclick={() =>
                            handleApproveSwap(item.id, item.substitute_user_id, item.user?.id)}
                        >
                          <CheckCircle class="mr-1 h-4 w-4" /> Setujui Swap
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                {/each}
              </TableBody>
            </Table>
          </CardContent>
        </Card>
      {/if}

      <Card class="border-yellow-200 shadow-sm">
        <CardHeader class="border-b border-yellow-100 bg-yellow-50/50 pb-4">
          <CardTitle class="flex items-center gap-2 text-lg text-yellow-800">
            <HelpCircle class="h-5 w-5" /> Tabel Pengganti Jadwal (Dicari Pengganti)
          </CardTitle>
        </CardHeader>
        <CardContent class="pt-4">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Peminta Izin</TableHead>
                <TableHead>Waktu & Kelas</TableHead>
                <TableHead>Tugas</TableHead>
                <TableHead class="text-right">Aksi</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {#if swapRequests.length === 0}
                <TableRow
                  ><TableCell colspan={4} class="text-muted-foreground py-6 text-center"
                    >Tidak ada jadwal yang membutuhkan pengganti saat ini.</TableCell
                  ></TableRow
                >
              {:else}
                {#each swapRequests as item (item.id)}
                  <TableRow>
                    <TableCell class="font-bold text-yellow-900">{item.user?.full_name}</TableCell>
                    <TableCell>
                      <div class="font-medium">
                        {item.activity_name} ({item.schedule?.major} Kls {item.schedule
                          ?.class_code})
                      </div>
                      <div class="text-muted-foreground mt-1 text-xs">
                        {shortDate(item.activity_date)} • Jam {item.schedule?.start_time?.slice(
                          0,
                          5,
                        )}
                      </div>
                    </TableCell>
                    <TableCell><Badge variant="outline">{item.task_role}</Badge></TableCell>
                    <TableCell class="text-right">
                      {#if item.user?.id === user.id}
                        <Badge class="bg-yellow-100 text-yellow-800">Menunggu Pengganti...</Badge>
                      {:else}
                        <Button
                          size="sm"
                          class="bg-blue-600 hover:bg-blue-700"
                          onclick={() => handleOfferSubstitute(item.id)}>Tawarkan Diri</Button
                        >
                      {/if}
                      {#if hasEditAccess}
                        <Button
                          size="sm"
                          variant="outline"
                          aria-label="Batalkan pencarian"
                          class="ml-2 h-8 px-2 text-gray-500 hover:text-red-600"
                          onclick={() => handleCancelSwapKoordinator(item.id)}
                        >
                          <RotateCcw class="h-3 w-3" />
                        </Button>
                      {/if}
                    </TableCell>
                  </TableRow>
                {/each}
              {/if}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      <div class="grid grid-cols-1 gap-6 md:grid-cols-3">
        <Card class="h-fit md:col-span-1">
          <CardHeader
            ><CardTitle class="flex items-center gap-2 text-base"
              ><RefreshCw class="h-4 w-4" /> Lempar Jadwal ke Tabel</CardTitle
            ></CardHeader
          >
          <CardContent class="space-y-4">
            <div class="space-y-2">
              <Label>Jadwal yang Ditinggalkan</Label>
              <Select type="single" bind:value={leaveAssignmentId}>
                <SelectTrigger class="w-full" aria-label="Jadwal yang Ditinggalkan">
                  {#if leaveAssignmentId}
                    {@const a = myActiveAssignments.find(
                      (x) => x.id.toString() === leaveAssignmentId,
                    )}
                    {a ? myAssignmentLabel(a) : ''}
                  {:else}<span class="text-muted-foreground">Pilih jadwal aktif Anda...</span>{/if}
                </SelectTrigger>
                <SelectContent>
                  {#if myActiveAssignments.length === 0}
                    <div class="text-muted-foreground p-2 text-center text-sm italic">
                      Tidak ada jadwal aktif
                    </div>
                  {:else}
                    {#each myActiveAssignments as item (item.id)}
                      <SelectItem value={item.id.toString()} label={myAssignmentLabel(item)} />
                    {/each}
                  {/if}
                </SelectContent>
              </Select>
            </div>
            <div class="space-y-2">
              <Label for="leave-reason">Alasan (Untuk Koordinator)</Label>
              <Input
                id="leave-reason"
                placeholder="Contoh: Sakit demam..."
                bind:value={leaveReason}
              />
            </div>
            <Button
              class="w-full bg-green-600 hover:bg-green-700"
              onclick={handleRequestLeave}
              disabled={submittingLeave}
            >
              {#if submittingLeave}<Loader2 class="mr-2 h-4 w-4 animate-spin" />{:else}<Smartphone
                  class="mr-2 h-4 w-4"
                />Kirim Alasan Swap{/if}
            </Button>
          </CardContent>
        </Card>
        <Card class="md:col-span-2">
          <CardHeader
            ><CardTitle class="flex items-center gap-2 text-base"
              ><Clock class="h-4 w-4" /> Status Permintaan Swap Saya</CardTitle
            ></CardHeader
          >
          <CardContent>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Jadwal Dilepas</TableHead>
                  <TableHead>Kandidat Pengganti</TableHead>
                  <TableHead>Status</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {#if mySwapRequestsStatus.length === 0}
                  <TableRow
                    ><TableCell colspan={3} class="text-muted-foreground py-6 text-center"
                      >Anda belum/tidak memiliki permintaan pertukaran jadwal yang sedang aktif.</TableCell
                    ></TableRow
                  >
                {:else}
                  {#each mySwapRequestsStatus as item (item.id)}
                    {@const waiting = item.status === 'menunggu_persetujuan'}
                    <TableRow>
                      <TableCell class="text-xs">
                        <div class="font-bold">{item.activity_name}</div>
                        <div class="text-muted-foreground">
                          {shortDate(item.activity_date, 'long')}
                        </div>
                      </TableCell>
                      <TableCell class="text-xs font-medium">
                        {#if waiting}
                          <span class="font-bold text-green-700"
                            >{nameOf(item.substitute_user_id) || '-'}</span
                          >
                        {:else}
                          <span class="text-muted-foreground italic">Menunggu...</span>
                        {/if}
                      </TableCell>
                      <TableCell>
                        <Badge
                          variant="outline"
                          class={waiting
                            ? 'border-blue-200 bg-blue-50 text-blue-700'
                            : 'border-yellow-200 bg-yellow-50 text-yellow-700'}
                        >
                          {waiting ? 'Menunggu ACC Koordinator' : 'Mencari Pengganti'}
                        </Badge>
                      </TableCell>
                    </TableRow>
                  {/each}
                {/if}
              </TableBody>
            </Table>
          </CardContent>
        </Card>
      </div>
    </TabsContent>
  </Tabs>
</div>
