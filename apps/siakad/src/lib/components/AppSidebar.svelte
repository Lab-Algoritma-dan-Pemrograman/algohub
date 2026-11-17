<script lang="ts">
  import { goto } from '$app/navigation';
  import { resolve } from '$app/paths';
  import { page } from '$app/state';
  import type { RouteId } from '$app/types';
  import {
    Button,
    Sidebar,
    SidebarContent,
    SidebarFooter,
    SidebarGroup,
    SidebarGroupContent,
    SidebarGroupLabel,
    SidebarMenu,
    SidebarMenuButton,
    SidebarMenuItem,
  } from '@algohub/ui';
  import Box from '@lucide/svelte/icons/box';
  import Calendar from '@lucide/svelte/icons/calendar';
  import CalendarCheck from '@lucide/svelte/icons/calendar-check';
  import CalendarDays from '@lucide/svelte/icons/calendar-days';
  import CheckCircle from '@lucide/svelte/icons/circle-check-big';
  import FileText from '@lucide/svelte/icons/file-text';
  import GraduationCap from '@lucide/svelte/icons/graduation-cap';
  import History from '@lucide/svelte/icons/history';
  import Home from '@lucide/svelte/icons/house';
  import ListChecks from '@lucide/svelte/icons/list-checks';
  import LogOut from '@lucide/svelte/icons/log-out';
  import MessageSquare from '@lucide/svelte/icons/message-square';
  import PackageSearch from '@lucide/svelte/icons/package-search';
  import QrCode from '@lucide/svelte/icons/qr-code';
  import Settings from '@lucide/svelte/icons/settings';
  import User from '@lucide/svelte/icons/user';
  import Users from '@lucide/svelte/icons/users';
  import { createQuery } from '@tanstack/svelte-query';
  import { tick } from 'svelte';
  import type { AuthState } from '#lib/auth.ts';
  import { signOut } from '#lib/auth.ts';
  import { rpcQueryOptions } from '#lib/queries/rpc.ts';
  import { ALWAYS_ALLOWED_MENUS, canonRole, roleAllowed } from '#lib/roles.ts';

  let { auth }: { auth: AuthState } = $props();
  const user = $derived(auth.user);
  const allowedPaths = $derived(auth.allowedPaths);

  const RESTRICTED_MENUS = [
    '/manajemen-user',
    '/manajemen-jadwal',
    '/manajemen-kelas',
    '/validasi-absensi',
    '/inventaris',
    '/laporan-keuangan',
    '/e-learning',
    '/ketersediaan',
    '/manajemen-sewa',
    '/audit-log',
  ];

  const menuItems = [
    {
      title: 'Beranda',
      url: '/beranda',
      icon: Home,
      roles: ['praktikan', 'asisten', 'koordinator', 'penyewa'],
    },
    {
      title: 'Profil Saya',
      url: '/profil',
      icon: User,
      roles: ['praktikan', 'asisten', 'koordinator', 'penyewa'],
    },
    { title: 'Sewa & Pinjam', url: '/sewa-barang', icon: PackageSearch, roles: ['penyewa'] },
    {
      title: 'Manajemen Sewa',
      url: '/manajemen-sewa',
      icon: ListChecks,
      roles: ['asisten', 'koordinator'],
    },
    {
      title: 'Absensi',
      url: '/absensi',
      icon: CalendarCheck,
      roles: ['praktikan', 'asisten', 'koordinator'],
    },
    { title: 'Buat QR Absen', url: '/buat-qr', icon: QrCode, roles: ['asisten', 'koordinator'] },
    {
      title: 'Kritik & Saran',
      url: '/kritik-saran',
      icon: MessageSquare,
      roles: ['praktikan', 'koordinator'],
    },
    {
      title: 'E-Learning',
      url: '/e-learning',
      icon: GraduationCap,
      roles: ['praktikan', 'asisten', 'koordinator'],
    },
    {
      title: 'Jadwal Saya',
      url: '/jadwal-saya',
      icon: Calendar,
      roles: ['praktikan', 'asisten', 'koordinator'],
    },
    {
      title: 'Manajemen User',
      url: '/manajemen-user',
      icon: Users,
      roles: ['koordinator', 'asisten'],
    },
    {
      title: 'Manajemen Jadwal',
      url: '/manajemen-jadwal',
      icon: CalendarDays,
      roles: ['asisten', 'koordinator'],
    },
    {
      title: 'Validasi Absensi',
      url: '/validasi-absensi',
      icon: CheckCircle,
      roles: ['asisten', 'koordinator'],
    },
    {
      title: 'Jadwal Jaga',
      url: '/jadwal-jaga',
      icon: Calendar,
      roles: ['asisten', 'koordinator'],
    },
    { title: 'Inventaris', url: '/inventaris', icon: Box, roles: ['asisten', 'koordinator'] },
    {
      title: 'Pilih Asisten',
      url: '/manajemen-kelas',
      icon: Users,
      roles: ['koordinator', 'asisten'],
    },
    {
      title: 'Input Jadwal Free',
      url: '/ketersediaan',
      icon: CalendarCheck,
      roles: ['koordinator', 'asisten'],
    },
    {
      title: 'Laporan Keuangan',
      url: '/laporan-keuangan',
      icon: FileText,
      roles: ['koordinator', 'asisten'],
    },
    { title: 'Audit Log', url: '/audit-log', icon: History, roles: ['koordinator', 'asisten'] },
    { title: 'Pengaturan', url: '/pengaturan', icon: Settings, roles: ['koordinator'] },
  ] as const;

  const pjQuery = createQuery(() =>
    rpcQueryOptions<boolean>(
      'check_pj_absen_today',
      { p_user_id: user.id },
      { enabled: user.role === 'asisten' },
    ),
  );
  $effect(() => {
    if (pjQuery.error) console.error('Gagal load PJ status:', pjQuery.error);
  });
  const isPJAbsenToday = $derived(!!pjQuery.data);

  const filteredMenu = $derived(
    menuItems.filter((item) => {
      const r = canonRole(user.role);
      if (r === 'admin' || r === 'koordinator') return true;
      if (r === 'praktikan' || r === 'penyewa') return roleAllowed(r, item.roles);
      if (r === 'asisten') {
        if (!roleAllowed(r, item.roles)) return false;
        if (ALWAYS_ALLOWED_MENUS.includes(item.url)) return true;
        if (item.url === '/absensi' && isPJAbsenToday) return true;
        if (RESTRICTED_MENUS.includes(item.url)) return allowedPaths.includes(item.url);
        return true;
      }
      return false;
    }),
  );

  const showTaskBadge = (url: string) =>
    url === '/absensi' && isPJAbsenToday && !allowedPaths.includes('/absensi');

  let scrollArea = $state<HTMLElement | null>(null);

  // Posisi scroll sidebar dipulihkan setiap pindah halaman.
  $effect(() => {
    void page.url.pathname;
    void tick().then(() => {
      const saved = sessionStorage.getItem('sidebarScrollPos');
      if (scrollArea && saved) scrollArea.scrollTop = parseInt(saved, 10);
    });
  });
</script>

<Sidebar
  class="text-dark border-r border-gray-100 bg-white shadow-[4px_0_24px_rgba(138,21,56,0.05)]"
>
  <SidebarContent
    bind:ref={scrollArea}
    id="sidebar-scroll-area"
    onscroll={(e) =>
      sessionStorage.setItem('sidebarScrollPos', e.currentTarget.scrollTop.toString())}
    class="overflow-y-auto"
  >
    <div class="mt-2 flex shrink-0 items-center space-x-3 p-6 pb-4">
      <img src="/logo.png" alt="Logo Lab AP" class="h-10 w-10 -rotate-6 transform object-contain" />
      <span class="text-dark text-2xl font-extrabold tracking-tight">Lab AP</span>
    </div>
    <SidebarGroup>
      <SidebarGroupLabel
        class="mb-2 px-6 text-xs font-black tracking-widest text-gray-400 uppercase"
      >
        Menu Navigasi
      </SidebarGroupLabel>
      <SidebarGroupContent class="px-3">
        <nav class="w-full" aria-label="Menu Utama">
          <SidebarMenu class="space-y-1">
            {#each filteredMenu as item (item.title)}
              {@const isActive = page.url.pathname === item.url}
              <SidebarMenuItem>
                <SidebarMenuButton
                  {isActive}
                  onclick={() => goto(resolve(`/(app)${item.url}` as RouteId))}
                  class="h-auto cursor-pointer bg-transparent p-0 hover:bg-transparent"
                >
                  {#if isActive}
                    <div
                      class="bg-maroon shadow-bubbly-maroon flex w-full transform items-center space-x-3 rounded-2xl px-4 py-3 font-black text-white transition-transform hover:-translate-y-0.5"
                    >
                      <item.icon class="h-5 w-5 text-center" />
                      <span class="text-sm font-black">{item.title}</span>
                      {#if showTaskBadge(item.url)}
                        <span
                          class="ml-auto rounded-full border border-white/30 bg-white/20 px-2 py-0.5 text-[10px] font-bold text-white"
                        >
                          TUGAS
                        </span>
                      {/if}
                    </div>
                  {:else}
                    <div
                      class="hover:bg-maroon-bg hover:text-maroon group flex w-full items-center space-x-3 rounded-xl px-4 py-2.5 text-sm font-bold text-gray-500 transition-colors"
                    >
                      <item.icon
                        class="group-hover:text-maroon h-5 w-5 text-center text-gray-400 transition-colors"
                      />
                      <span class="text-sm font-bold">{item.title}</span>
                      {#if showTaskBadge(item.url)}
                        <span
                          class="ml-auto rounded-full border border-red-500/30 bg-red-500/20 px-2 py-0.5 text-[10px] font-bold text-red-500"
                        >
                          TUGAS
                        </span>
                      {/if}
                    </div>
                  {/if}
                </SidebarMenuButton>
              </SidebarMenuItem>
            {/each}
          </SidebarMenu>
        </nav>
      </SidebarGroupContent>
    </SidebarGroup>
  </SidebarContent>

  <SidebarFooter class="border-t border-gray-100 p-4">
    <button
      type="button"
      class="bg-dark shadow-bubbly-dark group relative flex cursor-pointer flex-col gap-3 overflow-hidden rounded-2xl p-4 text-left"
      onclick={() => goto(resolve('/(app)/profil'))}
    >
      <div
        class="absolute -right-4 -bottom-4 h-16 w-16 rounded-full bg-white/5 transition-transform duration-500 group-hover:scale-150"
      ></div>
      <div class="relative z-10 flex items-center space-x-3">
        <div
          class="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-xl border border-white/20 bg-white/10 text-white"
        >
          <User class="h-5 w-5" />
        </div>
        <div class="overflow-hidden">
          <p class="truncate text-sm leading-tight font-black tracking-tight text-white uppercase">
            {user.full_name || 'Pengguna'}
          </p>
          <p class="mt-0.5 text-[10px] leading-none font-bold text-gray-400 capitalize">
            {user.role || 'Guest'}
          </p>
        </div>
      </div>
      {#if user.division}
        <div class="relative z-10 rounded-lg bg-red-100 py-1.5 text-center shadow-sm">
          <p class="text-maroon text-[10px] font-black tracking-widest uppercase">
            DIVISI {user.division.toUpperCase()}
          </p>
        </div>
      {/if}
    </button>
    <Button
      variant="ghost"
      onclick={signOut}
      class="text-maroon mt-4 flex h-auto w-full items-center justify-center space-x-2 rounded-xl border border-red-100 bg-red-50 py-2.5 text-sm font-bold transition-colors hover:border-red-200 hover:bg-red-100 hover:text-red-600"
    >
      <LogOut class="mr-1 h-4 w-4" />
      Keluar Aplikasi
    </Button>
  </SidebarFooter>
</Sidebar>
