<script lang="ts">
  import {
    Button,
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
    Input,
    Textarea,
    toast,
  } from '@algohub/ui';
  import ChevronDown from '@lucide/svelte/icons/chevron-down';
  import ChevronRight from '@lucide/svelte/icons/chevron-right';
  import ChevronUp from '@lucide/svelte/icons/chevron-up';
  import Loader from '@lucide/svelte/icons/loader-circle';
  import Move from '@lucide/svelte/icons/move';
  import Pencil from '@lucide/svelte/icons/pencil';
  import Plus from '@lucide/svelte/icons/plus';
  import Sparkles from '@lucide/svelte/icons/sparkles';
  import Trash from '@lucide/svelte/icons/trash-2';
  import Upload from '@lucide/svelte/icons/upload';
  import { createQuery } from '@tanstack/svelte-query';
  import { SvelteSet } from 'svelte/reactivity';
  import { perform } from '#lib/admin/action.ts';
  import { aiCurriculumService } from '#lib/admin/ai-curriculum.ts';
  import {
    cycleLevelAccess,
    draftFromTree,
    effectiveMode,
    moveItem,
    moveLessonTo,
    newLesson,
    newLevel,
    newModule,
    replaceLevel,
    updateLesson,
    updateModule,
    type LessonRef,
  } from '#lib/admin/curriculum-draft.ts';
  import type { ImportTarget, Lesson, Level } from '#lib/admin/types.ts';
  import ConfirmDialog from '#lib/components/admin/ConfirmDialog.svelte';
  import DocumentImportDialog from '#lib/components/admin/DocumentImportDialog.svelte';
  import LessonEditDialog from '#lib/components/admin/LessonEditDialog.svelte';
  import { buildCurriculum } from '#lib/curriculum.ts';
  import { getLevelLanguageLabel } from '#lib/levelLanguage.ts';
  import { curriculumActions } from '#lib/queries/admin.ts';
  import { levelsQuery, lessonsQuery, modulesQuery } from '#lib/queries/curriculum.ts';

  const levels = createQuery(levelsQuery);
  const modules = createQuery(modulesQuery);
  const lessons = createQuery(lessonsQuery);

  const server = $derived(
    draftFromTree(buildCurriculum(levels.data ?? [], modules.data ?? [], lessons.data ?? [])),
  );

  let draft = $state<Level[]>([]);
  let hasChanges = $state(false);
  let saving = $state(false);
  const expandedLevels = new SvelteSet<string>();
  const expandedModules = new SvelteSet<string>();

  // Draf mengikuti server selama belum ada perubahan lokal; setelah simpan, hasil refetch masuk.
  $effect(() => {
    const fresh = server;
    if (!hasChanges) draft = structuredClone(fresh);
  });

  let editingLevel = $state<string | null>(null);
  let levelForm = $state({ title: '', description: '' });
  let editingModuleKey = $state<string | null>(null);
  let moduleTitle = $state('');

  let lessonDialogOpen = $state(false);
  let lessonRef = $state<LessonRef | null>(null);
  let lessonForEdit = $state<Lesson | null>(null);

  let importOpen = $state(false);
  let importTarget = $state<ImportTarget | null>(null);

  let moving = $state<LessonRef | null>(null);
  let moveLevelId = $state('');
  let moveModId = $state('');

  let aiOpen = $state(false);
  let aiTarget = $state<{ type: 'module' | 'lesson'; levelId: string; modIdx?: number } | null>(
    null,
  );
  let aiContext = $state('');
  let aiBusy = $state(false);

  let confirmOpen = $state(false);
  let confirm = $state<{ title: string; message: string; run: () => void | Promise<void> }>({
    title: '',
    message: '',
    run: () => {},
  });

  const moveLevel = $derived(draft.find((l) => l.id === moveLevelId));

  function ask(title: string, message: string, run: () => void | Promise<void>) {
    confirm = { title, message, run };
    confirmOpen = true;
  }

  function setDraft(next: Level[]) {
    draft = next;
    hasChanges = true;
  }

  const levelById = (id: string) => draft.find((l) => l.id === id);

  function patchLevel(level: Level) {
    setDraft(replaceLevel(draft, level));
  }

  function toggle(set: SvelteSet<string>, key: string) {
    if (set.has(key)) set.delete(key);
    else set.add(key);
  }

  function addLevel() {
    const level = newLevel(draft.length);
    setDraft([...draft, level]);
    expandedLevels.add(level.id);
  }

  function saveLevelEdit(level: Level) {
    patchLevel({
      ...level,
      title: levelForm.title || level.title,
      description: levelForm.description || level.description,
    });
    editingLevel = null;
  }

  function saveModuleEdit(level: Level, modIdx: number) {
    patchLevel(
      updateModule(level, modIdx, { title: moduleTitle.trim() || level.modules[modIdx].title }),
    );
    editingModuleKey = null;
  }

  function openLessonEditor(ref: LessonRef) {
    const lesson = levelById(ref.levelId)?.modules[ref.modIdx]?.lessons[ref.lessonIdx];
    if (!lesson) return;
    lessonRef = ref;
    lessonForEdit = $state.snapshot(lesson) as Lesson;
    lessonDialogOpen = true;
  }

  function saveLesson(next: Lesson) {
    const ref = lessonRef;
    const level = ref && levelById(ref.levelId);
    if (!ref || !level) return;
    patchLevel(updateLesson(level, ref.modIdx, ref.lessonIdx, next));
  }

  function openImport(target: ImportTarget) {
    importTarget = target;
    importOpen = true;
  }

  function openMove(ref: LessonRef) {
    moving = ref;
    moveLevelId = ref.levelId;
    moveModId = levelById(ref.levelId)?.modules[0]?.id ?? '';
  }

  function pickMoveLevel(id: string) {
    moveLevelId = id;
    moveModId = levelById(id)?.modules[0]?.id ?? '';
  }

  function confirmMove() {
    if (!moving) return;
    const next = moveLessonTo(draft, moving, moveLevelId, moveModId);
    if (next) {
      setDraft(next);
      toast.success('Pelajaran dipindahkan. Jangan lupa simpan ke server.');
    }
    moving = null;
  }

  function removeLevel(level: Level) {
    ask(
      'Hapus Level',
      'Hapus level ini beserta semua modul dan pelajarannya? Tindakan ini tidak dapat dibatalkan!',
      () => setDraft(draft.filter((l) => l.id !== level.id)),
    );
  }

  function removeModule(level: Level, modIdx: number) {
    const title = level.modules[modIdx]?.title ?? 'modul';
    ask('Hapus Modul', `Hapus modul "${title}" beserta semua pelajarannya?`, () =>
      patchLevel({ ...level, modules: level.modules.filter((_, i) => i !== modIdx) }),
    );
  }

  function removeLesson(level: Level, modIdx: number, lessonIdx: number) {
    const title = level.modules[modIdx]?.lessons[lessonIdx]?.title ?? 'pelajaran';
    ask('Hapus Pelajaran', `Hapus pelajaran "${title}"?`, () =>
      patchLevel(
        updateModule(level, modIdx, {
          lessons: level.modules[modIdx].lessons.filter((_, i) => i !== lessonIdx),
        }),
      ),
    );
  }

  async function submitAi() {
    const target = aiTarget;
    const level = target && levelById(target.levelId);
    if (!target || !level || !aiContext.trim()) return;
    aiBusy = true;
    try {
      const language = getLevelLanguageLabel(level, draft.indexOf(level));
      if (target.type === 'module') {
        const created = await aiCurriculumService.generateSingleModule(
          aiContext,
          level.title,
          language,
        );
        patchLevel({ ...level, modules: [...level.modules, created] });
      } else if (typeof target.modIdx === 'number') {
        const mod = level.modules[target.modIdx];
        const created = await aiCurriculumService.generateSingleLesson(
          aiContext,
          mod.title,
          language,
        );
        patchLevel(updateModule(level, target.modIdx, { lessons: [...mod.lessons, created] }));
      }
      toast.success(
        `${target.type === 'module' ? 'Modul' : 'Pelajaran'} baru ditambahkan ke draft.`,
      );
      aiOpen = false;
      aiContext = '';
    } catch (error) {
      toast.error(error instanceof Error ? error.message : 'Gagal memanggil AI.');
    } finally {
      aiBusy = false;
    }
  }

  async function commit() {
    saving = true;
    const ok = await perform(
      () =>
        curriculumActions.save(
          $state.snapshot(draft) as Level[],
          'Mengubah struktur kurikulum (level, modul, atau pelajaran)',
        ),
      'Semua perubahan berhasil disimpan ke database.',
      'Gagal menyimpan perubahan ke database',
    );
    saving = false;
    if (ok) hasChanges = false;
  }

  function discard() {
    ask('Batalkan Perubahan', 'Semua draf perubahan akan hilang. Lanjutkan?', () => {
      draft = structuredClone(server);
      hasChanges = false;
    });
  }

  function clearAll() {
    ask(
      'Hapus Kurikulum',
      'PERINGATAN: Ini akan menghapus SEMUA kurikulum yang ada di database. Lanjutkan?',
      async () => {
        const ok = await perform(
          () => curriculumActions.clear(),
          'Kurikulum berhasil dikosongkan!',
          'Gagal menghapus kurikulum',
        );
        if (ok) hasChanges = false;
      },
    );
  }

  const mini = 'h-7 w-7 rounded-lg';
</script>

<div class="space-y-6">
  <div class="flex flex-wrap items-center justify-between gap-3">
    <div class="flex flex-wrap gap-2">
      <Button class="rounded-xl font-bold" onclick={addLevel}>
        <Plus class="mr-1 h-4 w-4" /> Level
      </Button>
      <Button variant="outline" class="rounded-xl font-bold" onclick={() => openImport('global')}>
        <Upload class="mr-1 h-4 w-4" /> Import Dokumen
      </Button>
      <Button variant="outline" class="rounded-xl font-bold text-red-600" onclick={clearAll}>
        Kosongkan Kurikulum
      </Button>
    </div>
    <div class="flex items-center gap-2">
      {#if hasChanges}
        <span class="text-xs font-bold text-amber-600">Ada perubahan belum disimpan</span>
        <Button variant="outline" class="rounded-xl font-bold" disabled={saving} onclick={discard}>
          Batalkan
        </Button>
        <Button class="rounded-xl font-bold" disabled={saving} onclick={commit}>
          {#if saving}<Loader class="mr-1 h-4 w-4 animate-spin" />{/if}
          Simpan ke Server
        </Button>
      {/if}
    </div>
  </div>

  {#if levels.isPending}
    <p class="py-12 text-center text-zinc-400">Memuat kurikulum...</p>
  {/if}

  {#each draft as level, lIdx (level.id)}
    <section class="rounded-3xl border border-zinc-200 bg-white shadow-sm">
      <header class="flex flex-wrap items-center gap-2 p-4">
        <button
          type="button"
          class="rounded-lg p-1 hover:bg-zinc-100"
          aria-label="Buka atau tutup level"
          onclick={() => toggle(expandedLevels, level.id)}
        >
          {#if expandedLevels.has(level.id)}<ChevronDown size={18} />{:else}<ChevronRight
              size={18}
            />{/if}
        </button>
        {#if editingLevel === level.id}
          <Input class="max-w-xs" bind:value={levelForm.title} placeholder="Judul level" />
          <Input class="max-w-sm" bind:value={levelForm.description} placeholder="Deskripsi" />
          <Button size="sm" class="rounded-lg" onclick={() => saveLevelEdit(level)}>Simpan</Button>
          <Button
            size="sm"
            variant="outline"
            class="rounded-lg"
            onclick={() => (editingLevel = null)}
          >
            Batal
          </Button>
        {:else}
          <div class="min-w-0 flex-1">
            <h3 class="truncate font-black text-zinc-900">{level.title}</h3>
            <p class="truncate text-xs text-zinc-500">{level.description}</p>
          </div>
          <Button
            variant="outline"
            size="sm"
            class="rounded-lg text-xs"
            onclick={() => patchLevel(cycleLevelAccess(level))}
          >
            Akses: {effectiveMode(level)}
          </Button>
          <Button
            variant="ghost"
            size="icon"
            class={mini}
            aria-label="Edit level"
            onclick={() => {
              editingLevel = level.id;
              levelForm = { title: level.title, description: level.description };
            }}
          >
            <Pencil class="h-4 w-4" />
          </Button>
          <Button
            variant="ghost"
            size="icon"
            class={mini}
            aria-label="Naikkan level"
            disabled={lIdx === 0}
            onclick={() => setDraft(moveItem(draft, lIdx, -1))}
          >
            <ChevronUp class="h-4 w-4" />
          </Button>
          <Button
            variant="ghost"
            size="icon"
            class={mini}
            aria-label="Turunkan level"
            disabled={lIdx === draft.length - 1}
            onclick={() => setDraft(moveItem(draft, lIdx, 1))}
          >
            <ChevronDown class="h-4 w-4" />
          </Button>
          <Button
            variant="ghost"
            size="icon"
            class={mini}
            aria-label="Hapus level"
            onclick={() => removeLevel(level)}
          >
            <Trash class="h-4 w-4 text-red-500" />
          </Button>
        {/if}
      </header>

      {#if expandedLevels.has(level.id)}
        <div class="space-y-3 border-t border-zinc-100 p-4">
          {#each level.modules as mod, mIdx (mod.id)}
            {@const modKey = `${level.id}:${mod.id}`}
            <div class="rounded-2xl border border-zinc-200 bg-zinc-50/50">
              <div class="flex flex-wrap items-center gap-2 p-3">
                <button
                  type="button"
                  class="rounded-lg p-1 hover:bg-zinc-100"
                  aria-label="Buka atau tutup modul"
                  onclick={() => toggle(expandedModules, modKey)}
                >
                  {#if expandedModules.has(modKey)}<ChevronDown size={16} />{:else}<ChevronRight
                      size={16}
                    />{/if}
                </button>
                {#if editingModuleKey === modKey}
                  <Input class="max-w-xs" bind:value={moduleTitle} />
                  <Button size="sm" class="rounded-lg" onclick={() => saveModuleEdit(level, mIdx)}
                    >Simpan</Button
                  >
                  <Button
                    size="sm"
                    variant="outline"
                    class="rounded-lg"
                    onclick={() => (editingModuleKey = null)}
                  >
                    Batal
                  </Button>
                {:else}
                  <span class="min-w-0 flex-1 truncate text-sm font-bold">
                    {mod.title}
                    <span class="text-xs font-normal text-zinc-400"
                      >({mod.lessons.length} pelajaran)</span
                    >
                  </span>
                  <Button
                    variant="ghost"
                    size="icon"
                    class={mini}
                    aria-label="Edit modul"
                    onclick={() => {
                      editingModuleKey = modKey;
                      moduleTitle = mod.title;
                    }}
                  >
                    <Pencil class="h-4 w-4" />
                  </Button>
                  <Button
                    variant="ghost"
                    size="icon"
                    class={mini}
                    aria-label="Import dokumen ke modul"
                    onclick={() => openImport({ levelId: level.id, modIdx: mIdx })}
                  >
                    <Upload class="h-4 w-4" />
                  </Button>
                  <Button
                    variant="ghost"
                    size="icon"
                    class={mini}
                    aria-label="Tambah pelajaran dengan AI"
                    onclick={() => {
                      aiTarget = { type: 'lesson', levelId: level.id, modIdx: mIdx };
                      aiOpen = true;
                    }}
                  >
                    <Sparkles class="h-4 w-4 text-amber-500" />
                  </Button>
                  <Button
                    variant="ghost"
                    size="icon"
                    class={mini}
                    aria-label="Tambah pelajaran"
                    onclick={() =>
                      patchLevel(
                        updateModule(level, mIdx, { lessons: [...mod.lessons, newLesson()] }),
                      )}
                  >
                    <Plus class="h-4 w-4" />
                  </Button>
                  <Button
                    variant="ghost"
                    size="icon"
                    class={mini}
                    aria-label="Naikkan modul"
                    disabled={mIdx === 0}
                    onclick={() =>
                      patchLevel({ ...level, modules: moveItem(level.modules, mIdx, -1) })}
                  >
                    <ChevronUp class="h-4 w-4" />
                  </Button>
                  <Button
                    variant="ghost"
                    size="icon"
                    class={mini}
                    aria-label="Turunkan modul"
                    disabled={mIdx === level.modules.length - 1}
                    onclick={() =>
                      patchLevel({ ...level, modules: moveItem(level.modules, mIdx, 1) })}
                  >
                    <ChevronDown class="h-4 w-4" />
                  </Button>
                  <Button
                    variant="ghost"
                    size="icon"
                    class={mini}
                    aria-label="Hapus modul"
                    onclick={() => removeModule(level, mIdx)}
                  >
                    <Trash class="h-4 w-4 text-red-500" />
                  </Button>
                {/if}
              </div>

              {#if expandedModules.has(modKey)}
                <ul class="space-y-1 border-t border-zinc-200 p-3">
                  {#each mod.lessons as lesson, lessonIdx (lesson.id)}
                    <li class="flex items-center gap-2 rounded-xl bg-white px-3 py-2 text-sm">
                      <span class="min-w-0 flex-1 truncate font-semibold">{lesson.title}</span>
                      <Button
                        variant="ghost"
                        size="icon"
                        class={mini}
                        aria-label="Edit pelajaran"
                        onclick={() =>
                          openLessonEditor({ levelId: level.id, modIdx: mIdx, lessonIdx })}
                      >
                        <Pencil class="h-4 w-4" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="icon"
                        class={mini}
                        aria-label="Pindahkan pelajaran"
                        onclick={() => openMove({ levelId: level.id, modIdx: mIdx, lessonIdx })}
                      >
                        <Move class="h-4 w-4" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="icon"
                        class={mini}
                        aria-label="Naikkan pelajaran"
                        disabled={lessonIdx === 0}
                        onclick={() =>
                          patchLevel(
                            updateModule(level, mIdx, {
                              lessons: moveItem(mod.lessons, lessonIdx, -1),
                            }),
                          )}
                      >
                        <ChevronUp class="h-4 w-4" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="icon"
                        class={mini}
                        aria-label="Turunkan pelajaran"
                        disabled={lessonIdx === mod.lessons.length - 1}
                        onclick={() =>
                          patchLevel(
                            updateModule(level, mIdx, {
                              lessons: moveItem(mod.lessons, lessonIdx, 1),
                            }),
                          )}
                      >
                        <ChevronDown class="h-4 w-4" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="icon"
                        class={mini}
                        aria-label="Hapus pelajaran"
                        onclick={() => removeLesson(level, mIdx, lessonIdx)}
                      >
                        <Trash class="h-4 w-4 text-red-500" />
                      </Button>
                    </li>
                  {:else}
                    <li class="px-3 py-2 text-xs text-zinc-400">Belum ada pelajaran.</li>
                  {/each}
                </ul>
              {/if}
            </div>
          {/each}

          <div class="flex gap-2">
            <Button
              variant="outline"
              size="sm"
              class="rounded-xl"
              onclick={() => patchLevel({ ...level, modules: [...level.modules, newModule()] })}
            >
              <Plus class="mr-1 h-4 w-4" /> Modul
            </Button>
            <Button
              variant="outline"
              size="sm"
              class="rounded-xl"
              onclick={() => {
                aiTarget = { type: 'module', levelId: level.id };
                aiOpen = true;
              }}
            >
              <Sparkles class="mr-1 h-4 w-4 text-amber-500" /> Modul dengan AI
            </Button>
          </div>
        </div>
      {/if}
    </section>
  {/each}
</div>

<LessonEditDialog bind:open={lessonDialogOpen} lesson={lessonForEdit} onsave={saveLesson} />

<DocumentImportDialog bind:open={importOpen} target={importTarget} {draft} onapply={setDraft} />

<ConfirmDialog
  bind:open={confirmOpen}
  title={confirm.title}
  message={confirm.message}
  onconfirm={confirm.run}
/>

<Dialog open={moving !== null} onOpenChange={(open) => !open && (moving = null)}>
  <DialogContent class="rounded-3xl sm:max-w-md">
    <DialogHeader>
      <DialogTitle class="text-xl font-black">Pindahkan Pelajaran</DialogTitle>
      <DialogDescription>Pilih level dan modul tujuan.</DialogDescription>
    </DialogHeader>
    <div class="space-y-3">
      <select
        class="w-full rounded-xl border border-zinc-200 bg-zinc-50 px-3 py-2 text-sm"
        value={moveLevelId}
        onchange={(e) => pickMoveLevel(e.currentTarget.value)}
      >
        {#each draft as l (l.id)}<option value={l.id}>{l.title}</option>{/each}
      </select>
      <select
        class="w-full rounded-xl border border-zinc-200 bg-zinc-50 px-3 py-2 text-sm"
        bind:value={moveModId}
      >
        {#each moveLevel?.modules ?? [] as m (m.id)}<option value={m.id}>{m.title}</option>
        {:else}<option value="">(Tidak ada modul di level ini)</option>{/each}
      </select>
    </div>
    <DialogFooter class="gap-2">
      <Button variant="outline" class="rounded-xl" onclick={() => (moving = null)}>Batal</Button>
      <Button class="rounded-xl font-bold" disabled={!moveModId} onclick={confirmMove}
        >Pindahkan</Button
      >
    </DialogFooter>
  </DialogContent>
</Dialog>

<Dialog bind:open={aiOpen}>
  <DialogContent class="rounded-3xl sm:max-w-md">
    <DialogHeader>
      <DialogTitle class="text-xl font-black">
        Buat {aiTarget?.type === 'module' ? 'Modul' : 'Pelajaran'} dengan AI
      </DialogTitle>
      <DialogDescription>Jelaskan topik yang diinginkan.</DialogDescription>
    </DialogHeader>
    <Textarea bind:value={aiContext} placeholder="Contoh: perulangan for dan while" rows={4} />
    <DialogFooter class="gap-2">
      <Button variant="outline" class="rounded-xl" onclick={() => (aiOpen = false)}>Batal</Button>
      <Button
        class="rounded-xl font-bold"
        disabled={aiBusy || !aiContext.trim()}
        onclick={submitAi}
      >
        {#if aiBusy}<Loader class="mr-1 h-4 w-4 animate-spin" />{/if}
        Buat
      </Button>
    </DialogFooter>
  </DialogContent>
</Dialog>
