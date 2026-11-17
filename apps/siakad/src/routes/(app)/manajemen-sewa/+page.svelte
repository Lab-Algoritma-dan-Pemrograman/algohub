<script lang="ts">
  import {
    Badge,
    Button,
    Card,
    Input,
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
    Tabs,
    TabsContent,
    TabsList,
    TabsTrigger,
  } from '@algohub/ui';
  import Check from '@lucide/svelte/icons/check';
  import CheckCircle2 from '@lucide/svelte/icons/circle-check';
  import Clock from '@lucide/svelte/icons/clock';
  import MessageSquare from '@lucide/svelte/icons/message-square';
  import PackageCheck from '@lucide/svelte/icons/package-check';
  import RefreshCcw from '@lucide/svelte/icons/refresh-ccw';
  import Search from '@lucide/svelte/icons/search';
  import ShoppingBag from '@lucide/svelte/icons/shopping-bag';
  import X from '@lucide/svelte/icons/x';
  import { createQuery } from '@tanstack/svelte-query';
  import { format } from 'date-fns';
  import { id as localeId } from 'date-fns/locale';
  import { createRpcMutation, rpcQueryOptions } from '#lib/queries/rpc.ts';
  import { confirm } from '#lib/stores/confirm.svelte.ts';
  import { toast } from '#lib/toast.ts';
  import type { PageProps } from './$types';

  interface Rental {
    id: number;
    renter_name: string;
    renter_phone: string;
    item_name: string;
    type: string;
    quantity: number;
    start_date: string;
    end_date: string;
    total_price: number;
    status: string;
  }

  let { data }: PageProps = $props();
  const user = $derived(data.auth.user);

  const rentalsQuery = createQuery(() =>
    rpcQueryOptions<Rental[]>('get_admin_rentals_secure', { p_caller_id: user.id }),
  );
  const updateStatus = createRpcMutation('update_rental_status_secure', [
    'get_admin_rentals_secure',
  ]);

  let search = $state('');

  async function handleUpdateStatus(rentalId: number, status: string) {
    if (status === 'rejected' && !(await confirm('Tolak pengajuan ini?'))) return;
    const { error } = await updateStatus.mutateAsync({
      p_caller_id: user.id,
      p_rental_id: rentalId,
      p_new_status: status,
    });
    if (error) toast.error('Gagal update: ' + error.message);
    else toast.success(`Status berhasil diperbarui menjadi ${status}`);
  }

  const statusColor: Record<string, string> = {
    pending: 'bg-yellow-100 text-yellow-700 border-yellow-200',
    approved: 'bg-blue-100 text-blue-700 border-blue-200',
    active: 'bg-green-100 text-green-700 border-green-200',
    returned: 'bg-slate-100 text-slate-700 border-slate-200',
    rejected: 'bg-red-100 text-red-700 border-red-200',
  };

  const filtered = $derived(
    (rentalsQuery.data || []).filter(
      (r) =>
        r.renter_name.toLowerCase().includes(search.toLowerCase()) ||
        r.item_name.toLowerCase().includes(search.toLowerCase()),
    ),
  );
  const pendingList = $derived(filtered.filter((r) => r.status === 'pending'));
  const approvedList = $derived(filtered.filter((r) => r.status === 'approved'));
  const activeList = $derived(filtered.filter((r) => r.status === 'active'));
  const finishedList = $derived(
    filtered.filter((r) => ['returned', 'rejected'].includes(r.status)),
  );

  function openWA(phoneNumber: string, rental: Rental) {
    const phone = phoneNumber.replace(/\D/g, '').replace(/^0/, '62');
    const text = `Halo ${rental.renter_name}, kami dari Admin Lab Algoritma menghubungi mengenai pengajuan ${rental.type} Anda untuk barang: ${rental.item_name}.`;
    window.open(`https://wa.me/${phone}?text=${encodeURIComponent(text)}`, '_blank');
  }

  const triggerClass =
    'rounded-lg gap-2 data-[state=active]:bg-white data-[state=active]:shadow-sm';
</script>

{#snippet rentalTable(list: Rental[])}
  <div class="overflow-x-auto">
    <Table>
      <TableHeader class="bg-slate-50">
        <TableRow>
          <TableHead class="font-bold whitespace-nowrap">Penyewa</TableHead>
          <TableHead class="font-bold whitespace-nowrap">Barang</TableHead>
          <TableHead class="font-bold whitespace-nowrap">Waktu</TableHead>
          <TableHead class="font-bold whitespace-nowrap">Biaya</TableHead>
          <TableHead class="text-right font-bold whitespace-nowrap">Aksi</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {#if list.length === 0}
          <TableRow>
            <TableCell colspan={5} class="text-muted-foreground py-10 text-center italic">
              Tidak ada data di kategori ini.
            </TableCell>
          </TableRow>
        {:else}
          {#each list as r (r.id)}
            <TableRow class="transition-colors hover:bg-slate-50/50">
              <TableCell>
                <div class="font-bold text-slate-900">{r.renter_name}</div>
                <div class="text-muted-foreground flex items-center gap-1 text-[10px]">
                  <MessageSquare class="h-3 w-3" />
                  {r.renter_phone}
                </div>
              </TableCell>
              <TableCell>
                <div class="font-medium">{r.item_name}</div>
                <div class="flex items-center gap-1">
                  <Badge variant="outline" class="h-4 bg-slate-50 py-0 text-[9px] uppercase"
                    >{r.type}</Badge
                  >
                  <span class="text-muted-foreground text-[10px]">x{r.quantity} Unit</span>
                </div>
              </TableCell>
              <TableCell>
                <div class="text-[11px] font-medium text-slate-700">
                  {format(new Date(r.start_date), 'dd MMM', { locale: localeId })} - {format(
                    new Date(r.end_date),
                    'dd MMM yyyy',
                    { locale: localeId },
                  )}
                </div>
              </TableCell>
              <TableCell>
                <div class="text-primary text-xs font-bold">
                  Rp {r.total_price.toLocaleString()}
                </div>
              </TableCell>
              <TableCell class="text-right">
                <div class="flex justify-end gap-1">
                  <Button
                    variant="outline"
                    size="icon"
                    aria-label="Hubungi via WhatsApp"
                    class="h-8 w-8 hover:bg-green-50 hover:text-green-600"
                    onclick={() => openWA(r.renter_phone, r)}
                  >
                    <MessageSquare class="h-4 w-4" />
                  </Button>
                  {#if r.status === 'pending'}
                    <Button
                      variant="outline"
                      size="icon"
                      aria-label="Setujui"
                      class="h-8 w-8 border-blue-200 bg-blue-50 text-blue-600 hover:bg-blue-100"
                      onclick={() => handleUpdateStatus(r.id, 'approved')}
                    >
                      <Check class="h-4 w-4" />
                    </Button>
                    <Button
                      variant="outline"
                      size="icon"
                      aria-label="Tolak"
                      class="h-8 w-8 border-red-200 bg-red-50 text-red-600 hover:bg-red-100"
                      onclick={() => handleUpdateStatus(r.id, 'rejected')}
                    >
                      <X class="h-4 w-4" />
                    </Button>
                  {/if}
                  {#if r.status === 'approved'}
                    <Button
                      variant="default"
                      size="sm"
                      class="h-8 bg-green-600 text-[11px] font-bold hover:bg-green-700"
                      onclick={() => handleUpdateStatus(r.id, 'active')}
                    >
                      Konfirmasi Ambil
                    </Button>
                  {/if}
                  {#if r.status === 'active'}
                    <Button
                      variant="default"
                      size="sm"
                      class="bg-primary hover:bg-primary/90 h-8 text-[11px] font-bold"
                      onclick={() => handleUpdateStatus(r.id, 'returned')}
                    >
                      Konfirmasi Kembali
                    </Button>
                  {/if}
                  {#if ['returned', 'rejected'].includes(r.status)}
                    <Badge class={statusColor[r.status] ?? ''} variant="outline">
                      {r.status === 'returned' ? 'Selesai' : 'Ditolak'}
                    </Badge>
                  {/if}
                </div>
              </TableCell>
            </TableRow>
          {/each}
        {/if}
      </TableBody>
    </Table>
  </div>
{/snippet}

<div class="animate-fade-in space-y-6">
  <div class="flex flex-col items-start justify-between gap-4 md:flex-row md:items-center">
    <h1 class="flex items-center gap-2 text-2xl font-bold">
      <ShoppingBag class="text-primary" /> Manajemen Sewa & Pinjam
    </h1>
    <div class="relative w-full md:w-64">
      <Search class="text-muted-foreground absolute top-2.5 left-2.5 h-4 w-4" />
      <Input placeholder="Cari penyewa/barang..." class="h-10 w-full pl-9" bind:value={search} />
    </div>
  </div>

  <Tabs value="pending" class="w-full">
    <TabsList class="grid h-11 w-full grid-cols-4 rounded-xl bg-slate-100 p-1">
      <TabsTrigger value="pending" class={triggerClass}>
        <Clock class="h-4 w-4" />
        <span class="hidden sm:inline">Pengajuan Baru</span>
        {#if pendingList.length > 0}<Badge
            class="ml-0 scale-75 bg-yellow-500 hover:bg-yellow-500 md:ml-1"
            >{pendingList.length}</Badge
          >{/if}
      </TabsTrigger>
      <TabsTrigger value="approved" class={triggerClass}>
        <CheckCircle2 class="h-4 w-4" />
        <span class="hidden sm:inline">Menunggu Ambil</span>
        {#if approvedList.length > 0}<Badge
            class="ml-0 scale-75 bg-blue-500 hover:bg-blue-500 md:ml-1">{approvedList.length}</Badge
          >{/if}
      </TabsTrigger>
      <TabsTrigger value="active" class={triggerClass}>
        <PackageCheck class="h-4 w-4" />
        <span class="hidden sm:inline">Sedang Dipinjam</span>
        {#if activeList.length > 0}<Badge
            class="ml-0 scale-75 bg-green-500 hover:bg-green-500 md:ml-1">{activeList.length}</Badge
          >{/if}
      </TabsTrigger>
      <TabsTrigger value="history" class={triggerClass}>
        <RefreshCcw class="h-4 w-4" />
        <span class="hidden sm:inline">Riwayat</span>
      </TabsTrigger>
    </TabsList>
    <Card class="shadow-elevated mt-4 overflow-hidden border-0">
      <TabsContent value="pending" class="m-0">{@render rentalTable(pendingList)}</TabsContent>
      <TabsContent value="approved" class="m-0">{@render rentalTable(approvedList)}</TabsContent>
      <TabsContent value="active" class="m-0">{@render rentalTable(activeList)}</TabsContent>
      <TabsContent value="history" class="m-0">{@render rentalTable(finishedList)}</TabsContent>
    </Card>
  </Tabs>
</div>
