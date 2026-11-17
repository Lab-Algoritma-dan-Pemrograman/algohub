<script lang="ts">
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
  import ArrowRightLeft from '@lucide/svelte/icons/arrow-right-left';
  import CalendarCheck from '@lucide/svelte/icons/calendar-check';
  import CheckCircle2 from '@lucide/svelte/icons/circle-check';
  import Clock from '@lucide/svelte/icons/clock';
  import FileDown from '@lucide/svelte/icons/file-down';
  import Filter from '@lucide/svelte/icons/filter';
  import Info from '@lucide/svelte/icons/info';
  import Lock from '@lucide/svelte/icons/lock';
  import QrCode from '@lucide/svelte/icons/qr-code';
  import Send from '@lucide/svelte/icons/send';
  import Trash2 from '@lucide/svelte/icons/trash-2';
  import Users from '@lucide/svelte/icons/users';
  import XCircle from '@lucide/svelte/icons/circle-x';
  import { createQuery } from '@tanstack/svelte-query';
  import QrScanner from '#lib/components/absensi/QrScanner.svelte';
  import {
    createRpcMutation,
    fetchRpc,
    invalidateRpcs,
    maybeSingle,
    rpcQueryOptions,
  } from '#lib/queries/rpc.ts';
  import { subscribe } from '#lib/realtime/client.ts';
  import { canonRole } from '#lib/roles.ts';
  import { confirm } from '#lib/stores/confirm.svelte.ts';
  import { toast } from '#lib/toast.ts';
  import type { PageProps } from './$types';

  // Samakan dengan meetingOptions di BuatQR dan kolom "Kehadiran (10%)" spreadsheet penilaian.
  const MEETING_OPTIONS = [
    'Pengarahan',
    'Modul 1',
    'Modul 2',
    'Modul 3&4',
    'Modul 5',
    'Modul 6',
    'Ujian Praktik',
  ];
  const DAYS_INDO = ['Minggu', 'Senin', 'Selasa', 'Rabu', 'Kamis', 'Jumat', 'Sabtu'];
  const STATUSES = ['Hadir', 'Izin', 'Sakit', 'Alpha'] as const;

  interface Schedule {
    id: number;
    title?: string;
    type?: string;
    day_of_week?: string;
    start_time?: string;
    end_time?: string;
    major?: string;
    class_code?: string;
  }

  interface LogRow {
    id: number;
    user_id: number;
    status: string;
    notes?: string | null;
    check_in_time: string;
    is_verified?: boolean;
    reschedule_status?: string | null;
    reschedule_schedule_id?: number | null;
    schedule_title?: string;
    schedule_day?: string;
    schedule_time?: string;
    recorded_by_name?: string | null;
    user_full_name?: string;
    user_role?: string;
    user_username?: string;
    user_major?: string;
    user_class_code?: string;
    user_shift?: string;
  }

  interface UserRow {
    id: number;
    username: string;
    full_name: string;
    role: string;
    class_code?: string;
    major?: string;
    shift?: string;
  }

  interface QrSession {
    id?: number;
    title?: string;
  }

  interface GroupRow {
    student_id: number;
    assistant_id: number | null;
    assistant_name?: string | null;
  }

  let { data }: PageProps = $props();
  const user = $derived(data.auth.user);
  const role = $derived(canonRole(user.role));
  // Asisten, koordinator, sekretaris, dan K3 dianggap staff di halaman ini (boleh input manual
  // dan melihat tabel rekap semua praktikan).
  const isStaff = $derived(['asisten', 'koordinator', 'sekretaris', 'k3'].includes(role));
  const division = $derived((user.division || '').toLowerCase());
  const isAslabAdmin = $derived(
    user.role === 'koordinator' || ['sekretaris', 'k3'].includes(division),
  );
  const isAslabK = $derived(
    user.role === 'asisten' && (user.assistant_code || '').toUpperCase().endsWith('K'),
  );
  const hasDeletePower = $derived(isAslabAdmin || isAslabK);
  const checksMenu = $derived(!isAslabAdmin && user.role === 'asisten');

  let filterDate = $state(new Date().toLocaleDateString('en-CA'));

  const absenAccess = createQuery(() =>
    rpcQueryOptions<boolean>(
      'check_menu_access_secure',
      { p_viewer_id: user.id, p_menu_key: '/absensi' },
      { enabled: checksMenu },
    ),
  );
  const validasiAccess = createQuery(() =>
    rpcQueryOptions<boolean>(
      'check_menu_access_secure',
      { p_viewer_id: user.id, p_menu_key: '/validasi-absensi' },
      { enabled: checksMenu },
    ),
  );
  const hasExportAccess = $derived(isAslabAdmin || (checksMenu && !!absenAccess.data));
  const hasFullAccess = $derived(isAslabAdmin || (checksMenu && !!validasiAccess.data));

  createQuery(() =>
    rpcQueryOptions<boolean>(
      'is_pj_absen_today',
      { p_user_id: user.id },
      { enabled: user.role === 'asisten' },
    ),
  );

  const schedulesQuery = createQuery(() =>
    rpcQueryOptions<Schedule[]>('get_schedules_secure', { p_viewer_id: user.id || 0 }),
  );
  const attendanceQuery = createQuery(() =>
    rpcQueryOptions<LogRow[]>('get_attendance_logs_secure', {
      p_viewer_id: user.id,
      p_date_filter: filterDate || null,
    }),
  );
  const myLogsQuery = createQuery(() =>
    rpcQueryOptions<LogRow[]>(
      'get_attendance_logs_secure',
      { p_viewer_id: user.id, p_date_filter: null },
      { enabled: !isStaff },
    ),
  );
  const deletionQuery = createQuery(() =>
    rpcQueryOptions<
      {
        id: number;
        deleted_at: string;
        target_user_id: string;
        deleted_by_name?: string;
        reason?: string;
      }[]
    >('get_deletion_history_secure', { p_viewer_id: user.id }, { enabled: hasFullAccess }),
  );
  const myOwnSchedulesQuery = createQuery(() =>
    rpcQueryOptions<
      {
        schedule_id: number;
        schedule_title?: string;
        schedule_day?: string;
        schedule_start?: string;
        schedule_end?: string;
        schedule_major?: string;
        schedule_class?: string;
      }[]
    >('get_personal_schedules_secure', { p_viewer_id: user.id }, { enabled: !isStaff }),
  );
  const contactQuery = createQuery(() =>
    rpcQueryOptions<{ phone_number?: string; assistant_code?: string }[]>(
      'get_smart_validation_contact_secure',
      { p_student_id: user.id },
      { enabled: !isStaff },
    ),
  );
  // RPC pengaturan lengkap hanya untuk staff; praktikan mendapat 400 "Akses ditolak.".
  const settingsQuery = createQuery(() =>
    rpcQueryOptions<{ wa_templates?: Record<string, string>; active_shift?: string }[]>(
      'get_system_settings_full_secure',
      { p_viewer_id: user.id },
      { enabled: isStaff },
    ),
  );

  const upsertLog = createRpcMutation('upsert_attendance_log_secure');
  const requestReschedule = createRpcMutation('request_reschedule_secure');
  const deleteLogMutation = createRpcMutation('delete_attendance_log_secure', [
    'get_attendance_logs_secure',
    'get_deletion_history_secure',
  ]);

  const toLog = (log: LogRow) => ({
    ...log,
    users: {
      full_name: log.user_full_name,
      role: log.user_role,
      username: log.user_username,
      major: log.user_major,
      class_code: log.user_class_code,
      shift: log.user_shift,
    },
    schedules: log.reschedule_schedule_id
      ? {
          id: log.reschedule_schedule_id,
          title: log.schedule_title,
          day_of_week: log.schedule_day,
          start_time: log.schedule_time,
        }
      : null,
  });

  const availableSchedules = $derived(
    (schedulesQuery.data || []).filter((s) => s.type === 'praktikum'),
  );
  const allAttendanceData = $derived((attendanceQuery.data || []).map(toLog));
  const myLogs = $derived((myLogsQuery.data || []).map(toLog));
  const deletionHistory = $derived(deletionQuery.data || []);
  const myOwnSchedules = $derived(
    (myOwnSchedulesQuery.data || []).map((s) => ({
      id: s.schedule_id,
      title: s.schedule_title,
      day_of_week: s.schedule_day,
      start_time: s.schedule_start,
      end_time: s.schedule_end,
      major: s.schedule_major,
      class_code: s.schedule_class,
    })),
  );
  const contact = $derived(
    contactQuery.data && contactQuery.data.length > 0 ? contactQuery.data[0] : null,
  );
  const adminPhone = $derived(contact?.phone_number ?? null);
  const adminCode = $derived(contact?.assistant_code ?? null);
  const settings = $derived(
    settingsQuery.data && settingsQuery.data.length > 0 ? settingsQuery.data[0] : null,
  );
  const waTemplates = $derived(settings?.wa_templates ?? null);

  let loading = $state(false);
  let activeTab = $state('scan');
  let scanLocked = $state(false);
  let cameraError = $state('');
  let targetNim = $state('');
  let targetStatus = $state('Hadir');
  let targetNote = $state('');
  let selectedClassSchedule = $state('');
  let draftAttendance = $state<Record<number, string>>({});
  let savingBatch = $state(false);
  let selectedShiftFilter = $state('all');
  let selectedMeeting = $state('');
  let izinReason = $state('');
  let izinType = $state('Izin');
  let izinDate = $state('');
  let selectedMySchedule = $state('');
  let selectedSchedule = $state('');
  let deletionOpen = $state(false);
  let rescheduleFor = $state<number | null>(null);

  let shiftFromSettings = false;
  $effect(() => {
    const sysShift = settings?.active_shift;
    if (shiftFromSettings || !settings) return;
    shiftFromSettings = true;
    if (sysShift === '1' || sysShift === '2') selectedShiftFilter = sysShift;
  });

  // Staff mendengar semua perubahan; praktikan hanya baris miliknya sendiri.
  $effect(() => {
    const filter = isStaff ? undefined : `custom_user_id=eq.${user.id}`;
    return subscribe('attendance_live_updates', { table: 'attendance_logs', filter }, () => {
      void invalidateRpcs(['get_attendance_logs_secure']);
    });
  });

  const selectedDayName = $derived(
    filterDate
      ? DAYS_INDO[new Date(filterDate + 'T00:00:00').getDay()]
      : DAYS_INDO[new Date().getDay()],
  );
  const dayClasses = $derived(
    availableSchedules.filter(
      (s) => s.day_of_week?.toString().trim().toLowerCase() === selectedDayName.toLowerCase(),
    ),
  );

  $effect(() => {
    if (dayClasses.length > 0) {
      if (!dayClasses.some((s) => s.id.toString() === selectedClassSchedule)) {
        selectedClassSchedule = dayClasses[0].id.toString();
      }
    } else {
      selectedClassSchedule = '';
    }
  });

  $effect(() => {
    void [selectedClassSchedule, filterDate, selectedShiftFilter];
    draftAttendance = {};
  });

  const schedId = $derived(parseInt(selectedClassSchedule));
  const hasClass = $derived(!!selectedClassSchedule && !isNaN(schedId));
  const groupQuery = createQuery(() =>
    rpcQueryOptions<GroupRow[]>(
      'get_group_members_secure',
      { p_viewer_id: user.id, p_schedule_id: schedId },
      { enabled: hasClass },
    ),
  );
  const usersQuery = createQuery(() =>
    rpcQueryOptions<UserRow[]>('get_users_secure', { p_viewer_id: user.id }, { enabled: hasClass }),
  );
  const fetchingMembers = $derived(hasClass && (groupQuery.isPending || usersQuery.isPending));

  const extractClassLetterOrCode = (val: string | undefined | null) => {
    if (!val) return '';
    const str = val.trim();
    const classMatch = str.match(/kelas\s+([a-z0-9-]+)/i);
    return classMatch ? classMatch[1].toUpperCase() : str.toUpperCase();
  };

  const isMatchingClassCode = (
    studentClassCode: string | undefined,
    scheduleClassCode: string | undefined,
    scheduleTitle: string | undefined,
  ) => {
    if (!studentClassCode) return false;
    const stdClass = extractClassLetterOrCode(studentClassCode);
    const schedClass =
      extractClassLetterOrCode(scheduleClassCode) || extractClassLetterOrCode(scheduleTitle);
    if (!stdClass || !schedClass) return false;
    if (stdClass === schedClass) return true;
    if (stdClass.endsWith(`-${schedClass}`) || stdClass.endsWith(` ${schedClass}`)) return true;
    return schedClass.endsWith(`-${stdClass}`) || schedClass.endsWith(` ${stdClass}`);
  };

  // Sumber kebenaran: users dengan class_code DAN major yang cocok. Format class_code users
  // ("TE B") berbeda dengan schedules ("B"), jadi dibandingkan lewat isMatchingClassCode.
  const classMembers = $derived.by(() => {
    if (!hasClass || !groupQuery.data || !usersQuery.data) return [];
    const targetSched = availableSchedules.find((s) => s.id.toString() === selectedClassSchedule);
    const groupData = groupQuery.data || [];
    const groupById = new Map(groupData.map((m) => [String(m.student_id), m]));
    return usersQuery.data
      .filter((u) => {
        if (canonRole(u.role) !== 'praktikan') return false;
        if (
          targetSched?.class_code &&
          !isMatchingClassCode(u.class_code, targetSched.class_code, targetSched.title)
        )
          return false;
        return !(targetSched?.major && u.major !== targetSched.major);
      })
      .map((u) => {
        const gm = groupById.get(String(u.id));
        return {
          id: u.id,
          schedule_id: schedId,
          student_id: u.id,
          assistant_id: gm?.assistant_id ?? null,
          student_name: u.full_name,
          student_nim: u.username,
          assistant_name: gm?.assistant_name ?? null,
          student_shift: u.shift || '1',
          student_class_code: u.class_code,
          student_major: u.major || targetSched?.major || '',
          _in_group: !!gm,
        };
      });
  });

  const getNormalizedShift = (shiftVal: unknown) => {
    if (!shiftVal) return '1';
    const str = String(shiftVal).toLowerCase().trim();
    if (str.includes('2')) return '2';
    if (str.includes('1')) return '1';
    return str;
  };

  const filteredClassMembers = $derived(
    selectedShiftFilter === 'all'
      ? classMembers
      : classMembers.filter((m) => getNormalizedShift(m.student_shift) === selectedShiftFilter),
  );

  const logFor = (student: { student_nim: string; student_id: number }) =>
    allAttendanceData.find(
      (l) => l.users.username === student.student_nim || l.user_id === student.student_id,
    );

  const classStats = $derived.by(() => {
    const stats = { hadir: 0, izin: 0, sakit: 0, alpha: 0, total: filteredClassMembers.length };
    for (const student of filteredClassMembers) {
      const draftStatus = draftAttendance[student.student_id];
      const log = logFor(student);
      const status = draftStatus !== undefined ? draftStatus : log ? log.status : null;
      if (!status || status === 'Alpha') stats.alpha++;
      else if (status === 'Hadir') stats.hadir++;
      else if (status === 'Izin') stats.izin++;
      else if (status === 'Sakit') stats.sakit++;
    }
    return stats;
  });

  function handleSelectDraftStatus(studentUserId: number, status: string) {
    if (draftAttendance[studentUserId] === status) {
      const next = { ...draftAttendance };
      delete next[studentUserId];
      draftAttendance = next;
    } else {
      draftAttendance = { ...draftAttendance, [studentUserId]: status };
    }
  }

  async function handleSaveBatchAttendance() {
    const entries = Object.entries(draftAttendance);
    if (entries.length === 0) {
      toast.info('Belum ada perubahan status absensi yang dipilih.');
      return;
    }
    if (!selectedMeeting) {
      toast.error('Pilih Pertemuan dulu — absen tanpa pertemuan tidak masuk rekap spreadsheet.');
      return;
    }
    savingBatch = true;
    let successCount = 0;
    let errorCount = 0;
    for (const [studentIdStr, status] of entries) {
      const studentUserId = Number(studentIdStr);
      try {
        const { error } = await upsertLog.mutateAsync({
          p_caller_id: user.id,
          p_target_user_id: studentUserId,
          p_status: status,
          p_notes: `Quick input (${status}) — ${selectedMeeting} — oleh ${user.full_name}`,
          p_check_in: new Date(`${filterDate}T08:00:00`).toISOString(),
          p_is_verified: true,
          p_type: 'staff_manual',
          p_meeting: selectedMeeting,
        });
        if (error) {
          console.error('Gagal menyimpan absensi user:', studentUserId, error);
          errorCount++;
        } else {
          successCount++;
        }
      } catch {
        errorCount++;
      }
    }
    savingBatch = false;
    if (successCount > 0) {
      toast.success(`Berhasil menyimpan ${successCount} data absensi!`);
      draftAttendance = {};
      void invalidateRpcs(['get_attendance_logs_secure']);
    }
    if (errorCount > 0) toast.error(`Gagal menyimpan ${errorCount} data absensi.`);
  }

  async function handleExportExcel() {
    if (allAttendanceData.length === 0) {
      return toast.warning('Tidak ada data untuk diekspor pada tanggal ini.');
    }
    const XLSX = await import('xlsx');
    const dataToExport = allAttendanceData.map((log, index) => ({
      No: index + 1,
      'Nama Mahasiswa': log.users.full_name || 'Unknown',
      NIM: log.users.username || '-',
      Jurusan: log.users.major || '-',
      Kelas: log.users.class_code || '-',
      Shift: log.users.shift || '-',
      'Waktu Absen': new Date(log.check_in_time).toLocaleTimeString('id-ID', {
        hour: '2-digit',
        minute: '2-digit',
      }),
      Status: log.status,
      Keterangan: log.notes || '-',
      'Petugas/Log': log.recorded_by_name || 'Mandiri',
    }));
    const ws = XLSX.utils.json_to_sheet(dataToExport);
    const wb = XLSX.utils.book_new();
    XLSX.utils.book_append_sheet(wb, ws, 'Daftar Hadir');
    XLSX.writeFile(wb, `Rekap_Absensi_Praktikan_${filterDate}.xlsx`);
    toast.success('Data berhasil diekspor ke Excel!');
  }

  const getGreeting = () => {
    const hour = new Date().getHours();
    if (hour < 11) return 'Pagi';
    if (hour < 15) return 'Siang';
    if (hour < 19) return 'Sore';
    return 'Malam';
  };

  const getHonorific = (code: string | null) => {
    if (!code) return 'Kak';
    return code.startsWith('P') ? 'Kak' : 'Bang';
  };

  function getWaProofLink(targetPhone: string, type: string, notes: string) {
    let cleanPhone = targetPhone.replace(/\D/g, '');
    if (cleanPhone.startsWith('0')) cleanPhone = '62' + cleanPhone.slice(1);
    let text = `Selamat ${getGreeting()} ${getHonorific(adminCode)}, Mohon maaf mengganggu waktunya. Saya ${user.full_name} (${user.username}) ingin mengirimkan bukti izin: ${type} - ${notes}`;
    if (waTemplates?.absen_izin) {
      text = waTemplates.absen_izin
        .replace(/{{waktu}}/g, getGreeting())
        .replace(/{{panggilan}}/g, getHonorific(adminCode))
        .replace(/{{nama}}/g, user.full_name || '')
        .replace(/{{nim}}/g, user.username || '')
        .replace(/{{tipe}}/g, type)
        .replace(/{{alasan}}/g, notes)
        .replace(/{{kelas}}/g, (user.class_code as string) || '')
        .replace(/{{jurusan}}/g, user.division || '');
    }
    return `https://wa.me/${cleanPhone}?text=${encodeURIComponent(text)}`;
  }

  async function handleReschedule(logId: number) {
    if (!selectedSchedule) return toast.error('Pilih jadwal pengganti!');
    loading = true;
    try {
      const { error } = await requestReschedule.mutateAsync({
        p_caller_id: user.id,
        p_log_id: logId,
        p_new_schedule_id: parseInt(selectedSchedule),
      });
      if (error) throw error;
      toast.success('Jadwal diajukan. Tunggu persetujuan Asisten.');
      selectedSchedule = '';
    } catch (err) {
      toast.error((err as Error).message);
    } finally {
      loading = false;
    }
  }

  // Kunci dilepas otomatis supaya user tidak perlu menekan "Scan Ulang".
  const autoResetScanner = () => setTimeout(() => (scanLocked = false), 3000);

  async function handleScanProcess(decodedText: string) {
    scanLocked = true;
    toast.info('Memproses Kehadiran...');
    loading = true;
    try {
      let token = decodedText;
      if (decodedText.includes('token=')) token = decodedText.split('token=')[1].split('&')[0];
      let session: QrSession | null = null;
      try {
        session = maybeSingle(
          await fetchRpc<{ id: number; title: string }[]>('get_qr_session_secure', {
            p_token: token,
          }),
        ) as QrSession | null;
      } catch {
        session = null;
      }
      if (!session) {
        toast.error('QR Code Salah atau Sesi Telah Berakhir!');
        autoResetScanner();
        return;
      }
      // Validasi jadwal dan duplikat dilakukan server di upsert_attendance_log_secure.
      const { error } = await upsertLog.mutateAsync({
        p_caller_id: user.id,
        p_target_user_id: user.id,
        p_status: 'Hadir',
        p_notes: `Scan QR: ${session.title}`,
        p_check_in: new Date().toISOString(),
        p_is_verified: true,
        p_type: 'scan',
        p_session_id: session.id,
      });
      if (error) throw error;
      toast.success('Berhasil Absen!');
    } catch (err) {
      toast.error((err as Error).message);
      autoResetScanner();
    } finally {
      loading = false;
    }
  }

  async function handleSubmitIzin() {
    if (!selectedMySchedule) return toast.error('Pilih jadwal kelas yang ditinggalkan!');
    if (!izinDate) return toast.error('Pilih tanggal absen (izin/sakit)!');
    if (!izinReason) return toast.error('Isi alasan izin secara detail!');
    loading = true;
    try {
      const schedInfo = myOwnSchedules.find((s) => s.id.toString() === selectedMySchedule);
      const dateDayName = DAYS_INDO[new Date(izinDate + 'T00:00:00').getDay()];
      if (schedInfo && schedInfo.day_of_week !== dateDayName) {
        toast.error(
          `Tanggal yang Anda pilih (${izinDate} jatuh pada hari ${dateDayName}) tidak cocok dengan hari jadwal praktikum Anda (${schedInfo.day_of_week})!`,
        );
        return;
      }
      const prefixInfo = schedInfo
        ? `[Kelas ${schedInfo.class_code} - ${schedInfo.day_of_week}] `
        : '';
      const { error } = await upsertLog.mutateAsync({
        p_caller_id: user.id,
        p_target_user_id: user.id,
        p_status: izinType,
        p_notes: prefixInfo + izinReason,
        p_check_in: new Date(`${izinDate}T08:00:00`).toISOString(),
        p_is_verified: false,
        p_type: 'izin',
      });
      if (error) throw error;
      toast.success('Terkirim! Silakan kirim bukti ke Asisten.');
      izinReason = '';
      izinDate = '';
      selectedMySchedule = '';
    } catch (err) {
      toast.error((err as Error).message);
    } finally {
      loading = false;
    }
  }

  async function handleStaffHelp() {
    loading = true;
    try {
      const allUsers = await fetchRpc<UserRow[]>('get_users_secure', {
        p_viewer_id: user.id,
      }).catch(() => [] as UserRow[]);
      const usr = (allUsers || []).find((u) => u.username === targetNim);
      if (!usr) {
        toast.error('NIM tidak ditemukan. Pastikan User sudah terdaftar.');
        return;
      }
      const { error } = await upsertLog.mutateAsync({
        p_caller_id: user.id,
        p_target_user_id: usr.id,
        p_status: targetStatus,
        p_notes:
          targetNote || `Input manual oleh Aslab${selectedMeeting ? ` — ${selectedMeeting}` : ''}`,
        p_check_in: new Date(`${filterDate}T08:00:00`).toISOString(),
        p_is_verified: true,
        p_type: 'staff_manual',
        p_meeting: selectedMeeting || null,
      });
      if (error) throw error;
      toast.success('Absensi Manual Berhasil!');
      targetNim = '';
      targetNote = '';
    } catch (err) {
      toast.error((err as Error).message);
    } finally {
      loading = false;
    }
  }

  async function deleteLog(log: (typeof allAttendanceData)[number]) {
    if (!(await confirm('Hapus data absen ini? Pastikan Anda tidak salah hapus!'))) return;
    loading = true;
    try {
      const { error } = await deleteLogMutation.mutateAsync({
        p_caller_id: user.id,
        p_log_id: log.id,
        p_reason: 'Dihapus manual oleh Asisten',
        p_target_nim: log.users.username || '-',
        p_snapshot_data: log,
      });
      if (error) throw error;
      toast.success('Data berhasil dihapus dan riwayat tersimpan');
    } catch (err) {
      toast.error(
        'Gagal menghapus: ' + ((err as Error).message || 'Akses ditolak atau terjadi kesalahan'),
      );
    } finally {
      loading = false;
    }
  }

  const displayAttendance = $derived(
    isStaff
      ? allAttendanceData
      : allAttendanceData.filter((log) => log.users.username === user.username),
  );
  const totalHadir = $derived(allAttendanceData.filter((log) => log.status === 'Hadir').length);
  const draftCount = $derived(Object.keys(draftAttendance).length);

  const statusButtonClass: Record<string, [string, string]> = {
    Hadir: [
      'bg-emerald-600 text-white border-emerald-600 shadow-xs',
      'bg-emerald-50 text-emerald-700 border-emerald-300 hover:bg-emerald-100',
    ],
    Izin: [
      'bg-amber-600 text-white border-amber-600 shadow-xs',
      'bg-amber-50 text-amber-700 border-amber-300 hover:bg-amber-100',
    ],
    Sakit: [
      'bg-blue-600 text-white border-blue-600 shadow-xs',
      'bg-blue-50 text-blue-700 border-blue-300 hover:bg-blue-100',
    ],
    Alpha: [
      'bg-rose-600 text-white border-rose-600 shadow-xs',
      'bg-rose-50 text-rose-700 border-rose-300 hover:bg-rose-100',
    ],
  };
  const shiftLabels: Record<string, string> = {
    all: 'Semua Shift',
    '1': 'Shift 1',
    '2': 'Shift 2',
  };
  const scheduleLabel = (s: Schedule) =>
    `Kelas ${s.class_code || s.title} (${s.start_time?.slice(0, 5)} - ${s.end_time?.slice(0, 5)}) - ${s.major || 'Praktikum'}`;
</script>

<div class="mx-auto max-w-6xl space-y-6">
  <h1 class="flex items-center gap-2 text-2xl font-bold">
    <CalendarCheck class="text-primary" /> Manajemen Absensi Praktikan
  </h1>

  <div class="grid grid-cols-1 gap-6 lg:grid-cols-12">
    <div class="space-y-6 lg:col-span-4">
      {#if !isStaff}
        <div class="w-full space-y-5">
          <Tabs bind:value={activeTab} class="w-full">
            <TabsList class="grid w-full grid-cols-2"
              ><TabsTrigger value="scan">Scan QR</TabsTrigger><TabsTrigger value="izin"
                >Izin</TabsTrigger
              ></TabsList
            >
            <TabsContent value="scan">
              <Card class="border-t-primary border-t-4 shadow-lg">
                <CardHeader class="pb-2 text-center"
                  ><CardTitle class="text-lg">Arahkan ke Layar</CardTitle></CardHeader
                >
                <CardContent>
                  <div
                    class="relative flex min-h-[300px] flex-col items-center justify-center overflow-hidden rounded-xl border-2 border-dashed border-gray-300 bg-slate-50"
                  >
                    {#if cameraError}
                      <div class="space-y-4 p-6 text-center text-red-500">
                        <XCircle class="mx-auto h-12 w-12" />
                        <p class="text-sm font-medium">{cameraError}</p>
                        <Button variant="outline" onclick={() => (cameraError = '')}
                          >Coba Lagi</Button
                        >
                      </div>
                    {:else if scanLocked}
                      <div class="space-y-4 p-6 text-center">
                        <CheckCircle2
                          class="animate-in zoom-in mx-auto h-16 w-16 text-green-500 duration-300"
                        />
                        <div>
                          <p class="text-lg font-bold text-slate-800">Scan Terkunci</p>
                          <p class="text-muted-foreground mt-1 text-xs">
                            Sistem sedang memproses atau Anda sudah berhasil absen.
                          </p>
                        </div>
                        <Button variant="outline" class="mt-2" onclick={() => (scanLocked = false)}
                          >Scan Ulang</Button
                        >
                      </div>
                    {:else}
                      <div class="relative h-full w-full">
                        {#if activeTab === 'scan'}
                          <QrScanner
                            onScan={handleScanProcess}
                            onCameraError={(m) => (cameraError = m)}
                          />
                        {/if}
                        <div class="absolute right-0 bottom-4 left-0 text-center">
                          <Badge class="bg-black/50 px-3 py-1 text-white backdrop-blur-sm"
                            ><QrCode class="mr-2 inline h-3 w-3" /> Mencari QR Code...</Badge
                          >
                        </div>
                      </div>
                    {/if}
                  </div>
                </CardContent>
              </Card>
            </TabsContent>
            <TabsContent value="izin">
              <Card class="shadow-sm">
                <CardHeader class="pb-2">
                  <CardTitle class="text-lg">Form Izin / Sakit</CardTitle>
                  <CardDescription class="text-xs"
                    >Ajukan izin pada jadwal praktikum Anda.</CardDescription
                  >
                </CardHeader>
                <CardContent class="space-y-4">
                  <div class="grid grid-cols-2 gap-4">
                    <div class="space-y-1.5">
                      <Label class="text-muted-foreground text-xs">Jenis Izin</Label>
                      <Select type="single" bind:value={izinType}>
                        <SelectTrigger class="w-full">{izinType}</SelectTrigger>
                        <SelectContent
                          ><SelectItem value="Izin" label="Izin" /><SelectItem
                            value="Sakit"
                            label="Sakit"
                          /></SelectContent
                        >
                      </Select>
                    </div>
                    <div class="space-y-1.5">
                      <Label class="text-muted-foreground text-xs">Pilih Jadwal Anda</Label>
                      <Select type="single" bind:value={selectedMySchedule}>
                        <SelectTrigger class="w-full">
                          {#if selectedMySchedule}
                            {@const s = myOwnSchedules.find(
                              (x) => x.id.toString() === selectedMySchedule,
                            )}
                            Kelas {s?.class_code} ({s?.day_of_week})
                          {:else}
                            <span class="text-muted-foreground"
                              >{myOwnSchedules.length === 0
                                ? 'Tidak ada jadwal'
                                : 'Pilih jadwal...'}</span
                            >
                          {/if}
                        </SelectTrigger>
                        <SelectContent>
                          {#each myOwnSchedules as s (s.id)}
                            <SelectItem
                              value={s.id.toString()}
                              label={`Kelas ${s.class_code} (${s.day_of_week})`}
                            />
                          {/each}
                        </SelectContent>
                      </Select>
                    </div>
                  </div>
                  <div class="space-y-1.5">
                    <Label for="izin-date" class="text-muted-foreground text-xs"
                      >Tanggal Absen</Label
                    >
                    <Input id="izin-date" type="date" bind:value={izinDate} />
                  </div>
                  <div class="space-y-1.5">
                    <Label for="izin-reason" class="text-muted-foreground text-xs"
                      >Alasan Detail</Label
                    >
                    <Input
                      id="izin-reason"
                      placeholder="Contoh: Ada acara keluarga / Dirawat di RS..."
                      bind:value={izinReason}
                    />
                  </div>
                  <Button
                    onclick={handleSubmitIzin}
                    disabled={loading || myOwnSchedules.length === 0}
                    class="mt-2 w-full"
                  >
                    Kirim Pengajuan
                  </Button>
                </CardContent>
              </Card>
            </TabsContent>
          </Tabs>

          <Card
            class="shadow-bubbly-yellow space-y-3 rounded-2xl border-2 border-amber-100 bg-amber-50/20 p-5"
          >
            <h4
              class="flex items-center gap-1.5 text-sm font-black tracking-wider text-amber-800 uppercase"
            >
              <Info class="h-4.5 w-4.5 shrink-0 text-amber-600" /> Alur Ganti Jadwal (Reschedule)
            </h4>
            <ol
              class="list-inside list-decimal space-y-2 pl-1 text-xs leading-relaxed font-semibold text-amber-900/80"
            >
              <li>
                <span class="font-bold text-amber-950">Ajukan Izin:</span> Isi form Izin/Sakit di tab
                sebelah untuk kelas yang ditinggalkan.
              </li>
              <li>
                <span class="font-bold text-amber-950">Kirim Bukti WA:</span> Klik tombol
                <span class="underline">Kirim Bukti WA</span> di riwayat bawah untuk mengirim bukti ke
                Asisten.
              </li>
              <li>
                <span class="font-bold text-amber-950">Tunggu Verifikasi:</span> Asisten akan memverifikasi
                izin Anda.
              </li>
              <li>
                <span class="font-bold text-amber-950">Pilih Jadwal Baru:</span> Klik tombol
                <span class="underline">Pilih Jadwal</span> pada riwayat izin Anda untuk memilih slot
                pengganti.
              </li>
              <li>
                <span class="font-bold text-amber-950">Tunggu ACC Jadwal:</span> Tunggu persetujuan jadwal
                pengganti dari Asisten.
              </li>
            </ol>
          </Card>
        </div>
      {:else}
        <div class="space-y-6">
          <Card class="border-l-primary border-l-4 shadow-md">
            <CardHeader class="border-b bg-slate-50/50 pb-3">
              <div class="flex items-center justify-between">
                <div>
                  <CardTitle class="flex items-center gap-2 text-lg font-bold">
                    <Users class="text-primary h-4 w-4" /> Quick Input & Rekap Kelas
                  </CardTitle>
                  <CardDescription class="text-xs">
                    Pilih kelas pada hari <span class="font-semibold text-slate-700"
                      >{selectedDayName}</span
                    >
                    ({filterDate}) untuk rekap & input cepat.
                  </CardDescription>
                </div>
                <Badge
                  variant="secondary"
                  class="bg-primary/10 text-primary border-primary/20 text-[10px] font-bold"
                >
                  {dayClasses.length} Kelas Tersedia
                </Badge>
              </div>
            </CardHeader>
            <CardContent class="space-y-4 pt-4">
              <div class="space-y-3">
                <div class="space-y-1.5">
                  <Label class="text-xs font-semibold text-slate-700">Pilih Jadwal Kelas</Label>
                  <Select type="single" bind:value={selectedClassSchedule}>
                    <SelectTrigger class="w-full bg-white" aria-label="Pilih Jadwal Kelas">
                      {#if selectedClassSchedule}
                        {@const s = dayClasses.find(
                          (x) => x.id.toString() === selectedClassSchedule,
                        )}
                        {s ? scheduleLabel(s) : ''}
                      {:else}
                        <span class="text-muted-foreground"
                          >{dayClasses.length === 0
                            ? `Tidak ada kelas hari ${selectedDayName}`
                            : 'Pilih kelas...'}</span
                        >
                      {/if}
                    </SelectTrigger>
                    <SelectContent>
                      {#each dayClasses as s (s.id)}
                        <SelectItem value={s.id.toString()} label={scheduleLabel(s)} />
                      {/each}
                    </SelectContent>
                  </Select>
                </div>
                <div class="space-y-1.5">
                  <Label class="text-xs font-semibold text-slate-700">Pertemuan</Label>
                  <Select type="single" bind:value={selectedMeeting}>
                    <SelectTrigger class="w-full bg-white" aria-label="Pertemuan">
                      {#if selectedMeeting}{selectedMeeting}{:else}<span
                          class="text-muted-foreground">Pilih pertemuan...</span
                        >{/if}
                    </SelectTrigger>
                    <SelectContent>
                      {#each MEETING_OPTIONS as m (m)}
                        <SelectItem value={m} label={m} />
                      {/each}
                    </SelectContent>
                  </Select>
                  <p class="text-[10px] leading-snug text-slate-400">
                    Menentukan kolom checkbox <span class="font-semibold">Kehadiran (10%)</span> yang
                    terisi di spreadsheet rekap.
                  </p>
                </div>
                <div class="flex items-center justify-between gap-2 border-t border-slate-100 pt-1">
                  <Label class="text-xs font-semibold text-slate-600">Filter Shift:</Label>
                  <div class="w-[140px]">
                    <Select type="single" bind:value={selectedShiftFilter}>
                      <SelectTrigger class="h-8 w-full border-slate-200 bg-white text-xs">
                        <Filter class="text-muted-foreground mr-1 h-3.5 w-3.5" />
                        {shiftLabels[selectedShiftFilter] ?? selectedShiftFilter}
                      </SelectTrigger>
                      <SelectContent>
                        {#each Object.entries(shiftLabels) as [value, label] (value)}
                          <SelectItem {value} {label} />
                        {/each}
                      </SelectContent>
                    </Select>
                  </div>
                </div>
              </div>

              {#if selectedClassSchedule}
                <div class="grid grid-cols-4 gap-2 text-center">
                  {#each [{ label: 'Hadir', value: classStats.hadir, c: 'emerald' }, { label: 'Izin', value: classStats.izin, c: 'amber' }, { label: 'Sakit', value: classStats.sakit, c: 'blue' }, { label: 'Alpha', value: classStats.alpha, c: 'rose' }] as stat (stat.label)}
                    <div
                      class="rounded-lg border p-2 shadow-xs {stat.c === 'emerald'
                        ? 'border-emerald-200/60 bg-emerald-50'
                        : stat.c === 'amber'
                          ? 'border-amber-200/60 bg-amber-50'
                          : stat.c === 'blue'
                            ? 'border-blue-200/60 bg-blue-50'
                            : 'border-rose-200/60 bg-rose-50'}"
                    >
                      <div
                        class="text-[10px] font-extrabold tracking-wider uppercase {stat.c ===
                        'emerald'
                          ? 'text-emerald-700'
                          : stat.c === 'amber'
                            ? 'text-amber-700'
                            : stat.c === 'blue'
                              ? 'text-blue-700'
                              : 'text-rose-700'}"
                      >
                        {stat.label}
                      </div>
                      <div
                        class="text-xl font-black {stat.c === 'emerald'
                          ? 'text-emerald-800'
                          : stat.c === 'amber'
                            ? 'text-amber-800'
                            : stat.c === 'blue'
                              ? 'text-blue-800'
                              : 'text-rose-800'}"
                      >
                        {stat.value}
                      </div>
                    </div>
                  {/each}
                </div>

                <div class="space-y-3">
                  <div
                    class="flex items-center justify-between px-1 text-[11px] font-bold text-slate-500 uppercase"
                  >
                    <span>Daftar Praktikan ({filteredClassMembers.length})</span>
                    <span class="text-[10px] font-normal text-slate-400"
                      >Pilih status lalu simpan</span
                    >
                  </div>
                  <div class="max-h-[300px] space-y-1.5 overflow-y-auto pr-1">
                    {#if fetchingMembers}
                      <p class="py-6 text-center text-xs text-slate-400 italic">
                        Memuat daftar mahasiswa...
                      </p>
                    {:else if filteredClassMembers.length === 0}
                      <div
                        class="rounded-lg border border-dashed bg-slate-50 px-4 py-6 text-center text-slate-400"
                      >
                        <p class="text-xs font-medium">
                          Belum ada praktikan terdaftar di shift/kelas ini.
                        </p>
                        <p class="mt-1 text-[10px] text-slate-400">
                          Gunakan filter Shift di atas atau sync dari Manajemen Kelas.
                        </p>
                      </div>
                    {:else}
                      {#each filteredClassMembers as student (student.id)}
                        {@const log = logFor(student)}
                        {@const draftStatus = draftAttendance[student.student_id]}
                        {@const effectiveStatus =
                          draftStatus !== undefined ? draftStatus : log ? log.status : null}
                        <div
                          class="flex items-center justify-between gap-2 rounded-lg border border-slate-100 bg-white p-2.5 text-xs shadow-2xs transition-all hover:bg-slate-50/80"
                          data-testid="class-member"
                        >
                          <div class="min-w-0 flex-1">
                            <div class="flex items-center gap-1.5">
                              <p class="truncate font-bold text-slate-800">
                                {student.student_name}
                              </p>
                              {#if draftStatus}
                                <span
                                  class="rounded border border-indigo-100 bg-indigo-50 px-1 text-[9px] font-semibold text-indigo-600"
                                  >Draft</span
                                >
                              {/if}
                            </div>
                            <div class="flex items-center gap-1.5 text-[10px]">
                              <span class="font-mono text-slate-400">{student.student_nim}</span>
                              <Badge
                                variant="outline"
                                class="h-4 border-slate-200 bg-slate-50 px-1 py-0 text-[9px] font-medium text-slate-600"
                              >
                                Shift {getNormalizedShift(
                                  student.student_shift || log?.users.shift,
                                )}
                              </Badge>
                              {#if student.student_class_code || log?.users.class_code}
                                <Badge
                                  variant="outline"
                                  class="h-4 border-blue-200 bg-blue-50 px-1 py-0 text-[9px] font-medium text-blue-700"
                                >
                                  {student.student_class_code || log?.users.class_code}
                                </Badge>
                              {/if}
                            </div>
                          </div>
                          <div class="flex shrink-0 items-center gap-1">
                            {#each STATUSES as status (status)}
                              <Button
                                size="sm"
                                variant="outline"
                                class="h-7 w-7 p-0 text-[11px] font-black transition-all active:scale-95 {statusButtonClass[
                                  status
                                ][effectiveStatus === status ? 0 : 1]}"
                                onclick={() => handleSelectDraftStatus(student.student_id, status)}
                                disabled={savingBatch}
                                title={`Pilih ${status} (${status[0]})`}
                              >
                                {status[0]}
                              </Button>
                            {/each}
                          </div>
                        </div>
                      {/each}
                    {/if}
                  </div>
                  <Button
                    onclick={handleSaveBatchAttendance}
                    disabled={savingBatch || draftCount === 0}
                    class="w-full font-bold shadow-sm transition-all"
                  >
                    {savingBatch
                      ? 'Menyimpan...'
                      : draftCount > 0
                        ? `Simpan Absensi (${draftCount} Perubahan)`
                        : 'Simpan Absensi Kelas'}
                  </Button>
                </div>
              {/if}
            </CardContent>
          </Card>

          <Card class="border bg-slate-50/40 shadow-xs">
            <CardHeader class="pb-2">
              <CardTitle class="text-sm font-semibold text-slate-700"
                >Input Manual per NIM (Fallback)</CardTitle
              >
            </CardHeader>
            <CardContent class="space-y-3">
              <p class="text-[10px] text-slate-500">
                Pertemuan: <span class="font-semibold text-slate-700"
                  >{selectedMeeting || 'belum dipilih'}</span
                >
                {!selectedMeeting
                  ? ' (pilih di panel kiri kalau ingin masuk rekap spreadsheet)'
                  : ''}
              </p>
              <div class="space-y-1.5">
                <Label for="target-nim" class="text-xs">NIM Praktikan</Label>
                <Input
                  id="target-nim"
                  placeholder="Masukkan NIM..."
                  bind:value={targetNim}
                  class="h-8 bg-white text-xs"
                />
              </div>
              <div class="grid grid-cols-2 gap-2">
                <div class="space-y-1.5">
                  <Label class="text-xs">Status</Label>
                  <Select type="single" bind:value={targetStatus}>
                    <SelectTrigger class="h-8 w-full bg-white text-xs">{targetStatus}</SelectTrigger
                    >
                    <SelectContent>
                      {#each STATUSES as status (status)}
                        <SelectItem value={status} label={status} />
                      {/each}
                    </SelectContent>
                  </Select>
                </div>
                <div class="space-y-1.5">
                  <Label for="target-note" class="text-xs">Ket (Opsional)</Label>
                  <Input
                    id="target-note"
                    placeholder="..."
                    bind:value={targetNote}
                    class="h-8 bg-white text-xs"
                  />
                </div>
              </div>
              <Button
                onclick={handleStaffHelp}
                disabled={loading}
                size="sm"
                class="h-8 w-full text-xs">Simpan Kehadiran</Button
              >
            </CardContent>
          </Card>
        </div>
      {/if}
    </div>

    <div class="flex flex-col lg:col-span-8">
      <Card class="flex-1 shadow-md">
        <CardHeader class="border-b bg-gray-50/50 pb-3">
          <div class="flex flex-col justify-between gap-6 md:flex-row md:items-center">
            <div class="space-y-1">
              <CardTitle class="flex items-center gap-2.5 text-xl font-bold text-slate-800">
                <div class="bg-primary/10 rounded-lg p-2">
                  <Users class="text-primary h-5 w-5" />
                </div>
                Daftar Kehadiran
              </CardTitle>
              <div class="mt-2 flex flex-wrap items-center gap-3">
                <CardDescription class="text-xs font-medium">Rekap harian praktikan</CardDescription
                >
                <div class="hidden h-4 w-[1px] bg-slate-200 sm:block"></div>
                {#if isStaff}
                  <div
                    class="flex items-center gap-1.5 rounded-full border border-blue-100 bg-blue-50 px-2.5 py-1 text-blue-700 shadow-sm transition-all hover:bg-blue-100"
                  >
                    <CheckCircle2 class="h-3 w-3" />
                    <span class="text-[10px] font-bold tracking-wider uppercase"
                      >Hadir: {totalHadir}</span
                    >
                  </div>
                {/if}
                <div
                  class="flex animate-pulse items-center gap-1.5 rounded-full border border-emerald-100 bg-emerald-50 px-2.5 py-1 text-emerald-700 shadow-sm"
                >
                  <span class="h-1.5 w-1.5 rounded-full bg-emerald-500"></span>
                  <span class="text-[10px] font-bold tracking-wider uppercase">Live Tracking</span>
                </div>
              </div>
            </div>
            {#if isStaff}
              <div class="flex flex-wrap items-center gap-3">
                <div
                  class="hover:border-primary/30 focus-within:ring-primary/10 flex items-center rounded-lg border border-slate-200 bg-white px-3 py-1.5 shadow-sm transition-all focus-within:ring-2"
                >
                  <Filter class="mr-2.5 h-4 w-4 text-slate-400" />
                  <Input
                    type="date"
                    aria-label="Filter tanggal"
                    class="h-6 w-[125px] border-none bg-transparent p-0 text-sm font-medium text-slate-700 shadow-none focus-visible:ring-0"
                    bind:value={filterDate}
                  />
                </div>
                <div class="flex items-center gap-2">
                  {#if hasExportAccess}
                    <Button
                      variant="outline"
                      size="sm"
                      class="flex h-9 items-center gap-2 border-emerald-200 px-4 text-emerald-700 shadow-sm transition-all hover:bg-emerald-50 hover:text-emerald-800 active:scale-95"
                      onclick={handleExportExcel}
                    >
                      <FileDown class="h-4 w-4" />
                      <span class="font-semibold">Export</span>
                    </Button>
                  {/if}
                  {#if hasDeletePower}
                    <Button
                      variant="outline"
                      size="sm"
                      onclick={() => (deletionOpen = true)}
                      class="flex h-9 items-center gap-2 border-rose-200 px-4 text-rose-600 shadow-sm transition-all hover:bg-rose-50 hover:text-rose-700 active:scale-95"
                    >
                      <Trash2 class="h-4 w-4" />
                      <span class="font-semibold">Audit Hapus</span>
                    </Button>
                    <Dialog bind:open={deletionOpen}>
                      <DialogContent
                        class="flex max-h-[80vh] max-w-3xl flex-col overflow-hidden rounded-2xl border-rose-100 shadow-2xl"
                      >
                        <DialogHeader>
                          <DialogTitle
                            class="flex items-center gap-3 text-xl font-bold text-rose-600"
                          >
                            <div class="rounded-lg bg-rose-50 p-2"><Trash2 class="h-5 w-5" /></div>
                            Log Riwayat Penghapusan
                          </DialogTitle>
                          <DialogDescription class="font-medium text-slate-500">
                            Daftar kehadiran yang telah dihapus oleh tim koordinator untuk keperluan
                            audit.
                          </DialogDescription>
                        </DialogHeader>
                        <div
                          class="mt-6 flex-1 overflow-y-auto rounded-xl border border-slate-100 bg-slate-50/30"
                        >
                          <Table>
                            <TableHeader class="sticky top-0 z-20 bg-slate-100/80 backdrop-blur-sm">
                              <TableRow class="hover:bg-transparent">
                                {#each ['Waktu', 'NIM Target', 'Dihapus Oleh', 'Keterangan'] as h (h)}
                                  <TableHead
                                    class="text-[10px] font-bold tracking-widest text-slate-500 uppercase"
                                    >{h}</TableHead
                                  >
                                {/each}
                              </TableRow>
                            </TableHeader>
                            <TableBody>
                              {#if deletionHistory.length === 0}
                                <TableRow
                                  ><TableCell
                                    colspan={4}
                                    class="py-16 text-center font-medium text-slate-400 italic"
                                    >Tidak ada riwayat penghapusan.</TableCell
                                  ></TableRow
                                >
                              {:else}
                                {#each deletionHistory as h (h.id)}
                                  <TableRow
                                    class="border-rose-50/50 transition-colors hover:bg-rose-50/30"
                                  >
                                    <TableCell
                                      class="font-mono text-[11px] font-medium text-slate-600"
                                      >{new Date(h.deleted_at).toLocaleString('id-ID')}</TableCell
                                    >
                                    <TableCell class="text-xs font-bold text-slate-800"
                                      >{h.target_user_id}</TableCell
                                    >
                                    <TableCell class="text-xs font-semibold text-slate-700"
                                      >{h.deleted_by_name || '-'}</TableCell
                                    >
                                    <TableCell
                                      class="text-[11px] leading-relaxed text-slate-500 italic"
                                      >{h.reason}</TableCell
                                    >
                                  </TableRow>
                                {/each}
                              {/if}
                            </TableBody>
                          </Table>
                        </div>
                      </DialogContent>
                    </Dialog>
                  {/if}
                </div>
              </div>
            {/if}
          </div>
        </CardHeader>
        <CardContent class="p-0">
          <div class="relative max-h-[500px] overflow-x-auto overflow-y-auto">
            <Table>
              <TableHeader
                class="sticky top-0 z-20 border-b bg-slate-50/80 shadow-sm backdrop-blur-md"
              >
                <TableRow class="hover:bg-transparent">
                  <TableHead
                    class="pl-6 text-[10px] font-bold tracking-widest text-slate-500 uppercase"
                    >Mahasiswa</TableHead
                  >
                  <TableHead class="text-[10px] font-bold tracking-widest text-slate-500 uppercase"
                    >Info Kelas</TableHead
                  >
                  <TableHead class="text-[10px] font-bold tracking-widest text-slate-500 uppercase"
                    >Waktu & Status</TableHead
                  >
                  {#if isStaff}<TableHead
                      class="text-[10px] font-bold tracking-widest text-slate-500 uppercase"
                      >Petugas</TableHead
                    >{/if}
                  {#if hasDeletePower}<TableHead
                      class="pr-6 text-center text-[10px] font-bold tracking-widest text-slate-500 uppercase"
                      >Aksi</TableHead
                    >{/if}
                </TableRow>
              </TableHeader>
              <TableBody>
                {#if displayAttendance.length === 0}
                  <TableRow>
                    <TableCell
                      colspan={hasDeletePower ? 5 : 3}
                      class="h-48 text-center text-sm text-slate-400 italic"
                    >
                      <div class="flex flex-col items-center gap-3">
                        <CalendarCheck class="h-12 w-12 text-slate-100" />
                        <p class="font-medium">Belum ada data absensi pada tanggal ini.</p>
                      </div>
                    </TableCell>
                  </TableRow>
                {:else}
                  {#each displayAttendance as log (log.id)}
                    <TableRow
                      class="group border-b border-slate-50 transition-all duration-200 hover:bg-slate-50/80"
                      data-testid="attendance-row"
                    >
                      <TableCell class="py-4 pl-6">
                        <div class="flex items-center gap-3">
                          <div
                            class="bg-primary/5 text-primary border-primary/10 flex h-9 w-9 items-center justify-center rounded-full border text-xs font-bold transition-transform group-hover:scale-110"
                          >
                            {log.users.full_name?.charAt(0) || '?'}
                          </div>
                          <div>
                            <div class="text-sm font-bold text-slate-800">
                              {log.users.full_name || 'Unknown'}
                            </div>
                            <div class="font-mono text-[11px] tracking-tighter text-slate-500">
                              {log.users.username}
                            </div>
                          </div>
                        </div>
                      </TableCell>
                      <TableCell class="py-4">
                        <div class="flex flex-col">
                          <span class="text-[11px] font-bold text-slate-700"
                            >{log.users.class_code}</span
                          >
                          <span
                            class="text-[10px] font-medium tracking-tight text-slate-500 uppercase"
                            >{log.users.major}</span
                          >
                          <Badge
                            variant="secondary"
                            class="mt-1 h-4 w-fit border-none bg-slate-100 px-1.5 text-[9px] font-semibold text-slate-600"
                          >
                            Shift {log.users.shift || '-'}
                          </Badge>
                        </div>
                      </TableCell>
                      <TableCell class="py-4">
                        <div class="flex flex-col items-start gap-2">
                          <div class="flex items-center gap-2">
                            <div
                              class="h-2 w-2 rounded-full {log.status === 'Hadir'
                                ? 'animate-pulse bg-emerald-500'
                                : log.status === 'Alpha'
                                  ? 'bg-rose-500'
                                  : 'bg-amber-500'}"
                            ></div>
                            <span
                              class="text-[11px] font-bold tracking-wide uppercase {log.status ===
                              'Hadir'
                                ? 'text-emerald-700'
                                : log.status === 'Alpha'
                                  ? 'text-rose-700'
                                  : 'text-amber-700'}"
                            >
                              {log.status}
                            </span>
                          </div>
                          <div
                            class="flex items-center gap-1.5 rounded border border-slate-100 bg-slate-50 px-2 py-0.5 font-mono text-[10px] text-slate-400"
                          >
                            <Clock class="h-3 w-3" />
                            {new Date(log.check_in_time).toLocaleTimeString('id-ID', {
                              hour: '2-digit',
                              minute: '2-digit',
                              timeZone: 'Asia/Jakarta',
                            })} WIB
                          </div>
                          {#if log.status !== 'Hadir' && log.notes}
                            <p
                              class="mt-1 line-clamp-1 max-w-[150px] border-l-2 border-slate-200 pl-2 text-[10px] text-slate-400 italic"
                            >
                              "{log.notes}"
                            </p>
                          {/if}
                        </div>
                      </TableCell>
                      {#if isStaff}
                        <TableCell class="py-4">
                          {#if log.recorded_by_name === log.users.full_name}
                            <Badge
                              variant="outline"
                              class="border-indigo-100 bg-indigo-50/30 px-2 py-0 text-[10px] font-medium text-indigo-600"
                            >
                              Mandiri
                            </Badge>
                          {:else}
                            <div class="flex items-center gap-1.5">
                              <div
                                class="flex h-5 w-5 items-center justify-center rounded-md bg-slate-100 text-[9px] font-bold text-slate-500 uppercase"
                              >
                                {log.recorded_by_name?.charAt(0) || '-'}
                              </div>
                              <span class="text-[11px] font-semibold text-slate-600"
                                >{log.recorded_by_name || '-'}</span
                              >
                            </div>
                          {/if}
                        </TableCell>
                      {/if}
                      {#if hasDeletePower}
                        <TableCell class="py-4 pr-6 text-center">
                          <Button
                            variant="ghost"
                            size="icon"
                            aria-label="Hapus absensi"
                            class="h-8 w-8 rounded-full text-rose-300 transition-all hover:bg-rose-50 hover:text-rose-600 active:scale-90"
                            onclick={() => deleteLog(log)}
                          >
                            <Trash2 class="h-4 w-4" />
                          </Button>
                        </TableCell>
                      {/if}
                    </TableRow>
                  {/each}
                {/if}
              </TableBody>
            </Table>
          </div>
        </CardContent>
      </Card>
    </div>
  </div>

  {#if !isStaff && myLogs.length > 0}
    <div class="mt-8 space-y-5 border-t pt-8">
      <h3 class="flex items-center gap-3 text-xl font-bold text-slate-800">
        <div class="bg-primary/10 rounded-lg p-2"><Clock class="text-primary h-5 w-5" /></div>
        Riwayat & Status Pengajuan Saya
      </h3>
      <div class="grid grid-cols-1 gap-5 md:grid-cols-2 lg:grid-cols-3">
        {#each myLogs as log (log.id)}
          <Card
            class="group overflow-hidden border-none bg-white shadow-md transition-all duration-300 hover:shadow-xl"
          >
            <div
              class="h-1.5 w-full {log.status === 'Hadir' ? 'bg-emerald-500' : 'bg-amber-500'}"
            ></div>
            <CardContent class="p-5">
              <div class="flex flex-col gap-4">
                <div class="flex items-start justify-between">
                  <div class="flex flex-col gap-1">
                    <Badge
                      class="{log.status === 'Hadir'
                        ? 'bg-emerald-500 hover:bg-emerald-600'
                        : 'bg-amber-500 hover:bg-amber-600'} w-fit text-[10px] font-bold tracking-wider uppercase"
                    >
                      {log.status}
                    </Badge>
                    <span class="mt-1 font-mono text-[10px] text-slate-400">
                      {new Date(log.check_in_time).toLocaleDateString('id-ID', {
                        day: '2-digit',
                        month: 'long',
                        year: 'numeric',
                      })}
                    </span>
                  </div>
                  <div
                    class="rounded-md border border-slate-100 bg-slate-50 px-2 py-1 text-[11px] font-bold text-slate-400"
                  >
                    {new Date(log.check_in_time).toLocaleTimeString('id-ID', {
                      hour: '2-digit',
                      minute: '2-digit',
                    })} WIB
                  </div>
                </div>
                <div class="space-y-2">
                  <p
                    class="group-hover:text-primary text-sm leading-snug font-semibold text-slate-700 transition-colors"
                  >
                    {log.notes}
                  </p>
                  <div class="mt-3 flex flex-wrap gap-2">
                    {#if log.reschedule_status === 'pending'}
                      <Badge
                        variant="outline"
                        class="border-amber-200 bg-amber-50 py-0.5 text-[10px] text-amber-700"
                        >Menunggu ACC Jadwal</Badge
                      >
                    {:else if log.reschedule_status === 'approved'}
                      <Badge
                        variant="outline"
                        class="border-emerald-200 bg-emerald-50 py-0.5 text-[10px] text-emerald-700"
                      >
                        Jadwal Fix: {log.schedules?.day_of_week}
                        {log.schedules?.start_time?.slice(0, 5)}
                      </Badge>
                    {:else if log.reschedule_status === 'rejected'}
                      <Badge
                        variant="outline"
                        class="border-rose-200 bg-rose-50 py-0.5 text-[10px] text-rose-700"
                        >Jadwal Ditolak</Badge
                      >
                    {/if}
                  </div>
                </div>
                <div class="flex flex-col gap-2 border-t border-slate-50 pt-2">
                  {#if log.status !== 'Hadir' && adminPhone}
                    <a
                      href={getWaProofLink(adminPhone, log.status, log.notes || '')}
                      target="_blank"
                      rel="noreferrer external"
                      class="w-full"
                    >
                      <Button
                        size="sm"
                        class="h-9 w-full bg-[#25D366] text-xs text-white shadow-sm transition-all hover:bg-[#1ebd5c] active:scale-95"
                      >
                        <Send class="mr-2 h-3 w-3" /> Kirim Bukti WA
                      </Button>
                    </a>
                  {/if}
                  {#if log.is_verified && log.status !== 'Hadir'}
                    {#if !log.reschedule_status || log.reschedule_status === 'rejected'}
                      <Button
                        size="sm"
                        variant="outline"
                        onclick={() => {
                          void invalidateRpcs(['get_schedules_secure']);
                          rescheduleFor = log.id;
                        }}
                        class="border-primary/20 text-primary hover:bg-primary/5 h-9 w-full text-xs transition-all active:scale-95"
                      >
                        <ArrowRightLeft class="mr-2 h-3 w-3" /> Pilih Jadwal
                      </Button>
                    {/if}
                    {#if log.reschedule_status === 'pending'}
                      <Button
                        size="sm"
                        disabled
                        class="h-8 w-full border-yellow-200 bg-yellow-100 text-xs text-yellow-700 opacity-100"
                      >
                        <Clock class="mr-2 h-3 w-3" /> Menunggu Verifikasi
                      </Button>
                    {/if}
                    {#if log.reschedule_status === 'approved'}
                      <Button
                        size="sm"
                        variant="ghost"
                        disabled
                        class="h-8 w-full text-xs text-green-700 opacity-100"
                      >
                        <Lock class="mr-2 h-3 w-3" /> Jadwal Terkunci
                      </Button>
                    {/if}
                  {/if}
                </div>
              </div>
            </CardContent>
          </Card>
        {/each}
      </div>
    </div>
  {/if}

  <Dialog
    open={rescheduleFor !== null}
    onOpenChange={(open) => {
      if (!open) rescheduleFor = null;
    }}
  >
    <DialogContent>
      <DialogHeader>
        <DialogTitle>Pilih Jadwal Pengganti</DialogTitle>
        <DialogDescription>
          Silakan pilih jadwal pengganti yang tersedia untuk menggantikan kehadiran Anda.
        </DialogDescription>
      </DialogHeader>
      <div class="space-y-4 py-2">
        <Label>Jadwal Tersedia:</Label>
        <Select type="single" bind:value={selectedSchedule}>
          <SelectTrigger class="w-full">
            {#if selectedSchedule}
              {@const s = availableSchedules.find((x) => x.id.toString() === selectedSchedule)}
              {s?.day_of_week} - {s?.start_time?.slice(0, 5)} ({s?.title})
            {:else}
              <span class="text-muted-foreground">Pilih Jadwal...</span>
            {/if}
          </SelectTrigger>
          <SelectContent>
            {#each availableSchedules as s (s.id)}
              <SelectItem
                value={s.id.toString()}
                label={`${s.day_of_week} - ${s.start_time?.slice(0, 5)} (${s.title})`}
              />
            {/each}
          </SelectContent>
        </Select>
        <Button
          onclick={() => rescheduleFor !== null && handleReschedule(rescheduleFor)}
          disabled={loading}
          class="w-full">Ajukan Jadwal Ini</Button
        >
      </div>
    </DialogContent>
  </Dialog>
</div>
