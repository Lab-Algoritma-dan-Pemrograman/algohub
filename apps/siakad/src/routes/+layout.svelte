<script lang="ts">
  import '../app.css';
  import { PUBLIC_APP_VERSION } from '$app/env/public';
  import { persistOptions } from '@algohub/query';
  import { Toaster, TooltipProvider } from '@algohub/ui';
  import { QueryClientProvider } from '@tanstack/svelte-query';
  import { PersistQueryClientProvider } from '@tanstack/svelte-query-persist-client';
  import ErrorFallback from '#lib/components/ErrorFallback.svelte';
  import GlobalConfirm from '#lib/components/GlobalConfirm.svelte';
  import { persisterFor, queryClient } from '#lib/query.ts';
  import { session } from '#lib/session.ts';

  let { children } = $props();

  // Hanya id yang dipantau: objek user baru dengan id sama tidak boleh memicu restore ulang
  // yang menimpa cache memori dengan isi IndexedDB yang lebih lama.
  const userId = $derived(session.user?.id);
  const options = $derived(
    userId === undefined ? null : persistOptions(persisterFor(userId), PUBLIC_APP_VERSION),
  );
</script>

<svelte:boundary onerror={(error) => console.error('ErrorBoundary caught an error:', error)}>
  <TooltipProvider>
    <Toaster />
    {#key userId}
      {#if options}
        <PersistQueryClientProvider client={queryClient} persistOptions={options}>
          {@render children()}
        </PersistQueryClientProvider>
      {:else}
        <QueryClientProvider client={queryClient}>
          {@render children()}
        </QueryClientProvider>
      {/if}
    {/key}
    <GlobalConfirm />
  </TooltipProvider>
  {#snippet failed(error)}
    <ErrorFallback {error} />
  {/snippet}
</svelte:boundary>
