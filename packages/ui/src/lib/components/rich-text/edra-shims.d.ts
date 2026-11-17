// edra (di-vendor ke rich-text/edra) mengakses storage dan API ProseMirror yang belum
// dideklarasikan tipenya di @tiptap v3; shim ini membersihkan error type-check tanpa
// mengubah kode edra.
import '@tiptap/core';
/* eslint-disable @typescript-eslint/no-explicit-any */
declare module '@tiptap/core' {
  interface Storage {
    searchAndReplace: any;
    slashCommand: any;
  }
}
declare module '@tiptap/pm/view' {
  export function __serializeForClipboard(
    view: any,
    slice: any,
  ): { dom: HTMLElement; text: string };
}
