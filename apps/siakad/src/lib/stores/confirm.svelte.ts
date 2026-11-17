export const confirmState = $state<{
  open: boolean;
  message: string;
  resolve: ((value: boolean) => void) | null;
}>({ open: false, message: '', resolve: null });

// Emoji dan simbol bawaan pesan lama dibuang supaya tidak dobel dengan ikon dialog.
const cleanMessage = (msg: string) => msg.replace(/^[🗑⚠👋✨\s]+/u, '').replace(/[\s🌟]+$/u, '');

export function confirm(message: string): Promise<boolean> {
  return new Promise<boolean>((resolve) => {
    // Pemanggilan beruntun membatalkan dialog sebelumnya.
    confirmState.resolve?.(false);
    confirmState.message = cleanMessage(message);
    confirmState.resolve = resolve;
    confirmState.open = true;
  });
}

export function settleConfirm(value: boolean): void {
  const resolve = confirmState.resolve;
  confirmState.resolve = null;
  confirmState.open = false;
  resolve?.(value);
}
