<script lang="ts">
  import { createQuery } from '@tanstack/svelte-query';
  import { DataTable, renderSnippet, type ColumnDef } from '@algohub/ui';
  import { FileSpreadsheet, Save } from '@lucide/svelte';
  import type { RekapMahasiswa, RekapSel } from '@algohub/api-client/praktikum';
  import { downloadCsv } from '#lib/download.ts';
  import { kelasQuery } from '#lib/queries/admin/kelas.ts';
  import { createKeaktifanMutation } from '#lib/queries/admin/penilaian.ts';
  import { rekapKelasQuery } from '#lib/queries/admin/rekap.ts';
  import { buildRekapCsv, rekapCsvFilename } from '#lib/rekap-nilai.ts';
  import { TBL_CELL_CLASS, TBL_HEAD_CLASS } from '#lib/table.ts';

  const kelasListQuery = createQuery(kelasQuery);
  const keaktifanMutation = createKeaktifanMutation();

  let kelasList = $derived(kelasListQuery.data ?? []);
  let selectedKelas = $state<number>(0);
  let errorMsg = $state('');
  let searchQuery = $state('');
  // Perubahan keaktifan yang belum disimpan: pengerjaan_id -> nilai baru.
  let edits = $state<Record<number, number>>({});
  let saving = $state(false);

  $effect(() => {
    if (kelasList.length > 0 && selectedKelas === 0) selectedKelas = kelasList[0].id;
  });

  const rekapQuery = createQuery(() => ({
    ...rekapKelasQuery(selectedKelas),
    enabled: selectedKelas > 0,
  }));
  let rekap = $derived(rekapQuery.data ?? null);
  let loading = $derived(rekapQuery.isLoading);

  $effect(() => {
    const e = kelasListQuery.error ?? rekapQuery.error;
    if (e) errorMsg = e.message;
  });

  function liveKeaktifan(sel: RekapSel): number {
    return Number(edits[sel.pengerjaan_id] ?? sel.keaktifan ?? 0);
  }

  async function saveKeaktifan() {
    const items = Object.entries(edits).map(([pid, nilai]) => ({
      pengerjaan_id: Number(pid),
      nilai: Number(nilai),
    }));
    if (items.length === 0) return;
    saving = true;
    errorMsg = '';
    try {
      await keaktifanMutation.mutateAsync({ items });
      edits = {};
    } catch (e) {
      errorMsg = (e as Error).message;
    } finally {
      saving = false;
    }
  }

  function selectKelas() {
    errorMsg = '';
  }

  let filteredData = $derived(
    rekap?.data.filter((m) => {
      const q = searchQuery.toLowerCase();
      return m.nim.toLowerCase().includes(q) || m.nama.toLowerCase().includes(q);
    }) || [],
  );

  let editCount = $derived(Object.keys(edits).length);

  function exportCSV() {
    if (!rekap || filteredData.length === 0) return;
    downloadCsv(buildRekapCsv(rekap, filteredData), rekapCsvFilename(selectedKelas));
  }

  let columns = $derived<ColumnDef<RekapMahasiswa>[]>([
    {
      id: 'nim',
      header: 'NIM',
      cell: ({ row }) => row.original.nim,
      meta: {
        headClass: 'sticky left-0 bg-gray-50 z-10 w-24 border-r border-gray-200',
        cellClass: 'sticky left-0 bg-white z-10 font-medium text-ink-body border-r border-gray-100',
      },
    },
    {
      id: 'nama',
      header: 'Nama',
      cell: ({ row }) => row.original.nama,
      meta: {
        headClass: 'sticky left-24 bg-gray-50 z-10 w-48 border-r border-gray-200',
        cellClass:
          'sticky left-24 bg-white z-10 text-ink-body border-r border-gray-100 whitespace-nowrap truncate max-w-xs',
      },
    },
    ...(rekap?.columns ?? []).map((col): ColumnDef<RekapMahasiswa> => ({
      id: col.key,
      header: col.label,
      cell: ({ row }) => renderSnippet(scoreCell, { sel: row.original.scores[col.key] }),
      meta: { headClass: 'text-center', cellClass: 'text-center text-ink-caption' },
    })),
    {
      id: 'total',
      header: 'Total',
      cell: ({ row }) => row.original.total_score,
      meta: {
        headClass: 'text-center bg-gray-50 border-l border-gray-200 text-brand-blue',
        cellClass:
          'text-center bg-gray-50/50 border-l border-gray-100 font-semibold text-brand-blue',
      },
    },
  ]);
</script>

{#snippet scoreCell({ sel }: { sel: RekapSel | undefined })}
  {#if !sel}
    -
  {:else if sel.edit_keaktifan}
    <div class="flex items-center justify-center gap-1 text-xs">
      <span>{sel.murni}</span>
      <span class="text-ink-caption">+</span>
      <input
        type="number"
        min="0"
        max="100"
        class="w-12 rounded border border-gray-200 px-1 py-0.5 text-center"
        value={edits[sel.pengerjaan_id] ?? sel.keaktifan ?? ''}
        oninput={(e) => (edits[sel.pengerjaan_id] = Number((e.target as HTMLInputElement).value))}
      />
      <span class="font-semibold text-ink-body">= {sel.murni + liveKeaktifan(sel)}</span>
    </div>
  {:else if sel.nilai_akhir !== null && sel.nilai_akhir !== undefined}
    <span class="font-semibold text-brand-blue">{sel.nilai_akhir}</span>
    <span class="text-xs text-ink-caption">({sel.total})</span>
  {:else}
    {sel.total}
  {/if}
{/snippet}

<h1 class="mb-4 text-2xl font-bold">Rekap Nilai</h1>

{#if errorMsg}
  <div class="mb-4 rounded-lg bg-state-error-bg p-3 text-state-error">{errorMsg}</div>
{/if}

<div
  class="mb-6 rounded-xl bg-white p-5 shadow-sm border border-gray-100 flex flex-wrap items-end gap-4"
>
  <div class="flex-1 min-w-[200px]">
    <label for="kelas" class="mb-1 block text-sm font-medium text-ink-caption">Pilih Kelas</label>
    <select id="kelas" class="input w-full" bind:value={selectedKelas} onchange={selectKelas}>
      {#each kelasList as k (k.id)}
        <option value={k.id}>{k.nama_kelas}</option>
      {/each}
    </select>
  </div>
  <div class="flex-1 min-w-[200px]">
    <label for="search" class="mb-1 block text-sm font-medium text-ink-caption"
      >Cari Mahasiswa</label
    >
    <input
      type="text"
      id="search"
      placeholder="NIM atau Nama..."
      class="input w-full"
      bind:value={searchQuery}
    />
  </div>
  <div class="flex-none flex gap-2 items-end">
    <button
      class="btn-primary flex items-center gap-1.5 py-2.5 text-sm font-semibold"
      onclick={saveKeaktifan}
      disabled={saving || editCount === 0}
    >
      <Save size={16} />
      {saving ? 'Menyimpan…' : `Simpan Keaktifan${editCount ? ` (${editCount})` : ''}`}
    </button>
    <button
      class="btn-outline flex items-center gap-1.5 py-2.5 text-sm font-semibold hover:bg-slate-50 transition-colors"
      onclick={exportCSV}
      disabled={!rekap || filteredData.length === 0}
    >
      <FileSpreadsheet size={16} class="text-emerald-600" />
      Export CSV
    </button>
  </div>
</div>

{#if loading}
  <div class="py-10 text-center text-ink-caption">Memuat data rekapitulasi...</div>
{:else if rekap}
  {#if filteredData.length === 0}
    <div class="py-10 text-center text-ink-caption">
      Tidak ada data yang cocok dengan pencarian.
    </div>
  {:else}
    <DataTable
      data={filteredData}
      {columns}
      class="table-wrap rounded-xl border border-gray-100 bg-white shadow-sm overflow-x-auto"
      tableClass="tbl w-full min-w-max"
      headClass={TBL_HEAD_CLASS}
      cellClass={TBL_CELL_CLASS}
      rowClass={() => 'hover:bg-gray-50/50 transition-colors'}
    />
  {/if}
{/if}
