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
    RichTextEditor,
  } from '@algohub/ui';
  import Plus from '@lucide/svelte/icons/plus';
  import Trash from '@lucide/svelte/icons/trash-2';
  import { untrack } from 'svelte';
  import { codeKeydown } from '#lib/admin/code-keys.ts';
  import { uploadFromForm } from '#lib/admin/upload.ts';
  import type { Lesson } from '#lib/admin/types.ts';

  let {
    open = $bindable(false),
    lesson,
    onsave,
  }: { open?: boolean; lesson: Lesson | null; onsave: (lesson: Lesson) => void } = $props();

  // Salinan dalam: draf lesson di induk baru berubah setelah tombol Simpan.
  let form = $state<Lesson | null>(null);

  $effect(() => {
    if (!open) return;
    untrack(() => {
      if (!lesson) return;
      const copy = structuredClone($state.snapshot(lesson)) as Lesson;
      // Lesson dari AI parsial bisa kehilangan field; dilengkapi supaya form tidak error.
      copy.quiz ??= { question: '', options: ['', '', '', ''], correctAnswer: 0 };
      copy.testCases = Array.isArray(copy.testCases) ? copy.testCases : [];
      copy.validationRules = Array.isArray(copy.validationRules) ? copy.validationRules : [];
      form = copy;
    });
  });

  const CODE_FIELDS = [
    ['codeExample', 'Contoh Kode'],
    ['initialCode', 'Kode Awal (Latihan)'],
    ['solution', 'Solusi'],
  ] as const;

  function save() {
    if (!form) return;
    onsave($state.snapshot(form) as Lesson);
    open = false;
  }
</script>

<Dialog bind:open>
  <DialogContent class="max-h-[90vh] overflow-y-auto rounded-3xl sm:max-w-3xl">
    <DialogHeader>
      <DialogTitle class="text-xl font-black">Edit Pelajaran</DialogTitle>
      <DialogDescription
        >Perubahan masuk ke draf; simpan ke server dari tab Struktur.</DialogDescription
      >
    </DialogHeader>

    {#if form}
      <div class="space-y-5">
        <label class="block space-y-1">
          <span class="text-xs font-bold text-zinc-400 uppercase">Judul</span>
          <Input bind:value={form.title} />
        </label>

        <div class="space-y-1">
          <span class="text-xs font-bold text-zinc-400 uppercase">Penjelasan</span>
          <RichTextEditor bind:value={form.explanation} uploadImage={uploadFromForm} />
        </div>

        {#each CODE_FIELDS as [field, label] (field)}
          {@const current = form[field]}
          <label class="block space-y-1">
            <span class="text-xs font-bold text-zinc-400 uppercase">{label}</span>
            <textarea
              class="min-h-32 w-full rounded-xl border border-zinc-200 bg-zinc-950 p-3 font-mono text-xs text-zinc-100"
              spellcheck="false"
              value={current}
              oninput={(e) => form && (form[field] = e.currentTarget.value)}
              onkeydown={(e) =>
                form && codeKeydown(e, form[field], (next) => form && (form[field] = next))}
            ></textarea>
          </label>
        {/each}

        <label class="block space-y-1">
          <span class="text-xs font-bold text-zinc-400 uppercase">Petunjuk</span>
          <Input bind:value={form.hint} />
        </label>

        <fieldset class="space-y-2 rounded-2xl border border-zinc-200 p-4">
          <legend class="px-2 text-xs font-bold text-zinc-400 uppercase">Kuis</legend>
          <Input placeholder="Pertanyaan" bind:value={form.quiz.question} />
          {#each form.quiz.options as option, i (i)}
            <div class="flex items-center gap-2">
              <input
                type="radio"
                name="quiz-answer"
                checked={form.quiz.correctAnswer === i}
                onchange={() => form && (form.quiz.correctAnswer = i)}
                aria-label="Jawaban benar {i + 1}"
              />
              <Input
                placeholder="Opsi {i + 1}"
                value={option}
                oninput={(e) => form && (form.quiz.options[i] = e.currentTarget.value)}
              />
            </div>
          {/each}
        </fieldset>

        <fieldset class="space-y-2 rounded-2xl border border-zinc-200 p-4">
          <legend class="px-2 text-xs font-bold text-zinc-400 uppercase">Test Case</legend>
          {#each form.testCases as testCase, i (i)}
            <div class="grid grid-cols-1 gap-2 rounded-xl bg-zinc-50 p-3 sm:grid-cols-3">
              <Input placeholder="Deskripsi" bind:value={testCase.description} />
              <Input placeholder="Input (opsional)" bind:value={testCase.input} />
              <div class="flex gap-2">
                <Input placeholder="Output yang diharapkan" bind:value={testCase.expectedOutput} />
                <Button
                  variant="ghost"
                  size="icon"
                  aria-label="Hapus test case"
                  onclick={() => form?.testCases.splice(i, 1)}
                >
                  <Trash class="h-4 w-4 text-red-500" />
                </Button>
              </div>
            </div>
          {/each}
          <Button
            variant="outline"
            size="sm"
            class="rounded-xl"
            onclick={() => form?.testCases.push({ expectedOutput: '', description: '' })}
          >
            <Plus class="mr-1 h-4 w-4" /> Tambah test case
          </Button>
        </fieldset>
      </div>
    {/if}

    <DialogFooter class="gap-2">
      <Button variant="outline" class="rounded-xl" onclick={() => (open = false)}>Batal</Button>
      <Button class="rounded-xl font-bold" onclick={save}>Simpan ke Draf</Button>
    </DialogFooter>
  </DialogContent>
</Dialog>
