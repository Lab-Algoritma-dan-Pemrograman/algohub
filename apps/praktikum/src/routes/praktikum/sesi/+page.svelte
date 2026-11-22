<script lang="ts">
  import { createQuery } from '@tanstack/svelte-query';
  import { Lock, Search, ArrowRight, Filter, BookOpen, Layers } from '@lucide/svelte';
  import { mahasiswaQueries } from '#lib/queries/mahasiswa.ts';
  import { labelJenis, labelStatus, statusBadgeClass } from '#lib/utils.ts';
  import { toast } from '@algohub/ui';

  const sesiQuery = createQuery(mahasiswaQueries.sesi);

  let sesi = $derived(sesiQuery.data ?? []);
  let loading = $derived(sesiQuery.isPending);
  // Seperti swrGet lama: error hanya tampil bila belum ada data yang bisa ditampilkan.
  // Jika cache ada tapi revalidasi gagal, err kosong oleh sengaja; toast yang memberi tahu.
  let err = $derived(sesiQuery.data ? '' : (sesiQuery.error?.message ?? ''));

  let lastToastError = $state<string | null>(null);
  $effect(() => {
    const msg = sesiQuery.error?.message;
    if (!msg || !sesiQuery.data) return;
    if (lastToastError === msg) return;
    lastToastError = msg;
    toast.error('Gagal memperbarui daftar sesi', {
      description: 'Menampilkan data terakhir yang tersimpan. Tarik untuk coba lagi.',
    });
  });

  let searchQuery = $state('');
  let filterMode = $state<'semua' | 'aktif' | 'terkunci'>('semua');

  let filteredSesi = $derived(
    sesi.filter((s) => {
      const q = searchQuery.toLowerCase().trim();
      const matchesSearch =
        q === '' || s.judul.toLowerCase().includes(q) || s.deskripsi.toLowerCase().includes(q);
      const matchesFilter =
        filterMode === 'semua' ||
        (filterMode === 'aktif' && s.aktif) ||
        (filterMode === 'terkunci' && !s.aktif);
      return matchesSearch && matchesFilter;
    }),
  );

  let countAktif = $derived(sesi.filter((s) => s.aktif).length);
  let countTerkunci = $derived(sesi.filter((s) => !s.aktif).length);
</script>

<div class="space-y-6">
  <div class="flex flex-col gap-1">
    <div class="flex items-center gap-3">
      <div class="flex h-10 w-10 items-center justify-center rounded-xl bg-primary/10 text-primary">
        <Layers size={20} />
      </div>
      <div>
        <h1 class="text-2xl font-bold text-ink-heading">Daftar Sesi Praktikum</h1>
        <p class="text-sm text-ink-caption">
          {sesi.length} sesi tersedia · {countAktif} aktif · {countTerkunci} terkunci
        </p>
      </div>
    </div>
  </div>

  {#if loading}
    <p class="text-ink-caption">Memuat…</p>
  {:else if err}
    <p class="rounded-lg bg-state-error-bg p-3 text-state-error">{err}</p>
  {:else}
    <div
      class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between rounded-xl border border-gray-200 bg-white p-4 shadow-sm"
    >
      <div class="relative flex-1 max-w-sm">
        <Search size={16} class="absolute left-3 top-1/2 -translate-y-1/2 text-ink-caption" />
        <input
          type="text"
          placeholder="Cari sesi..."
          bind:value={searchQuery}
          class="input pl-9 w-full"
        />
      </div>
      <div class="flex items-center gap-2">
        <Filter size={14} class="text-ink-caption" />
        <div class="flex rounded-lg border border-gray-200 overflow-hidden">
          <button
            class="px-3 py-1.5 text-xs font-semibold transition-colors active:scale-[0.97] {filterMode === 'semua'
              ? 'bg-primary/80 text-white active:bg-primary/80'
              : 'bg-white text-ink-body hover:bg-slate-100 active:bg-slate-100'}"
            onclick={() => (filterMode = 'semua')}>Semua ({sesi.length})</button
          >
          <button
            class="px-3 py-1.5 text-xs font-semibold transition-colors border-l border-gray-200 active:scale-[0.97] {filterMode ===
            'aktif'
              ? 'bg-primary/80 text-white active:bg-primary/80'
              : 'bg-white text-ink-body hover:bg-slate-100 active:bg-slate-100'}"
            onclick={() => (filterMode = 'aktif')}>Aktif ({countAktif})</button
          >
          <button
            class="px-3 py-1.5 text-xs font-semibold transition-colors border-l border-gray-200 active:scale-[0.97] {filterMode ===
            'terkunci'
              ? 'bg-ink-heading text-white active:bg-ink-heading'
              : 'bg-white text-ink-body hover:bg-slate-100 active:bg-slate-100'}"
            onclick={() => (filterMode = 'terkunci')}>Terkunci ({countTerkunci})</button
          >
        </div>
      </div>
    </div>

    {#if filteredSesi.length === 0}
      <div class="rounded-xl border border-dashed border-gray-300 bg-gray-50 p-10 text-center">
        <p class="text-sm font-semibold text-ink-caption">Tidak ada sesi yang sesuai filter.</p>
      </div>
    {:else}
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {#each filteredSesi as s (s.sesi_id)}
          <div
            class="rounded-2xl border bg-white shadow-sm flex flex-col justify-between overflow-hidden transition-[transform,box-shadow,border-color] duration-200 hover:-translate-y-1 hover:shadow-md hover:border-primary/20 {s.aktif
              ? 'border-slate-200 border-l-4 border-l-primary'
              : 'border-slate-200 border-l-4 border-l-slate-300 bg-slate-50/40 opacity-80'}"
          >
            <div class="p-6 flex flex-col flex-grow min-w-0">
              <div class="flex items-start justify-between gap-3 mb-4">
                <div
                  class="flex h-11 w-11 flex-shrink-0 items-center justify-center rounded-xl {s.aktif
                    ? 'bg-primary/10 text-primary border border-primary/20'
                    : 'bg-slate-100 text-slate-400 border border-slate-200'}"
                >
                  {#if s.aktif}
                    <BookOpen size={20} />
                  {:else}
                    <Lock size={20} />
                  {/if}
                </div>

                <div class="flex flex-wrap gap-1.5 justify-end">
                  {#if s.aktif}
                    <span
                      class="badge bg-primary/10 text-primary border-primary/20 text-[10px] px-2 py-0.5"
                      >Aktif</span
                    >
                  {:else}
                    <span
                      class="badge inline-flex items-center gap-1 bg-slate-100 text-slate-400 border border-slate-200 text-[10px] px-2 py-0.5"
                      ><Lock size={10} /> Terkunci</span
                    >
                  {/if}
                  {#if s.susulan}
                    <span
                      class="badge bg-state-warning-bg text-state-warning text-[10px] px-2 py-0.5"
                      >Susulan</span
                    >
                  {/if}
                </div>
              </div>

              <h3 class="text-base font-bold text-ink-heading mb-1.5 line-clamp-2">{s.judul}</h3>
              <p class="text-xs text-ink-caption mb-5 line-clamp-3 leading-relaxed">
                {s.deskripsi}
              </p>

              {#if s.courses && s.courses.length > 0}
                <div class="mt-auto pt-4 border-t border-slate-100 space-y-2">
                  <h4 class="text-[10px] font-bold text-ink-caption uppercase tracking-wider mb-2">
                    Daftar Modul
                  </h4>
                  {#each s.courses as c (c.course_id)}
                    <div
                      class="flex items-center justify-between rounded-xl border border-slate-100/80 bg-slate-50/50 px-3 py-2 text-xs"
                    >
                      <div class="flex items-center gap-2.5 min-w-0">
                        <span class="text-sm flex-shrink-0">
                          {#if c.jenis === 'pretest'}📝{:else if c.jenis === 'posttest'}✅{:else if c.jenis === 'keterampilan'}🧪{:else}📋{/if}
                        </span>
                        <span class="font-semibold text-ink-heading truncate"
                          >{labelJenis(c.jenis)}</span
                        >
                      </div>
                      <span
                        class="badge text-[9px] px-1.5 py-0.5 flex-shrink-0 {statusBadgeClass(
                          c.status,
                        )}">{labelStatus(c.status)}{c.is_open ? '' : ' · terkunci'}</span
                      >
                    </div>
                  {/each}
                </div>
              {/if}
            </div>

            <div class="px-6 pb-6 pt-2 flex-shrink-0">
              {#if s.aktif}
                <a
                  href={`/praktikum/sesi/${s.sesi_id}`}
                  class="btn-primary w-full text-xs py-2 shadow-sm"
                >
                  Masuk Sesi <ArrowRight size={14} />
                </a>
              {:else}
                <button
                  class="btn w-full text-xs py-2 bg-slate-100 text-slate-400 border border-slate-200 cursor-not-allowed inline-flex items-center justify-center gap-1.5"
                  disabled
                >
                  <Lock size={12} /> Belum Diaktifkan
                </button>
              {/if}
            </div>
          </div>
        {/each}
      </div>
    {/if}
  {/if}
</div>
