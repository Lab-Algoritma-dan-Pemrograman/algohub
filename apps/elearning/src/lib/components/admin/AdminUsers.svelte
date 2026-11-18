<script lang="ts">
  import type { ElearningUser } from '@algohub/api-client/elearning';
  import { Button, DataTable, Input, renderSnippet, toast, type ColumnDef } from '@algohub/ui';
  import ChevronLeft from '@lucide/svelte/icons/chevron-left';
  import ChevronRight from '@lucide/svelte/icons/chevron-right';
  import Search from '@lucide/svelte/icons/search';
  import { createQuery } from '@tanstack/svelte-query';
  import { perform } from '#lib/admin/action.ts';
  import ConfirmDialog from '#lib/components/admin/ConfirmDialog.svelte';
  import {
    buildCurriculum,
    overridesOf,
    type AccessMode,
    type AccessOverrides,
  } from '#lib/curriculum.ts';
  import { levelsQuery, lessonsQuery, modulesQuery } from '#lib/queries/curriculum.ts';
  import {
    ADMIN_PAGE_SIZE,
    adminUserActions,
    adminUserProgressQuery,
    adminUsersQuery,
  } from '#lib/queries/admin.ts';
  import { normalizeRole, type AppRole } from '#lib/roles.ts';
  import type { ElearningSessionUser } from '#lib/session.ts';

  let { currentUser }: { currentUser: ElearningSessionUser } = $props();

  const ROLES: AppRole[] = ['admin', 'kordas', 'asisten', 'praktikan'];
  const ACCESS_LABEL: Record<AccessMode, string> = {
    auto: 'Otomatis',
    unlocked: 'Dibuka',
    locked: 'Dikunci',
  };

  let searchTerm = $state('');
  let search = $state('');
  let page = $state(1);
  let selected = $state<ElearningUser | null>(null);
  let xpValue = $state('');
  let busy = $state(false);
  let confirmOpen = $state(false);
  let confirm = $state({
    title: '',
    message: '',
    label: 'Lanjutkan',
    run: async () => {},
  });

  // Pencarian dikirim ke server (?q=) setelah jeda; mengetik tidak memicu satu request per huruf.
  $effect(() => {
    const term = searchTerm;
    const timer = setTimeout(() => {
      search = term.trim();
      page = 1;
    }, 400);
    return () => clearTimeout(timer);
  });

  const users = createQuery(() => adminUsersQuery(search, page));
  const levels = createQuery(levelsQuery);
  const modules = createQuery(modulesQuery);
  const lessons = createQuery(lessonsQuery);

  const items = $derived(users.data?.items ?? []);
  const total = $derived(users.data?.total ?? 0);
  const pageCount = $derived(Math.max(1, Math.ceil(total / ADMIN_PAGE_SIZE)));
  const curriculum = $derived(
    buildCurriculum(levels.data ?? [], modules.data ?? [], lessons.data ?? []),
  );
  // Data terbaru dari daftar menang atas salinan yang dipilih, supaya XP dan role ikut segar.
  const user = $derived(selected ? (items.find((u) => u.nim === selected?.nim) ?? selected) : null);

  const progress = createQuery(() => ({
    ...adminUserProgressQuery(selected?.nim ?? ''),
    enabled: !!selected,
  }));
  const completedIds = $derived(new Set((progress.data ?? []).map((p) => p.lesson_id)));

  const roleOf = (u: ElearningUser): AppRole =>
    normalizeRole(u.elearning_role === 'admin' ? 'admin' : u.role);
  const overrides = $derived<AccessOverrides>(overridesOf(user?.level_access_overrides));

  function lessonTitle(id: string): string {
    return lessons.data?.find((l) => l.id === id)?.title ?? id;
  }

  function ask(title: string, message: string, run: () => Promise<void>, label = 'Lanjutkan') {
    confirm = { title, message, label, run };
    confirmOpen = true;
  }

  async function act(task: () => Promise<unknown>, ok: string, fail: string) {
    busy = true;
    try {
      return await perform(task, ok, fail);
    } finally {
      busy = false;
    }
  }

  function changeRole(target: ElearningUser, role: AppRole) {
    if (target.nim === currentUser.nim) {
      toast.error('Aksi ditolak: Anda tidak dapat mengubah peran akun Anda sendiri.');
      return;
    }
    ask('Ubah Peran', `Ubah peran ${target.nama} menjadi ${role}?`, async () => {
      await act(
        () => adminUserActions.setRole(target.nim, role),
        'Peran diubah.',
        'Gagal mengubah peran user',
      );
    });
  }

  async function cycleAccess(levelId: string) {
    if (!user) return;
    const current = overrides[levelId] ?? 'auto';
    const next: AccessMode =
      current === 'auto' ? 'unlocked' : current === 'unlocked' ? 'locked' : 'auto';
    await act(
      () => adminUserActions.setLevelAccess(user.nim, { ...overrides, [levelId]: next }),
      'Akses level diperbarui.',
      'Gagal update akses level user',
    );
  }

  function resetAll() {
    if (!user) return;
    const target = user;
    ask(
      'Reset Semua Progress',
      `PERINGATAN: Ini akan menghapus SEMUA progress dan mengeset XP ke 0 untuk ${target.nama}. Tindakan ini TIDAK BISA dibatalkan!`,
      async () => {
        await act(
          () => adminUserActions.resetUser(target.nim),
          `Progress ${target.nama} berhasil direset.`,
          'Gagal mereset progress',
        );
      },
    );
  }

  function resetLevel(levelId: string, levelTitle: string) {
    if (!user) return;
    const target = user;
    ask(
      'Reset Level Progress',
      `Reset progress level "${levelTitle}" untuk ${target.nama}? XP akan dikurangi sesuai jumlah pelajaran yang dihapus.`,
      async () => {
        await act(
          () => adminUserActions.resetLevel(target.nim, levelId),
          `Level "${levelTitle}" berhasil direset.`,
          'Gagal mereset level',
        );
      },
    );
  }

  function resetLesson(lessonId: string) {
    if (!user) return;
    const target = user;
    const title = lessonTitle(lessonId);
    ask(
      'Reset Progress Pelajaran',
      `Reset progress pelajaran "${title}" untuk ${target.nama}? XP akan dikurangi 60.`,
      async () => {
        await act(
          () => adminUserActions.resetLesson(target.nim, lessonId, title),
          `Pelajaran "${title}" berhasil direset.`,
          'Gagal mereset pelajaran',
        );
      },
    );
  }

  function adjustXp() {
    if (!user) return;
    const target = user;
    const value = Number.parseInt(xpValue, 10);
    if (Number.isNaN(value) || value < 0) {
      toast.error('Masukkan angka XP yang valid (>= 0).');
      return;
    }
    ask('Ubah XP', `Ubah XP ${target.nama} dari ${target.xp} menjadi ${value}?`, async () => {
      if (
        await act(
          () => adminUserActions.adjustXp(target.nim, value),
          `XP berhasil diubah menjadi ${value}.`,
          'Gagal mengubah XP',
        )
      ) {
        xpValue = '';
      }
    });
  }

  function removeUser() {
    if (!user) return;
    const target = user;
    if (target.nim === currentUser.nim) {
      toast.error('Aksi ditolak: Anda tidak dapat menghapus akun Anda sendiri.');
      return;
    }
    ask(
      'Hapus Akun Peserta',
      `PERINGATAN: Ini akan menghapus akun ${target.nama} dan SEMUA datanya secara permanen dari sistem. Tindakan ini tidak dapat dibatalkan!`,
      async () => {
        if (
          await act(
            () => adminUserActions.deleteUser(target.nim),
            `Akun ${target.nama} telah berhasil dihapus.`,
            'Gagal menghapus akun',
          )
        ) {
          selected = null;
        }
      },
      'Hapus',
    );
  }

  const columns: ColumnDef<ElearningUser>[] = [
    {
      id: 'peserta',
      header: 'Peserta',
      cell: ({ row }) => renderSnippet(pesertaCell, row.original),
    },
    {
      accessorKey: 'kelas',
      header: 'Kelas',
      meta: { headClass: 'text-center', cellClass: 'text-center' },
    },
    {
      id: 'jurusan',
      header: 'Jurusan',
      cell: ({ row }) => row.original.jurusan ?? '-',
      meta: { headClass: 'text-center', cellClass: 'text-center' },
    },
    {
      id: 'xp',
      header: 'XP',
      cell: ({ row }) => row.original.xp.toLocaleString('id-ID'),
      meta: { headClass: 'text-center', cellClass: 'text-center font-bold text-rose-700' },
    },
    {
      id: 'role',
      header: 'Role',
      cell: ({ row }) => roleOf(row.original),
      meta: { headClass: 'text-center', cellClass: 'text-center text-xs font-bold uppercase' },
    },
  ];
</script>

{#snippet pesertaCell(u: ElearningUser)}
  <button
    type="button"
    class="text-left"
    onclick={() => {
      selected = u;
      xpValue = '';
    }}
  >
    <span class="block font-bold text-zinc-900">{u.nama}</span>
    <span class="text-xs text-zinc-400">{u.nim}</span>
  </button>
{/snippet}

<div class="grid grid-cols-1 gap-8 lg:grid-cols-3">
  <div class="space-y-4 lg:col-span-2">
    <div class="relative">
      <Search class="absolute top-1/2 left-4 -translate-y-1/2 text-zinc-400" size={20} />
      <Input
        type="text"
        placeholder="Cari nama atau email peserta..."
        bind:value={searchTerm}
        class="h-12 rounded-2xl pl-12"
      />
    </div>

    <div class="overflow-hidden rounded-3xl border border-zinc-200 bg-white shadow-sm">
      {#if users.isPending}
        <p class="px-6 py-12 text-center text-zinc-400">Memuat data peserta...</p>
      {:else if users.isError}
        <p class="px-6 py-12 text-center text-red-600">Gagal memuat peserta.</p>
      {:else}
        <DataTable
          data={items}
          {columns}
          getRowId={(u) => u.nim}
          rowClass={(u) => (u.nim === selected?.nim ? 'bg-rose-50/40' : undefined)}
        >
          {#snippet empty()}
            <tr
              ><td colspan={5} class="px-6 py-12 text-center text-zinc-400">Tidak ada peserta.</td
              ></tr
            >
          {/snippet}
        </DataTable>
      {/if}
    </div>

    <div class="flex items-center justify-between">
      <p class="text-xs font-bold text-zinc-400">
        Halaman {page} dari {pageCount} · {total} peserta
      </p>
      <div class="flex gap-2">
        <Button
          variant="outline"
          size="sm"
          class="rounded-xl font-bold"
          disabled={page <= 1}
          onclick={() => (page -= 1)}
        >
          <ChevronLeft class="mr-1 h-4 w-4" /> Sebelumnya
        </Button>
        <Button
          variant="outline"
          size="sm"
          class="rounded-xl font-bold"
          disabled={page >= pageCount}
          onclick={() => (page += 1)}
        >
          Selanjutnya <ChevronRight class="ml-1 h-4 w-4" />
        </Button>
      </div>
    </div>
  </div>

  <aside class="space-y-4">
    {#if user}
      <div class="space-y-5 rounded-3xl border border-zinc-200 bg-white p-6 shadow-sm">
        <div>
          <h3 class="text-lg font-black text-zinc-900">{user.nama}</h3>
          <p class="text-xs text-zinc-500">
            {user.nim} · {user.kelas}{user.jurusan ? ` · ${user.jurusan}` : ''}
          </p>
          <p class="mt-1 text-xs font-bold text-rose-700">
            {user.xp.toLocaleString('id-ID')} XP · Level {user.level} · Streak {user.streak}
          </p>
        </div>

        <label class="block space-y-1">
          <span class="text-xs font-bold text-zinc-400 uppercase">Peran</span>
          <select
            class="w-full rounded-xl border border-zinc-200 bg-zinc-50 px-3 py-2 text-sm"
            value={roleOf(user)}
            disabled={busy}
            onchange={(e) => changeRole(user, e.currentTarget.value as AppRole)}
          >
            {#each ROLES as role (role)}
              <option value={role}>{role}</option>
            {/each}
          </select>
        </label>

        <div class="space-y-2">
          <span class="text-xs font-bold text-zinc-400 uppercase">Atur XP</span>
          <div class="flex gap-2">
            <Input type="number" min="0" placeholder="XP baru" bind:value={xpValue} />
            <Button class="rounded-xl font-bold" disabled={busy || !xpValue} onclick={adjustXp}>
              Simpan
            </Button>
          </div>
        </div>

        <div class="space-y-2">
          <span class="text-xs font-bold text-zinc-400 uppercase">Akses level</span>
          {#each curriculum as level (level.id)}
            <div class="flex items-center justify-between gap-2 text-sm">
              <span class="truncate font-semibold">{level.title}</span>
              <div class="flex shrink-0 gap-1">
                <Button
                  variant="outline"
                  size="sm"
                  class="rounded-lg text-xs"
                  disabled={busy}
                  onclick={() => cycleAccess(level.id)}
                >
                  {ACCESS_LABEL[overrides[level.id] ?? 'auto']}
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  class="rounded-lg text-xs text-amber-700"
                  disabled={busy}
                  onclick={() => resetLevel(level.id, level.title)}
                >
                  Reset
                </Button>
              </div>
            </div>
          {/each}
        </div>

        <div class="space-y-2">
          <span class="text-xs font-bold text-zinc-400 uppercase">
            Pelajaran selesai ({completedIds.size})
          </span>
          {#if progress.isPending}
            <p class="text-xs text-zinc-400">Memuat progress...</p>
          {:else if completedIds.size === 0}
            <p class="text-xs text-zinc-400">Belum ada pelajaran yang selesai.</p>
          {:else}
            <ul class="max-h-56 space-y-1 overflow-y-auto">
              {#each progress.data ?? [] as p (p.id)}
                <li class="flex items-center justify-between gap-2 text-xs">
                  <span class="truncate">{lessonTitle(p.lesson_id)}</span>
                  <button
                    type="button"
                    class="shrink-0 font-bold text-amber-700 hover:underline"
                    disabled={busy}
                    onclick={() => resetLesson(p.lesson_id)}
                  >
                    Reset
                  </button>
                </li>
              {/each}
            </ul>
          {/if}
        </div>

        <div class="flex flex-wrap gap-2 border-t border-zinc-100 pt-4">
          <Button
            variant="outline"
            class="rounded-xl font-bold text-amber-700"
            disabled={busy}
            onclick={resetAll}
          >
            Reset Semua Progress
          </Button>
          <Button
            variant="outline"
            class="rounded-xl font-bold text-red-600"
            disabled={busy}
            onclick={removeUser}
          >
            Hapus Akun
          </Button>
        </div>
      </div>
    {:else}
      <div
        class="rounded-3xl border border-dashed border-zinc-200 p-8 text-center text-sm text-zinc-400"
      >
        Pilih peserta untuk melihat detail dan mengelola progress.
      </div>
    {/if}
  </aside>
</div>

<ConfirmDialog
  bind:open={confirmOpen}
  title={confirm.title}
  message={confirm.message}
  confirmLabel={confirm.label}
  {busy}
  onconfirm={confirm.run}
/>
