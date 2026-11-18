<script lang="ts">
  import Loader from '@lucide/svelte/icons/loader-circle';
  import { createQuery } from '@tanstack/svelte-query';
  import LessonView from '#lib/components/LessonView.svelte';
  import { buildCurriculum, overridesOf } from '#lib/curriculum.ts';
  import { locateLesson } from '#lib/lesson.ts';
  import { levelsQuery, lessonsQuery, modulesQuery } from '#lib/queries/curriculum.ts';
  import { completedLessonsQuery } from '#lib/queries/progress.ts';
  import { appStore } from '#lib/stores/app.svelte.ts';
  import type { PageProps } from './$types';

  let { params }: PageProps = $props();

  const levels = createQuery(levelsQuery);
  const modules = createQuery(modulesQuery);
  const lessons = createQuery(lessonsQuery);
  const completedQuery = createQuery(completedLessonsQuery);

  const curriculum = $derived(
    buildCurriculum(levels.data ?? [], modules.data ?? [], lessons.data ?? []),
  );
  const completed = $derived(new Set(completedQuery.data ?? []));
  const overrides = $derived(overridesOf(appStore.profile?.levelAccessOverrides));
  const pos = $derived(locateLesson(curriculum, params.id));
</script>

<!-- Daftar selesai harus sudah ada: lesson yang sudah selesai dibuka dalam mode tinjau, bukan
     dilanjutkan dari langkah tersimpan. Id yang tidak ada di kurikulum menampilkan spinner,
     sama seperti halaman lesson lama. -->
{#if pos && !completedQuery.isPending}
  {#key pos.lesson.id}
    <LessonView {curriculum} {pos} {completed} {overrides} />
  {/key}
{:else}
  <div class="flex h-64 items-center justify-center" role="status" aria-label="Memuat pelajaran">
    <Loader class="h-8 w-8 animate-spin text-rose-700" />
  </div>
{/if}
