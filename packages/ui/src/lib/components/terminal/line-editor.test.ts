import { describe, expect, it } from 'vitest';
import { createLineEditor, type KeyModifiers } from './line-editor';

const plain = (keyCode: number): KeyModifiers => ({
  keyCode,
  ctrlKey: false,
  altKey: false,
  metaKey: false,
});

const type = (editor: ReturnType<typeof createLineEditor>, text: string) =>
  [...text].map((ch) => editor.handle(ch, plain(ch.charCodeAt(0))).echo).join('');

describe('createLineEditor', () => {
  it('mengirim baris saat Enter lalu mengosongkan buffer', () => {
    const editor = createLineEditor();
    expect(type(editor, '12')).toBe('12');
    expect(editor.handle('\r', plain(13))).toEqual({ echo: '\r\n', line: '12' });
    expect(editor.handle('\r', plain(13))).toEqual({ echo: '\r\n', line: '' });
  });

  it('menghapus satu karakter dengan Backspace dan diam saat buffer kosong', () => {
    const editor = createLineEditor();
    expect(editor.handle('\x7f', plain(8))).toEqual({ echo: '' });
    type(editor, 'ab');
    expect(editor.handle('\x7f', plain(8))).toEqual({ echo: '\b \b' });
    expect(editor.handle('\r', plain(13)).line).toBe('a');
  });

  it('mengabaikan kombinasi Ctrl, Alt, dan Meta', () => {
    const editor = createLineEditor();
    expect(editor.handle('\x03', { ...plain(67), ctrlKey: true })).toEqual({ echo: '' });
    expect(editor.handle('c', { ...plain(67), altKey: true })).toEqual({ echo: '' });
    expect(editor.handle('c', { ...plain(67), metaKey: true })).toEqual({ echo: '' });
    expect(editor.handle('\r', plain(13)).line).toBe('');
  });

  it('reset membuang isi buffer', () => {
    const editor = createLineEditor();
    type(editor, 'halo');
    editor.reset();
    expect(editor.handle('\r', plain(13)).line).toBe('');
  });
});
