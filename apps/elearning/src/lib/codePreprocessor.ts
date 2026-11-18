import type { Language } from '@algohub/runner';

// Pembersih kode untuk validasi berbasis regex: komentar dan string literal dibuang supaya
// aturan seperti "harus memakai for" tidak lolos karena kata itu hanya ada di komentar.
export function stripComments(code: string, language: Language): string {
  return language === 'c'
    ? code.replace(/\/\*[\s\S]*?\*\/|\/\/.*/g, '')
    : code.replace(/#.*|"""[\s\S]*?"""|'''[\s\S]*?'''/g, '');
}

// Mencegah hardcode keluaran: printf("Hasil: 100") menjadi printf("").
export function stripStringLiterals(code: string, language: Language): string {
  return language === 'c'
    ? code.replace(/"(\\.|[^"\\])*"/g, '""')
    : code.replace(/"(\\.|[^"\\])*"|'(\\.|[^'\\])*'/g, '""');
}

export function preprocessCode(
  code: string,
  language: Language,
  { removeComments = true, removeStrings = false } = {},
): string {
  let clean = code;
  if (removeComments) clean = stripComments(clean, language);
  if (removeStrings) clean = stripStringLiterals(clean, language);
  return clean;
}
