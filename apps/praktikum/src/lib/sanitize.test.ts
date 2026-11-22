import { describe, expect, it } from 'vitest';
import { sanitizeSoal } from './sanitize.ts';
import { firstName } from './utils.ts';

describe('sanitizeSoal', () => {
  it('buang <script> tag', () => {
    expect(sanitizeSoal('<p>hai</p><script>alert(1)</script>')).toBe('<p>hai</p>');
  });

  it('buang onerror / onload attribute', () => {
    expect(sanitizeSoal('<img src="x" onerror="alert(1)">')).not.toContain('onerror');
    expect(sanitizeSoal('<p onload="alert(1)">x</p>')).not.toContain('onload');
  });

  it('buang javascript: URL', () => {
    const out = sanitizeSoal('<a href="javascript:alert(1)">klik</a>');
    expect(out).not.toContain('javascript:');
  });

  it('buang iframe/object/style', () => {
    expect(sanitizeSoal('<iframe src="https://evil.test"></iframe>')).toBe('');
    expect(sanitizeSoal('<object data="x"></object>')).toBe('');
    expect(sanitizeSoal('<style>body{}</style>')).toBe('');
  });

  it('string aman lewat utuh', () => {
    expect(sanitizeSoal('<p><strong>hai</strong></p>')).toBe('<p><strong>hai</strong></p>');
  });

  it('null/kosong → string kosong', () => {
    expect(sanitizeSoal(null)).toBe('');
    expect(sanitizeSoal('')).toBe('');
  });
});

describe('firstName', () => {
  it('null/undefined → Pengguna', () => {
    expect(firstName(null)).toBe('Pengguna');
    expect(firstName(undefined)).toBe('Pengguna');
  });

  it('string kosong / spasi saja → Pengguna', () => {
    expect(firstName('')).toBe('Pengguna');
    expect(firstName('   ')).toBe('Pengguna');
  });

  it('ambil kata pertama', () => {
    expect(firstName('Budi Santoso')).toBe('Budi');
  });

  it('spasi berlebih → tetap token pertama valid', () => {
    expect(firstName('   Budi    Santoso  ')).toBe('Budi');
  });

  it('token emoji dilewati, ambil token berikutnya', () => {
    expect(firstName('🎉 Budi')).toBe('Budi');
    expect(firstName('🎉')).toBe('🎉'); // tak ada token huruf/angka: pakai nama penuh
  });

  it('token kosong di antara spasi non-breaking', () => {
    expect(firstName(' Budi')).toBe('Budi');
  });
});
