export function avatarUrl(name: string): string {
  const query = new URLSearchParams({
    name: name.trim() || '?',
    background: '0D8ABC',
    color: 'fff',
    bold: 'true',
  });
  return `https://ui-avatars.com/api/?${query.toString()}`;
}

// Foto lama di Supabase dapat membalas 402 selama project dibatasi. Ganti hanya
// tampilan yang gagal; URL asli di database tidak dihapus sehingga bisa pulih.
export function useAvatarFallback(event: Event, name: string): void {
  const image = event.currentTarget as HTMLImageElement | null;
  if (!image || image.dataset.fallbackApplied === 'true') return;
  image.dataset.fallbackApplied = 'true';
  image.src = avatarUrl(name);
}
