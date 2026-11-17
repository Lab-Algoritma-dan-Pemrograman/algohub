<script lang="ts">
  import {
    Button,
    Card,
    CardContent,
    CardHeader,
    CardTitle,
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
    Input,
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
  } from '@algohub/ui';
  import DollarSign from '@lucide/svelte/icons/dollar-sign';
  import FileDown from '@lucide/svelte/icons/file-down';
  import Plus from '@lucide/svelte/icons/plus';
  import Trash2 from '@lucide/svelte/icons/trash-2';
  import TrendingDown from '@lucide/svelte/icons/trending-down';
  import TrendingUp from '@lucide/svelte/icons/trending-up';
  import { createQuery } from '@tanstack/svelte-query';
  import { createRpcMutation, rpcQueryOptions } from '#lib/queries/rpc.ts';
  import { confirm } from '#lib/stores/confirm.svelte.ts';
  import { toast } from '#lib/toast.ts';
  import type { PageProps } from './$types';

  interface FinancialRecord {
    id: number;
    date?: string;
    created_at?: string;
    title?: string;
    category?: string;
    type: string;
    amount: number | string;
  }

  const TYPE_LABELS: Record<string, string> = {
    pemasukan: 'Pemasukan',
    pengeluaran: 'Pengeluaran',
  };

  let { data }: PageProps = $props();
  const user = $derived(data.auth.user);

  const accessQuery = createQuery(() =>
    rpcQueryOptions<boolean>('check_menu_access_secure', {
      p_viewer_id: user.id,
      p_menu_key: '/laporan-keuangan',
    }),
  );
  const hasExportAccess = $derived(!accessQuery.error && !!accessQuery.data);
  const recordsQuery = createQuery(() =>
    rpcQueryOptions<FinancialRecord[]>('get_financial_records_secure', { p_viewer_id: user.id }),
  );
  const refetched = ['get_financial_records_secure'];
  const upsertRecord = createRpcMutation('upsert_financial_record_secure', refetched);
  const deleteRecord = createRpcMutation('delete_financial_record_secure', refetched);

  const records = $derived(recordsQuery.data || []);
  let isOpen = $state(false);
  let form = $state({ title: '', type: 'pemasukan', amount: 0, category: 'Kas' });

  const sum = (type: string) =>
    records.filter((r) => r.type === type).reduce((acc, curr) => acc + Number(curr.amount), 0);
  const totalMasuk = $derived(sum('pemasukan'));
  const totalKeluar = $derived(sum('pengeluaran'));
  const saldo = $derived(totalMasuk - totalKeluar);

  async function handleSave() {
    const { error } = await upsertRecord.mutateAsync({
      p_caller_id: user.id,
      p_id: 0,
      p_title: form.title,
      p_amount: form.amount,
      p_type: form.type,
      p_category: form.category,
    });
    if (error) toast.error('Gagal menyimpan: ' + error.message);
    else {
      toast.success('Data tersimpan');
      isOpen = false;
    }
  }

  async function handleDelete(id: number) {
    if (!(await confirm('Hapus data ini?'))) return;
    const { error } = await deleteRecord.mutateAsync({ p_caller_id: user.id, p_id: id });
    if (error) toast.error('Gagal menghapus: ' + error.message);
    else toast.success('Data dihapus');
  }

  async function handleExportExcel() {
    if (records.length === 0) return toast.warning('Tidak ada data transaksi untuk diekspor.');
    const XLSX = await import('xlsx');
    const dataToExport: Record<string, unknown>[] = records.map((r, index) => ({
      No: index + 1,
      Tanggal: new Date(r.date as string).toLocaleDateString('id-ID'),
      Keterangan: r.title,
      Kategori: r.category,
      Tipe: r.type.toUpperCase(),
      'Pemasukan (Rp)': r.type === 'pemasukan' ? Number(r.amount) : 0,
      'Pengeluaran (Rp)': r.type === 'pengeluaran' ? Number(r.amount) : 0,
    }));
    dataToExport.push({});
    dataToExport.push({ Keterangan: 'TOTAL PEMASUKAN', 'Pemasukan (Rp)': totalMasuk });
    dataToExport.push({ Keterangan: 'TOTAL PENGELUARAN', 'Pengeluaran (Rp)': totalKeluar });
    dataToExport.push({ Keterangan: 'SALDO AKHIR TERSISA', 'Pemasukan (Rp)': saldo });
    const ws = XLSX.utils.json_to_sheet(dataToExport);
    const wb = XLSX.utils.book_new();
    XLSX.utils.book_append_sheet(wb, ws, 'Laporan Kas Lab');
    XLSX.writeFile(wb, `Laporan_Keuangan_Lab_${new Date().toISOString().split('T')[0]}.xlsx`);
    toast.success('File Excel berhasil di-download!');
  }

  const recordDate = (r: FinancialRecord) =>
    r.date
      ? new Date(r.date).toLocaleDateString('id-ID')
      : r.created_at
        ? new Date(r.created_at).toLocaleDateString('id-ID')
        : '-';
</script>

<div class="space-y-6">
  <h1 class="flex items-center gap-2 text-2xl font-bold">
    <DollarSign class="text-primary" /> Laporan Keuangan
  </h1>

  <div class="grid grid-cols-1 gap-4 md:grid-cols-3">
    <Card>
      <CardHeader class="pb-2"
        ><CardTitle class="text-muted-foreground text-sm font-medium">Total Pemasukan</CardTitle
        ></CardHeader
      >
      <CardContent
        ><div class="flex items-center gap-2 text-2xl font-bold text-green-600">
          <TrendingUp class="h-5 w-5" /> Rp {totalMasuk.toLocaleString()}
        </div></CardContent
      >
    </Card>
    <Card>
      <CardHeader class="pb-2"
        ><CardTitle class="text-muted-foreground text-sm font-medium">Total Pengeluaran</CardTitle
        ></CardHeader
      >
      <CardContent
        ><div class="flex items-center gap-2 text-2xl font-bold text-red-600">
          <TrendingDown class="h-5 w-5" /> Rp {totalKeluar.toLocaleString()}
        </div></CardContent
      >
    </Card>
    <Card class="bg-primary/5 border-primary/20">
      <CardHeader class="pb-2"
        ><CardTitle class="text-primary text-sm font-medium">Saldo Akhir</CardTitle></CardHeader
      >
      <CardContent
        ><div class="text-primary text-2xl font-bold">Rp {saldo.toLocaleString()}</div></CardContent
      >
    </Card>
  </div>

  <Card>
    <CardHeader class="flex flex-col justify-between gap-4 sm:flex-row sm:items-center">
      <CardTitle>Riwayat Transaksi</CardTitle>
      <div class="flex flex-wrap gap-2">
        {#if hasExportAccess}
          <Button
            variant="outline"
            class="border-green-600 text-green-700 hover:bg-green-50"
            onclick={handleExportExcel}
          >
            <FileDown class="mr-2 h-4 w-4" /> Export Excel
          </Button>
        {/if}
        <Button onclick={() => (isOpen = true)}
          ><Plus class="mr-2 h-4 w-4" /> Catat Transaksi</Button
        >
      </div>
    </CardHeader>
    <CardContent>
      <div class="overflow-x-auto">
        <Table>
          <TableHeader
            ><TableRow
              ><TableHead>Tanggal</TableHead><TableHead>Keterangan</TableHead><TableHead
                >Kategori</TableHead
              ><TableHead>Tipe</TableHead><TableHead class="text-right">Jumlah</TableHead><TableHead
              ></TableHead></TableRow
            ></TableHeader
          >
          <TableBody>
            {#if records.length === 0}
              <TableRow
                ><TableCell colspan={6} class="text-muted-foreground py-6 text-center"
                  >Belum ada catatan transaksi.</TableCell
                ></TableRow
              >
            {:else}
              {#each records as r (r.id)}
                <TableRow>
                  <TableCell>{recordDate(r)}</TableCell>
                  <TableCell class="font-medium">{r.title || 'Tanpa Keterangan'}</TableCell>
                  <TableCell>{r.category}</TableCell>
                  <TableCell>
                    <span
                      class="rounded px-2 py-1 text-xs font-bold {r.type === 'pemasukan'
                        ? 'bg-green-50 text-green-600'
                        : 'bg-red-50 text-red-600'}">{r.type.toUpperCase()}</span
                    >
                  </TableCell>
                  <TableCell class="text-right font-mono"
                    >Rp {Number(r.amount).toLocaleString()}</TableCell
                  >
                  <TableCell class="text-right"
                    ><Button
                      variant="ghost"
                      size="sm"
                      aria-label="Hapus transaksi"
                      onclick={() => handleDelete(r.id)}
                      ><Trash2 class="h-4 w-4 text-red-400 hover:text-red-600" /></Button
                    ></TableCell
                  >
                </TableRow>
              {/each}
            {/if}
          </TableBody>
        </Table>
      </div>
    </CardContent>
  </Card>

  <Dialog bind:open={isOpen}>
    <DialogContent>
      <DialogHeader>
        <DialogTitle>Catat Keuangan</DialogTitle>
        <DialogDescription>Masukkan detail transaksi keuangan kas laboratorium.</DialogDescription>
      </DialogHeader>
      <div class="space-y-4 py-2">
        <Input placeholder="Keterangan (misal: Beli Spidol)" bind:value={form.title} />
        <div class="grid grid-cols-2 gap-4">
          <Select type="single" bind:value={form.type}>
            <SelectTrigger class="w-full">{TYPE_LABELS[form.type]}</SelectTrigger>
            <SelectContent>
              {#each Object.entries(TYPE_LABELS) as [value, label] (value)}<SelectItem
                  {value}
                  {label}
                />{/each}
            </SelectContent>
          </Select>
          <Input
            type="number"
            placeholder="Jumlah (Rp)"
            oninput={(e) => (form.amount = Number(e.currentTarget.value))}
          />
        </div>
        <Input placeholder="Kategori (misal: Kas, Inventaris)" bind:value={form.category} />
      </div>
      <DialogFooter><Button onclick={handleSave}>Simpan</Button></DialogFooter>
    </DialogContent>
  </Dialog>
</div>
