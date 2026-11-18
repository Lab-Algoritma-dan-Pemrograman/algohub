<script lang="ts">
  import { QueryClientProvider } from '@tanstack/svelte-query';
  import type { Component } from 'svelte';
  import { queryClient } from '../src/lib/query';
  import type { ElearningSessionUser } from '../src/lib/session';

  type PageData = { auth: { user: ElearningSessionUser } };
  // Halaman memakai createQuery, jadi harus dirender di dalam provider seperti di root layout.
  let {
    page: Page,
    data,
  }: { page: Component<{ data: PageData; params: Record<string, never> }>; data: PageData } =
    $props();
</script>

<QueryClientProvider client={queryClient}>
  <Page {data} params={{}} />
</QueryClientProvider>
