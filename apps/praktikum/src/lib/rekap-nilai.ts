import type { RekapKelasResponse, RekapMahasiswa } from '@algohub/api-client/praktikum';

// Format sama persis dengan exportCSV lama: header tanpa kutip, sel data dikutip apa adanya
// (tanpa escape), course yang belum dikerjakan bernilai 0.
export function buildRekapCsv(rekap: Pick<RekapKelasResponse, 'columns'>, rows: RekapMahasiswa[]) {
  const header = ['NIM', 'Nama', ...rekap.columns.map((c) => c.label), 'Total Nilai'];
  const body = rows.map((m) => {
    const row = [m.nim, m.nama];
    for (const c of rekap.columns) {
      const sel = m.scores[c.key];
      row.push(sel ? sel.total.toString() : '0');
    }
    row.push(m.total_score.toString());
    return row;
  });
  return [header.join(','), ...body.map((r) => r.map((v) => `"${v}"`).join(','))].join('\n');
}

export const rekapCsvFilename = (kelasId: number) => `rekap_nilai_kelas_${kelasId}.csv`;
