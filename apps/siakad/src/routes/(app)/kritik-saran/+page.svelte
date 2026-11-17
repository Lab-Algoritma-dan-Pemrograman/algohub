<script lang="ts">
  import {
    Badge,
    Button,
    Card,
    CardContent,
    CardDescription,
    CardHeader,
    CardTitle,
    Dialog,
    DialogContent,
    DialogHeader,
    DialogTitle,
    Input,
    Label,
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
    Textarea,
  } from '@algohub/ui';
  import Clock from '@lucide/svelte/icons/clock';
  import DownloadCloud from '@lucide/svelte/icons/cloud-download';
  import ExternalLink from '@lucide/svelte/icons/external-link';
  import Filter from '@lucide/svelte/icons/filter';
  import LinkIcon from '@lucide/svelte/icons/link';
  import MessageSquare from '@lucide/svelte/icons/message-square';
  import Plus from '@lucide/svelte/icons/plus';
  import Send from '@lucide/svelte/icons/send';
  import Trash2 from '@lucide/svelte/icons/trash-2';
  import User from '@lucide/svelte/icons/user';
  import { createQuery } from '@tanstack/svelte-query';
  import { createRpcMutation, rpcQueryOptions } from '#lib/queries/rpc.ts';
  import { canonRole } from '#lib/roles.ts';
  import { confirm } from '#lib/stores/confirm.svelte.ts';
  import { toast } from '#lib/toast.ts';
  import type { PageProps } from './$types';

  interface Feedback {
    id: number;
    created_at: string;
    category: string;
    content: string;
    user_full_name?: string;
    user_username?: string;
  }

  interface ExternalLinkRow {
    id: number;
    title: string;
    url: string;
  }

  let { data }: PageProps = $props();
  const user = $derived(data.auth.user);
  const isPraktikan = $derived(canonRole(user.role) === 'praktikan');
  const isStaff = $derived(user.role === 'koordinator' || user.role === 'asisten');

  // Akses tabel division_access langsung selalu 401 di produksi; check_menu_access_secure
  // menangani lookup divisi + menu yang sama lewat jalur RPC.
  const accessQuery = createQuery(() =>
    rpcQueryOptions<boolean>(
      'check_menu_access_secure',
      { p_viewer_id: user.id, p_menu_key: '/kritik-saran' },
      { enabled: user.role !== 'koordinator' && !!user.division },
    ),
  );
  const feedbackQuery = createQuery(() =>
    rpcQueryOptions<Feedback[]>('get_feedback_secure', { p_viewer_id: user.id }),
  );
  const linksQuery = createQuery(() =>
    rpcQueryOptions<ExternalLinkRow[]>('get_external_links_secure', { p_viewer_id: user.id }),
  );
  const refetched = ['get_feedback_secure', 'get_external_links_secure'];
  const insertFeedback = createRpcMutation('insert_feedback_secure', refetched);
  const upsertLink = createRpcMutation('upsert_external_link_secure', refetched);
  const deleteLink = createRpcMutation('delete_external_link_secure', refetched);

  const hasEditAccess = $derived(user.role === 'koordinator' || accessQuery.data === true);
  const feedbacks = $derived(
    (feedbackQuery.data || []).map((fb) => ({
      ...fb,
      users: { full_name: fb.user_full_name, username: fb.user_username },
    })),
  );
  const externalLinks = $derived(linksQuery.data || []);

  let loading = $state(false);
  let category = $state('Fasilitas');
  let content = $state('');
  let isLinkDialogOpen = $state(false);
  let linkTitle = $state('');
  let linkUrl = $state('');
  let filterCategory = $state('all');

  const categoryLabels: Record<string, string> = {
    Fasilitas: 'Fasilitas Lab (AC, PC, Kursi)',
    Pelayanan: 'Pelayanan Asisten',
    Modul: 'Modul Praktikum',
    Lainnya: 'Lainnya',
  };
  const filterLabels: Record<string, string> = {
    all: 'Semua Kategori',
    Fasilitas: 'Fasilitas',
    Pelayanan: 'Pelayanan',
    Modul: 'Modul',
    Lainnya: 'Lainnya',
  };

  const filteredFeedbacks = $derived(
    feedbacks.filter((fb) => filterCategory === 'all' || fb.category === filterCategory),
  );

  async function handleSubmit(e: SubmitEvent) {
    e.preventDefault();
    if (!content.trim()) return toast.error('Pesan tidak boleh kosong');
    loading = true;
    const { error } = await insertFeedback.mutateAsync({
      p_caller_id: user.id,
      p_category: category,
      p_content: content,
    });
    if (error) {
      toast.error('Gagal mengirim pesan: ' + error.message);
    } else {
      toast.success('Terima kasih! Masukan Anda telah terkirim.');
      content = '';
    }
    loading = false;
  }

  async function handleExportExcel() {
    if (feedbacks.length === 0) return toast.error('Tidak ada data untuk diexport');
    // T-30: pembungkus toast lama tidak meneruskan loading(), jadi di prod baris ini melempar
    // TypeError dan file tidak pernah diunduh. Dibiarkan sama sampai temuan diputuskan.
    const loadingToast = (toast as unknown as { loading: (m: string) => string }).loading(
      'Menyiapkan file Excel...',
    );
    try {
      const XLSX = await import('xlsx');
      const cleanData = filteredFeedbacks.map((fb, idx) => ({
        No: idx + 1,
        Tanggal: new Date(fb.created_at).toLocaleString('id-ID'),
        'Nama Pengirim': fb.users?.full_name || 'Anonim',
        'NIM/ID': fb.users?.username || '-',
        Kategori: fb.category,
        'Isi Pesan': fb.content,
      }));
      const worksheet = XLSX.utils.json_to_sheet(cleanData);
      const workbook = XLSX.utils.book_new();
      worksheet['!cols'] = [
        { wch: 5 },
        { wch: 20 },
        { wch: 25 },
        { wch: 15 },
        { wch: 20 },
        { wch: 50 },
      ];
      XLSX.utils.book_append_sheet(workbook, worksheet, 'Kritik dan Saran');
      const today = new Date().toISOString().split('T')[0];
      XLSX.writeFile(workbook, `Rekap_KritikSaran_${today}.xlsx`);
      toast.dismiss(loadingToast);
      toast.success('File Excel berhasil diunduh');
    } catch (err) {
      toast.dismiss(loadingToast);
      toast.error('Gagal export: ' + (err as Error).message);
    }
  }

  async function handleAddLink() {
    if (!linkTitle || !linkUrl) return toast.error('Judul dan URL harus diisi!');
    let finalUrl = linkUrl;
    if (!/^https?:\/\//i.test(finalUrl)) finalUrl = 'https://' + finalUrl;
    loading = true;
    try {
      const { error } = await upsertLink.mutateAsync({
        p_caller_id: user.id,
        p_id: 0,
        p_title: linkTitle,
        p_url: finalUrl,
        p_is_active: true,
      });
      if (error) throw error;
      toast.success('Link berhasil ditambahkan!');
      linkTitle = '';
      linkUrl = '';
      isLinkDialogOpen = false;
    } catch (err) {
      toast.error('Gagal menambah link: ' + (err as { message?: string }).message);
    } finally {
      loading = false;
    }
  }

  async function handleDeleteLink(id: number) {
    if (!(await confirm('Hapus link ini?'))) return;
    try {
      const { error } = await deleteLink.mutateAsync({ p_caller_id: user.id, p_id: id });
      if (error) throw error;
      toast.success('Link dihapus.');
    } catch (err) {
      toast.error('Gagal menghapus: ' + (err as { message?: string }).message);
    }
  }
</script>

<div class="mx-auto max-w-6xl space-y-6">
  <div class="flex flex-col items-start justify-between gap-4 sm:flex-row sm:items-center">
    <div>
      <h1 class="flex items-center gap-2 text-2xl font-bold">
        <MessageSquare class="text-primary h-8 w-8" />
        Kritik & Saran
      </h1>
      <p class="text-muted-foreground">
        {isPraktikan
          ? 'Sampaikan masukan untuk kemajuan laboratorium.'
          : 'Kelola dan pantau masukan dari praktikan.'}
      </p>
    </div>
    {#if hasEditAccess}
      <div class="flex gap-2">
        <Button
          onclick={handleExportExcel}
          variant="outline"
          class="border-green-600 text-green-700 hover:bg-green-50"
        >
          <DownloadCloud class="mr-2 h-4 w-4" /> Export Excel
        </Button>
      </div>
    {/if}
  </div>

  <div class="grid grid-cols-1 gap-6 lg:grid-cols-12">
    <div class="space-y-6 lg:col-span-4">
      <Card class="border-l-4 border-l-blue-500 shadow-md">
        <CardHeader class="flex flex-row items-center justify-between pb-3">
          <div>
            <CardTitle class="flex items-center gap-2 text-base"
              ><LinkIcon class="h-4 w-4 text-blue-600" /> Tautan Penting</CardTitle
            >
          </div>
          {#if hasEditAccess}
            <Button
              variant="ghost"
              size="sm"
              class="h-8 w-8 p-0 text-blue-600"
              aria-label="Tambah Link Form"
              onclick={() => (isLinkDialogOpen = true)}><Plus class="h-5 w-5" /></Button
            >
            <Dialog bind:open={isLinkDialogOpen}>
              <DialogContent>
                <DialogHeader><DialogTitle>Tambah Link Form</DialogTitle></DialogHeader>
                <div class="space-y-4 py-3">
                  <div class="space-y-2">
                    <Label for="link-title">Nama Form / Judul</Label>
                    <Input
                      id="link-title"
                      placeholder="Contoh: Form Pendaftaran Calon Asisten"
                      bind:value={linkTitle}
                    />
                  </div>
                  <div class="space-y-2">
                    <Label for="link-url">URL / Link (Google Form, dsb)</Label>
                    <Input id="link-url" placeholder="https://forms.gle/..." bind:value={linkUrl} />
                  </div>
                  <Button onclick={handleAddLink} disabled={loading} class="w-full"
                    >Simpan Link</Button
                  >
                </div>
              </DialogContent>
            </Dialog>
          {/if}
        </CardHeader>
        <CardContent>
          {#if externalLinks.length === 0}
            <p class="text-muted-foreground text-sm italic">Belum ada tautan aktif saat ini.</p>
          {:else}
            <div class="space-y-3">
              {#each externalLinks as link (link.id)}
                <div
                  class="group flex items-center justify-between rounded-md border border-transparent p-2 transition-all hover:border-slate-100 hover:bg-slate-50"
                >
                  <a
                    href={link.url}
                    target="_blank"
                    rel="noreferrer external"
                    class="flex flex-1 items-center gap-2 truncate text-sm font-medium text-blue-700 hover:underline"
                  >
                    <ExternalLink class="h-3.5 w-3.5 flex-shrink-0" />
                    <span class="truncate">{link.title}</span>
                  </a>
                  {#if hasEditAccess}
                    <Button
                      variant="ghost"
                      size="icon"
                      aria-label="Hapus link"
                      class="h-6 w-6 text-red-400 opacity-0 transition-opacity group-hover:opacity-100"
                      onclick={() => handleDeleteLink(link.id)}
                    >
                      <Trash2 class="h-3 w-3" />
                    </Button>
                  {/if}
                </div>
              {/each}
            </div>
          {/if}
        </CardContent>
      </Card>

      {#if isPraktikan}
        <Card class="border-l-primary border-l-4 shadow-md">
          <CardHeader>
            <CardTitle>Kirim Masukan Baru</CardTitle>
            <CardDescription class="text-xs"
              >Identitas Anda tercatat namun privasi tetap kami jaga.</CardDescription
            >
          </CardHeader>
          <CardContent>
            <form onsubmit={handleSubmit} class="space-y-4">
              <div class="space-y-2">
                <span class="text-sm font-medium">Kategori</span>
                <Select type="single" bind:value={category}>
                  <SelectTrigger class="w-full">{categoryLabels[category]}</SelectTrigger>
                  <SelectContent>
                    {#each Object.entries(categoryLabels) as [value, label] (value)}
                      <SelectItem {value} {label} />
                    {/each}
                  </SelectContent>
                </Select>
              </div>
              <div class="space-y-2">
                <label for="fb-content" class="text-sm font-medium">Isi Pesan</label>
                <Textarea
                  id="fb-content"
                  rows={5}
                  placeholder="Tuliskan kritik atau saran Anda secara detail..."
                  bind:value={content}
                />
              </div>
              <Button type="submit" disabled={loading} class="w-full">
                <Send class="mr-2 h-4 w-4" /> Kirim Masukan
              </Button>
            </form>
          </CardContent>
        </Card>
      {/if}
    </div>

    <div class="flex flex-col lg:col-span-8">
      <Card class="flex-1 shadow-md">
        <CardHeader
          class="flex flex-col justify-between gap-4 border-b bg-gray-50/50 pb-3 sm:flex-row sm:items-center"
        >
          <CardTitle class="text-lg">
            {isPraktikan ? 'Riwayat Masukan Saya' : 'Kotak Masuk (Semua Masukan)'}
          </CardTitle>
          {#if hasEditAccess}
            <div class="flex items-center gap-2 rounded-md border bg-white px-2 py-1 shadow-sm">
              <Filter class="text-muted-foreground h-4 w-4" />
              <Select type="single" bind:value={filterCategory}>
                <SelectTrigger
                  class="h-7 w-[150px] border-none px-1 text-sm font-medium shadow-none focus:ring-0"
                  >{filterLabels[filterCategory]}</SelectTrigger
                >
                <SelectContent>
                  {#each Object.entries(filterLabels) as [value, label] (value)}
                    <SelectItem {value} {label} />
                  {/each}
                </SelectContent>
              </Select>
            </div>
          {/if}
        </CardHeader>
        <CardContent class="px-0 pt-0">
          <div class="max-h-[600px] overflow-x-auto overflow-y-auto">
            <Table>
              <TableHeader class="sticky top-0 z-10 bg-gray-100 shadow-sm">
                <TableRow>
                  <TableHead class="w-[120px] text-xs">Tanggal</TableHead>
                  {#if isStaff}<TableHead class="text-xs">Pengirim</TableHead>{/if}
                  <TableHead class="w-[120px] text-xs">Kategori</TableHead>
                  <TableHead class="text-xs">Isi Pesan</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {#if filteredFeedbacks.length === 0}
                  <TableRow>
                    <TableCell
                      colspan={isStaff ? 4 : 3}
                      class="text-muted-foreground py-12 text-center"
                    >
                      Belum ada data masukan.
                    </TableCell>
                  </TableRow>
                {:else}
                  {#each filteredFeedbacks as fb (fb.id)}
                    <TableRow class="transition-colors hover:bg-gray-50/50">
                      <TableCell
                        class="text-muted-foreground pt-4 align-top text-[11px] whitespace-nowrap"
                      >
                        <div class="mt-0.5 flex items-center gap-1">
                          <Clock class="h-3 w-3" />
                          {new Date(fb.created_at).toLocaleDateString('id-ID')}
                        </div>
                      </TableCell>
                      {#if isStaff}
                        <TableCell class="pt-3 align-top whitespace-nowrap">
                          <div class="flex items-center gap-1.5 text-sm font-medium text-gray-800">
                            <User class="text-muted-foreground h-3.5 w-3.5" />
                            {fb.users?.full_name || 'Anonim'}
                          </div>
                          <div class="text-muted-foreground mt-0.5 pl-5 font-mono text-[11px]">
                            {fb.users?.username || '-'}
                          </div>
                        </TableCell>
                      {/if}
                      <TableCell class="pt-3 align-top">
                        <Badge variant="outline" class="bg-white text-[10px]">{fb.category}</Badge>
                      </TableCell>
                      <TableCell
                        class="max-w-md pt-3 align-top text-sm leading-relaxed break-words text-gray-700"
                      >
                        {fb.content}
                      </TableCell>
                    </TableRow>
                  {/each}
                {/if}
              </TableBody>
            </Table>
          </div>
        </CardContent>
      </Card>
    </div>
  </div>
</div>
