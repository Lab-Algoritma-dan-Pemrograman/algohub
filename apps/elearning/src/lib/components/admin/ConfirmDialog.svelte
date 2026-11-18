<script lang="ts">
  import {
    Button,
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
  } from '@algohub/ui';
  let {
    open = $bindable(false),
    title,
    message,
    confirmLabel = 'Lanjutkan',
    busy = false,
    onconfirm,
  }: {
    open?: boolean;
    title: string;
    message: string;
    confirmLabel?: string;
    busy?: boolean;
    onconfirm: () => void | Promise<void>;
  } = $props();
  async function confirm() {
    await onconfirm();
    open = false;
  }
</script>

<Dialog bind:open>
  <DialogContent class="rounded-3xl sm:max-w-sm">
    <DialogHeader>
      <DialogTitle class="text-xl font-black text-zinc-900">{title}</DialogTitle>
      <DialogDescription class="text-sm font-semibold text-zinc-500">{message}</DialogDescription>
    </DialogHeader>
    <DialogFooter class="gap-2">
      <Button variant="outline" class="rounded-xl font-bold" onclick={() => (open = false)}>
        Batal
      </Button>
      <Button
        class="rounded-xl bg-red-600 font-bold text-white hover:bg-red-700"
        disabled={busy}
        onclick={confirm}
      >
        {busy ? 'Memproses...' : confirmLabel}
      </Button>
    </DialogFooter>
  </DialogContent>
</Dialog>
