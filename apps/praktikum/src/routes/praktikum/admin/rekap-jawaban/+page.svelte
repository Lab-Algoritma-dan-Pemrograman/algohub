<script lang="ts">
  import { untrack } from 'svelte';
  import { createQuery } from '@tanstack/svelte-query';
  import { SvelteSet } from 'svelte/reactivity';
  import { DataTable, renderSnippet, type ColumnDef } from '@algohub/ui';
  import { RotateCcw, Trash2, X, LockOpen, Syringe } from '@lucide/svelte';
  import type { AktivasiSesi, RekapJawabanItem, User } from '@algohub/api-client/praktikum';
  import { aktivasiListQuery } from '#lib/queries/admin/aktivasi.ts';
  import { kelasQuery } from '#lib/queries/admin/kelas.ts';
  import {
    createBulkActionMutation,
    createInjectJawabanMutation,
    createSaveNilaiMutation,
    penilaianRekapQuery,
  } from '#lib/queries/admin/penilaian.ts';
  import { rekapJawabanQuery } from '#lib/queries/admin/rekap.ts';
  import { sesiListQuery } from '#lib/queries/admin/sesi.ts';
  import { usersQuery } from '#lib/queries/admin/users.ts';
  import { queryClient } from '#lib/query.ts';
  import { sanitizeSoal } from '#lib/sanitize.ts';
  import { confirmAction } from '#lib/stores/confirm.svelte.ts';
  import { TBL_CELL_CLASS, TBL_HEAD_CLASS } from '#lib/table.ts';
  import { labelJenis, renderMath } from '#lib/utils.ts';

  const kelasListQuery = createQuery(kelasQuery);
  const sesiQuery = createQuery(sesiListQuery);
  const bulkActionMutation = createBulkActionMutation();
  const saveNilaiMutation = createSaveNilaiMutation();
  const injectMutation = createInjectJawabanMutation();

  let kelasList = $derived(kelasListQuery.data ?? []);
  let sesiList = $derived(sesiQuery.data ?? []);

  let selectedKelas = $state<number>(0);
  let selectedSesi = $state<number>(0);
  let selectedJenis = $state<string>('');
  let searchQuery = $state<string>('');
  // Pencarian baru dikirim saat Enter, sama dengan app lama.
  let appliedSearch = $state<string>('');

  const rekapQuery = createQuery(() =>
    rekapJawabanQuery({
      kelas_id: selectedKelas,
      sesi_id: selectedSesi,
      jenis: selectedJenis,
      search: appliedSearch,
    }),
  );
  let rekap = $derived(rekapQuery.data ?? null);
  let bulkBusy = $state(false);
  let loading = $derived(rekapQuery.isFetching || bulkBusy);
  let errorMsg = $state('');
  let successMsg = $state('');

  $effect(() => {
    const e = kelasListQuery.error ?? sesiQuery.error ?? rekapQuery.error;
    if (e) errorMsg = e.message;
  });

  const selectedIds = new SvelteSet<number>();

  // Setiap hasil baru membuang pilihan, seperti fetchRekap lama.
  $effect(() => {
    void rekapQuery.dataUpdatedAt;
    untrack(() => selectedIds.clear());
  });

  let showDetailModal = $state(false);
  let detailData = $state<RekapJawabanItem | null>(null);

  let showEditModal = $state(false);
  let editData = $state<RekapJawabanItem | null>(null);
  let editNilai = $state<number | null>(null);
  let editFeedback = $state<string>('');

  function fetchRekap() {
    errorMsg = '';
    selectedIds.clear();
    rekapQuery.refetch();
  }

  function filterChanged() {
    errorMsg = '';
  }

  function handleSearch(e: KeyboardEvent) {
    if (e.key === 'Enter') {
      errorMsg = '';
      if (appliedSearch === searchQuery) fetchRekap();
      else appliedSearch = searchQuery;
    }
  }

  function toggleSelectAll() {
    if (!rekap || rekap.items.length === 0) return;
    if (selectedIds.size === rekap.items.length) {
      selectedIds.clear();
    } else {
      rekap.items.forEach((i) => selectedIds.add(i.jawaban_id));
    }
  }

  function toggleSelect(id: number) {
    if (selectedIds.has(id)) {
      selectedIds.delete(id);
    } else {
      selectedIds.add(id);
    }
  }

  async function doBulkAction(action: 'reset_nilai' | 'delete' | 'buka_kunci') {
    if (selectedIds.size === 0) return;
    const count = selectedIds.size;
    const msgConfirm =
      action === 'delete'
        ? `Hapus permanen ${count} jawaban terpilih?`
        : action === 'buka_kunci'
          ? `Buka kunci pengerjaan untuk ${count} jawaban terpilih? Mahasiswa terkait bisa mengerjakan & submit ulang (selama course masih dibuka).`
          : `Reset nilai (jadi null) untuk ${count} jawaban terpilih?`;

    if (
      !(await confirmAction({
        title:
          action === 'delete'
            ? 'Hapus Jawaban?'
            : action === 'buka_kunci'
              ? 'Buka Kunci?'
              : 'Reset Nilai?',
        message: msgConfirm,
      }))
    )
      return;

    errorMsg = '';
    successMsg = '';
    bulkBusy = true;

    try {
      await bulkActionMutation.mutateAsync({ action, jawaban_ids: Array.from(selectedIds) });
      successMsg = `Berhasil melakukan bulk ${action} pada ${count} data.`;
    } catch (e) {
      errorMsg = (e as Error).message;
    } finally {
      bulkBusy = false;
    }
  }

  function openDetail(item: RekapJawabanItem) {
    detailData = item;
    showDetailModal = true;
  }

  function openEdit(item: RekapJawabanItem) {
    editData = item;
    editNilai = item.nilai;
    editFeedback = item.feedback || '';
    showEditModal = true;
  }

  async function saveNilai() {
    if (!editData) return;
    errorMsg = '';
    successMsg = '';
    try {
      await saveNilaiMutation.mutateAsync({
        jawaban_id: editData.jawaban_id,
        nilai: Number(editNilai),
        feedback: editFeedback || null,
      });
      successMsg = 'Nilai berhasil disimpan.';
      showEditModal = false;
    } catch (e) {
      errorMsg = (e as Error).message;
    }
  }

  function formatDate(ts: string | null) {
    if (!ts) return '-';
    return new Date(ts).toLocaleString('id-ID', { dateStyle: 'short', timeStyle: 'short' });
  }

  function truncate(str: string, len: number) {
    if (!str) return '-';
    if (str.length <= len) return str;
    return str.slice(0, len) + '...';
  }

  interface SoalTerpilihSimple {
    soal_terpilih_id: number;
    urutan: number;
    jenis_soal: string;
    teks_soal: string;
    poin: number;
  }

  let showInjectModal = $state(false);
  let injectAktivasiList = $state<AktivasiSesi[]>([]);
  let injectUsers = $state<User[]>([]);
  let injectSoalList = $state<SoalTerpilihSimple[]>([]);
  let injectLoading = $state(false);

  let injectForm = $state({
    aktivasi_sesi_id: 0,
    course_id: 0,
    mahasiswa_id: 0,
    soal_terpilih_id: 0,
    jawaban_teks: '',
    auto_submit: false,
  });

  let injectCourses = $derived(
    injectAktivasiList.find((a) => a.id === injectForm.aktivasi_sesi_id)?.aktivasi_courses ?? [],
  );

  async function openInjectModal() {
    showInjectModal = true;
    injectForm = {
      aktivasi_sesi_id: 0,
      course_id: 0,
      mahasiswa_id: 0,
      soal_terpilih_id: 0,
      jawaban_teks: '',
      auto_submit: false,
    };
    injectSoalList = [];
    try {
      injectAktivasiList = await queryClient.fetchQuery(aktivasiListQuery());
      injectUsers = await queryClient.fetchQuery(usersQuery());
    } catch (e) {
      errorMsg = (e as Error).message;
    }
  }

  async function loadInjectSoal() {
    if (!injectForm.aktivasi_sesi_id || !injectForm.course_id) {
      injectSoalList = [];
      return;
    }
    try {
      const res = await queryClient.fetchQuery(
        penilaianRekapQuery(injectForm.aktivasi_sesi_id, injectForm.course_id),
      );
      // Satu soal terpilih muncul di jawaban banyak mahasiswa, jadi diambil sekali saja.
      const seen: Record<number, boolean> = {};
      const list: SoalTerpilihSimple[] = [];
      for (const item of res?.items ?? []) {
        if (item.soal_terpilih_id && !seen[item.soal_terpilih_id]) {
          seen[item.soal_terpilih_id] = true;
          list.push({
            soal_terpilih_id: item.soal_terpilih_id,
            urutan: item.urutan ?? 0,
            jenis_soal: item.jenis_soal ?? '',
            teks_soal: item.teks_soal ?? '',
            poin: item.poin ?? 0,
          });
        }
      }
      injectSoalList = list.sort((a, b) => a.urutan - b.urutan);
    } catch {
      injectSoalList = [];
    }
  }

  async function doInject() {
    if (!injectForm.mahasiswa_id || !injectForm.soal_terpilih_id) return;
    injectLoading = true;
    errorMsg = '';
    successMsg = '';
    try {
      await injectMutation.mutateAsync({
        mahasiswa_id: injectForm.mahasiswa_id,
        soal_terpilih_id: injectForm.soal_terpilih_id,
        jawaban_teks: injectForm.jawaban_teks,
        auto_submit: injectForm.auto_submit,
      });
      successMsg = 'Jawaban berhasil diinject/diedit.';
      showInjectModal = false;
    } catch (e) {
      errorMsg = (e as Error).message;
    } finally {
      injectLoading = false;
    }
  }

  const columns: ColumnDef<RekapJawabanItem>[] = [
    {
      id: 'pilih',
      header: () => renderSnippet(selectAllHead),
      cell: ({ row }) => renderSnippet(selectCell, { row: row.original }),
      meta: { headClass: 'w-12 text-center', cellClass: 'text-center' },
    },
    {
      id: 'mahasiswa',
      header: 'NIM / Nama',
      cell: ({ row }) => renderSnippet(mahasiswaCell, { row: row.original }),
    },
    {
      id: 'kelas',
      header: 'Kelas / Sesi',
      cell: ({ row }) => renderSnippet(kelasCell, { row: row.original }),
    },
    {
      id: 'soal',
      header: 'Soal & Modul',
      cell: ({ row }) => renderSnippet(soalCell, { row: row.original }),
    },
    {
      id: 'waktu',
      header: 'Waktu',
      cell: ({ row }) => renderSnippet(waktuCell, { row: row.original }),
      meta: { cellClass: 'text-sm text-ink-caption' },
    },
    {
      id: 'nilai',
      header: 'Nilai',
      cell: ({ row }) => renderSnippet(nilaiCell, { row: row.original }),
      meta: { headClass: 'text-center', cellClass: 'text-center' },
    },
    {
      id: 'aksi',
      header: 'Aksi',
      cell: ({ row }) => renderSnippet(aksiCell, { row: row.original }),
      meta: { headClass: 'text-center', cellClass: 'text-center space-x-1' },
    },
  ];
</script>

{#snippet selectAllHead()}
  <input
    type="checkbox"
    checked={!!rekap && selectedIds.size === rekap.items.length && rekap.items.length > 0}
    onchange={toggleSelectAll}
    class="rounded border-gray-300 text-primary focus:ring-primary"
  />
{/snippet}

{#snippet selectCell({ row }: { row: RekapJawabanItem })}
  <input
    type="checkbox"
    checked={selectedIds.has(row.jawaban_id)}
    onchange={() => toggleSelect(row.jawaban_id)}
    class="rounded border-gray-300 text-primary focus:ring-primary"
  />
{/snippet}

{#snippet mahasiswaCell({ row }: { row: RekapJawabanItem })}
  <div class="font-medium text-ink-body">{row.nim}</div>
  <div class="text-xs text-ink-caption truncate max-w-[150px]">{row.nama_mahasiswa}</div>
{/snippet}

{#snippet kelasCell({ row }: { row: RekapJawabanItem })}
  <div class="text-sm text-ink-body">{row.nama_kelas}</div>
  <div class="text-xs text-ink-caption">{row.judul_sesi}</div>
{/snippet}

{#snippet soalCell({ row }: { row: RekapJawabanItem })}
  <div class="flex items-center gap-2 mb-1">
    <span
      class="badge {row.jenis_course === 'keterampilan' || row.jenis_course === 'ujian_praktik'
        ? 'bg-amber-100 text-amber-800'
        : 'bg-primary/10 text-primary'}"
    >
      {labelJenis(row.jenis_course)}
    </span>
    <span class="text-xs font-medium text-ink-caption">({row.jenis_soal})</span>
  </div>
  <div class="text-xs font-medium text-ink-body mb-1">{row.judul_course}</div>
  <div class="text-xs text-ink-body truncate max-w-[200px]" title={row.teks_soal}>
    {truncate(row.teks_soal, 40)}
  </div>
{/snippet}

{#snippet waktuCell({ row }: { row: RekapJawabanItem })}
  {#if row.is_submitted}
    <span class="block text-state-success">Submitted</span>
    <span class="text-xs">{formatDate(row.waktu_submit)}</span>
  {:else}
    <span class="text-state-warning">Belum Submit</span>
  {/if}
{/snippet}

{#snippet nilaiCell({ row }: { row: RekapJawabanItem })}
  {#if row.nilai === null}
    <span class="text-ink-caption">-</span>
  {:else}
    <span
      class="font-bold {row.nilai < row.poin_maksimal * 0.5
        ? 'text-state-error'
        : 'text-state-success'}"
    >
      {row.nilai}
    </span>
    <span class="text-xs text-ink-caption block">/{row.poin_maksimal}</span>
  {/if}
{/snippet}

{#snippet aksiCell({ row }: { row: RekapJawabanItem })}
  <button class="btn-outline px-2 py-1 text-xs" onclick={() => openDetail(row)}>Detail</button>
  <button
    class="btn-outline px-2 py-1 text-xs border-amber-500 text-amber-600 hover:bg-amber-50"
    onclick={() => openEdit(row)}>Nilai</button
  >
{/snippet}

<h1 class="mb-4 text-2xl font-bold">Rekap Jawaban Global</h1>

{#if errorMsg}
  <div class="mb-4 rounded-lg bg-state-error-bg p-3 text-state-error">{errorMsg}</div>
{/if}
{#if successMsg}
  <div class="mb-4 rounded-lg bg-state-success-bg p-3 text-state-success">{successMsg}</div>
{/if}

<div
  class="mb-6 rounded-xl border border-gray-100 bg-white p-5 shadow-sm flex flex-wrap items-end gap-4"
>
  <div class="flex-1 min-w-[150px]">
    <label for="kelas" class="mb-1 block text-sm font-medium text-ink-caption">Kelas</label>
    <select id="kelas" class="input w-full" bind:value={selectedKelas} onchange={filterChanged}>
      <option value={0}>Semua Kelas</option>
      {#each kelasList as k (k.id)}
        <option value={k.id}>{k.nama_kelas}</option>
      {/each}
    </select>
  </div>
  <div class="flex-1 min-w-[150px]">
    <label for="sesi" class="mb-1 block text-sm font-medium text-ink-caption">Sesi Praktikum</label>
    <select id="sesi" class="input w-full" bind:value={selectedSesi} onchange={filterChanged}>
      <option value={0}>Semua Sesi</option>
      {#each sesiList as s (s.id)}
        <option value={s.id}>{s.judul_sesi}</option>
      {/each}
    </select>
  </div>
  <div class="flex-1 min-w-[150px]">
    <label for="jenis" class="mb-1 block text-sm font-medium text-ink-caption">Jenis Tes</label>
    <select id="jenis" class="input w-full" bind:value={selectedJenis} onchange={filterChanged}>
      <option value="">Semua Jenis</option>
      <option value="pretest">Pre-test</option>
      <option value="posttest">Post-test</option>
      <option value="keterampilan">Keterampilan</option>
      <option value="ujian_praktik">Ujian Praktik</option>
    </select>
  </div>
  <div class="flex-1 min-w-[200px]">
    <label for="search" class="mb-1 block text-sm font-medium text-ink-caption"
      >Cari (NIM/Nama)</label
    >
    <input
      type="text"
      id="search"
      placeholder="Tekan Enter untuk cari..."
      class="input w-full"
      bind:value={searchQuery}
      onkeydown={handleSearch}
    />
  </div>
  <div class="flex-none flex gap-2">
    <button class="btn-outline inline-flex items-center gap-1" onclick={openInjectModal}>
      <Syringe size={16} /> Inject Jawaban
    </button>
    <button class="btn-primary" onclick={fetchRekap} disabled={loading}>
      {#if loading}
        Memuat...
      {:else}
        Refresh
      {/if}
    </button>
  </div>
</div>

{#if selectedIds.size > 0}
  <div
    class="mb-4 flex flex-wrap items-center gap-4 rounded-xl border border-primary/20 bg-primary/5 p-4 shadow-sm animate-fade-in"
  >
    <div class="font-medium text-primary"><strong>{selectedIds.size}</strong> jawaban dipilih</div>
    <div class="ml-auto flex gap-2">
      <button
        class="btn-outline inline-flex items-center gap-1 border-state-warning text-state-warning hover:bg-state-warning hover:text-white"
        onclick={() => doBulkAction('reset_nilai')}
      >
        <RotateCcw size={14} /> Reset Nilai
      </button>
      <button
        class="btn-outline inline-flex items-center gap-1 border-state-error text-state-error hover:bg-state-error hover:text-white"
        onclick={() => doBulkAction('delete')}
      >
        <Trash2 size={14} /> Hapus
      </button>
      <button
        class="btn-outline inline-flex items-center gap-1 border-primary text-primary hover:bg-primary hover:text-white"
        onclick={() => doBulkAction('buka_kunci')}
      >
        <LockOpen size={14} /> Buka Kunci
      </button>
      <button class="btn-outline" onclick={() => selectedIds.clear()}>Batal</button>
    </div>
  </div>
{/if}

{#if loading && !rekap}
  <div class="py-10 text-center text-ink-caption">Memuat data jawaban...</div>
{:else if rekap}
  <div class="mb-2 text-sm text-ink-caption">
    {#if rekap.limited}
      Menampilkan <strong>{rekap.total}</strong> jawaban terbaru. Pakai
      <strong>Cari (NIM/Nama)</strong> untuk menemukan data lainnya.
    {:else}
      Total: <strong>{rekap.total}</strong> jawaban
    {/if}
  </div>

  {#if rekap.items.length === 0}
    <div class="py-10 text-center text-ink-caption">Tidak ada data jawaban yang cocok.</div>
  {:else}
    <DataTable
      data={rekap.items}
      {columns}
      getRowId={(row) => String(row.jawaban_id)}
      estimateRowHeight={96}
      class="table-wrap rounded-xl border border-gray-100 bg-white shadow-sm overflow-x-auto"
      tableClass="tbl w-full min-w-max"
      headClass={TBL_HEAD_CLASS}
      cellClass={TBL_CELL_CLASS}
      rowClass={(row) =>
        `hover:bg-gray-50/50 transition-colors ${selectedIds.has(row.jawaban_id) ? 'bg-primary/5' : ''}`}
    />
  {/if}
{/if}

{#if showDetailModal && detailData}
  <div class="fixed inset-0 z-50 flex items-center justify-center p-4 sm:p-6">
    <button
      class="absolute inset-0 bg-black/60 backdrop-blur-sm"
      onclick={() => (showDetailModal = false)}
      aria-label="Tutup Modal"
    ></button>
    <div
      class="relative flex w-full max-w-4xl flex-col rounded-2xl bg-white shadow-2xl max-h-[90vh]"
    >
      <div class="flex items-center justify-between border-b px-6 py-4">
        <h3 class="text-lg font-bold text-ink-body">
          Detail Jawaban - {detailData.nama_mahasiswa}
        </h3>
        <button
          class="text-ink-caption hover:text-ink-body"
          onclick={() => (showDetailModal = false)}><X size={20} /></button
        >
      </div>

      <div class="overflow-y-auto p-6">
        <div class="grid grid-cols-2 gap-4 mb-6">
          <div class="rounded-lg bg-surface-soft p-4 text-sm">
            <div class="text-ink-caption mb-1">Soal / Modul</div>
            <div class="font-medium text-ink-body">
              {detailData.judul_course} ({detailData.jenis_soal})
            </div>
            <div
              class="mt-2 prose prose-sm max-w-none bg-white p-3 rounded border"
              use:renderMath={detailData.teks_soal}
            >
              <!-- eslint-disable-next-line svelte/no-at-html-tags -- teks soal sudah disanitasi sanitizeSoal -->
              {@html sanitizeSoal(detailData.teks_soal)}
            </div>
          </div>
          <div class="rounded-lg bg-surface-soft p-4 text-sm">
            <div class="grid grid-cols-2 gap-y-3">
              <div>
                <span class="text-ink-caption block text-xs">NIM</span>
                <span class="font-medium">{detailData.nim}</span>
              </div>
              <div>
                <span class="text-ink-caption block text-xs">Kelas</span>
                <span class="font-medium">{detailData.nama_kelas}</span>
              </div>
              <div>
                <span class="text-ink-caption block text-xs">Status</span>
                <span
                  class="font-medium {detailData.is_submitted
                    ? 'text-state-success'
                    : 'text-state-warning'}"
                >
                  {detailData.is_submitted ? 'Submitted' : 'Belum Submit'}
                </span>
              </div>
              <div>
                <span class="text-ink-caption block text-xs">Nilai Akhir</span>
                <span class="font-bold text-lg"
                  >{detailData.nilai ?? '-'}
                  <span class="text-sm font-normal text-ink-caption"
                    >/ {detailData.poin_maksimal}</span
                  ></span
                >
              </div>
            </div>
          </div>
        </div>

        <div class="rounded-lg border border-gray-200">
          <div class="bg-gray-50 px-4 py-2 border-b font-medium text-ink-body">
            Jawaban Mahasiswa
          </div>
          <div class="p-4 bg-white">
            {#if detailData.jenis_soal === 'coding'}
              <pre
                class="bg-[#1e1e1e] text-white p-4 rounded-lg overflow-x-auto text-sm font-mono">{detailData.jawaban_teks ||
                  '(Kosong)'}</pre>
            {:else}
              <div
                class="prose max-w-none text-ink-body bg-gray-50/50 p-4 rounded-lg min-h-[100px] whitespace-pre-wrap"
              >
                {detailData.jawaban_teks || '(Kosong)'}
              </div>
            {/if}
          </div>
        </div>

        {#if detailData.feedback}
          <div class="mt-4 rounded-lg border border-amber-200 bg-amber-50 p-4">
            <div class="font-medium text-amber-800 mb-1">Feedback:</div>
            <p class="text-sm text-amber-900">{detailData.feedback}</p>
          </div>
        {/if}
      </div>

      <div class="border-t bg-gray-50 px-6 py-4 flex justify-end gap-3 rounded-b-2xl">
        <button class="btn-outline" onclick={() => (showDetailModal = false)}>Tutup</button>
        <button
          class="btn-primary"
          onclick={() => {
            showDetailModal = false;
            openEdit(detailData!);
          }}>Edit Nilai</button
        >
      </div>
    </div>
  </div>
{/if}

{#if showEditModal && editData}
  <div class="fixed inset-0 z-[60] flex items-center justify-center p-4">
    <button
      class="absolute inset-0 bg-black/60 backdrop-blur-sm"
      onclick={() => (showEditModal = false)}
      aria-label="Tutup Modal"
    ></button>
    <div class="relative w-full max-w-md rounded-2xl bg-white p-6 shadow-2xl">
      <h3 class="mb-4 text-lg font-bold text-ink-body">Beri Nilai - {editData.nama_mahasiswa}</h3>

      <div class="mb-4">
        <label class="label block mb-1" for="input_nilai"
          >Nilai (Maks: {editData.poin_maksimal})</label
        >
        <input
          id="input_nilai"
          type="number"
          class="input w-full"
          bind:value={editNilai}
          min="0"
          max={editData.poin_maksimal}
        />
      </div>
      <div class="mb-6">
        <label class="label block mb-1" for="input_feedback">Feedback (Opsional)</label>
        <textarea id="input_feedback" class="input w-full min-h-[80px]" bind:value={editFeedback}
        ></textarea>
      </div>

      <div class="flex justify-end gap-3">
        <button class="btn-outline" onclick={() => (showEditModal = false)}>Batal</button>
        <button class="btn-primary" onclick={saveNilai}>Simpan Nilai</button>
      </div>
    </div>
  </div>
{/if}

{#if showInjectModal}
  <div class="fixed inset-0 z-[70] flex items-center justify-center p-4">
    <button
      class="absolute inset-0 bg-black/60 backdrop-blur-sm"
      onclick={() => (showInjectModal = false)}
      aria-label="Tutup Modal"
    ></button>
    <div
      class="relative w-full max-w-lg rounded-2xl bg-white shadow-2xl max-h-[90vh] flex flex-col"
    >
      <div class="flex items-center justify-between border-b px-6 py-4">
        <h3 class="text-lg font-bold text-ink-body">Inject / Edit Jawaban</h3>
        <button
          class="text-ink-caption hover:text-ink-body"
          onclick={() => (showInjectModal = false)}><X size={20} /></button
        >
      </div>

      <div class="overflow-y-auto p-6 space-y-4">
        <div>
          <label class="label block mb-1" for="inject_aktivasi">Aktivasi Sesi</label>
          <select
            id="inject_aktivasi"
            class="input w-full"
            bind:value={injectForm.aktivasi_sesi_id}
            onchange={() => {
              injectForm.course_id = 0;
              injectForm.soal_terpilih_id = 0;
              injectSoalList = [];
            }}
          >
            <option value={0}>— Pilih Aktivasi —</option>
            {#each injectAktivasiList as a (a.id)}
              <option value={a.id}
                >{a.sesi?.judul_sesi ?? `Sesi ${a.sesi_praktikum_id}`} — {a.kelas?.nama_kelas ??
                  `Kelas ${a.kelas_id}`} Shift {a.shift}</option
              >
            {/each}
          </select>
        </div>

        {#if injectForm.aktivasi_sesi_id}
          <div>
            <label class="label block mb-1" for="inject_course">Course</label>
            <select
              id="inject_course"
              class="input w-full"
              bind:value={injectForm.course_id}
              onchange={() => {
                injectForm.soal_terpilih_id = 0;
                loadInjectSoal();
              }}
            >
              <option value={0}>— Pilih Course —</option>
              {#each injectCourses as ac (ac.id)}
                <option value={ac.course_id}
                  >{ac.course?.judul ?? labelJenis(ac.course?.jenis ?? '')}</option
                >
              {/each}
            </select>
          </div>
        {/if}

        {#if injectForm.course_id}
          <div>
            <label class="label block mb-1" for="inject_mhs">Mahasiswa</label>
            <select id="inject_mhs" class="input w-full" bind:value={injectForm.mahasiswa_id}>
              <option value={0}>— Pilih Mahasiswa —</option>
              {#each injectUsers as u (u.id)}
                <option value={u.id}>{u.nama} ({u.nim})</option>
              {/each}
            </select>
          </div>
        {/if}

        {#if injectSoalList.length > 0}
          <div>
            <label class="label block mb-1" for="inject_soal">Soal</label>
            <select id="inject_soal" class="input w-full" bind:value={injectForm.soal_terpilih_id}>
              <option value={0}>— Pilih Soal —</option>
              {#each injectSoalList as s (s.soal_terpilih_id)}
                <option value={s.soal_terpilih_id}
                  >Soal {s.urutan} ({s.jenis_soal}, {s.poin} poin)</option
                >
              {/each}
            </select>
          </div>
        {/if}

        {#if injectForm.soal_terpilih_id}
          <div>
            <label class="label block mb-1" for="inject_jawaban">Jawaban</label>
            <textarea
              id="inject_jawaban"
              class="input w-full min-h-[120px] font-mono text-sm"
              placeholder="Tulis jawaban yang akan diinject..."
              bind:value={injectForm.jawaban_teks}></textarea>
          </div>

          <label class="flex items-center gap-2 cursor-pointer">
            <input
              type="checkbox"
              class="rounded border-gray-300 text-primary focus:ring-primary"
              bind:checked={injectForm.auto_submit}
            />
            <span class="text-sm text-ink-body">Langsung submit (mahasiswa tidak bisa edit)</span>
          </label>
        {/if}
      </div>

      <div class="border-t bg-gray-50 px-6 py-4 flex justify-end gap-3 rounded-b-2xl">
        <button class="btn-outline" onclick={() => (showInjectModal = false)}>Batal</button>
        <button
          class="btn-primary"
          onclick={doInject}
          disabled={!injectForm.mahasiswa_id || !injectForm.soal_terpilih_id || injectLoading}
        >
          {injectLoading ? 'Menyimpan...' : 'Inject Jawaban'}
        </button>
      </div>
    </div>
  </div>
{/if}
