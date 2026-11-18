<script lang="ts">
  import type { GameQuestion } from '@algohub/api-client/elearning';
  import {
    Button,
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
    Input,
    Switch,
    Textarea,
    toast,
  } from '@algohub/ui';
  import Loader from '@lucide/svelte/icons/loader-circle';
  import Pencil from '@lucide/svelte/icons/pencil';
  import Plus from '@lucide/svelte/icons/plus';
  import Sparkles from '@lucide/svelte/icons/sparkles';
  import Trash from '@lucide/svelte/icons/trash-2';
  import Upload from '@lucide/svelte/icons/upload';
  import { createQuery } from '@tanstack/svelte-query';
  import { perform } from '#lib/admin/action.ts';
  import { aiCurriculumService } from '#lib/admin/ai-curriculum.ts';
  import { codeKeydown } from '#lib/admin/code-keys.ts';
  import { extractDocxText } from '#lib/admin/document-parser.ts';
  import ConfirmDialog from '#lib/components/admin/ConfirmDialog.svelte';
  import { gameSettingsQuery } from '#lib/queries/game.ts';
  import { adminGameQuestionsQuery, gameActions } from '#lib/queries/admin.ts';

  type Filter = 'all' | 'c' | 'python';
  interface Draft {
    id: string;
    language: string;
    difficulty: string;
    title: string;
    code: string;
    bugLine: number | null;
    explanation: string;
  }

  const questions = createQuery(adminGameQuestionsQuery);
  const settings = createQuery(gameSettingsQuery);

  let filter = $state<Filter>('all');
  let editing = $state<Draft | null>(null);
  let editOpen = $state(false);
  let saving = $state(false);
  let limitValue = $state('');
  let confirmOpen = $state(false);
  let confirm = $state<{ title: string; message: string; run: () => void | Promise<void> }>({
    title: '',
    message: '',
    run: () => {},
  });
  let aiOpen = $state(false);
  let aiTopic = $state('');
  let aiBusy = $state(false);

  const visible = $derived(
    (questions.data ?? []).filter((q) => filter === 'all' || q.language === filter),
  );

  $effect(() => {
    const limit = settings.data?.bugHuntWeeklyLimit;
    if (typeof limit === 'number') limitValue = String(limit);
  });

  const blank = (language = 'python'): Draft => ({
    id: '',
    language,
    difficulty: 'medium',
    title: '',
    code: '',
    bugLine: 0,
    explanation: '',
  });

  function edit(q: GameQuestion) {
    editing = {
      id: q.id,
      language: q.language,
      difficulty: q.difficulty ?? 'medium',
      title: q.title ?? '',
      code: q.code ?? '',
      bugLine: q.bugLine,
      explanation: q.explanation ?? '',
    };
    editOpen = true;
  }

  function ask(title: string, message: string, run: () => void | Promise<void>) {
    confirm = { title, message, run };
    confirmOpen = true;
  }

  async function saveQuestion() {
    const q = editing;
    if (!q) return;
    saving = true;
    const body = {
      language: q.language,
      difficulty: q.difficulty,
      title: q.title,
      code: q.code,
      bugLine: q.bugLine,
      explanation: q.explanation,
    };
    const ok = await perform(
      () =>
        q.id
          ? gameActions.updateQuestion(q.id, { ...body, bugLine: q.bugLine })
          : gameActions.addQuestion(body),
      'Soal disimpan.',
      'Gagal menyimpan soal',
    );
    saving = false;
    if (ok) editOpen = false;
  }

  function removeQuestion(q: GameQuestion) {
    ask(
      'Hapus Soal',
      `Hapus soal "${q.title ?? q.id}"?`,
      () =>
        void perform(
          () => gameActions.deleteQuestion(q.id),
          'Soal dihapus.',
          'Gagal menghapus soal',
        ),
    );
  }

  async function importCode(event: Event) {
    const input = event.currentTarget as HTMLInputElement;
    const file = input.files?.[0];
    if (!file || !editing) return;
    try {
      const code = file.name.toLowerCase().endsWith('.docx')
        ? (await extractDocxText(file)).filter((p) => p.trim()).join('\n')
        : await file.text();
      editing.code = code.trim();
    } catch {
      toast.error('Tidak dapat membaca file. Pastikan format file benar (.docx, .md, atau .txt).');
    }
    input.value = '';
  }

  async function generate() {
    if (!aiTopic.trim()) return;
    aiBusy = true;
    try {
      const language = filter === 'all' ? 'python' : filter;
      const result = await aiCurriculumService.generateBugHuntQuestion(aiTopic, language, 'medium');
      editing = {
        ...blank(language),
        title: result.title,
        code: (result.code || '').replace(/\\n/g, '\n'),
        bugLine: result.bugLine,
        explanation: result.explanation,
      };
      aiOpen = false;
      aiTopic = '';
      editOpen = true;
    } catch (error) {
      toast.error(error instanceof Error ? error.message : 'Gagal generate soal via AI.');
    } finally {
      aiBusy = false;
    }
  }

  async function seed() {
    const { default: data } = await import('#lib/data/gameQuestions.json');
    await perform(
      async () => {
        const seeded = await gameActions.seedQuestions(data);
        if (!seeded) toast.info('Bank soal sudah terisi, seed dilewati.');
      },
      'Seed selesai.',
      'Gagal seed bank soal',
    );
  }

  function resetBank() {
    ask(
      'Reset Bank Soal Bug Hunt',
      'Reset bank soal ke data default JSON? Aksi ini akan me-replace semua soal di database.',
      async () => {
        const { default: data } = await import('#lib/data/gameQuestions.json');
        await perform(
          () => gameActions.replaceQuestions(data),
          'Bank soal berhasil disinkronkan ke data default.',
          'Gagal mereset bank soal game',
        );
      },
    );
  }

  const flag = (
    key: 'bugHuntActive' | 'bugHuntCActive' | 'bugHuntPythonActive',
    checked: boolean,
  ) =>
    perform(
      () => gameActions.updateSettings({ [key]: checked }),
      'Pengaturan disimpan.',
      'Gagal menyimpan pengaturan',
    );

  function saveLimit() {
    const limit = Number.parseInt(limitValue, 10);
    if (Number.isNaN(limit) || limit < 0) {
      toast.error('Batas mingguan harus angka 0 atau lebih.');
      return;
    }
    void perform(
      () => gameActions.updateSettings({ bugHuntWeeklyLimit: limit }),
      'Pengaturan disimpan.',
      'Gagal menyimpan pengaturan',
    );
  }
</script>

<div class="space-y-6">
  <section class="space-y-4 rounded-3xl border border-zinc-200 bg-white p-6 shadow-sm">
    <h3 class="font-black text-zinc-900">Pengaturan Bug Hunt</h3>
    {#if settings.data}
      <div class="grid grid-cols-1 gap-3 sm:grid-cols-3">
        <label
          class="flex items-center justify-between gap-3 rounded-xl bg-zinc-50 p-3 text-sm font-bold"
        >
          Game aktif
          <Switch
            checked={settings.data.bugHuntActive ?? true}
            onCheckedChange={(v) => flag('bugHuntActive', v)}
          />
        </label>
        <label
          class="flex items-center justify-between gap-3 rounded-xl bg-zinc-50 p-3 text-sm font-bold"
        >
          Level C
          <Switch
            checked={settings.data.bugHuntCActive ?? true}
            onCheckedChange={(v) => flag('bugHuntCActive', v)}
          />
        </label>
        <label
          class="flex items-center justify-between gap-3 rounded-xl bg-zinc-50 p-3 text-sm font-bold"
        >
          Level Python
          <Switch
            checked={settings.data.bugHuntPythonActive ?? true}
            onCheckedChange={(v) => flag('bugHuntPythonActive', v)}
          />
        </label>
      </div>
      <div class="flex max-w-sm items-end gap-2">
        <label class="flex-1 space-y-1">
          <span class="text-xs font-bold text-zinc-400 uppercase"
            >Batas main per minggu (0 = tanpa batas)</span
          >
          <Input type="number" min="0" bind:value={limitValue} />
        </label>
        <Button class="rounded-xl font-bold" onclick={saveLimit}>Simpan</Button>
      </div>
    {:else}
      <p class="text-sm text-zinc-400">Memuat pengaturan...</p>
    {/if}
  </section>

  <section class="space-y-4 rounded-3xl border border-zinc-200 bg-white p-6 shadow-sm">
    <div class="flex flex-wrap items-center justify-between gap-3">
      <h3 class="font-black text-zinc-900">Bank Soal ({visible.length})</h3>
      <div class="flex flex-wrap gap-2">
        {#each ['all', 'c', 'python'] as const as option (option)}
          <Button
            variant={filter === option ? 'default' : 'outline'}
            size="sm"
            class="rounded-xl font-bold"
            onclick={() => (filter = option)}
          >
            {option === 'all' ? 'Semua' : option === 'c' ? 'C' : 'Python'}
          </Button>
        {/each}
        <Button
          size="sm"
          class="rounded-xl font-bold"
          onclick={() => {
            editing = blank(filter === 'all' ? 'python' : filter);
            editOpen = true;
          }}
        >
          <Plus class="mr-1 h-4 w-4" /> Soal
        </Button>
        <Button
          variant="outline"
          size="sm"
          class="rounded-xl font-bold"
          onclick={() => (aiOpen = true)}
        >
          <Sparkles class="mr-1 h-4 w-4 text-amber-500" /> Soal dengan AI
        </Button>
        <Button variant="outline" size="sm" class="rounded-xl font-bold" onclick={seed}
          >Seed jika kosong</Button
        >
        <Button
          variant="outline"
          size="sm"
          class="rounded-xl font-bold text-red-600"
          onclick={resetBank}
        >
          Reset ke default
        </Button>
      </div>
    </div>

    {#if questions.isPending}
      <p class="text-sm text-zinc-400">Memuat soal...</p>
    {:else}
      <ul class="max-h-[60vh] space-y-2 overflow-y-auto">
        {#each visible as q (q.id)}
          <li class="flex items-center gap-3 rounded-2xl border border-zinc-100 bg-zinc-50/60 p-3">
            <span
              class="rounded-md bg-zinc-900 px-2 py-0.5 text-[10px] font-black text-white uppercase"
              >{q.language}</span
            >
            <span
              class="rounded-md bg-amber-100 px-2 py-0.5 text-[10px] font-black text-amber-800 uppercase"
              >{q.difficulty}</span
            >
            <span class="min-w-0 flex-1 truncate text-sm font-semibold">{q.title}</span>
            <Button variant="ghost" size="icon" aria-label="Edit soal" onclick={() => edit(q)}>
              <Pencil class="h-4 w-4" />
            </Button>
            <Button
              variant="ghost"
              size="icon"
              aria-label="Hapus soal"
              onclick={() => removeQuestion(q)}
            >
              <Trash class="h-4 w-4 text-red-500" />
            </Button>
          </li>
        {:else}
          <li class="py-8 text-center text-sm text-zinc-400">Belum ada soal.</li>
        {/each}
      </ul>
    {/if}
  </section>
</div>

<Dialog bind:open={editOpen}>
  <DialogContent class="max-h-[90vh] overflow-y-auto rounded-3xl sm:max-w-2xl">
    <DialogHeader>
      <DialogTitle class="text-xl font-black"
        >{editing?.id ? 'Edit Soal' : 'Tambah Soal Baru'}</DialogTitle
      >
      <DialogDescription>Konfigurasi materi untuk tantangan Bug Hunt.</DialogDescription>
    </DialogHeader>
    {#if editing}
      <div class="space-y-4">
        <div class="grid grid-cols-2 gap-3">
          <label class="space-y-1">
            <span class="text-xs font-bold text-zinc-400 uppercase">Bahasa</span>
            <select
              class="w-full rounded-xl border border-zinc-200 bg-zinc-50 px-3 py-2 text-sm"
              bind:value={editing.language}
            >
              <option value="c">Bahasa C</option>
              <option value="python">Python</option>
            </select>
          </label>
          <label class="space-y-1">
            <span class="text-xs font-bold text-zinc-400 uppercase">Kesulitan</span>
            <select
              class="w-full rounded-xl border border-zinc-200 bg-zinc-50 px-3 py-2 text-sm"
              bind:value={editing.difficulty}
            >
              <option value="easy">Mudah</option>
              <option value="medium">Sedang</option>
              <option value="hard">Sulit</option>
            </select>
          </label>
        </div>
        <label class="block space-y-1">
          <span class="text-xs font-bold text-zinc-400 uppercase">Judul</span>
          <Input bind:value={editing.title} />
        </label>
        <div class="space-y-1">
          <div class="flex items-center justify-between">
            <span class="text-xs font-bold text-zinc-400 uppercase">Kode</span>
            <label class="flex cursor-pointer items-center gap-1 text-xs font-bold text-blue-700">
              <Upload class="h-3.5 w-3.5" /> Import
              <input type="file" accept=".docx,.md,.txt" class="hidden" onchange={importCode} />
            </label>
          </div>
          <textarea
            class="min-h-48 w-full rounded-xl border border-zinc-200 bg-zinc-950 p-3 font-mono text-xs text-zinc-100"
            spellcheck="false"
            value={editing.code}
            oninput={(e) => editing && (editing.code = e.currentTarget.value)}
            onkeydown={(e) =>
              editing && codeKeydown(e, editing.code, (next) => editing && (editing.code = next))}
          ></textarea>
        </div>
        <label class="block space-y-1">
          <span class="text-xs font-bold text-zinc-400 uppercase">Baris bug (mulai dari 0)</span>
          <Input
            type="number"
            min="0"
            value={editing.bugLine ?? 0}
            oninput={(e) =>
              editing && (editing.bugLine = Number.parseInt(e.currentTarget.value, 10) || 0)}
          />
        </label>
        <label class="block space-y-1">
          <span class="text-xs font-bold text-zinc-400 uppercase">Penjelasan</span>
          <Textarea bind:value={editing.explanation} rows={3} />
        </label>
      </div>
    {/if}
    <DialogFooter class="gap-2">
      <Button variant="outline" class="rounded-xl" onclick={() => (editOpen = false)}>Batal</Button>
      <Button
        class="rounded-xl font-bold"
        disabled={saving || !editing?.title || !editing?.code}
        onclick={saveQuestion}
      >
        {saving ? 'Menyimpan...' : 'Simpan'}
      </Button>
    </DialogFooter>
  </DialogContent>
</Dialog>

<Dialog bind:open={aiOpen}>
  <DialogContent class="rounded-3xl sm:max-w-md">
    <DialogHeader>
      <DialogTitle class="text-xl font-black">Buat Soal dengan AI</DialogTitle>
      <DialogDescription
        >Hasilnya dibuka di editor soal untuk ditinjau sebelum disimpan.</DialogDescription
      >
    </DialogHeader>
    <Input bind:value={aiTopic} placeholder="Topik, mis. perulangan atau pointer" />
    <DialogFooter class="gap-2">
      <Button variant="outline" class="rounded-xl" onclick={() => (aiOpen = false)}>Batal</Button>
      <Button class="rounded-xl font-bold" disabled={aiBusy || !aiTopic.trim()} onclick={generate}>
        {#if aiBusy}<Loader class="mr-1 h-4 w-4 animate-spin" />{/if}
        Buat
      </Button>
    </DialogFooter>
  </DialogContent>
</Dialog>

<ConfirmDialog
  bind:open={confirmOpen}
  title={confirm.title}
  message={confirm.message}
  onconfirm={confirm.run}
/>
