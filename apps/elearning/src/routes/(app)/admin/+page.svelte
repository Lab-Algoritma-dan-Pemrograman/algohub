<script lang="ts">
  import { Button } from '@algohub/ui';
  import AdminAi from '#lib/components/admin/AdminAi.svelte';
  import AdminGames from '#lib/components/admin/AdminGames.svelte';
  import AdminPlayground from '#lib/components/admin/AdminPlayground.svelte';
  import AdminStructure from '#lib/components/admin/AdminStructure.svelte';
  import AdminUsers from '#lib/components/admin/AdminUsers.svelte';
  import type { PageProps } from './$types';

  type Tab = 'users' | 'structure' | 'games' | 'curriculum';
  const TABS: [Tab, string][] = [
    ['users', 'Peserta'],
    ['structure', 'Struktur'],
    ['games', 'Game & Soal'],
    ['curriculum', 'Kurikulum AI'],
  ];

  let { data }: PageProps = $props();
  let tab = $state<Tab>('users');
</script>

<div class="space-y-8">
  <div class="flex flex-col justify-between gap-4 md:flex-row md:items-center">
    <div>
      <h1 class="text-3xl font-bold tracking-tight">Admin Dashboard</h1>
      <p class="mt-1 text-zinc-500">Kelola peserta, kurikulum, dan struktur kursus.</p>
    </div>
    <div class="flex items-center gap-1 rounded-2xl bg-zinc-100 p-1" role="tablist">
      {#each TABS as [key, label] (key)}
        <Button
          variant="ghost"
          role="tab"
          aria-selected={tab === key}
          class="rounded-xl px-5 font-bold {tab === key
            ? 'bg-white text-zinc-900 shadow-sm'
            : 'text-zinc-500'}"
          onclick={() => (tab = key)}
        >
          {label}
        </Button>
      {/each}
    </div>
  </div>

  {#if tab === 'users'}
    <AdminUsers currentUser={data.auth.user} />
  {:else if tab === 'structure'}
    <AdminStructure />
  {:else if tab === 'games'}
    <div class="space-y-8">
      <AdminGames />
      <AdminPlayground />
    </div>
  {:else}
    <AdminAi />
  {/if}
</div>
