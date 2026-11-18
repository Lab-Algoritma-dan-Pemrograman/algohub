<script lang="ts">
  import type { PlaygroundExample } from '@algohub/api-client/elearning';
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
  } from '@algohub/ui';
  import Pencil from '@lucide/svelte/icons/pencil';
  import Plus from '@lucide/svelte/icons/plus';
  import Trash from '@lucide/svelte/icons/trash-2';
  import { createQuery } from '@tanstack/svelte-query';
  import { perform } from '#lib/admin/action.ts';
  import { codeKeydown } from '#lib/admin/code-keys.ts';
  import ConfirmDialog from '#lib/components/admin/ConfirmDialog.svelte';
  import { adminPlaygroundQuery, playgroundActions } from '#lib/queries/admin.ts';

  interface Draft {
    id: string;
    title: string;
    language: string;
    code: string;
    description: string;
    sortOrder: number;
  }

  const examples = createQuery(adminPlaygroundQuery);

  let editing = $state<Draft | null>(null);
  let open = $state(false);
  let saving = $state(false);
  let removing = $state<PlaygroundExample | null>(null);
  let removeOpen = $state(false);

  function startNew() {
    const next = Math.max(0, ...(examples.data ?? []).map((e) => e.sortOrder)) + 1;
    editing = { id: '', title: '', language: 'python', code: '', description: '', sortOrder: next };
    open = true;
  }

  function edit(example: PlaygroundExample) {
    editing = {
      id: example.id,
      title: example.title,
      language: example.language,
      code: example.code,
      description: example.description ?? '',
      sortOrder: example.sortOrder,
    };
    open = true;
  }

  async function save() {
    const draft = editing;
    if (!draft) return;
    saving = true;
    const { id, ...body } = draft;
    const ok = await perform(
      () => (id ? playgroundActions.update(id, body) : playgroundActions.add(body)),
      'Contoh disimpan.',
      'Gagal menyimpan contoh',
    );
    saving = false;
    if (ok) open = false;
  }

  async function remove() {
    const target = removing;
    if (!target) return;
    await perform(
      () => playgroundActions.remove(target.id),
      'Contoh dihapus.',
      'Gagal menghapus contoh',
    );
  }
</script>

<section class="space-y-4 rounded-3xl border border-zinc-200 bg-white p-6 shadow-sm">
  <div class="flex items-center justify-between">
    <h3 class="font-black text-zinc-900">Contoh Playground ({examples.data?.length ?? 0})</h3>
    <Button size="sm" class="rounded-xl font-bold" onclick={startNew}>
      <Plus class="mr-1 h-4 w-4" /> Contoh
    </Button>
  </div>
  {#if examples.isPending}
    <p class="text-sm text-zinc-400">Memuat contoh...</p>
  {:else}
    <ul class="space-y-2">
      {#each examples.data ?? [] as example (example.id)}
        <li class="flex items-center gap-3 rounded-2xl border border-zinc-100 bg-zinc-50/60 p-3">
          <span
            class="rounded-md bg-zinc-900 px-2 py-0.5 text-[10px] font-black text-white uppercase"
            >{example.language}</span
          >
          <span class="min-w-0 flex-1 truncate text-sm font-semibold">{example.title}</span>
          <Button
            variant="ghost"
            size="icon"
            aria-label="Edit contoh"
            onclick={() => edit(example)}
          >
            <Pencil class="h-4 w-4" />
          </Button>
          <Button
            variant="ghost"
            size="icon"
            aria-label="Hapus contoh"
            onclick={() => {
              removing = example;
              removeOpen = true;
            }}
          >
            <Trash class="h-4 w-4 text-red-500" />
          </Button>
        </li>
      {:else}
        <li class="py-6 text-center text-sm text-zinc-400">Belum ada contoh.</li>
      {/each}
    </ul>
  {/if}
</section>

<Dialog bind:open>
  <DialogContent class="max-h-[90vh] overflow-y-auto rounded-3xl sm:max-w-2xl">
    <DialogHeader>
      <DialogTitle class="text-xl font-black"
        >{editing?.id ? 'Edit Contoh' : 'Tambah Contoh'}</DialogTitle
      >
      <DialogDescription>Contoh kode yang tampil di modal Playground.</DialogDescription>
    </DialogHeader>
    {#if editing}
      <div class="space-y-4">
        <div class="grid grid-cols-3 gap-3">
          <label class="col-span-2 space-y-1">
            <span class="text-xs font-bold text-zinc-400 uppercase">Judul</span>
            <Input bind:value={editing.title} />
          </label>
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
        </div>
        <label class="block space-y-1">
          <span class="text-xs font-bold text-zinc-400 uppercase">Deskripsi</span>
          <Textarea bind:value={editing.description} rows={2} />
        </label>
        <label class="block space-y-1">
          <span class="text-xs font-bold text-zinc-400 uppercase">Kode</span>
          <textarea
            class="min-h-48 w-full rounded-xl border border-zinc-200 bg-zinc-950 p-3 font-mono text-xs text-zinc-100"
            spellcheck="false"
            value={editing.code}
            oninput={(e) => editing && (editing.code = e.currentTarget.value)}
            onkeydown={(e) =>
              editing && codeKeydown(e, editing.code, (next) => editing && (editing.code = next))}
          ></textarea>
        </label>
        <label class="block space-y-1">
          <span class="text-xs font-bold text-zinc-400 uppercase">Urutan</span>
          <Input
            type="number"
            value={editing.sortOrder}
            oninput={(e) =>
              editing && (editing.sortOrder = Number.parseInt(e.currentTarget.value, 10) || 0)}
          />
        </label>
      </div>
    {/if}
    <DialogFooter class="gap-2">
      <Button variant="outline" class="rounded-xl" onclick={() => (open = false)}>Batal</Button>
      <Button
        class="rounded-xl font-bold"
        disabled={saving || !editing?.title || !editing?.code}
        onclick={save}
      >
        {saving ? 'Menyimpan...' : 'Simpan'}
      </Button>
    </DialogFooter>
  </DialogContent>
</Dialog>

<ConfirmDialog
  bind:open={removeOpen}
  title="Hapus Contoh"
  message={`Hapus contoh "${removing?.title ?? ''}"?`}
  confirmLabel="Hapus"
  onconfirm={remove}
/>
