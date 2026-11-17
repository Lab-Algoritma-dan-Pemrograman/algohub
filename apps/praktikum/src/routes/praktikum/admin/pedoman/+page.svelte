<script lang="ts">
  import { createQuery } from '@tanstack/svelte-query';
  import { Check } from '@lucide/svelte';
  import type { Pedoman } from '@algohub/api-client/praktikum';
  import {
    createDeletePedomanMutation,
    createSavePedomanMutation,
    pedomanQuery,
  } from '#lib/queries/admin/pedoman.ts';
  import { createAdminUploadMutation } from '#lib/queries/upload.ts';
  import { confirmAction } from '#lib/stores/confirm.svelte.ts';

  const listQuery = createQuery(pedomanQuery);
  const saveMutation = createSavePedomanMutation();
  const deleteMutation = createDeletePedomanMutation();
  const uploadMutation = createAdminUploadMutation();

  let list = $derived(listQuery.data ?? []);
  let err = $state('');
  let msg = $state('');
  let editId = $state<number | null>(null);
  let form = $state({ nama_dokumen: '', file_url: '' });

  $effect(() => {
    if (listQuery.error) err = listQuery.error.message;
  });

  function reset() {
    editId = null;
    form = { nama_dokumen: '', file_url: '' };
  }
  function edit(p: Pedoman) {
    editId = p.id;
    form = { nama_dokumen: p.nama_dokumen, file_url: p.file_url };
  }
  async function uploadFile(ev: Event) {
    const input = ev.target as HTMLInputElement;
    if (!input.files?.[0]) return;
    try {
      const res = await uploadMutation.mutateAsync({ file: input.files[0], folder: 'pedoman' });
      form.file_url = res.url;
      msg = 'File berhasil diunggah.';
    } catch (e) {
      err = (e as Error).message;
    }
  }
  async function save() {
    err = '';
    msg = '';
    try {
      await saveMutation.mutateAsync({ id: editId, body: form });
      msg = 'Tersimpan.';
      reset();
    } catch (e) {
      err = (e as Error).message;
    }
  }
  async function del(id: number) {
    if (
      !(await confirmAction({
        title: 'Hapus Pedoman?',
        message: 'Apakah Anda yakin ingin menghapus pedoman laporan ini?',
      }))
    )
      return;
    try {
      await deleteMutation.mutateAsync(id);
    } catch (e) {
      err = (e as Error).message;
    }
  }
</script>

<h1 class="mb-4 text-2xl">Manajemen Pedoman Laporan</h1>
<p class="mb-4 text-sm text-ink-caption">
  File yang diunggah di sini akan tampil sebagai tombol download di halaman /info/laporan.
</p>
{#if msg}<p class="mb-3 rounded-lg bg-state-success-bg p-3 text-sm text-state-success">
    {msg}
  </p>{/if}
{#if err}<p class="mb-3 rounded-lg bg-state-error-bg p-3 text-sm text-state-error">{err}</p>{/if}

<div class="grid gap-4 lg:grid-cols-3">
  <div class="card">
    <h2 class="mb-3 text-lg">{editId ? 'Edit' : 'Tambah'} Pedoman</h2>
    <label class="label" for="nd">Nama Dokumen</label>
    <input
      id="nd"
      class="input"
      bind:value={form.nama_dokumen}
      placeholder="mis. Template Laporan Akhir"
    />
    <label class="label mt-2" for="fu">File</label>
    <input id="fu" type="file" onchange={uploadFile} />
    {#if form.file_url}
      <p class="mt-1 inline-flex items-center gap-1 text-xs text-state-success">
        <Check size={14} /> File:
        <a href={form.file_url} target="_blank" rel="noopener">{form.file_url.split('/').pop()}</a>
      </p>
    {/if}
    <div class="mt-3 flex gap-2">
      <button class="btn-primary" onclick={save}>Simpan</button>
      {#if editId}<button class="btn-outline" onclick={reset}>Batal</button>{/if}
    </div>
  </div>
  <div class="lg:col-span-2">
    <div class="table-wrap">
      <table class="tbl">
        <thead><tr><th>Nama Dokumen</th><th>File</th><th>Aksi</th></tr></thead>
        <tbody>
          {#each list as p (p.id)}
            <tr>
              <td>{p.nama_dokumen}</td>
              <td
                ><a
                  href={p.file_url}
                  target="_blank"
                  rel="noopener"
                  class="text-primary hover:underline">Download</a
                ></td
              >
              <td class="space-x-2">
                <button class="text-primary hover:underline" onclick={() => edit(p)}>Edit</button>
                <button class="text-state-error hover:underline" onclick={() => del(p.id)}
                  >Hapus</button
                >
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  </div>
</div>
