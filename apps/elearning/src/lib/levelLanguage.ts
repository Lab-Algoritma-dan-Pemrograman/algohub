import type { Language } from '@algohub/runner';
interface LevelLike {
  id?: string | null;
  title?: string | null;
}
// Bahasa level ditentukan dari judul, bukan prefix ID: sejak level di-rename menjadi
// level-1..level-6, prefix 'c-'/'py-' tidak ada lagi dan filter berbasis prefix mengembalikan
// nol level. Prefix ID lama dan urutan kurikulum hanya fallback untuk data historis.
export function getLevelLanguage(level: LevelLike | null | undefined, index?: number): Language {
  if (!level) return 'python';
  const title = String(level.title || '').toUpperCase();
  if (title.includes('PYTHON')) return 'python';
  if (/\bC\b/.test(title)) return 'c';
  const id = String(level.id || '');
  if (id.startsWith('c-')) return 'c';
  if (id.startsWith('py-') || id.startsWith('p-')) return 'python';
  return typeof index === 'number' && index <= 1 ? 'c' : 'python';
}
export function getLevelLanguageLabel(level: LevelLike | null | undefined, index?: number): string {
  return getLevelLanguage(level, index) === 'c' ? 'Bahasa C' : 'Python';
}
