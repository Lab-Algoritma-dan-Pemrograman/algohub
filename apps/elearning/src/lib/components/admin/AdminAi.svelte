<script lang="ts">
  import { Button, Textarea, toast } from '@algohub/ui';
  import Loader from '@lucide/svelte/icons/loader-circle';
  import Sparkles from '@lucide/svelte/icons/sparkles';
  import { perform } from '#lib/admin/action.ts';
  import { aiCurriculumService } from '#lib/admin/ai-curriculum.ts';
  import type { Level } from '#lib/admin/types.ts';
  import ConfirmDialog from '#lib/components/admin/ConfirmDialog.svelte';
  import { curriculumActions } from '#lib/queries/admin.ts';

  let material = $state('');
  let file = $state<File | null>(null);
  let generating = $state(false);
  let progress = $state('');
  let generated = $state<Level[] | null>(null);
  let confirmOpen = $state(false);

  function pickFile(event: Event) {
    const input = event.currentTarget as HTMLInputElement;
    const picked = input.files?.[0];
    if (!picked) return;
    if (picked.type !== 'application/pdf') {
      toast.error('File tidak valid. Mohon unggah file PDF.');
      input.value = '';
      return;
    }
    file = picked;
  }

  // Backend menerima PDF sebagai base64 tanpa prefix data URL.
  function readBase64(source: File): Promise<string> {
    return new Promise((resolve, reject) => {
      const reader = new FileReader();
      reader.onload = () => resolve(String(reader.result).split(',')[1] ?? '');
      reader.onerror = () => reject(reader.error);
      reader.readAsDataURL(source);
    });
  }

  async function generate() {
    if (!material.trim() && !file) return;
    generating = true;
    progress = 'Menganalisis isi materi...';
    try {
      const fileData = file ? await readBase64(file) : undefined;
      generated = await aiCurriculumService.generateCurriculum(
        material,
        fileData,
        (msg) => (progress = msg),
      );
    } catch (error) {
      toast.error(
        `Gagal membuat kurikulum: ${error instanceof Error ? error.message : 'Error tidak diketahui'}`,
      );
    } finally {
      generating = false;
      progress = '';
    }
  }

  async function save() {
    const levels = generated;
    if (!levels) return;
    const ok = await perform(
      () => curriculumActions.replace(levels, 'Membuat kurikulum baru menggunakan AI'),
      'Kurikulum berhasil disimpan!',
      'Gagal menyimpan kurikulum',
    );
    if (ok) {
      generated = null;
      material = '';
      file = null;
    }
  }
</script>

<section class="space-y-4 rounded-3xl border border-zinc-200 bg-white p-6 shadow-sm">
  <h3 class="flex items-center gap-2 font-black text-zinc-900">
    <Sparkles class="h-5 w-5 text-amber-500" /> Kurikulum AI
  </h3>
  <p class="text-sm text-zinc-500">
    Tempel materi atau unggah PDF. AI menyusun 6 level lengkap dengan pelajaran, kuis, dan test
    case.
  </p>
  <Textarea bind:value={material} rows={8} placeholder="Tempel materi di sini..." />
  <div class="flex flex-wrap items-center gap-3">
    <input type="file" accept="application/pdf" class="text-sm" onchange={pickFile} />
    {#if file}<span class="text-xs font-bold text-zinc-500">{file.name}</span>{/if}
    <Button
      class="rounded-xl font-bold"
      disabled={generating || (!material.trim() && !file)}
      onclick={generate}
    >
      {#if generating}<Loader class="mr-1 h-4 w-4 animate-spin" />{/if}
      Buat Kurikulum
    </Button>
  </div>
  {#if progress}<p class="text-sm font-semibold text-zinc-500">{progress}</p>{/if}
</section>

{#if generated}
  <section class="mt-6 space-y-4 rounded-3xl border border-emerald-200 bg-emerald-50/50 p-6">
    <h3 class="font-black text-emerald-900">Hasil ({generated.length} level)</h3>
    <ul class="space-y-2">
      {#each generated as level (level.id)}
        <li class="rounded-xl bg-white p-3 text-sm">
          <span class="font-bold">{level.title}</span>
          <span class="text-xs text-zinc-500">
            · {level.modules.length} modul · {level.modules.reduce(
              (n, m) => n + m.lessons.length,
              0,
            )} pelajaran
          </span>
        </li>
      {/each}
    </ul>
    <div class="flex gap-2">
      <Button class="rounded-xl font-bold" onclick={() => (confirmOpen = true)}
        >Simpan Kurikulum</Button
      >
      <Button variant="outline" class="rounded-xl font-bold" onclick={() => (generated = null)}
        >Buang</Button
      >
    </div>
  </section>
{/if}

<ConfirmDialog
  bind:open={confirmOpen}
  title="Simpan Kurikulum"
  message="Kurikulum yang ada akan diganti seluruhnya dengan hasil AI. Lanjutkan?"
  onconfirm={save}
/>
