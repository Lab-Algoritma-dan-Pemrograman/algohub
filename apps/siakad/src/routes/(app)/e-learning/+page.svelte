<script lang="ts">
  import { PUBLIC_ELEARNING_URL } from '$app/env/public';
  import { Badge, Button, Card, CardContent, CardHeader, CardTitle } from '@algohub/ui';
  import AlertCircle from '@lucide/svelte/icons/circle-alert';
  import BookOpen from '@lucide/svelte/icons/book-open';
  import CheckCircle2 from '@lucide/svelte/icons/circle-check';
  import Clock from '@lucide/svelte/icons/clock';
  import ExternalLink from '@lucide/svelte/icons/external-link';
  import GraduationCap from '@lucide/svelte/icons/graduation-cap';
  import Loader2 from '@lucide/svelte/icons/loader-2';
  import RefreshCw from '@lucide/svelte/icons/refresh-cw';
  import TrendingUp from '@lucide/svelte/icons/trending-up';
  import Trophy from '@lucide/svelte/icons/trophy';
  import { createQuery } from '@tanstack/svelte-query';
  import { generateJwt } from '#lib/api.ts';
  import { elearningProgressQuery } from '#lib/queries/elearning.ts';
  import { fetchRpc } from '#lib/queries/rpc.ts';
  import { subscribe } from '#lib/realtime/client.ts';
  import { toast } from '#lib/toast.ts';
  import type { PageProps } from './$types';

  // Env lama pernah berisi deployment *.vercel.app yang sudah mati (DEPLOYMENT_NOT_FOUND)
  // sehingga SSO gagal; URL seperti itu diabaikan dan diganti domain resmi.
  const RAW_ELEARNING_URL = PUBLIC_ELEARNING_URL || '';
  const ELEARNING_URL = /vercel\.app\/?$/.test(RAW_ELEARNING_URL.trim())
    ? 'https://elearning.algohub.web.id/'
    : RAW_ELEARNING_URL.trim() || 'https://elearning.algohub.web.id/';

  let { data }: PageProps = $props();
  const user = $derived(data.auth.user);
  const userNim = $derived(user.nim || user.username || '');

  const progressQuery = createQuery(() => elearningProgressQuery(userNim));
  const progress = $derived(progressQuery.data ?? null);
  const loadingProgress = $derived(progressQuery.isPending);
  const lastSyncTime = $derived(
    progressQuery.dataUpdatedAt ? new Date(progressQuery.dataUpdatedAt) : null,
  );

  $effect(() => {
    if (progressQuery.error) console.error('Gagal load progress E-Learning:', progressQuery.error);
  });

  $effect(() => {
    if (!userNim) return;
    return subscribe(
      'elearning-student-progress-sync',
      { table: 'elearning_progress', filter: `nim=eq.${userNim}` },
      () => void progressQuery.refetch(),
    );
  });

  let isLoading = $state(false);
  let isRefreshing = $state(false);

  async function handleRefresh() {
    isRefreshing = true;
    const result = await progressQuery.refetch();
    if (result.error) {
      toast.error('Gagal memuat data progress', { description: result.error.message });
    } else {
      toast.success('Data progres berhasil disinkronkan!');
    }
    setTimeout(() => (isRefreshing = false), 600);
  }

  // Secret JWT tidak pernah sampai ke browser: handshake dari DB ditukar token di backend.
  async function requestJWT(userId: number): Promise<string> {
    const handshakeCode = await fetchRpc<string | null>('get_elearning_handshake_secure', {
      p_viewer_id: userId,
    }).catch(() => null);
    if (!handshakeCode) throw new Error('Gagal melakukan verifikasi keamanan. Harap coba lagi.');
    const res = await generateJwt<{ token?: string; error?: string }>({
      handshake_code: handshakeCode,
    });
    if (!res.ok) {
      const err = res.body ?? { error: 'Unknown error' };
      throw new Error(err.error || `Server error: ${res.status}`);
    }
    return res.body?.token as string;
  }

  // Elearning eksternal mengharapkan ?token=...; token aman karena handshake sekali pakai
  // dan kedaluwarsa 2 jam.
  async function handleOpenELearning() {
    isLoading = true;
    try {
      const token = await requestJWT(user.id);
      window.open(`${ELEARNING_URL}?token=${token}`, '_blank');
      toast.success('E-Learning dibuka di tab baru!', { description: 'Login otomatis via SSO.' });
    } catch (err) {
      console.error('Gagal generate token:', err);
      toast.error('Gagal membuka E-Learning', { description: (err as Error).message });
    } finally {
      isLoading = false;
    }
  }

  const completionPercent = $derived(progress?.completion_percentage ?? 0);
  const completedLessons = $derived(progress?.completed_lessons ?? 0);
  const totalLessons = $derived(progress?.total_lessons ?? 0);
  const levels = $derived(progress?.completed_levels ?? []);
  const hasCompletedLevel = $derived(levels.length > 0);

  function formatRelativeTime(date: Date | null) {
    if (!date) return 'Belum pernah';
    const diff = Math.floor((Date.now() - date.getTime()) / 1000);
    if (diff < 60) return 'Baru saja';
    if (diff < 3600) return `${Math.floor(diff / 60)} menit lalu`;
    if (diff < 86400) return `${Math.floor(diff / 3600)} jam lalu`;
    return date.toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric' });
  }

  function formatLastAccessed(dateStr: string | null | undefined) {
    if (!dateStr) return 'Belum pernah diakses';
    return new Date(dateStr).toLocaleDateString('id-ID', {
      day: 'numeric',
      month: 'long',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  }
</script>

{#snippet loadingRow(iconClass: string)}
  <div class="flex items-center gap-2">
    <Loader2 class="h-5 w-5 animate-spin {iconClass}" />
    <span class="text-muted-foreground text-sm">Memuat...</span>
  </div>
{/snippet}

<div class="space-y-6">
  <div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
    <div>
      <h1 class="text-foreground flex items-center gap-2.5 text-2xl font-bold">
        <div
          class="rounded-xl bg-linear-to-br from-red-600 to-red-700 p-2 shadow-lg shadow-red-500/20"
        >
          <GraduationCap class="h-6 w-6 text-white" />
        </div>
        E-Learning Praktikum
      </h1>
      <p class="text-muted-foreground mt-1.5 text-sm">
        Selesaikan pembelajaran mandiri sebelum mengikuti praktikum.
      </p>
    </div>
    <div class="flex items-center gap-2">
      <div
        class="text-muted-foreground bg-muted/50 hidden items-center gap-1.5 rounded-full px-3 py-1.5 text-xs sm:flex"
      >
        <div
          class="h-2 w-2 rounded-full {progress ? 'animate-pulse bg-green-500' : 'bg-gray-400'}"
        ></div>
        <span
          >{lastSyncTime ? `Sync: ${formatRelativeTime(lastSyncTime)}` : 'Belum tersinkron'}</span
        >
      </div>
      <Button
        variant="outline"
        size="sm"
        onclick={handleRefresh}
        disabled={isRefreshing}
        class="gap-1.5 rounded-full"
      >
        <RefreshCw class="h-3.5 w-3.5 {isRefreshing ? 'animate-spin' : ''}" />
        Refresh
      </Button>
    </div>
  </div>

  <div class="grid gap-4 md:grid-cols-3">
    <Card class="relative overflow-hidden border-0 bg-linear-to-br from-blue-50 to-white shadow-sm">
      <div
        class="absolute top-0 right-0 h-24 w-24 translate-x-1/2 -translate-y-1/2 rounded-full bg-blue-100/50"
      ></div>
      <CardHeader class="relative flex flex-row items-center justify-between pb-2">
        <CardTitle class="text-muted-foreground text-sm font-medium">Progress Materi</CardTitle>
        <div class="rounded-lg bg-blue-100 p-2"><BookOpen class="h-4 w-4 text-blue-600" /></div>
      </CardHeader>
      <CardContent class="relative">
        {#if loadingProgress}
          {@render loadingRow('text-blue-500')}
        {:else}
          <div class="flex items-end gap-1" data-testid="elearning-percent">
            <span class="text-3xl font-bold text-blue-700">{Math.round(completionPercent)}</span>
            <span class="mb-0.5 text-lg font-semibold text-blue-500/70">%</span>
          </div>
          <div class="mt-3 h-2.5 w-full overflow-hidden rounded-full bg-blue-100">
            <div
              class="h-2.5 rounded-full bg-linear-to-r from-blue-500 to-blue-600 transition-all duration-700 ease-out"
              style="width: {Math.min(completionPercent, 100)}%"
            ></div>
          </div>
          <p class="text-muted-foreground mt-2 flex items-center gap-1 text-xs">
            <BookOpen class="h-3 w-3" />
            {completedLessons} / {totalLessons || '?'} materi diselesaikan
          </p>
        {/if}
      </CardContent>
    </Card>

    <Card
      class="relative overflow-hidden border-0 shadow-sm {hasCompletedLevel
        ? 'bg-linear-to-br from-green-50 to-white'
        : 'bg-linear-to-br from-orange-50 to-white'}"
    >
      <div
        class="absolute top-0 right-0 h-24 w-24 translate-x-1/2 -translate-y-1/2 rounded-full {hasCompletedLevel
          ? 'bg-green-100/50'
          : 'bg-orange-100/50'}"
      ></div>
      <CardHeader class="relative flex flex-row items-center justify-between pb-2">
        <CardTitle class="text-muted-foreground text-sm font-medium">Status Kurikulum</CardTitle>
        <div class="rounded-lg p-2 {hasCompletedLevel ? 'bg-green-100' : 'bg-orange-100'}">
          {#if hasCompletedLevel}
            <CheckCircle2 class="h-4 w-4 text-green-600" />
          {:else}
            <Clock class="h-4 w-4 text-orange-600" />
          {/if}
        </div>
      </CardHeader>
      <CardContent class="relative">
        {#if loadingProgress}
          {@render loadingRow('text-muted-foreground')}
        {:else}
          <div class="mb-2 flex items-end gap-1">
            <span
              class="text-2xl font-bold {hasCompletedLevel ? 'text-green-700' : 'text-orange-700'}"
              >{levels.length}</span
            >
            <span
              class="mb-0.5 text-sm font-medium {hasCompletedLevel
                ? 'text-green-500/70'
                : 'text-orange-500/70'}">level selesai</span
            >
          </div>
          {#if hasCompletedLevel}
            <div class="flex flex-col gap-1">
              {#each levels as lvl, i (i)}
                <div class="flex items-center gap-1.5 text-xs text-green-700">
                  <CheckCircle2 class="h-3 w-3 flex-shrink-0" />
                  <span class="truncate">{lvl}</span>
                </div>
              {/each}
            </div>
          {:else}
            <p class="text-muted-foreground flex items-center gap-1 text-xs">
              <TrendingUp class="h-3 w-3 text-orange-500" />
              Selesaikan level untuk unlock absensi.
            </p>
          {/if}
        {/if}
      </CardContent>
    </Card>

    <Card
      class="relative overflow-hidden border-0 bg-linear-to-br from-purple-50 to-white shadow-sm"
    >
      <div
        class="absolute top-0 right-0 h-24 w-24 translate-x-1/2 -translate-y-1/2 rounded-full bg-purple-100/50"
      ></div>
      <CardHeader class="relative flex flex-row items-center justify-between pb-2">
        <CardTitle class="text-muted-foreground text-sm font-medium">Level Saat Ini</CardTitle>
        <div class="rounded-lg bg-purple-100 p-2"><Trophy class="h-4 w-4 text-purple-600" /></div>
      </CardHeader>
      <CardContent class="relative">
        {#if loadingProgress}
          {@render loadingRow('text-purple-500')}
        {:else}
          <div class="truncate text-base font-bold text-purple-700">
            {progress?.current_level || 'Belum mulai'}
          </div>
          {#if levels.length > 0}
            <div class="mt-2 flex flex-wrap gap-1">
              {#each levels as lvl, i (i)}
                <Badge
                  variant="outline"
                  class="border-green-200 bg-green-50 text-[10px] text-green-700"
                >
                  <CheckCircle2 class="mr-0.5 h-2.5 w-2.5" />{lvl}
                </Badge>
              {/each}
            </div>
          {/if}
          <p class="text-muted-foreground mt-2 flex items-center gap-1 text-xs">
            <Clock class="h-3 w-3" />
            {progress?.last_accessed_at
              ? `Terakhir: ${formatLastAccessed(progress.last_accessed_at)}`
              : 'Belum pernah mengakses'}
          </p>
        {/if}
      </CardContent>
    </Card>
  </div>

  <Card class="border-0 bg-linear-to-r from-red-50 via-white to-red-50/50 shadow-sm">
    <CardContent class="pt-6">
      <div class="flex flex-col items-start justify-between gap-4 sm:flex-row sm:items-center">
        <div class="flex items-start gap-4">
          <div
            class="rounded-2xl bg-linear-to-br from-red-600 to-red-700 p-3.5 shadow-lg shadow-red-500/25"
          >
            <GraduationCap class="h-8 w-8 text-white" />
          </div>
          <div>
            <h3 class="text-lg font-semibold">Buka Platform E-Learning</h3>
            <p class="text-muted-foreground mt-1 text-sm">
              Akses materi, kerjakan quiz, dan pantau progress belajar Anda.
            </p>
            {#if progress?.last_accessed_at}
              <p class="text-muted-foreground mt-1.5 flex items-center gap-1 text-xs">
                <Clock class="h-3 w-3" />
                Terakhir diakses: {formatLastAccessed(progress.last_accessed_at)}
              </p>
            {/if}
          </div>
        </div>
        <Button
          onclick={handleOpenELearning}
          disabled={isLoading}
          size="lg"
          class="min-w-[200px] bg-linear-to-r from-red-600 to-red-700 shadow-lg shadow-red-500/20 transition-all duration-300 hover:from-red-700 hover:to-red-800 hover:shadow-red-500/30"
        >
          {#if isLoading}
            <Loader2 class="mr-2 h-4 w-4 animate-spin" />Menyiapkan...
          {:else}
            <ExternalLink class="mr-2 h-4 w-4" />Buka E-Learning
          {/if}
        </Button>
      </div>
    </CardContent>
  </Card>

  <Card class="bg-muted/30 border-0 shadow-sm">
    <CardContent class="pt-6">
      <h4 class="mb-3 flex items-center gap-2 text-sm font-medium">
        <AlertCircle class="text-muted-foreground h-4 w-4" />
        Informasi Penting
      </h4>
      <ul class="text-muted-foreground space-y-2.5 text-sm">
        <li class="flex items-start gap-2">
          <span class="mt-0.5 text-red-500">•</span>
          <span
            >Pembelajaran <strong>mandiri</strong> wajib diselesaikan sebelum mengikuti praktikum</span
          >
        </li>
        <li class="flex items-start gap-2">
          <span class="mt-0.5 text-green-500">•</span>
          <span>Selesaikan setiap level yang terbuka</span>
        </li>
        <li class="flex items-start gap-2">
          <span class="mt-0.5 text-blue-500">•</span>
          <span>Jika mengalami kendala, hubungi asisten atau koordinator asisten</span>
        </li>
      </ul>
    </CardContent>
  </Card>
</div>
