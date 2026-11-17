<script lang="ts">
  import { page } from '$app/state';
  import AppShell from '#lib/components/AppShell.svelte';
  import { PUBLIC_PATHS } from '#lib/auth.ts';

  let { children } = $props();

  let isAuthPage = $derived(PUBLIC_PATHS.includes(page.url.pathname));
</script>

{#if isAuthPage}
  {@render children()}
{:else}
  <AppShell>
    {@render children()}
  </AppShell>
{/if}
