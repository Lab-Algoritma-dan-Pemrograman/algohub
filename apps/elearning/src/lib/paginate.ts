import type { Page } from '@algohub/api-client/elearning';
const PAGE_LIMIT = 100;
// Backend membatasi limit maksimum 100 per halaman; daftar yang lebih panjang (lesson) harus
// diambil berurutan sampai total terpenuhi.
export async function fetchAllPages<T>(
  load: (page: number, limit: number) => Promise<Page<T>>,
): Promise<T[]> {
  const items: T[] = [];
  for (let page = 1; ; page++) {
    const res = await load(page, PAGE_LIMIT);
    items.push(...res.items);
    if (res.items.length === 0 || items.length >= res.total) return items;
  }
}
