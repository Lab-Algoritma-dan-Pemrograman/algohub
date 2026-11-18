<script lang="ts">
  import type { ElearningUser } from '@algohub/api-client/elearning';
  import { Button, Input } from '@algohub/ui';
  import CheckCircle from '@lucide/svelte/icons/circle-check';
  import ChevronDown from '@lucide/svelte/icons/chevron-down';
  import ChevronRight from '@lucide/svelte/icons/chevron-right';
  import Clock from '@lucide/svelte/icons/clock';
  import { lessonIds, type TreeLevel } from '#lib/curriculum.ts';
  import { getLevelLanguage } from '#lib/levelLanguage.ts';

  let {
    student,
    online,
    heartbeat,
    completed,
    totalLessons,
    curriculum,
    canManage,
    expanded,
    ontoggle,
    onadjustxp,
    onreset,
    ondelete,
  }: {
    student: ElearningUser;
    online: boolean;
    heartbeat: string | null;
    completed: readonly string[];
    totalLessons: number;
    curriculum: TreeLevel[];
    canManage: boolean;
    expanded: boolean;
    ontoggle: () => void;
    onadjustxp: (xp: number) => void;
    onreset: () => void;
    ondelete: () => void;
  } = $props();

  let editingXp = $state(false);
  let xpValue = $state('');

  const doneSet = $derived(new Set(completed));
  const percent = $derived(
    totalLessons > 0 ? Math.round((completed.length / totalLessons) * 100) : 0,
  );

  const levelStats = (level: TreeLevel) => {
    const ids = lessonIds(level);
    return { done: ids.filter((id) => doneSet.has(id)).length, total: ids.length };
  };

  function formatTime(iso: string | null): string {
    if (!iso) return '-';
    return new Intl.DateTimeFormat('id-ID', {
      timeZone: 'Asia/Jakarta',
      day: '2-digit',
      month: 'short',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    }).format(new Date(iso));
  }

  function studyText(seconds: number | null): string {
    if (!seconds || seconds <= 0) return '0 menit';
    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    return hours > 0 ? `${hours} jam ${minutes} menit` : `${minutes} menit`;
  }

  function saveXp() {
    const value = Number.parseInt(xpValue, 10);
    if (Number.isNaN(value) || value < 0) return;
    onadjustxp(value);
    editingXp = false;
    xpValue = '';
  }
</script>

<article
  class="mb-3 overflow-hidden rounded-[2rem] border bg-white {expanded
    ? 'border-rose-200 shadow-lg'
    : 'border-zinc-200'}"
>
  <button
    type="button"
    class="flex w-full flex-col gap-4 p-5 text-left sm:flex-row sm:items-center"
    aria-expanded={expanded}
    onclick={ontoggle}
  >
    <div class="flex min-w-0 flex-1 items-center gap-4">
      <div
        class="relative flex h-12 w-12 shrink-0 items-center justify-center rounded-2xl text-sm font-black {online
          ? 'bg-emerald-50 text-emerald-700 ring-2 ring-emerald-500/20'
          : 'bg-zinc-100 text-zinc-500'}"
      >
        {student.nama.charAt(0).toUpperCase()}
        <span
          class="absolute -right-0.5 -bottom-0.5 h-3.5 w-3.5 rounded-full border-2 border-white {online
            ? 'bg-emerald-500'
            : 'bg-zinc-300'}"
        ></span>
      </div>
      <div class="min-w-0">
        <div class="flex flex-wrap items-center gap-2">
          <span class="truncate text-base font-extrabold text-zinc-950">{student.nama}</span>
          <span
            class="rounded-lg border border-zinc-200 bg-zinc-100 px-2 py-0.5 text-[9px] font-black text-zinc-500"
          >
            {student.kelas}
          </span>
          {#if online}
            <span
              class="rounded-lg border border-emerald-200 bg-emerald-50 px-2 py-0.5 text-[8px] font-black text-emerald-700 uppercase"
            >
              Aktif
            </span>
          {/if}
        </div>
        <p class="mt-0.5 flex flex-wrap items-center gap-x-2 text-xs font-semibold text-zinc-500">
          <span>NIM: {student.nim}</span>
          <span class="font-bold text-rose-700">{student.xp.toLocaleString('id-ID')} XP</span>
          <span class="font-bold text-amber-600">Streak {student.streak} Hari</span>
          <span class="flex items-center gap-1 text-zinc-400">
            <Clock size={11} />
            {studyText(student.study_time)}
          </span>
        </p>
      </div>
    </div>
    <div class="flex shrink-0 items-center gap-3">
      <div
        class="hidden h-2.5 w-28 overflow-hidden rounded-full border border-zinc-200/50 bg-zinc-100 sm:block"
      >
        <div
          class="h-full rounded-full {percent === 100 ? 'bg-emerald-500' : 'bg-rose-700'}"
          style:width="{percent}%"
        ></div>
      </div>
      <div class="text-right">
        <div class="text-xs font-black text-zinc-900">{percent}%</div>
        <div class="text-[9px] font-bold text-zinc-400 uppercase">Selesai</div>
      </div>
      {#if expanded}<ChevronDown size={16} />{:else}<ChevronRight size={16} />{/if}
    </div>
  </button>

  {#if expanded}
    <div class="space-y-6 border-t border-zinc-100 bg-zinc-50/20 p-6">
      <div class="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
        {#each curriculum as level, idx (level.id)}
          {@const stats = levelStats(level)}
          {@const isC = getLevelLanguage(level, idx) === 'c'}
          <div class="rounded-2xl border border-zinc-200/60 bg-white p-4 shadow-sm">
            <div class="mb-2 flex items-center justify-between">
              <span
                class="rounded-lg border px-2 py-0.5 text-[9px] font-black uppercase {isC
                  ? 'border-blue-100 bg-blue-50 text-blue-700'
                  : 'border-rose-100 bg-rose-50 text-rose-700'}"
              >
                {isC ? 'Bahasa C' : 'Python'}
              </span>
              <span class="text-[10px] font-black text-zinc-500"
                >{stats.done} / {stats.total} Lesson</span
              >
            </div>
            <h4 class="truncate text-xs font-extrabold" title={level.title}>{level.title}</h4>
          </div>
        {/each}
      </div>

      <div
        class="custom-scrollbar max-h-72 space-y-5 overflow-y-auto rounded-2xl border border-zinc-200/70 bg-white p-4"
      >
        {#each curriculum as level, lIdx (level.id)}
          <div class="space-y-3">
            <div class="flex items-center gap-2 text-xs font-black tracking-wider uppercase">
              <span
                class="flex h-6 w-6 items-center justify-center rounded-lg bg-zinc-900 text-[10px] text-white"
              >
                {lIdx + 1}
              </span>
              <span class="truncate">{level.title}</span>
            </div>
            {#each level.modules as mod (mod.id)}
              <div class="ml-4 space-y-2 border-l-2 border-zinc-200 pl-4">
                <div class="text-[10px] font-black text-zinc-700 uppercase">{mod.title}</div>
                <div class="grid grid-cols-1 gap-2 sm:grid-cols-2 md:grid-cols-3">
                  {#each mod.lessons as lesson (lesson.id)}
                    {@const done = doneSet.has(lesson.id)}
                    <div
                      class="flex items-center gap-2 rounded-xl border p-2 text-[11px] {done
                        ? 'border-emerald-100 bg-emerald-50/40 font-semibold text-emerald-800'
                        : 'border-zinc-200/60 bg-white text-zinc-500'}"
                    >
                      {#if done}
                        <CheckCircle size={14} class="shrink-0 text-emerald-500" />
                      {:else}
                        <span class="h-3.5 w-3.5 shrink-0 rounded-full border border-zinc-300"
                        ></span>
                      {/if}
                      <span class="flex-1 truncate" title={lesson.title}>{lesson.title}</span>
                    </div>
                  {/each}
                </div>
              </div>
            {/each}
          </div>
        {/each}
      </div>

      <div class="flex flex-wrap items-center justify-between gap-4 border-t border-zinc-100 pt-4">
        <span class="flex items-center gap-2 text-xs font-medium text-zinc-500">
          <Clock size={14} /> Terakhir Aktif: {formatTime(heartbeat ?? student.last_active)}
        </span>
        {#if canManage}
          {#if editingXp}
            <div class="flex items-center gap-2">
              <Input
                type="number"
                min="0"
                class="w-28"
                placeholder="XP baru"
                bind:value={xpValue}
              />
              <Button size="sm" class="rounded-xl" onclick={saveXp}>Simpan</Button>
              <Button
                size="sm"
                variant="outline"
                class="rounded-xl"
                onclick={() => (editingXp = false)}>Batal</Button
              >
            </div>
          {:else}
            <div class="flex gap-2">
              <Button
                size="sm"
                variant="outline"
                class="rounded-xl text-[10px] font-black uppercase"
                onclick={() => {
                  editingXp = true;
                  xpValue = String(student.xp);
                }}
              >
                Atur XP
              </Button>
              <Button
                size="sm"
                variant="outline"
                class="rounded-xl text-[10px] font-black text-amber-700 uppercase"
                onclick={onreset}
              >
                Reset Progres
              </Button>
              <Button
                size="sm"
                variant="outline"
                class="rounded-xl text-[10px] font-black text-red-600 uppercase"
                onclick={ondelete}
              >
                Hapus
              </Button>
            </div>
          {/if}
        {/if}
      </div>
    </div>
  {/if}
</article>
