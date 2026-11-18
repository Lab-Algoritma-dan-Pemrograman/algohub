// @vitest-environment node
import { describe, expect, it } from 'vitest';
import {
  branchGroupOf,
  buildBank,
  buildSlots,
  parseDistractors,
  parseFlow,
  sameCard,
} from '../src/lib/flowchart';

const v2 = JSON.stringify({
  version: 2,
  flow: [
    { at: 'main', shape: 'terminator', label: 'Mulai' },
    { at: 'main', shape: 'decision', label: 'x > 0?' },
    { at: 'yes', shape: 'io', label: 'Cetak positif' },
    { at: 'no', shape: 'io', label: 'Cetak non-positif' },
    { at: 'main', shape: 'terminator', label: 'Selesai' },
  ],
});

describe('parseFlow', () => {
  it('versi 2 memisahkan alur utama dan cabang decision', () => {
    const flow = parseFlow(v2);
    expect(flow.main.map((s) => s.label)).toEqual(['Mulai', 'x > 0?', 'Selesai']);
    expect(flow.yes[0].map((s) => s.label)).toEqual(['Cetak positif']);
    expect(flow.no[0].map((s) => s.label)).toEqual(['Cetak non-positif']);
  });

  it('versi 1 (array datar) seluruhnya alur utama', () => {
    const flow = parseFlow(
      JSON.stringify([
        { shape: 'terminator', label: 'Mulai' },
        { shape: 'process', label: 'a = 1' },
      ]),
    );
    expect(flow.main).toHaveLength(2);
    expect(flow.yes).toEqual([]);
  });

  it('JSON rusak, bentuk tak dikenal, dan cabang tanpa decision menghasilkan alur kosong atau diabaikan', () => {
    expect(parseFlow('{rusak')).toEqual({ main: [], yes: [], no: [] });
    expect(parseFlow('{"a":1}')).toEqual({ main: [], yes: [], no: [] });
    const orphan = parseFlow(
      JSON.stringify({ flow: [{ at: 'yes', shape: 'process', label: 'x' }] }),
    );
    expect(orphan.main).toEqual([]);
    expect(orphan.yes).toEqual([]);
  });
});

describe('buildSlots dan bank', () => {
  it('slot utama diikuti slot cabang tiap decision', () => {
    const slots = buildSlots(parseFlow(v2));
    expect(slots.map((s) => s.id)).toEqual(['m-0', 'm-1', 'm-2', 'b0-yes-0', 'b0-no-0']);
    expect(slots.find((s) => s.id === 'b0-yes-0')).toMatchObject({ branch: 'yes', owner: 0 });
  });

  it('bank berisi semua jawaban dan pengecoh dengan kunci unik', () => {
    const slots = buildSlots(parseFlow(v2));
    const bank = buildBank(slots, [{ shape: 'process', label: 'pengecoh' }]);
    expect(bank).toHaveLength(slots.length + 1);
    expect(new Set(bank.map((b) => b.key)).size).toBe(bank.length);
  });

  it('pengacakan deterministik bila sumber acak diberikan', () => {
    const slots = buildSlots(parseFlow(v2));
    expect(buildBank(slots, [], () => 0).map((b) => b.label)).toEqual(
      buildBank(slots, [], () => 0).map((b) => b.label),
    );
  });

  it('branchGroupOf menghitung decision sampai indeks', () => {
    const { main } = parseFlow(v2);
    expect(branchGroupOf(main, 0)).toBe(-1);
    expect(branchGroupOf(main, 1)).toBe(0);
  });
});

describe('sameCard dan parseDistractors', () => {
  it('kartu sama bila bentuk dan label sama', () => {
    expect(sameCard({ shape: 'io', label: 'a' }, { shape: 'io', label: 'a' })).toBe(true);
    expect(sameCard({ shape: 'io', label: 'a' }, { shape: 'process', label: 'a' })).toBe(false);
    expect(sameCard(undefined, { shape: 'io', label: 'a' })).toBe(false);
  });

  it('pengecoh tidak valid dibuang', () => {
    expect(parseDistractors(null)).toEqual([]);
    expect(parseDistractors('{rusak')).toEqual([]);
    expect(parseDistractors('{"a":1}')).toEqual([]);
  });
});
