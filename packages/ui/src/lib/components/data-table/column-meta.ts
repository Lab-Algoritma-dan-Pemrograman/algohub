import type { RowData } from '@tanstack/table-core';

// Kelas per kolom (lebar, rata tengah, kolom sticky) ditaruh di meta agar definisi kolom tetap satu tempat.
declare module '@tanstack/table-core' {
  // eslint-disable-next-line @typescript-eslint/no-unused-vars -- parameter wajib sama dengan deklarasi asli
  interface ColumnMeta<TData extends RowData, TValue> {
    headClass?: string;
    cellClass?: string;
  }
}
