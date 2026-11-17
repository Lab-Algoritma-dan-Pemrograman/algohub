<script lang="ts">
  import { Separator, SidebarInset, SidebarProvider, SidebarTrigger } from '@algohub/ui';
  import { signOut } from '#lib/auth.ts';
  import AppSidebar from '#lib/components/AppSidebar.svelte';
  import { useShiftNotifications } from '#lib/shift-notifications.svelte.ts';
  import type { LayoutProps } from './$types';

  let { data, children }: LayoutProps = $props();

  useShiftNotifications(() => data.auth.user);

  const INACTIVITY_LIMIT = 15 * 60 * 1000;
  const ACTIVITY_EVENTS = ['mousemove', 'keydown', 'click', 'scroll', 'touchstart'];

  $effect(() => {
    let timeoutId: ReturnType<typeof setTimeout> | undefined;
    const resetTimer = () => {
      clearTimeout(timeoutId);
      timeoutId = setTimeout(() => void signOut(), INACTIVITY_LIMIT);
    };
    for (const event of ACTIVITY_EVENTS) window.addEventListener(event, resetTimer);
    resetTimer();
    return () => {
      clearTimeout(timeoutId);
      for (const event of ACTIVITY_EVENTS) window.removeEventListener(event, resetTimer);
    };
  });
</script>

<SidebarProvider>
  <AppSidebar auth={data.auth} />
  <SidebarInset class="flex h-screen flex-col overflow-hidden bg-transparent">
    <header
      class="sticky top-0 z-10 flex h-16 shrink-0 items-center gap-2 border-b bg-white px-4 shadow-sm transition-[width,height] ease-linear group-has-data-[collapsible=icon]/sidebar-wrapper:h-12"
    >
      <div class="flex items-center gap-2 px-4">
        <SidebarTrigger class="-ml-1" />
        <Separator orientation="vertical" class="mr-2 h-4" />
        <span class="text-primary max-w-[200px] truncate font-semibold sm:max-w-none">
          <span class="hidden sm:inline">Laboratorium</span> Algoritma Pemrograman
        </span>
      </div>
    </header>
    <main class="flex-1 overflow-y-auto p-4 md:p-6">
      {@render children()}
    </main>
  </SidebarInset>
</SidebarProvider>
