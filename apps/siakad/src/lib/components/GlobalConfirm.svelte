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
  import HelpCircle from '@lucide/svelte/icons/circle-question-mark';
  import { confirmState, settleConfirm } from '#lib/stores/confirm.svelte.ts';
</script>

<Dialog
  open={confirmState.open}
  onOpenChange={(open) => {
    if (!open) settleConfirm(false);
  }}
>
  <DialogContent
    class="max-w-md rounded-2xl border-2 border-slate-200 bg-[#FDFBF7] p-6 shadow-[0_8px_0_#CBD5E1]"
  >
    <DialogHeader class="space-y-3">
      <DialogTitle class="text-dark flex items-center gap-2 text-lg font-black">
        <div
          class="text-maroon flex h-8 w-8 items-center justify-center rounded-lg border border-red-200 bg-red-50 shadow-[0_2px_0_#5C0E25]"
        >
          <HelpCircle class="h-5 w-5" />
        </div>
        Konfirmasi Aksi
      </DialogTitle>
      <DialogDescription class="pt-2 text-sm leading-relaxed font-bold text-gray-600">
        {confirmState.message}
      </DialogDescription>
    </DialogHeader>
    <DialogFooter class="mt-6 flex gap-3 sm:justify-end">
      <Button
        variant="outline"
        onclick={() => settleConfirm(false)}
        class="h-auto rounded-xl border-2 border-slate-200 px-5 py-2.5 font-bold text-gray-500 shadow-sm hover:bg-slate-50"
      >
        Batal
      </Button>
      <Button
        onclick={() => settleConfirm(true)}
        class="bg-maroon hover:bg-maroon-light h-auto rounded-xl px-6 py-2.5 font-black text-white shadow-[0_4px_0_#5C0E25] transition-transform hover:-translate-y-0.5 active:translate-y-0"
      >
        Lanjutkan
      </Button>
    </DialogFooter>
  </DialogContent>
</Dialog>
