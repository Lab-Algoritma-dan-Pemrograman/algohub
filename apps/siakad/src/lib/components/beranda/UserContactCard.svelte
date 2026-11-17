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
    Label,
  } from '@algohub/ui';
  import Loader2 from '@lucide/svelte/icons/loader-2';
  import Pencil from '@lucide/svelte/icons/pencil';
  import Phone from '@lucide/svelte/icons/phone';
  import Save from '@lucide/svelte/icons/save';
  import { createQuery } from '@tanstack/svelte-query';
  import { dashboardStatsQuery, type DashboardStats } from '#lib/queries/beranda.ts';
  import { createRpcMutation } from '#lib/queries/rpc.ts';
  import { toast } from '#lib/toast.ts';

  let { userId }: { userId: number } = $props();

  const stats = createQuery(() => dashboardStatsQuery(userId));
  const updateProfile = createRpcMutation('update_user_profile_secure', [
    'get_dashboard_stats_secure',
  ]);
  const phone = $derived((stats.data as DashboardStats | null)?.user_phone || '');
  let newPhone = $state('');
  let isDialogOpen = $state(false);

  $effect(() => {
    if (phone) newPhone = phone;
  });

  async function handleSavePhone() {
    if (!newPhone) return toast.error('Nomor HP tidak boleh kosong');
    try {
      const { error } = await updateProfile.mutateAsync({
        p_caller_id: userId,
        p_phone_number: newPhone,
      });
      if (error) throw error;
      toast.success('Nomor WhatsApp berhasil diperbarui!');
      isDialogOpen = false;
    } catch (error) {
      toast.error('Gagal update: ' + (error as { message?: string }).message);
    }
  }
</script>

<Card class="card-hover border-fun-green shadow-bubbly-green bg-card rounded-2xl border-2">
  <CardHeader class="flex flex-row items-center justify-between pb-2">
    <CardTitle class="text-muted-foreground text-sm font-medium">Kontak WhatsApp Aktif</CardTitle>
    <Phone class="h-4 w-4 text-green-600" />
  </CardHeader>
  <CardContent>
    <div class="flex items-end justify-between">
      <div>
        <div class="text-xl font-bold">{phone || '-'}</div>
        <p class="text-muted-foreground mt-1 text-xs">
          {phone ? 'Digunakan untuk validasi izin' : 'Harap lengkapi nomor HP!'}
        </p>
      </div>
      <Button
        variant="outline"
        size="sm"
        class="h-8 w-8 rounded-full p-0"
        aria-label="Ubah Nomor WhatsApp"
        onclick={() => (isDialogOpen = true)}><Pencil class="h-3 w-3" /></Button
      >
      <Dialog bind:open={isDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Ubah Nomor WhatsApp</DialogTitle>
            <DialogDescription>
              Perbarui nomor WhatsApp Anda untuk mempermudah koordinasi dan validasi absensi.
            </DialogDescription>
          </DialogHeader>
          <div class="space-y-4 py-4">
            <div class="space-y-2">
              <Label for="new-phone">Nomor HP (Format: 08xx / 62xx)</Label>
              <Input
                id="new-phone"
                value={newPhone}
                oninput={(e) => (newPhone = e.currentTarget.value.replace(/\D/g, ''))}
                placeholder="Contoh: 08123456789"
                type="tel"
              />
              <p class="text-muted-foreground text-xs">Pastikan nomor ini aktif di WhatsApp.</p>
            </div>
          </div>
          <DialogFooter>
            <Button onclick={handleSavePhone} disabled={updateProfile.isPending} class="w-full">
              {#if updateProfile.isPending}
                <Loader2 class="mr-2 h-4 w-4 animate-spin" />
              {:else}
                <Save class="mr-2 h-4 w-4" />
              {/if}
              Simpan Perubahan
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  </CardContent>
</Card>
