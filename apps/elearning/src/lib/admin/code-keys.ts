import { tick } from 'svelte';

// Port handleCodeKeyDown di AdminDashboard.tsx lama: textarea kode menyisipkan dua spasi untuk
// Tab dan mempertahankan indentasi baris sebelumnya saat Enter.
export function codeKeydown(event: KeyboardEvent, value: string, set: (next: string) => void) {
  const area = event.currentTarget as HTMLTextAreaElement;
  const start = area.selectionStart;
  const end = area.selectionEnd;
  let insert: string | null = null;
  if (event.key === 'Tab') {
    insert = '  ';
  } else if (event.key === 'Enter') {
    const indent = /^\s*/.exec(value.slice(0, start).split('\n').pop() ?? '')?.[0] ?? '';
    if (indent) insert = `\n${indent}`;
  }
  if (insert === null) return;
  event.preventDefault();
  set(value.slice(0, start) + insert + value.slice(end));
  const caret = start + insert.length;
  // Nilai textarea diperbarui Svelte setelah tick; kursor baru valid sesudahnya.
  void tick().then(() => area.setSelectionRange(caret, caret));
}
