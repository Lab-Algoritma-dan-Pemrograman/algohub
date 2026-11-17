import type { RekapKelasResponse, RekapSel } from '@algohub/api-client/praktikum';
import { describe, expect, it } from 'vitest';
import { buildRekapCsv, rekapCsvFilename } from './rekap-nilai';

const sel = (total: number): RekapSel => ({
  pengerjaan_id: 1,
  murni: total,
  keaktifan: null,
  total,
  edit_keaktifan: false,
  nilai_akhir: null,
});

const rekap: RekapKelasResponse = {
  columns: [
    { key: 'course_1', label: 'Modul 1 - Pre-test' },
    { key: 'course_2', label: 'Modul 1 - Post-test' },
  ],
  data: [
    {
      nim: '2211001',
      nama: 'Budi Santoso',
      scores: { course_1: sel(80), course_2: sel(92.5) },
      total_score: 172.5,
    },
    { nim: '2211002', nama: 'Ani, Wati', scores: { course_2: sel(0) }, total_score: 0 },
  ],
};

// Salinan algoritma exportCSV di OLD/src/routes/praktikum/admin/rekap-nilai/+page.svelte.
function exportCsvLama(r: RekapKelasResponse, filteredData: RekapKelasResponse['data']) {
  const header = ['NIM', 'Nama', ...r.columns.map((c) => c.label), 'Total Nilai'];
  const rows = filteredData.map((m) => {
    const row = [m.nim, m.nama];
    r.columns.forEach((c) => {
      const s = m.scores[c.key];
      row.push(s ? s.total.toString() : '0');
    });
    row.push(m.total_score.toString());
    return row;
  });
  return [header.join(','), ...rows.map((x) => x.map((v) => `"${v}"`).join(','))].join('\n');
}

describe('buildRekapCsv', () => {
  it('menghasilkan string yang sama dengan exportCSV lama', () => {
    expect(buildRekapCsv(rekap, rekap.data)).toBe(exportCsvLama(rekap, rekap.data));
  });

  it('header tanpa kutip, sel dikutip, course kosong bernilai 0', () => {
    expect(buildRekapCsv(rekap, rekap.data)).toBe(
      [
        'NIM,Nama,Modul 1 - Pre-test,Modul 1 - Post-test,Total Nilai',
        '"2211001","Budi Santoso","80","92.5","172.5"',
        '"2211002","Ani, Wati","0","0","0"',
      ].join('\n'),
    );
  });

  it('hanya memuat baris yang diberikan (hasil pencarian)', () => {
    expect(buildRekapCsv(rekap, [rekap.data[1]])).toBe(exportCsvLama(rekap, [rekap.data[1]]));
    expect(buildRekapCsv(rekap, []).split('\n')).toHaveLength(1);
  });

  it('nama file memakai id kelas', () => {
    expect(rekapCsvFilename(7)).toBe('rekap_nilai_kelas_7.csv');
  });
});
