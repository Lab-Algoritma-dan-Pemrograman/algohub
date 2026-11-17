<script lang="ts">
  import {
    Badge,
    Button,
    Card,
    CardContent,
    CardHeader,
    DataTable,
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
    Input,
    Label,
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    Switch,
    renderSnippet,
    type ColumnDef,
  } from '@algohub/ui';
  import Briefcase from '@lucide/svelte/icons/briefcase';
  import Download from '@lucide/svelte/icons/download';
  import Eye from '@lucide/svelte/icons/eye';
  import EyeOff from '@lucide/svelte/icons/eye-off';
  import FileSpreadsheet from '@lucide/svelte/icons/file-spreadsheet';
  import Filter from '@lucide/svelte/icons/filter';
  import GraduationCap from '@lucide/svelte/icons/graduation-cap';
  import Loader2 from '@lucide/svelte/icons/loader-2';
  import Lock from '@lucide/svelte/icons/lock';
  import Pencil from '@lucide/svelte/icons/pencil';
  import Phone from '@lucide/svelte/icons/phone';
  import Plus from '@lucide/svelte/icons/plus';
  import Save from '@lucide/svelte/icons/save';
  import Search from '@lucide/svelte/icons/search';
  import ShieldCheck from '@lucide/svelte/icons/shield-check';
  import Trash2 from '@lucide/svelte/icons/trash-2';
  import UserCheck from '@lucide/svelte/icons/user-check';
  import UserCog from '@lucide/svelte/icons/user-cog';
  import { siakadKeys } from '@algohub/api-client/siakad';
  import { createQuery } from '@tanstack/svelte-query';
  import { queryClient } from '#lib/query.ts';
  import { createRpcMutation, invalidateRpcs, rpcQueryOptions } from '#lib/queries/rpc.ts';
  import { canonRole } from '#lib/roles.ts';
  import { confirm } from '#lib/stores/confirm.svelte.ts';
  import { toast } from '#lib/toast.ts';
  import type { PageProps } from './$types';

  const MASTER_MENUS = [
    { key: '/jadwal-jaga', label: 'Jadwal Jaga' },
    { key: '/validasi-absensi', label: 'Validasi Absensi' },
    { key: '/manajemen-jadwal', label: 'Manajemen Jadwal' },
    { key: '/manajemen-kelas', label: 'Manajemen Kelas' },
    { key: '/inventaris', label: 'Inventaris Barang' },
    { key: '/manajemen-user', label: 'Manajemen User' },
    { key: '/laporan-keuangan', label: 'Laporan Keuangan' },
    { key: '/ketersediaan', label: 'Input Jadwal Free' },
    { key: '/e-learning', label: 'Modul E-Learning' },
    { key: '/absensi', label: 'Kelola Absensi (Export & Hapus)' },
    { key: '/audit-log', label: 'Audit Log Aktivitas' },
  ];

  interface UserData {
    id: number;
    username: string;
    password?: string;
    full_name: string;
    phone_number?: string;
    role: string;
    nim?: string;
    assistant_code?: string;
    division?: string;
    class_code?: string;
    shift?: string;
    major?: string;
    is_active: boolean;
  }

  const DEFAULT_FORM: UserData = {
    id: 0,
    username: '',
    password: '',
    full_name: '',
    phone_number: '',
    role: 'praktikan',
    is_active: true,
    nim: '',
    class_code: '',
    shift: '',
    assistant_code: '',
    division: '',
  };

  const ROLE_FILTERS: Record<string, string> = {
    all: 'Semua Peran',
    praktikan: 'Praktikan',
    penyewa: 'Penyewa Umum',
    asisten: 'Asisten',
    koordinator: 'Koordinator',
  };
  const SHIFT_FILTERS: Record<string, string> = {
    all: 'Semua Shift',
    '1': 'Shift 1',
    '2': 'Shift 2',
  };
  const SHIFT_OPTIONS = [
    { val: '1', label: 'S1' },
    { val: '2', label: 'S2' },
    { val: 'all', label: 'SEMUA' },
    { val: 'none', label: 'TUTUP' },
  ];

  function getJurusanByNIM(nim: string | undefined) {
    if (!nim || nim.length < 6) return '-';
    switch (nim.substring(4, 6)) {
      case '14':
        return 'S1 Teknik Tenaga Listrik';
      case '15':
        return 'S1 Teknik Sistem Energi';
      case '71':
        return 'D3 Teknologi Listrik';
      case '11':
        return 'S1 Teknik Elektro';
      default:
        return 'Jurusan Lain';
    }
  }

  let { data }: PageProps = $props();
  const currentUser = $derived(data.auth.user);
  const isKoordinator = $derived(currentUser.role === 'koordinator');
  const isStaff = $derived(['koordinator', 'asisten'].includes(currentUser.role || ''));

  const usersParams = $derived({ p_viewer_id: currentUser.id });
  const usersQuery = createQuery(() =>
    rpcQueryOptions<UserData[]>('get_users_secure', usersParams),
  );
  const shiftQuery = createQuery(() =>
    rpcQueryOptions<{ active_shift?: string }[]>('get_system_settings_secure', {
      p_viewer_id: currentUser.id,
    }),
  );

  $effect(() => {
    if (usersQuery.error) {
      console.error(usersQuery.error);
      toast.error('Gagal mengambil data user.');
    }
  });
  $effect(() => {
    if (shiftQuery.error) console.error('Error fetching global shift:', shiftQuery.error);
  });

  const users = $derived(usersQuery.data || []);
  const loading = $derived(usersQuery.isFetching);
  // Tabel divisi tidak diberi GRANT untuk anon di produksi, jadi daftar diturunkan dari users.
  // Casing asli dipertahankan karena hak akses dicari memakai users.division apa adanya.
  const divisions = $derived.by(() => {
    const byLower: Record<string, string> = {};
    for (const u of users) {
      const div = (u.division || '').trim();
      if (div && !(div.toLowerCase() in byLower)) byLower[div.toLowerCase()] = div;
    }
    return Object.values(byLower).sort();
  });

  const refetchUsers = () => invalidateRpcs(['get_users_secure']);
  const deleteUser = createRpcMutation('admin_delete_user');
  const updateShift = createRpcMutation('admin_update_system_setting', [
    'get_system_settings_secure',
  ]);
  const saveAccess = createRpcMutation('admin_save_division_access', [
    'get_division_access_secure',
  ]);
  const updateUser = createRpcMutation('admin_update_user');
  const updatePassword = createRpcMutation('update_password');
  const createUser = createRpcMutation('admin_create_user');
  const toggleStatus = createRpcMutation('admin_toggle_user_status');
  const registerBatch = createRpcMutation('register_users_batch');

  let busy = $state(false);
  let search = $state('');
  let roleFilter = $state('all');
  let shiftFilter = $state('all');
  let showPassword = $state(false);
  let isOpen = $state(false);
  let isEdit = $state(false);
  let formData = $state<UserData>({ ...DEFAULT_FORM });
  let saving = $state(false);
  let localShift = $state<string | null>(null);
  let isUpdatingShift = $state(false);
  let isAccessOpen = $state(false);
  let selectedDivision = $state('');
  let accessList = $state<string[]>([]);
  let savingAccess = $state(false);
  let selectedIds = $state<number[]>([]);
  let fileInput = $state<HTMLInputElement | null>(null);

  const globalActiveShift = $derived(
    localShift ??
      (shiftQuery.data && shiftQuery.data.length > 0 ? shiftQuery.data[0].active_shift : 'all'),
  );

  const accessQuery = createQuery(() =>
    rpcQueryOptions<{ menu_key: string }[]>(
      'get_division_access_secure',
      { p_viewer_id: currentUser.id, p_division: selectedDivision },
      { enabled: !!selectedDivision && isAccessOpen, staleTime: 0 },
    ),
  );
  // Tanpa memeriksa error, sesi mati tampil sebagai "0 menu" seolah hak akses hilang.
  $effect(() => {
    if (accessQuery.error) {
      accessList = [];
      toast.error('Gagal memuat hak akses: ' + accessQuery.error.message);
    } else if (accessQuery.data) {
      accessList = accessQuery.data.map((d) => d.menu_key);
    }
  });
  const loadingAccess = $derived(savingAccess || (accessQuery.isFetching && !!selectedDivision));

  const filteredUsers = $derived(
    users.filter((u) => {
      const q = search.toLowerCase();
      const matchesSearch =
        u.full_name.toLowerCase().includes(q) ||
        u.username.toLowerCase().includes(q) ||
        (u.nim && u.nim.includes(search)) ||
        (u.phone_number && u.phone_number.includes(search)) ||
        (u.class_code && u.class_code.toLowerCase().includes(q)) ||
        (u.division && u.division.toLowerCase().includes(q)) ||
        getJurusanByNIM(u.nim).toLowerCase().includes(q);
      return (
        matchesSearch &&
        (roleFilter === 'all' || u.role === roleFilter) &&
        (shiftFilter === 'all' || u.shift === shiftFilter)
      );
    }),
  );

  function toggleSelect(id: number) {
    selectedIds = selectedIds.includes(id)
      ? selectedIds.filter((i) => i !== id)
      : [...selectedIds, id];
  }

  function toggleSelectAll() {
    selectedIds = selectedIds.length === filteredUsers.length ? [] : filteredUsers.map((u) => u.id);
  }

  async function handleBulkDelete() {
    if (selectedIds.length === 0) return;
    if (
      !(await confirm(
        `Hapus ${selectedIds.length} user terpilih? Seluruh riwayat absensi, bimbingan, dan jadwal yang berkaitan akan ikut terhapus secara permanen.`,
      ))
    )
      return;
    busy = true;
    let successCount = 0;
    let failCount = 0;
    // Serial supaya database tidak terbebani dan pembersihan data terkait berurutan.
    for (const id of selectedIds) {
      const { error } = await deleteUser.mutateAsync({
        p_caller_id: currentUser.id,
        p_target_id: id,
      });
      if (error) {
        console.error(`Gagal menghapus ID ${id}:`, error);
        failCount++;
      } else {
        successCount++;
      }
    }
    toast.success(`Proses selesai. ${successCount} berhasil dihapus, ${failCount} gagal.`);
    selectedIds = [];
    void refetchUsers();
    busy = false;
  }

  async function handleUpdateActiveShift(val: string) {
    isUpdatingShift = true;
    try {
      const { error } = await updateShift.mutateAsync({
        p_caller_id: currentUser.id,
        p_active_shift: val,
      });
      if (error) throw error;
      localShift = val;
      toast.success(
        `Akses praktikan diatur ke: ${val === 'all' ? 'Semua Aktif' : val === 'none' ? 'Semua Tertutup' : 'Hanya Shift ' + val}`,
      );
    } catch (err) {
      toast.error('Gagal update shift global: ' + (err as Error).message);
    } finally {
      isUpdatingShift = false;
    }
  }

  async function handleSaveAccess() {
    if (!isKoordinator) {
      toast.error('Akses Ditolak', {
        description: 'Hanya Koordinator yang bisa mengelola hak akses divisi.',
      });
      return;
    }
    savingAccess = true;
    try {
      const { error } = await saveAccess.mutateAsync({
        p_caller_id: currentUser.id,
        p_division: selectedDivision,
        p_menu_keys: accessList.length > 0 ? accessList : [],
      });
      if (error) throw error;
      toast.success('Hak akses disimpan!');
    } catch (err) {
      toast.error((err as Error).message);
    } finally {
      savingAccess = false;
    }
  }

  function toggleMenuAccess(menuKey: string) {
    accessList = accessList.includes(menuKey)
      ? accessList.filter((k) => k !== menuKey)
      : [...accessList, menuKey];
  }

  async function handleSubmit(e: SubmitEvent) {
    e.preventDefault();
    if (!isStaff) {
      toast.error('Akses Ditolak', {
        description: 'Anda tidak memiliki izin untuk menyimpan data user.',
      });
      return;
    }
    saving = true;
    try {
      const isPraktikan = canonRole(formData.role) === 'praktikan';
      const staffRole = formData.role === 'asisten' || formData.role === 'koordinator';
      const payload = {
        username: formData.username,
        password: formData.password,
        full_name: formData.full_name,
        phone_number: formData.phone_number || null,
        role: isPraktikan ? 'praktikan' : formData.role,
        is_active: formData.is_active,
        shift: isPraktikan ? formData.shift || null : null,
        nim: formData.nim || formData.username || null,
        class_code: isPraktikan ? formData.class_code || null : null,
        division: staffRole ? formData.division || null : null,
        assistant_code: formData.role === 'asisten' ? formData.assistant_code || null : null,
      };
      if (!isEdit && (!payload.password || payload.password.trim().length < 6)) {
        toast.error('Password wajib diisi dan minimal 6 karakter.');
        return;
      }
      if (isEdit) {
        const { error: updateError } = await updateUser.mutateAsync({
          p_caller_id: currentUser.id,
          p_target_id: formData.id,
          p_username: payload.username,
          p_full_name: payload.full_name,
          p_phone_number: payload.phone_number || null,
          p_role: payload.role,
          p_is_active: payload.is_active,
          p_shift: payload.shift || null,
          p_nim: payload.nim || null,
          p_class_code: payload.class_code || null,
          p_division: payload.division || null,
          p_assistant_code: payload.assistant_code || null,
        });
        if (updateError) throw updateError;
        if (payload.password && payload.password.length > 0) {
          const { error: pwdError } = await updatePassword.mutateAsync({
            p_caller_id: currentUser.id,
            p_target_id: formData.id,
            p_new_password: payload.password,
          });
          if (pwdError) throw pwdError;
        }
      } else {
        const { error: insertError } = await createUser.mutateAsync({
          p_caller_id: currentUser.id,
          p_username: payload.username,
          p_password: payload.password,
          p_full_name: payload.full_name,
          p_role: payload.role,
          p_nim: payload.nim,
          p_assistant_code: payload.assistant_code,
          p_division: payload.division,
          p_phone_number: payload.phone_number,
          p_class_code: payload.class_code,
          p_shift: payload.shift,
          p_is_active: payload.is_active,
          p_major: formData.major || null,
        });
        if (insertError) throw insertError;
      }
      toast.success('Data berhasil disimpan!');
      isOpen = false;
      void refetchUsers();
    } catch (err) {
      toast.error((err as Error).message);
    } finally {
      saving = false;
    }
  }

  async function handleDelete(id: number, name: string) {
    if (!isKoordinator) {
      toast.error('Akses Ditolak', { description: 'Hanya Koordinator yang dapat menghapus user.' });
      return;
    }
    if (
      !(await confirm(
        `Hapus user "${name}"? Seluruh riwayat absensi, bimbingan, dan jadwal yang berkaitan dengan orang ini akan ikut terhapus secara permanen.`,
      ))
    )
      return;
    busy = true;
    try {
      const { error } = await deleteUser.mutateAsync({
        p_caller_id: currentUser.id,
        p_target_id: id,
      });
      if (error) throw error;
      toast.success('User dan seluruh data terkait berhasil dihapus.');
      void refetchUsers();
    } catch (err) {
      console.error('Delete error:', err);
      toast.error('Gagal menghapus user', {
        description:
          (err as Error).message || 'Terjadi kesalahan saat mencoba menghapus data terkait.',
      });
    } finally {
      busy = false;
    }
  }

  async function handleToggleStatus(id: number, currentStatus: boolean) {
    if (!isStaff) {
      toast.error('Akses Ditolak', {
        description: 'Anda tidak memiliki izin untuk mengubah status user.',
      });
      return;
    }
    try {
      const { error } = await toggleStatus.mutateAsync({
        p_caller_id: currentUser.id,
        p_target_id: id,
        p_is_active: !currentStatus,
      });
      if (error) throw error;
      queryClient.setQueryData<UserData[]>(
        siakadKeys.rpc('get_users_secure', usersParams),
        (rows) => rows?.map((u) => (u.id === id ? { ...u, is_active: !currentStatus } : u)),
      );
      toast.success(`Status user berhasil ${!currentStatus ? 'diaktifkan' : 'dinonaktifkan'}.`);
    } catch (err) {
      toast.error('Gagal mengubah status: ' + (err as Error).message);
    }
  }

  function handleFileUpload(e: Event & { currentTarget: HTMLInputElement }) {
    const file = e.currentTarget.files?.[0];
    if (!file) return;
    if (!isStaff) {
      toast.error('Akses Ditolak', {
        description: 'Anda tidak memiliki izin untuk mengimport data.',
      });
      return;
    }
    const reader = new FileReader();
    reader.readAsArrayBuffer(file);
    reader.onload = async (evt) => {
      try {
        const XLSX = await import('xlsx');
        const wb = XLSX.read(evt.target?.result, { type: 'array' });
        const rawData = XLSX.utils.sheet_to_json<Record<string, unknown>>(
          wb.Sheets[wb.SheetNames[0]],
        );
        if (rawData.length === 0) return toast.error('File kosong!');
        const rows = rawData.map((row) =>
          Object.fromEntries(Object.entries(row).map(([k, v]) => [k.trim().toLowerCase(), v])),
        );
        const formatted = rows
          .map((row) => {
            const u = row['username'] || row['nim'];
            if (!u) return null;
            const phone = row['no hp'] || row['telepon'] || row['whatsapp'];
            return {
              username: String(u),
              password: String(u),
              full_name: row['nama'] || row['nama lengkap'],
              phone_number: phone ? String(phone) : null,
              role: 'praktikan',
              nim: String(u),
              class_code: row['kelas'],
              shift: row['shift'] ? String(row['shift']) : null,
              is_active: true,
              assistant_code: null,
              division: null,
              major: row['jurusan'] || row['prodi'] || row['program studi'] || null,
            };
          })
          .filter((item) => item !== null);
        const unique = [...new Map(formatted.map((item) => [item.username, item])).values()];
        busy = true;
        const { error } = await registerBatch.mutateAsync({
          p_caller_id: currentUser.id,
          p_users: unique,
        });
        if (error) throw error;
        toast.success(`Import berhasil: ${unique.length} data.`);
        void refetchUsers();
      } catch (err) {
        toast.error((err as Error).message);
      } finally {
        busy = false;
        if (fileInput) fileInput.value = '';
      }
    };
  }

  async function handleDownloadTemplateUser() {
    const XLSX = await import('xlsx');
    const ws = XLSX.utils.json_to_sheet([
      {
        NIM: '202414001',
        Nama: 'Ahmad Praktikan',
        'No HP': '081234567890',
        Kelas: 'S1-A',
        Shift: '1',
        Jurusan: 'S1 Teknik Tenaga Listrik',
      },
      {
        NIM: '202414002',
        Nama: 'Budi Praktikan',
        'No HP': '081298765432',
        Kelas: 'S1-B',
        Shift: '2',
        Jurusan: 'S1 Teknik Sistem Energi',
      },
    ]);
    const wb = XLSX.utils.book_new();
    XLSX.utils.book_append_sheet(wb, ws, 'Template Praktikan');
    XLSX.writeFile(wb, 'Template_Import_Praktikan.xlsx');
    toast.success('Template berhasil diunduh.');
  }

  function openAdd() {
    formData = { ...DEFAULT_FORM };
    isEdit = false;
    showPassword = false;
    isOpen = true;
  }

  function openEdit(target: UserData) {
    formData = { ...target, role: canonRole(target.role), password: '' };
    isEdit = true;
    showPassword = false;
    isOpen = true;
  }

  function canEdit(target: UserData) {
    if (currentUser.role === 'koordinator') return true;
    if (currentUser.role === 'asisten') {
      return target.id === currentUser.id || canonRole(target.role) === 'praktikan';
    }
    return false;
  }

  const roleOptions = $derived([
    { value: 'praktikan', label: 'Praktikan' },
    { value: 'penyewa', label: 'Penyewa Umum' },
    ...(isKoordinator || formData.role === 'asisten'
      ? [{ value: 'asisten', label: 'Asisten' }]
      : []),
    ...(isKoordinator ? [{ value: 'koordinator', label: 'Koordinator' }] : []),
  ]);

  const roleBadge = (role: string) =>
    role === 'koordinator'
      ? 'destructive'
      : role === 'asisten'
        ? 'default'
        : role === 'penyewa'
          ? 'outline'
          : 'secondary';

  const columns: ColumnDef<UserData>[] = [
    {
      id: 'select',
      header: () => renderSnippet(selectAllHeader),
      cell: ({ row }) => renderSnippet(selectCell, row.original),
    },
    {
      id: 'name',
      header: 'Nama Lengkap',
      cell: ({ row }) => renderSnippet(nameCell, row.original),
    },
    {
      id: 'username',
      header: 'NIM / Username',
      cell: ({ row }) => renderSnippet(usernameCell, row.original),
    },
    {
      id: 'contact',
      header: 'Kontak',
      cell: ({ row }) => renderSnippet(contactCell, row.original),
    },
    { id: 'shift', header: 'Shift', cell: ({ row }) => renderSnippet(shiftCell, row.original) },
    {
      id: 'detail',
      header: 'Detail (Kelas)',
      cell: ({ row }) => renderSnippet(detailCell, row.original),
    },
    { id: 'role', header: 'Role', cell: ({ row }) => renderSnippet(roleCell, row.original) },
    { id: 'status', header: 'Status', cell: ({ row }) => renderSnippet(statusCell, row.original) },
    { id: 'actions', header: 'Aksi', cell: ({ row }) => renderSnippet(actionsCell, row.original) },
  ];
</script>

{#snippet selectAllHeader()}
  <input
    type="checkbox"
    aria-label="Pilih semua user"
    class="h-4 w-4 rounded border-gray-300"
    checked={filteredUsers.length > 0 && selectedIds.length === filteredUsers.length}
    onchange={toggleSelectAll}
  />
{/snippet}

{#snippet selectCell(u: UserData)}
  <input
    type="checkbox"
    aria-label="Pilih user"
    class="h-4 w-4 rounded border-gray-300"
    checked={selectedIds.includes(u.id)}
    onchange={() => toggleSelect(u.id)}
  />
{/snippet}

{#snippet nameCell(u: UserData)}
  <span class="font-medium whitespace-nowrap" data-testid="user-row">{u.full_name}</span>
{/snippet}

{#snippet usernameCell(u: UserData)}
  <span class="font-mono text-xs whitespace-nowrap">{u.username}</span>
{/snippet}

{#snippet contactCell(u: UserData)}
  {#if u.phone_number}
    <div class="flex items-center gap-1 text-sm whitespace-nowrap text-gray-600">
      <Phone class="h-3 w-3" />
      {u.phone_number}
    </div>
  {:else}
    <span class="text-muted-foreground text-xs italic">-</span>
  {/if}
{/snippet}

{#snippet shiftCell(u: UserData)}
  {#if canonRole(u.role) === 'praktikan'}
    <Badge
      variant="outline"
      class="font-bold {u.shift === '1'
        ? 'border-orange-200 bg-orange-50 text-orange-700'
        : 'border-purple-200 bg-purple-50 text-purple-700'}">Shift {u.shift || '-'}</Badge
    >
  {:else}
    <span class="text-muted-foreground text-xs italic">-</span>
  {/if}
{/snippet}

{#snippet detailCell(u: UserData)}
  <div class="text-sm">
    {#if canonRole(u.role) === 'praktikan'}
      <div class="flex flex-col gap-0.5">
        <span
          class="w-fit rounded-md border border-blue-200 bg-blue-50 px-2 py-0.5 text-xs font-semibold"
          >Kelas: {u.class_code || '-'}</span
        >
        <div class="text-muted-foreground mt-1 flex items-center gap-1 text-xs">
          <GraduationCap class="h-3 w-3" /><span>{getJurusanByNIM(u.nim || u.username)}</span>
        </div>
      </div>
    {:else if u.role === 'peminjam'}
      <div class="flex flex-col gap-1">
        <Badge
          variant="outline"
          class="w-fit border-slate-200 bg-slate-50 text-[10px] text-slate-500">PENYEWA UMUM</Badge
        >
        <span class="text-muted-foreground ml-1 text-[10px] italic">Katalog Publik Aktif</span>
      </div>
    {:else}
      <div class="flex flex-col gap-0.5">
        <div class="text-foreground flex items-center gap-1.5 font-medium">
          <Briefcase class="text-muted-foreground h-3 w-3" /><span>Divisi: {u.division || '-'}</span
          >
        </div>
        {#if u.role === 'asisten'}<span class="text-muted-foreground pl-5 text-xs"
            >Kode: {u.assistant_code || '-'}</span
          >{/if}
      </div>
    {/if}
  </div>
{/snippet}

{#snippet roleCell(u: UserData)}
  <Badge variant={roleBadge(u.role)}>{canonRole(u.role).toUpperCase()}</Badge>
{/snippet}

{#snippet statusCell(u: UserData)}
  <div class="flex items-center justify-center gap-2">
    <Switch
      aria-label="Status aktif"
      checked={u.is_active}
      onCheckedChange={() => handleToggleStatus(u.id, u.is_active)}
      disabled={!canEdit(u) || (u.id === currentUser.id && u.role === 'koordinator')}
    />
    <span
      class="w-12 text-[10px] font-bold uppercase {u.is_active ? 'text-green-600' : 'text-red-500'}"
      >{u.is_active ? 'Aktif' : 'Non-Aktif'}</span
    >
  </div>
{/snippet}

{#snippet actionsCell(u: UserData)}
  <div class="flex justify-center gap-2">
    {#if canEdit(u)}<Button
        variant="ghost"
        size="icon"
        aria-label="Edit user"
        onclick={() => openEdit(u)}><Pencil class="h-4 w-4 text-blue-500" /></Button
      >{/if}
    {#if isKoordinator && u.id !== currentUser.id}<Button
        variant="ghost"
        size="icon"
        aria-label="Hapus user"
        onclick={() => handleDelete(u.id, u.full_name)}
        ><Trash2 class="h-4 w-4 text-red-500" /></Button
      >{/if}
  </div>
{/snippet}

<div class="space-y-6">
  <div class="flex flex-col items-start justify-between gap-4 md:flex-row md:items-center">
    <div>
      <h1 class="flex items-center gap-2 text-2xl font-bold">
        <UserCog class="text-primary h-8 w-8" /> Manajemen User
      </h1>
      <p class="text-muted-foreground">Kelola akun, shift, dan hak akses divisi.</p>
    </div>
    {#if isKoordinator}
      <div
        class="bg-primary/5 border-primary/20 flex flex-col items-center gap-3 rounded-lg border p-3 sm:flex-row"
      >
        <div class="text-primary flex items-center gap-2 text-sm font-semibold">
          <UserCheck class="h-4 w-4" />
          <span>Akses Praktikan:</span>
        </div>
        <div class="flex rounded-md border bg-white p-1 shadow-sm">
          {#each SHIFT_OPTIONS as opt (opt.val)}
            <button
              disabled={isUpdatingShift}
              onclick={() => handleUpdateActiveShift(opt.val)}
              class="rounded px-3 py-1 text-xs font-bold transition-all {globalActiveShift ===
              opt.val
                ? 'bg-primary text-white shadow-sm'
                : 'text-muted-foreground hover:bg-muted'}"
            >
              {opt.label}
            </button>
          {/each}
        </div>
        {#if isUpdatingShift}<Loader2 class="text-primary h-4 w-4 animate-spin" />{/if}
      </div>
    {/if}
    <div class="flex gap-2">
      <input
        type="file"
        accept=".xlsx, .xls"
        bind:this={fileInput}
        class="hidden"
        onchange={handleFileUpload}
      />
      {#if isStaff}
        {#if isKoordinator}
          <Button
            variant="secondary"
            onclick={() => {
              selectedDivision = '';
              isAccessOpen = true;
            }}
          >
            <ShieldCheck class="mr-2 h-4 w-4" /> Kelola Akses Divisi
          </Button>
        {/if}
        <div class="flex gap-1">
          <Button
            variant="outline"
            class="border-green-600 text-green-700 hover:bg-green-50"
            onclick={() => fileInput?.click()}
          >
            <FileSpreadsheet class="mr-2 h-4 w-4" /> Import Excel
          </Button>
          <Button
            variant="ghost"
            size="icon"
            class="text-muted-foreground"
            title="Download Template Excel"
            onclick={handleDownloadTemplateUser}
          >
            <Download class="h-4 w-4" />
          </Button>
        </div>
        {#if selectedIds.length > 0}
          <Button variant="destructive" onclick={handleBulkDelete}>
            <Trash2 class="mr-2 h-4 w-4" /> Hapus {selectedIds.length} User
          </Button>
        {/if}
        <Button onclick={openAdd}><Plus class="mr-2 h-4 w-4" /> Tambah Manual</Button>
      {/if}
    </div>
  </div>

  <Card>
    <CardHeader class="pb-3">
      <div class="flex flex-col gap-3 sm:flex-row">
        <div class="relative flex-1">
          <Search class="text-muted-foreground absolute top-2.5 left-2 h-4 w-4" />
          <Input placeholder="Cari Nama, NIM, No HP, Kelas..." class="pl-8" bind:value={search} />
        </div>
        <div class="w-full sm:w-[200px]">
          <Select type="single" bind:value={roleFilter}>
            <SelectTrigger class="w-full"
              ><Filter class="text-muted-foreground mr-2 h-4 w-4" />{ROLE_FILTERS[
                roleFilter
              ]}</SelectTrigger
            >
            <SelectContent>
              {#each Object.entries(ROLE_FILTERS) as [value, label] (value)}<SelectItem
                  {value}
                  {label}
                />{/each}
            </SelectContent>
          </Select>
        </div>
        <div class="w-full sm:w-[150px]">
          <Select type="single" bind:value={shiftFilter}>
            <SelectTrigger class="w-full"
              ><Filter class="text-muted-foreground mr-2 h-4 w-4" />{SHIFT_FILTERS[
                shiftFilter
              ]}</SelectTrigger
            >
            <SelectContent>
              {#each Object.entries(SHIFT_FILTERS) as [value, label] (value)}<SelectItem
                  {value}
                  {label}
                />{/each}
            </SelectContent>
          </Select>
        </div>
      </div>
    </CardHeader>
    <CardContent>
      {#if loading || busy}
        <div class="flex justify-center py-8"><Loader2 class="text-primary animate-spin" /></div>
      {:else if filteredUsers.length === 0}
        <div class="text-muted-foreground rounded-md border py-8 text-center">
          Data tidak ditemukan.
        </div>
      {:else}
        <DataTable
          class="overflow-x-auto rounded-md border"
          data={filteredUsers}
          {columns}
          getRowId={(u) => String(u.id)}
        />
      {/if}
    </CardContent>
  </Card>

  <Dialog bind:open={isOpen}>
    <DialogContent class="sm:max-w-[550px]">
      <DialogHeader>
        <DialogTitle>{isEdit ? 'Edit Data User' : 'Tambah User Baru'}</DialogTitle>
        <DialogDescription>
          {isEdit
            ? 'Perbarui informasi akun pengguna di bawah ini.'
            : 'Silakan lengkapi formulir untuk mendaftarkan akun baru ke sistem.'}
        </DialogDescription>
      </DialogHeader>
      <form onsubmit={handleSubmit} class="space-y-4 py-4">
        <div class="grid grid-cols-2 gap-4">
          <div class="space-y-2">
            <Label for="u-username">Username / NIM</Label><Input
              id="u-username"
              bind:value={formData.username}
              placeholder="202314..."
              required
              disabled={isEdit && !isKoordinator}
            />
          </div>
          <div class="space-y-2">
            <Label>Role</Label>
            <Select
              type="single"
              value={canonRole(formData.role)}
              onValueChange={(val) => (formData.role = val)}
              disabled={currentUser.role === 'asisten'}
            >
              <SelectTrigger class="w-full" aria-label="Role"
                >{roleOptions.find((o) => o.value === canonRole(formData.role))?.label ??
                  canonRole(formData.role)}</SelectTrigger
              >
              <SelectContent>
                {#each roleOptions as opt (opt.value)}<SelectItem
                    value={opt.value}
                    label={opt.label}
                  />{/each}
              </SelectContent>
            </Select>
          </div>
        </div>
        <div class="grid grid-cols-2 gap-4">
          <div class="space-y-2">
            <Label for="u-name">Nama Lengkap</Label>
            <Input id="u-name" bind:value={formData.full_name} required />
          </div>
          <div class="space-y-2">
            <Label for="u-phone">No. WhatsApp / HP</Label>
            <Input
              id="u-phone"
              value={formData.phone_number || ''}
              oninput={(e) => (formData.phone_number = e.currentTarget.value)}
              placeholder="Contoh: 08123456..."
            />
          </div>
        </div>
        <div class="space-y-2">
          <Label for="u-password"
            >Password {#if isEdit}<span class="text-muted-foreground text-[10px] font-normal italic"
                >(Kosongkan jika tidak ingin ganti)</span
              >{/if}</Label
          >
          <div class="relative">
            <Input
              id="u-password"
              type={showPassword ? 'text' : 'password'}
              bind:value={formData.password}
              placeholder={isEdit ? 'Masukkan password baru' : 'Masukkan password'}
              required={!isEdit}
            />
            <button
              type="button"
              aria-label="Tampilkan password"
              onclick={() => (showPassword = !showPassword)}
              class="text-muted-foreground hover:text-foreground absolute top-2.5 right-3"
            >
              {#if showPassword}<EyeOff class="h-4 w-4" />{:else}<Eye class="h-4 w-4" />{/if}
            </button>
          </div>
        </div>
        {#if canonRole(formData.role) === 'praktikan'}
          <div class="bg-muted/50 grid grid-cols-2 gap-4 rounded-md p-3">
            <div class="space-y-2">
              <Label for="u-class">Kelas</Label>
              <Input
                id="u-class"
                value={formData.class_code || ''}
                oninput={(e) => (formData.class_code = e.currentTarget.value)}
                placeholder="Contoh: IF-A-2024"
              />
            </div>
            <div class="space-y-2">
              <Label>Shift</Label>
              <Select
                type="single"
                value={formData.shift || ''}
                onValueChange={(val) => (formData.shift = val)}
              >
                <SelectTrigger class="w-full"
                  >{#if formData.shift}Shift {formData.shift}{:else}<span
                      class="text-muted-foreground">Pilih Shift</span
                    >{/if}</SelectTrigger
                >
                <SelectContent>
                  <SelectItem value="1" label="Shift 1" />
                  <SelectItem value="2" label="Shift 2" />
                </SelectContent>
              </Select>
            </div>
          </div>
        {/if}
        {#if formData.role === 'asisten' || formData.role === 'koordinator'}
          <div
            class="bg-muted/50 grid gap-4 rounded-md p-3 {formData.role === 'asisten'
              ? 'grid-cols-2'
              : 'grid-cols-1'}"
          >
            <div class="space-y-2">
              <Label for="u-division">Divisi</Label>
              <Input
                id="u-division"
                value={formData.division || ''}
                oninput={(e) => (formData.division = e.currentTarget.value)}
                placeholder="Divisi"
              />
            </div>
            {#if formData.role === 'asisten'}
              <div class="space-y-2">
                <Label for="u-code">Kode Asisten</Label>
                <Input
                  id="u-code"
                  value={formData.assistant_code || ''}
                  oninput={(e) => (formData.assistant_code = e.currentTarget.value.toUpperCase())}
                  placeholder="Contoh: SA"
                />
              </div>
            {/if}
          </div>
        {/if}
        <div class="bg-primary/5 flex items-center justify-between rounded-md border p-3">
          <div class="space-y-0.5">
            <Label for="u-active">Status Akun</Label>
            <p class="text-muted-foreground text-xs">
              Aktifkan atau nonaktifkan akses login user ini.
            </p>
          </div>
          <Switch id="u-active" bind:checked={formData.is_active} />
        </div>
        <DialogFooter class="pt-4"
          ><Button type="button" variant="outline" onclick={() => (isOpen = false)}>Batal</Button
          ><Button type="submit" disabled={saving}
            >{#if saving}<Loader2 class="mr-2 h-4 w-4 animate-spin" />{/if} Simpan</Button
          ></DialogFooter
        >
      </form>
    </DialogContent>
  </Dialog>

  <Dialog bind:open={isAccessOpen}>
    <DialogContent class="sm:max-w-[600px]">
      <DialogHeader>
        <DialogTitle class="flex items-center gap-2"
          ><ShieldCheck class="text-primary h-5 w-5" /> Kelola Hak Akses Divisi</DialogTitle
        >
        <DialogDescription
          >Tentukan menu apa saja yang dapat diakses oleh divisi tertentu.</DialogDescription
        >
      </DialogHeader>
      <div class="space-y-4 py-2">
        <div class="space-y-2">
          <Label>Pilih Divisi</Label>
          <Select type="single" bind:value={selectedDivision}>
            <SelectTrigger class="w-full"
              >{#if selectedDivision}{selectedDivision}{:else}<span class="text-muted-foreground"
                  >-- Pilih Divisi --</span
                >{/if}</SelectTrigger
            >
            <SelectContent>
              {#if divisions.length === 0}
                <SelectItem value="empty" label="Belum ada divisi terdaftar" disabled />
              {:else}
                {#each divisions as d (d)}<SelectItem value={d} label={d} />{/each}
              {/if}
            </SelectContent>
          </Select>
        </div>
        <div class="min-h-[200px] rounded-md border p-4">
          {#if !selectedDivision}
            <div
              class="text-muted-foreground flex h-full flex-col items-center justify-center gap-2"
            >
              <Lock class="h-8 w-8 opacity-20" />
              <span class="text-sm">Pilih divisi di atas terlebih dahulu.</span>
            </div>
          {:else if loadingAccess}
            <div class="flex h-full items-center justify-center">
              <Loader2 class="text-primary animate-spin" />
            </div>
          {:else}
            <div class="grid grid-cols-2 gap-3">
              {#each MASTER_MENUS as menu (menu.key)}
                <label
                  class="hover:bg-muted flex cursor-pointer items-center space-x-2 rounded border p-2"
                >
                  <input
                    type="checkbox"
                    class="text-primary focus:ring-primary h-4 w-4 rounded border-gray-300"
                    checked={accessList.includes(menu.key)}
                    onchange={() => toggleMenuAccess(menu.key)}
                  />
                  <span class="flex-1 cursor-pointer text-sm font-medium">{menu.label}</span>
                </label>
              {/each}
            </div>
          {/if}
        </div>
      </div>
      <DialogFooter>
        <Button onclick={handleSaveAccess} disabled={!selectedDivision || loadingAccess}>
          {#if loadingAccess}<Loader2 class="mr-2 animate-spin" />{:else}<Save
              class="mr-2 h-4 w-4"
            />{/if}
          Simpan Perubahan
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</div>
