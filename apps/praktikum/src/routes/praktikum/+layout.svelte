<script lang="ts">
  import { page } from '$app/state';
  import AppShell from '#lib/components/AppShell.svelte';
  import { PUBLIC_PATHS } from '#lib/auth.ts';
  import { Toaster } from '@algohub/ui';

  let { children } = $props();

  let isAuthPage = $derived(PUBLIC_PATHS.includes(page.url.pathname));
</script>

<Toaster />
{#if isAuthPage}
  {@render children()}
{:else}
  <AppShell>
    {@render children()}
  </AppShell>
{/if}
