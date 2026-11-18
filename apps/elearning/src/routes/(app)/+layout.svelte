<script lang="ts">
  import { resolve } from '$app/paths';
  import { page } from '$app/state';
  import ArrowLeft from '@lucide/svelte/icons/arrow-left';
  import BookOpen from '@lucide/svelte/icons/book-open';
  import ClipboardList from '@lucide/svelte/icons/clipboard-list';
  import LogOut from '@lucide/svelte/icons/log-out';
  import Menu from '@lucide/svelte/icons/menu';
  import Shapes from '@lucide/svelte/icons/shapes';
  import ShieldCheck from '@lucide/svelte/icons/shield-check';
  import Star from '@lucide/svelte/icons/star';
  import Terminal from '@lucide/svelte/icons/terminal';
  import Trophy from '@lucide/svelte/icons/trophy';
  import Users from '@lucide/svelte/icons/users';
  import X from '@lucide/svelte/icons/x';
  import { fade, fly } from 'svelte/transition';
  import { onMount } from 'svelte';
  import { startActivity } from '#lib/activity.ts';
  import { signOut } from '#lib/auth.ts';
  import SidebarItem from '#lib/components/SidebarItem.svelte';
  import type { StaticRouteId } from '#lib/paths.ts';
  import { loadProfile } from '#lib/profile.ts';
  import { startRealtimeToken, stopRealtimeToken } from '#lib/realtime/token.ts';
  import { isAdminOrKordas, isStaff } from '#lib/roles.ts';
  import { elearningRunner } from '#lib/runner.ts';
  import { appStore } from '#lib/stores/app.svelte.ts';
  import type { LayoutProps } from './$types';

  let { data, children }: LayoutProps = $props();

  const user = $derived(data.auth.user);
  const profile = $derived(appStore.profile);
  const adminOrKordas = $derived(isAdminOrKordas(user.role));
  const staff = $derived(isStaff(user.role));
  const pathname = $derived(page.url.pathname);
  const isActive = (path: string) => pathname === path || pathname.startsWith(`${path}/`);
  const isMobile = () => window.innerWidth < 1024;

  onMount(() => {
    if (isMobile()) appStore.sidebarOpen = false;
    loadProfile().catch(() => undefined);
    startRealtimeToken();
    const stopActivity = startActivity({ user, pathname: () => page.url.pathname });
    // File compiler diunduh pelan-pelan saat idle supaya lesson pertama tidak menunggu.
    elearningRunner().prefetch();
    return () => {
      stopActivity();
      stopRealtimeToken();
    };
  });

  // Di layar kecil sidebar menutup sendiri setelah pindah halaman.
  $effect(() => {
    void pathname;
    if (isMobile()) appStore.sidebarOpen = false;
  });
</script>

<div class="text-dark flex min-h-screen overflow-hidden bg-[#FDFBF7] font-sans">
  {#if appStore.sidebarOpen}
    <aside
      transition:fly={{ x: -256, duration: 200 }}
      class="bg-sidebar border-maroon/5 fixed inset-y-0 left-0 z-50 flex h-full w-64 flex-col justify-between border-r shadow-[4px_0_24px_rgba(138,21,56,0.05)]"
    >
      <div class="flex min-h-0 flex-1 flex-col">
        <div class="mt-2 flex items-center justify-between p-6">
          <a
            href={resolve('/(app)/dashboard')}
            class="bg-maroon shadow-bubbly-maroon flex h-11 w-11 rotate-[-6deg] items-center justify-center rounded-xl text-2xl font-black text-white transition-transform hover:rotate-0"
            aria-label="E-Learning"
          >
            E
          </a>
          <a
            href={resolve('/(app)/dashboard')}
            class="text-maroon ml-3 flex-1 text-xl font-black tracking-tight select-none"
            >E-Learning</a
          >
          <button
            onclick={() => appStore.toggleSidebar()}
            class="hover:bg-maroon-bg text-maroon cursor-pointer rounded-lg p-2 transition-colors lg:hidden"
            aria-label="Tutup menu"
          >
            <X size={18} />
          </button>
        </div>
        <nav class="custom-scrollbar mt-6 flex-1 space-y-3 overflow-y-auto px-4">
          <SidebarItem
            href="/dashboard"
            label="Beranda"
            icon={Shapes}
            active={isActive('/dashboard')}
          />
          <SidebarItem
            href="/courses"
            label="Belajar"
            icon={BookOpen}
            active={isActive('/courses') || isActive('/lesson')}
          />
          <SidebarItem
            href="/playground"
            label="Playground"
            icon={Terminal}
            active={isActive('/playground')}
            hoverClass="group-hover:text-fun-blue"
          />
          <SidebarItem
            href="/leaderboard"
            label="Papan Peringkat"
            icon={Trophy}
            active={isActive('/leaderboard')}
            hoverClass="group-hover:text-fun-yellow"
          />
          {#if staff}
            <div class="border-maroon/5 space-y-3 border-t pt-4">
              <div class="mb-2 px-4">
                <span
                  class="text-maroon/40 bg-maroon-bg rounded-full px-3 py-1 text-[10px] font-black tracking-widest uppercase"
                  >Panel Staf</span
                >
              </div>
              {#if adminOrKordas}
                <SidebarItem
                  href="/admin"
                  label="Admin Panel"
                  icon={ShieldCheck}
                  active={isActive('/admin')}
                />
              {/if}
              <SidebarItem
                href="/monitoring"
                label="Monitoring"
                icon={Users}
                active={isActive('/monitoring')}
                hoverClass="group-hover:text-fun-blue"
              />
              {#if adminOrKordas}
                <SidebarItem
                  href="/auditlog"
                  label="Audit Log"
                  icon={ClipboardList}
                  active={isActive('/auditlog')}
                  hoverClass="group-hover:text-fun-green"
                />
              {/if}
            </div>
          {/if}
        </nav>
      </div>
      <div class="border-maroon/5 border-t p-6">
        <button
          onclick={() => signOut()}
          class="group flex w-full cursor-pointer items-center space-x-3 px-2 text-left font-black text-gray-400 transition-colors hover:text-red-500"
        >
          <div
            class="flex h-9 w-9 items-center justify-center rounded-full bg-gray-100 transition-all group-hover:bg-red-100 group-hover:text-red-500"
          >
            <LogOut size={16} />
          </div>
          <span class="text-sm font-black tracking-wide">Keluar</span>
        </button>
      </div>
    </aside>
    <button
      transition:fade={{ duration: 200 }}
      onclick={() => appStore.toggleSidebar()}
      class="fixed inset-0 z-40 cursor-pointer bg-black/40 lg:hidden"
      aria-label="Tutup menu"
    ></button>
  {/if}

  <main
    class="relative flex h-screen flex-1 flex-col overflow-hidden transition-all duration-300 {appStore.sidebarOpen
      ? 'lg:ml-64'
      : 'ml-0'}"
  >
    <div
      class="bg-fun-yellow/20 pointer-events-none absolute top-[-50px] right-[-50px] z-0 h-64 w-64 rounded-full blur-3xl"
    ></div>
    <div
      class="bg-maroon/5 pointer-events-none absolute bottom-1/4 left-1/4 z-0 h-96 w-96 rounded-full blur-3xl"
    ></div>

    <header
      class="border-maroon/5 sticky top-0 z-40 flex h-20 items-center border-b bg-[#FDFBF7]/80 px-8 py-4 backdrop-blur-md"
    >
      <div class="mx-auto flex w-full max-w-7xl items-center justify-between">
        <div class="flex items-center space-x-4">
          <button
            onclick={() => appStore.toggleSidebar()}
            class="hover:bg-maroon-bg text-maroon flex h-10 w-10 cursor-pointer items-center justify-center rounded-full transition-colors focus:outline-none"
            aria-label="Buka atau tutup menu"
          >
            <Menu size={18} />
          </button>
          {#if pathname !== '/dashboard'}
            <a
              href={resolve('/(app)/dashboard')}
              class="hover:text-maroon flex items-center space-x-2 rounded-full border border-gray-100 bg-white px-4 py-2 text-sm font-bold text-gray-400 shadow-sm transition-colors"
            >
              <ArrowLeft size={12} />
              <span class="hidden sm:inline">Kembali ke Dashboard</span>
            </a>
          {/if}
        </div>
        <a
          href={resolve('/profile' as StaticRouteId)}
          class="hover:bg-maroon-bg hover:border-maroon/20 group flex items-center space-x-3 rounded-full border-2 border-transparent bg-white p-1.5 pl-5 shadow-sm transition-colors"
        >
          <div class="hidden text-right sm:block">
            <p
              class="text-dark group-hover:text-maroon mb-1 text-sm leading-none font-black uppercase transition-colors"
            >
              {profile?.nama || user.nama || 'User'}
            </p>
            <div class="flex items-center justify-end">
              <span class="bg-maroon mr-2 rounded-md px-2 py-0.5 text-[10px] font-bold text-white">
                LVL {profile?.level || 1}
              </span>
              <p class="flex items-center text-xs font-bold text-gray-500">
                <Star size={12} class="text-fun-yellow fill-fun-yellow mr-1" />
                {profile?.xp || 0} XP
              </p>
            </div>
          </div>
          <div
            class="bg-fun-yellow text-maroon flex h-10 w-10 items-center justify-center rounded-full border-2 border-white text-lg font-black shadow-inner transition-transform select-none group-hover:scale-105"
          >
            {(profile?.nama || user.nama || 'U').charAt(0).toUpperCase()}
          </div>
        </a>
      </div>
    </header>

    <div class="custom-scrollbar relative z-10 flex-1 overflow-y-auto p-8">
      <div class="mx-auto max-w-7xl pb-12">
        {@render children()}
      </div>
    </div>
  </main>
</div>
