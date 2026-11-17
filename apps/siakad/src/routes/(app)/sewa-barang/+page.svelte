<script lang="ts">
  import {
    Badge,
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
  } from '@algohub/ui';
  import AlertCircle from '@lucide/svelte/icons/circle-alert';
  import Calendar from '@lucide/svelte/icons/calendar';
  import CheckCircle2 from '@lucide/svelte/icons/circle-check';
  import Clock from '@lucide/svelte/icons/clock';
  import History from '@lucide/svelte/icons/history';
  import Info from '@lucide/svelte/icons/info';
  import MessageSquare from '@lucide/svelte/icons/message-square';
  import Search from '@lucide/svelte/icons/search';
  import ShoppingBag from '@lucide/svelte/icons/shopping-bag';
  import Tag from '@lucide/svelte/icons/tag';
  import XCircle from '@lucide/svelte/icons/circle-x';
  import { createQuery } from '@tanstack/svelte-query';
  import { format } from 'date-fns';
  import { id as localeId } from 'date-fns/locale';
  import { createRpcMutation, rpcQueryOptions } from '#lib/queries/rpc.ts';
  import { toast } from '#lib/toast.ts';
  import type { PageProps } from './$types';

  interface RenterItem {
    id: number;
    name: string;
    quantity: number;
    price_per_day: number;
    location?: string;
    condition?: string;
  }

  interface Rental {
    id: number;
    item_name: string;
    status: string;
    start_date: string;
    end_date: string;
    type: string;
    total_price: number;
  }

  let { data }: PageProps = $props();
  const user = $derived(data.auth.user);

  const itemsQuery = createQuery(() => rpcQueryOptions<RenterItem[]>('get_renter_items_secure'));
  const rentalsQuery = createQuery(() =>
    rpcQueryOptions<Rental[]>('get_my_rentals_secure', { p_user_id: user.id }),
  );
  const contactQuery = createQuery(() =>
    rpcQueryOptions<{ phone_number?: string }[]>('get_smart_validation_contact_secure', {
      p_student_id: user.id,
    }),
  );
  const submitRental = createRpcMutation('submit_rental_request_secure', [
    'get_renter_items_secure',
    'get_my_rentals_secure',
    'get_smart_validation_contact_secure',
  ]);

  const items = $derived(itemsQuery.data || []);
  const myRentals = $derived(rentalsQuery.data || []);
  const adminContact = $derived(
    contactQuery.data && contactQuery.data.length > 0 ? contactQuery.data[0] : null,
  );

  let search = $state('');
  let isRequestOpen = $state(false);
  let selectedItem = $state<RenterItem | null>(null);
  let form = $state({
    quantity: 1,
    startDate: format(new Date(), 'yyyy-MM-dd'),
    endDate: format(new Date(Date.now() + 86400000), 'yyyy-MM-dd'),
    type: 'sewa',
    notes: '',
  });

  const typeLabels: Record<string, string> = {
    sewa: 'Sewa (Berbayar)',
    pinjam: 'Pinjam (Praktik/Gratis)',
  };

  const filteredItems = $derived(
    items.filter((i) => i.name.toLowerCase().includes(search.toLowerCase())),
  );

  async function handleSubmitRequest() {
    if (!selectedItem) return;
    if (form.quantity > selectedItem.quantity) {
      toast.error('Stok tidak mencukupi');
      return;
    }
    const { error } = await submitRental.mutateAsync({
      p_user_id: user.id,
      p_item_id: selectedItem.id,
      p_quantity: parseInt(form.quantity.toString()),
      p_start_date: new Date(form.startDate).toISOString(),
      p_end_date: new Date(form.endDate).toISOString(),
      p_type: form.type,
      p_notes: form.notes,
    });
    if (error) {
      toast.error('Gagal mengirim permintaan: ' + error.message);
    } else {
      toast.success('Permintaan berhasil terkirim!');
      isRequestOpen = false;
    }
  }

  function openWA(rental: Rental | null = null) {
    if (!adminContact?.phone_number) {
      toast.error('Kontak admin tidak tersedia saat ini.');
      return;
    }
    const phone = adminContact.phone_number.replace(/\D/g, '').replace(/^0/, '62');
    const text = rental
      ? `Halo Admin Lab, saya ${user.full_name} ingin memverifikasi identitas (KTM) untuk pengajuan sewa/pinjam barang: ${rental.item_name} (ID: ${rental.id}).`
      : `Halo Admin Lab, saya ${user.full_name} ingin bertanya mengenai prosedur sewa barang di Lab Algoritma.`;
    window.open(`https://wa.me/${phone}?text=${encodeURIComponent(text)}`, '_blank');
  }
</script>

{#snippet statusBadge(status: string)}
  {#if status === 'pending'}
    <Badge variant="outline" class="border-yellow-200 bg-yellow-50 text-yellow-700"
      ><Clock class="mr-1 h-3 w-3" /> Pending</Badge
    >
  {:else if status === 'approved'}
    <Badge variant="outline" class="border-blue-200 bg-blue-50 text-blue-700"
      ><CheckCircle2 class="mr-1 h-3 w-3" /> Disetujui</Badge
    >
  {:else if status === 'active'}
    <Badge variant="outline" class="border-green-200 bg-green-50 text-green-700"
      ><ShoppingBag class="mr-1 h-3 w-3" /> Dipinjam</Badge
    >
  {:else if status === 'returned'}
    <Badge variant="outline" class="border-slate-200 bg-slate-50 text-slate-700">Selesai</Badge>
  {:else if status === 'rejected'}
    <Badge variant="outline" class="border-red-200 bg-red-50 text-red-700"
      ><XCircle class="mr-1 h-3 w-3" /> Ditolak</Badge
    >
  {:else}
    <Badge variant="secondary">{status}</Badge>
  {/if}
{/snippet}

<div class="animate-fade-in space-y-8">
  <div
    class="from-primary/10 border-primary/10 flex flex-col justify-between gap-4 rounded-2xl border bg-linear-to-r to-transparent p-6 md:flex-row md:items-center"
  >
    <div>
      <h1 class="text-3xl font-bold tracking-tight text-slate-900">Katalog Sewa & Pinjam</h1>
      <p class="mt-1 text-slate-500">Penyewaan barang untuk publik dan peminjaman alat praktik.</p>
    </div>
    <div class="flex gap-2">
      <Button
        variant="outline"
        onclick={() => openWA()}
        class="bg-white transition-all hover:border-green-200 hover:bg-green-50 hover:text-green-600"
      >
        <MessageSquare class="mr-2 h-4 w-4 text-green-500" /> Hubungi Admin
      </Button>
    </div>
  </div>

  <div class="grid grid-cols-1 gap-6 md:grid-cols-3">
    <Card class="border-primary/20 group overflow-hidden shadow-sm md:col-span-2">
      <CardHeader class="border-b bg-slate-50">
        <div class="flex items-center justify-between">
          <CardTitle class="flex items-center gap-2 text-lg">
            <Tag class="text-primary h-5 w-5" /> Katalog Barang
          </CardTitle>
          <div class="relative w-48 md:w-64">
            <Search class="text-muted-foreground absolute top-2.5 left-2.5 h-4 w-4" />
            <Input placeholder="Cari barang..." class="h-9 bg-white pl-9" bind:value={search} />
          </div>
        </div>
      </CardHeader>
      <CardContent class="p-0">
        <div class="overflow-x-auto">
          <Table>
            <TableHeader class="bg-slate-50/50">
              <TableRow>
                <TableHead class="font-bold">Nama Barang</TableHead>
                <TableHead class="font-bold">Tersedia</TableHead>
                <TableHead class="font-bold">Harga / Hari</TableHead>
                <TableHead class="text-right font-bold">Aksi</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {#if filteredItems.length === 0}
                <TableRow>
                  <TableCell colspan={4} class="text-muted-foreground py-10 text-center">
                    Tidak ada barang yang tersedia untuk disewa saat ini.
                  </TableCell>
                </TableRow>
              {:else}
                {#each filteredItems as item (item.id)}
                  <TableRow class="transition-colors hover:bg-slate-50/50">
                    <TableCell>
                      <div class="font-bold text-slate-800">{item.name}</div>
                      <div class="text-muted-foreground flex items-center gap-1 text-[10px] italic">
                        <Info class="h-3 w-3" />
                        {item.location} - {item.condition}
                      </div>
                    </TableCell>
                    <TableCell>
                      <Badge variant="secondary" class="bg-slate-200 font-medium">
                        {item.quantity} Unit
                      </Badge>
                    </TableCell>
                    <TableCell class="text-primary font-mono font-bold">
                      {item.price_per_day > 0
                        ? `Rp ${item.price_per_day.toLocaleString()}`
                        : 'Gratis (Pinjam)'}
                    </TableCell>
                    <TableCell class="text-right">
                      <Button
                        size="sm"
                        onclick={() => {
                          selectedItem = item;
                          isRequestOpen = true;
                        }}
                        class="rounded-full px-4 transition-all hover:scale-105 active:scale-95"
                      >
                        Ajukan
                      </Button>
                    </TableCell>
                  </TableRow>
                {/each}
              {/if}
            </TableBody>
          </Table>
        </div>
      </CardContent>
    </Card>

    <div class="space-y-6">
      <Card class="border-yellow-200 bg-yellow-50/30">
        <CardHeader class="pb-2">
          <CardTitle class="flex items-center gap-2 text-sm text-yellow-800">
            <AlertCircle class="h-4 w-4" /> Prosedur Sewa
          </CardTitle>
        </CardHeader>
        <CardContent class="space-y-2 text-[11px] leading-relaxed text-yellow-900/80 italic">
          <p>1. Pilih barang dari katalog dan isi detail peminjaman.</p>
          <p>
            2. Setelah mengajukan, klik tombol <strong>Hubungi Admin</strong> di riwayat untuk verifikasi
            identitas (Kirim foto KTM).
          </p>
          <p>3. Tunggu persetujuan status hingga berubah menjadi <strong>Disetujui</strong>.</p>
          <p>4. Barang dapat diambil di lab sesuai jadwal yang diajukan.</p>
        </CardContent>
      </Card>

      <Card class="shadow-sm">
        <CardHeader class="flex flex-row items-center justify-between space-y-0 border-b pb-3">
          <CardTitle class="flex items-center gap-2 text-sm font-bold">
            <History class="text-primary h-4 w-4" /> Pengajuan Saya
          </CardTitle>
        </CardHeader>
        <CardContent class="max-h-[400px] overflow-y-auto p-0">
          {#if myRentals.length === 0}
            <div class="text-muted-foreground px-4 py-10 text-center text-[11px]">
              Belum ada riwayat pengajuan.
            </div>
          {:else}
            <div class="divide-y">
              {#each myRentals as rental (rental.id)}
                <div class="space-y-2 p-4 transition-colors hover:bg-slate-50">
                  <div class="flex items-start justify-between">
                    <span class="text-xs font-bold">{rental.item_name}</span>
                    {@render statusBadge(rental.status)}
                  </div>
                  <div class="text-muted-foreground flex flex-col gap-1 text-[10px]">
                    <div class="flex items-center gap-1">
                      <Calendar class="h-3 w-3" />
                      {format(new Date(rental.start_date), 'dd MMM', { locale: localeId })} - {format(
                        new Date(rental.end_date),
                        'dd MMM yyyy',
                        { locale: localeId },
                      )}
                    </div>
                    <div class="flex items-center justify-between">
                      <Badge variant="outline" class="h-4 py-0 text-[9px] font-bold uppercase">
                        {rental.type}
                      </Badge>
                      <span class="text-primary font-bold"
                        >Rp {rental.total_price.toLocaleString()}</span
                      >
                    </div>
                  </div>
                  {#if rental.status === 'pending'}
                    <Button
                      variant="outline"
                      size="sm"
                      class="mt-1 h-7 w-full border-green-200 bg-green-50 text-[10px] text-green-700"
                      onclick={() => openWA(rental)}
                    >
                      <MessageSquare class="mr-1 h-3 w-3" /> Verifikasi KTM (WA)
                    </Button>
                  {/if}
                </div>
              {/each}
            </div>
          {/if}
        </CardContent>
      </Card>
    </div>
  </div>

  <Dialog bind:open={isRequestOpen}>
    <DialogContent class="max-w-md bg-white">
      <DialogHeader>
        <DialogTitle>Form Pengajuan {selectedItem?.name}</DialogTitle>
        <DialogDescription>Isi detail peminjaman/sewa barang di bawah ini.</DialogDescription>
      </DialogHeader>
      <div class="space-y-4 py-4">
        <div class="grid grid-cols-2 gap-4">
          <div class="space-y-1.5">
            <Label class="text-xs">Tipe Pengajuan</Label>
            <Select type="single" bind:value={form.type}>
              <SelectTrigger class="h-9 w-full">{typeLabels[form.type]}</SelectTrigger>
              <SelectContent class="bg-white">
                <SelectItem value="sewa" label="Sewa (Berbayar)" />
                <SelectItem value="pinjam" label="Pinjam (Praktik/Gratis)" />
              </SelectContent>
            </Select>
          </div>
          <div class="space-y-1.5">
            <Label for="rent-qty" class="text-xs">Jumlah Barang</Label>
            <Input
              id="rent-qty"
              type="number"
              value={form.quantity}
              oninput={(e) => (form.quantity = parseInt(e.currentTarget.value))}
              class="h-9"
              min={1}
              max={selectedItem?.quantity}
            />
          </div>
        </div>
        <div class="grid grid-cols-2 gap-4">
          <div class="space-y-1.5">
            <Label for="rent-start" class="text-xs">Tanggal Mulai</Label>
            <Input id="rent-start" type="date" bind:value={form.startDate} class="h-9" />
          </div>
          <div class="space-y-1.5">
            <Label for="rent-end" class="text-xs">Tanggal Selesai</Label>
            <Input id="rent-end" type="date" bind:value={form.endDate} class="h-9" />
          </div>
        </div>
        <div class="space-y-1.5">
          <Label for="rent-notes" class="text-muted-foreground text-xs"
            >Catatan / Tujuan (Opsional)</Label
          >
          <Input
            id="rent-notes"
            placeholder="Contoh: Untuk keperluan praktikum mata kuliah X"
            class="h-9"
            bind:value={form.notes}
          />
        </div>
        {#if form.type === 'sewa' && selectedItem && selectedItem.price_per_day > 0}
          <div
            class="bg-primary/5 border-primary/20 flex items-center justify-between rounded-lg border p-3"
          >
            <span class="text-xs font-medium text-slate-600">Estimasi Biaya:</span>
            <span class="text-primary text-lg font-bold">
              Rp {(selectedItem.price_per_day * form.quantity).toLocaleString()}
              <span class="text-muted-foreground text-[10px] font-normal">/ hari</span>
            </span>
          </div>
        {/if}
      </div>
      <DialogFooter>
        <Button variant="ghost" onclick={() => (isRequestOpen = false)}>Batal</Button>
        <Button onclick={handleSubmitRequest}>Ajukan Permintaan</Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</div>
