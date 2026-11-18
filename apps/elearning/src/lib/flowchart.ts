export type FlowchartShape =
  'terminator' | 'process' | 'io' | 'decision' | 'connector' | 'predefined';

export interface FlowchartSymbol {
  shape: FlowchartShape;
  label: string;
}

// Cabang alur: utama, atau cabang Ya/Tidak milik sebuah decision.
export type FlowchartBranch = 'main' | 'yes' | 'no';

export interface FlowchartStep extends FlowchartSymbol {
  at: FlowchartBranch;
}

export interface Flow {
  main: FlowchartSymbol[];
  yes: FlowchartSymbol[][];
  no: FlowchartSymbol[][];
}

export interface SlotRef {
  id: string;
  branch: FlowchartBranch;
  // Untuk cabang: indeks decision pemiliknya di alur utama.
  owner?: number;
  expected: FlowchartSymbol;
}

export type BankItem = FlowchartSymbol & { key: string };

export const SHAPE_LABEL: Record<FlowchartShape, string> = {
  terminator: 'Terminator (Mulai/Selesai)',
  process: 'Proses',
  io: 'Input/Output',
  decision: 'Keputusan',
  connector: 'Konektor',
  predefined: 'Proses Terdefinisi',
};

const SHAPES = new Set<string>(Object.keys(SHAPE_LABEL));

const toSymbol = (raw: unknown): FlowchartSymbol | null => {
  if (typeof raw !== 'object' || raw === null) return null;
  const { shape, label } = raw as { shape?: unknown; label?: unknown };
  if (typeof shape !== 'string' || !SHAPES.has(shape)) return null;
  return { shape: shape as FlowchartShape, label: String(label ?? '') };
};

export const sameCard = (
  a: FlowchartSymbol | null | undefined,
  b: FlowchartSymbol | null | undefined,
) => !!a && !!b && a.shape === b.shape && a.label === b.label;

// Simbol pengecoh tersimpan sebagai teks JSON di kolom flowchart_distractors.
export function parseDistractors(raw: string | null | undefined): FlowchartSymbol[] {
  if (!raw) return [];
  try {
    const parsed: unknown = JSON.parse(raw);
    return Array.isArray(parsed)
      ? parsed.map(toSymbol).filter((s): s is FlowchartSymbol => s !== null)
      : [];
  } catch {
    return [];
  }
}

// Urai `solution` menjadi alur utama dan cabang per decision.
// Versi 2: { version: 2, flow: [{ at, shape, label }] }. Versi 1: array datar, semuanya `main`.
// Setiap decision di alur utama mengambil entri yes/no berikutnya sebagai cabangnya.
export function parseFlow(solutionJson: string): Flow {
  const empty: Flow = { main: [], yes: [], no: [] };
  let parsed: unknown;
  try {
    parsed = JSON.parse(solutionJson);
  } catch {
    return empty;
  }
  let steps: FlowchartStep[];
  if (Array.isArray(parsed)) {
    steps = parsed.flatMap((s) => {
      const symbol = toSymbol(s);
      return symbol ? [{ ...symbol, at: 'main' as const }] : [];
    });
  } else if (typeof parsed === 'object' && parsed !== null && 'flow' in parsed) {
    const flow = (parsed as { flow: unknown }).flow;
    if (!Array.isArray(flow)) return empty;
    steps = flow.flatMap((s) => {
      const symbol = toSymbol(s);
      if (!symbol) return [];
      const at = (s as { at?: unknown }).at;
      return [{ ...symbol, at: at === 'yes' || at === 'no' ? at : ('main' as const) }];
    });
  } else {
    return empty;
  }
  const result: Flow = { main: [], yes: [], no: [] };
  let decisions = 0;
  for (const { at, ...card } of steps) {
    if (at === 'main') {
      result.main.push(card);
      if (card.shape === 'decision') {
        decisions++;
        result.yes.push([]);
        result.no.push([]);
      }
    } else if (decisions > 0) {
      // Cabang tanpa decision induk diabaikan.
      result[at][decisions - 1].push(card);
    }
  }
  return result;
}

// Semua slot penilaian: alur utama, lalu cabang tiap decision.
export function buildSlots({ main, yes, no }: Flow): SlotRef[] {
  const refs: SlotRef[] = main.map((expected, i) => ({ id: `m-${i}`, branch: 'main', expected }));
  let k = -1;
  for (const step of main) {
    if (step.shape !== 'decision') continue;
    k++;
    yes[k].forEach((expected, j) =>
      refs.push({ id: `b${k}-yes-${j}`, branch: 'yes', owner: k, expected }),
    );
    no[k].forEach((expected, j) =>
      refs.push({ id: `b${k}-no-${j}`, branch: 'no', owner: k, expected }),
    );
  }
  return refs;
}

// Bank berisi semua jawaban benar ditambah pengecoh, diacak.
export function buildBank(
  slots: SlotRef[],
  distractors: FlowchartSymbol[],
  random: () => number = Math.random,
): BankItem[] {
  const items = [...slots.map((r) => r.expected), ...distractors].map((s, i) => ({
    ...s,
    key: `${s.shape}-${i}-${s.label.slice(0, 8)}`,
  }));
  for (let i = items.length - 1; i > 0; i--) {
    const j = Math.floor(random() * (i + 1));
    [items[i], items[j]] = [items[j], items[i]];
  }
  return items;
}

// Jumlah decision di alur utama sampai indeks i (inklusif) dikurangi satu: indeks grup cabangnya.
export const branchGroupOf = (main: FlowchartSymbol[], i: number) =>
  main.slice(0, i + 1).filter((x) => x.shape === 'decision').length - 1;
