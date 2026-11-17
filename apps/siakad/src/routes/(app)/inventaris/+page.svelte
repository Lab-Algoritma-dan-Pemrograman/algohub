<script lang="ts">
  import {
    Button,
    Card,
    CardContent,
    CardHeader,
    Dialog,
    DialogContent,
    DialogFooter,
    DialogHeader,
    DialogTitle,
    Input,
    Label,
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    Switch,
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
  } from '@algohub/ui';
  import Box from '@lucide/svelte/icons/box';
  import DollarSign from '@lucide/svelte/icons/dollar-sign';
  import Pencil from '@lucide/svelte/icons/pencil';
  import Plus from '@lucide/svelte/icons/plus';
  import Search from '@lucide/svelte/icons/search';
  import Share2 from '@lucide/svelte/icons/share-2';
  import Trash2 from '@lucide/svelte/icons/trash-2';
  import { createQuery } from '@tanstack/svelte-query';
  import { createRpcMutation, rpcQueryOptions } from '#lib/queries/rpc.ts';
  import { confirm } from '#lib/stores/confirm.svelte.ts';
  import { toast } from '#lib/toast.ts';
  import type { PageProps } from './$types';

  interface Item {
    id?: number | string;
    name: string;
    condition: string;
    quantity: number | string;
    location: string;
    is_rentable: boolean;
    price_per_day: number | string;
  }

  const CONDITIONS = ['Baik', 'Rusak Ringan', 'Rusak Berat'];
  const emptyForm = (): Item => ({
    name: '',
    condition: 'Baik',
    quantity: 1,
    location: '',
    is_rentable: false,
    price_per_day: 0,
  });

  let { data }: PageProps = $props();
  const user = $derived(data.auth.user);

  const itemsQuery = createQuery(() =>
    rpcQueryOptions<Item[]>('get_inventory_items_secure', { p_viewer_id: user.id }),
  );
  const refetched = ['get_inventory_items_secure'];
  const upsertItem = createRpcMutation('upsert_inventory_item_secure', refetched);
  const deleteItem = createRpcMutation('delete_inventory_item_secure', refetched);

  let search = $state('');
  let isOpen = $state(false);
  let form = $state<Item>(emptyForm());

  const filtered = $derived(
    (itemsQuery.data || []).filter((i) => i.name.toLowerCase().includes(search.toLowerCase())),
  );

  async function handleSave() {
    const { error } = await upsertItem.mutateAsync({
      p_caller_id: user.id,
      p_id: form.id?.toString() || '0',
      p_name: form.name,
      p_condition: form.condition,
      p_quantity: parseInt(String(form.quantity)),
      p_location: form.location,
      p_is_rentable: form.is_rentable || false,
      p_price_per_day: parseFloat(String(form.price_per_day)) || 0,
    });
    if (error) toast.error('Gagal menyimpan: ' + error.message);
    else {
      toast.success('Berhasil disimpan');
      isOpen = false;
    }
  }

  async function handleDelete(id: string | number) {
    if (!(await confirm('Hapus barang ini?'))) return;
    const { error } = await deleteItem.mutateAsync({ p_caller_id: user.id, p_id: id.toString() });
    if (error) toast.error('Gagal menghapus: ' + error.message);
    else toast.success('Barang dihapus');
  }
</script>

<div class="space-y-6">
  <div class="flex items-center justify-between">
    <h1 class="flex gap-2 text-2xl font-bold"><Box class="text-primary" /> Inventaris Lab</h1>
    <Button
      onclick={() => {
        form = emptyForm();
        isOpen = true;
      }}><Plus class="mr-2 h-4 w-4" /> Tambah Barang</Button
    >
  </div>

  <Card>
    <CardHeader class="pb-3">
      <div class="relative">
        <Search class="text-muted-foreground absolute top-2.5 left-2 h-4 w-4" />
        <Input placeholder="Cari barang..." class="pl-8" bind:value={search} />
      </div>
    </CardHeader>
    <CardContent>
      <div class="overflow-x-auto">
        <Table>
          <TableHeader
            ><TableRow
              ><TableHead>Nama Barang</TableHead><TableHead>Kondisi</TableHead><TableHead
                >Jml</TableHead
              ><TableHead>Lokasi</TableHead><TableHead>Sewa</TableHead><TableHead class="text-right"
                >Aksi</TableHead
              ></TableRow
            ></TableHeader
          >
          <TableBody>
            {#each filtered as item (item.id)}
              <TableRow data-testid="inventory-row">
                <TableCell class="font-medium">{item.name}</TableCell>
                <TableCell
                  ><span
                    class="rounded border px-2 py-1 text-xs {item.condition === 'Baik'
                      ? 'border-green-200 bg-green-50 text-green-700'
                      : 'border-red-200 bg-red-50 text-red-700'}">{item.condition}</span
                  ></TableCell
                >
                <TableCell>{item.quantity}</TableCell>
                <TableCell>{item.location}</TableCell>
                <TableCell>
                  {#if item.is_rentable}
                    <div class="flex flex-col items-start gap-1">
                      <span
                        class="bg-primary/10 text-primary rounded px-1.5 py-0.5 text-[10px] font-bold"
                        >YES</span
                      >
                      <span class="text-muted-foreground text-[10px] whitespace-nowrap"
                        >Rp {item.price_per_day?.toLocaleString()}</span
                      >
                    </div>
                  {:else}
                    <span class="text-[10px] font-bold text-slate-300">NO</span>
                  {/if}
                </TableCell>
                <TableCell class="text-right">
                  <Button
                    variant="ghost"
                    size="icon"
                    aria-label="Edit barang"
                    onclick={() => {
                      form = { ...item };
                      isOpen = true;
                    }}><Pencil class="h-4 w-4 text-blue-500" /></Button
                  >
                  <Button
                    variant="ghost"
                    size="icon"
                    aria-label="Hapus barang"
                    onclick={() => item.id !== undefined && handleDelete(item.id)}
                    ><Trash2 class="h-4 w-4 text-red-500" /></Button
                  >
                </TableCell>
              </TableRow>
            {/each}
          </TableBody>
        </Table>
      </div>
    </CardContent>
  </Card>

  <Dialog bind:open={isOpen}>
    <DialogContent>
      <DialogHeader><DialogTitle>Data Barang</DialogTitle></DialogHeader>
      <div class="space-y-4 py-2">
        <Input placeholder="Nama Barang" bind:value={form.name} />
        <div class="grid grid-cols-2 gap-4">
          <Select type="single" bind:value={form.condition}>
            <SelectTrigger class="w-full">{form.condition}</SelectTrigger>
            <SelectContent>
              {#each CONDITIONS as c (c)}<SelectItem value={c} label={c} />{/each}
            </SelectContent>
          </Select>
          <Input type="number" placeholder="Jumlah" bind:value={form.quantity} />
        </div>
        <Input placeholder="Lokasi Penyimpanan" bind:value={form.location} />
        <div
          class="flex items-center justify-between rounded-lg border border-dashed bg-slate-50 p-3 shadow-sm"
        >
          <div class="flex items-center gap-3">
            <div
              class="rounded-full p-2 {form.is_rentable
                ? 'bg-primary/20 text-primary'
                : 'bg-slate-200 text-slate-400'}"
            >
              <Share2 class="h-4 w-4" />
            </div>
            <div>
              <Label for="is-rentable" class="text-sm font-bold">Disediakan untuk Sewa/Pinjam</Label
              >
              <p class="text-muted-foreground text-[10px]">
                Aktifkan agar muncul di katalog penyewa umum.
              </p>
            </div>
          </div>
          <Switch id="is-rentable" bind:checked={form.is_rentable} />
        </div>
        {#if form.is_rentable}
          <div class="animate-in fade-in slide-in-from-top-2 duration-300">
            <Label for="price-per-day" class="mb-1 block text-xs">Harga Sewa (Rp / Hari)</Label>
            <div class="relative">
              <DollarSign class="text-muted-foreground absolute top-2.5 left-2.5 h-4 w-4" />
              <Input
                id="price-per-day"
                type="number"
                placeholder="Contoh: 10000"
                class="pl-9"
                bind:value={form.price_per_day}
              />
            </div>
          </div>
        {/if}
      </div>
      <DialogFooter><Button onclick={handleSave}>Simpan</Button></DialogFooter>
    </DialogContent>
  </Dialog>
</div>
