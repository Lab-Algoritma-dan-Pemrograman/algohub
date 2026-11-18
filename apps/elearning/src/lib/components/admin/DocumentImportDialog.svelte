<script lang="ts">
  import {
    Button,
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
    toast,
  } from '@algohub/ui';
  import Loader from '@lucide/svelte/icons/loader-circle';
  import { importLessonsIntoModule } from '#lib/admin/curriculum-draft.ts';
  import {
    documentImportService,
    mergeCurriculum,
    type ParseResult,
  } from '#lib/admin/document-import.ts';
  import { extractDocxText } from '#lib/admin/document-parser.ts';
  import type { ImportTarget, Level } from '#lib/admin/types.ts';

  let {
    open = $bindable(false),
    target,
    draft,
    onapply,
  }: {
    open?: boolean;
    target: ImportTarget | null;
    draft: Level[];
    onapply: (next: Level[]) => void;
  } = $props();

  let fileName = $state('');
  let content = $state('');
  let mode = $state<'regex' | 'ai'>('regex');
  let result = $state<ParseResult | null>(null);
  let reading = $state(false);
  let processing = $state(false);

  $effect(() => {
    if (open) return;
    fileName = '';
    content = '';
    result = null;
    mode = 'regex';
  });

  const counts = $derived({
    levels: result?.levels.length ?? 0,
    modules: result?.levels.reduce((sum, l) => sum + l.modules.length, 0) ?? 0,
    lessons:
      result?.levels.reduce(
        (sum, l) => sum + l.modules.reduce((n, m) => n + m.lessons.length, 0),
        0,
      ) ?? 0,
  });

  async function pick(event: Event) {
    const input = event.currentTarget as HTMLInputElement;
    const file = input.files?.[0];
    if (!file) return;
    reading = true;
    try {
      content = file.name.endsWith('.docx')
        ? (await extractDocxText(file)).join('\n')
        : await file.text();
      fileName = file.name;
      result = null;
    } catch {
      toast.error('Gagal membaca file. Pastikan file .txt, .docx, atau .md yang valid.');
    } finally {
      reading = false;
      input.value = '';
    }
  }

  async function process() {
    processing = true;
    try {
      result = await documentImportService.importFromText(content, mode === 'ai');
    } catch (error) {
      toast.error(
        error instanceof Error ? error.message : 'Gagal mengekstrak konten dari dokumen.',
      );
    } finally {
      processing = false;
    }
  }

  function apply() {
    if (!result || !target) return;
    try {
      const next =
        target === 'global'
          ? mergeCurriculum(structuredClone($state.snapshot(draft)) as Level[], result.levels)
          : importLessonsIntoModule(draft, target.levelId, target.modIdx, result.levels);
      onapply(next);
      toast.success('Import berhasil. Simpan ke server untuk menerapkannya.');
      open = false;
    } catch (error) {
      toast.error(
        error instanceof Error ? error.message : 'Terjadi kesalahan saat menerapkan data.',
      );
    }
  }
</script>

<Dialog bind:open>
  <DialogContent class="rounded-3xl sm:max-w-lg">
    <DialogHeader>
      <DialogTitle class="text-xl font-black">Import Dokumen</DialogTitle>
      <DialogDescription>
        {target === 'global'
          ? 'Level, modul, dan pelajaran dari dokumen digabung ke seluruh draf.'
          : 'Pelajaran dari dokumen digabung ke modul ini.'}
      </DialogDescription>
    </DialogHeader>

    <div class="space-y-4">
      <input
        type="file"
        accept=".txt,.md,.docx"
        class="block w-full text-sm"
        disabled={reading}
        onchange={pick}
      />
      {#if fileName}
        <p class="text-xs font-bold text-zinc-500">{fileName}</p>
        <fieldset class="flex gap-4 text-sm">
          <label class="flex items-center gap-2">
            <input type="radio" name="parse-mode" value="regex" bind:group={mode} /> Cepat (regex)
          </label>
          <label class="flex items-center gap-2">
            <input type="radio" name="parse-mode" value="ai" bind:group={mode} /> Pintar (AI)
          </label>
        </fieldset>
        <Button
          variant="outline"
          class="rounded-xl font-bold"
          disabled={processing}
          onclick={process}
        >
          {#if processing}<Loader class="mr-2 h-4 w-4 animate-spin" />{/if}
          Proses Dokumen
        </Button>
      {/if}
      {#if result}
        <p class="rounded-xl bg-emerald-50 p-3 text-sm font-semibold text-emerald-800">
          Terdeteksi {counts.levels} level, {counts.modules} modul, {counts.lessons} pelajaran.
        </p>
      {/if}
    </div>

    <DialogFooter class="gap-2">
      <Button variant="outline" class="rounded-xl" onclick={() => (open = false)}>Batal</Button>
      <Button class="rounded-xl font-bold" disabled={!result} onclick={apply}
        >Terapkan ke Draf</Button
      >
    </DialogFooter>
  </DialogContent>
</Dialog>
