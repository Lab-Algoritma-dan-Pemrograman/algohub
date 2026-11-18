// @vitest-environment node
import { messages, type RunResult } from '@algohub/runner';
import { describe, expect, it, vi } from 'vitest';
import { preprocessCode, stripComments, stripStringLiterals } from '../src/lib/codePreprocessor';
import {
  checkRules,
  gradeExercise,
  toLessonOutput,
  type LessonOutput,
} from '../src/lib/lesson-run';

const result = (extra: Partial<RunResult>): RunResult => ({
  status: 'ok',
  output: '',
  error: null,
  images: [],
  waitingForInput: false,
  ...extra,
});

const out = (output: string, error: string | null = null): LessonOutput => ({
  output,
  error,
  images: [],
});

describe('codePreprocessor', () => {
  it('membuang komentar C dan Python', () => {
    expect(stripComments('int a; // x\n/* y */ int b;', 'c')).toBe('int a; \n int b;');
    expect(stripComments('a = 1 # x\n"""doc"""\nb = 2', 'python')).toBe('a = 1 \n\nb = 2');
  });

  it('mengosongkan string literal agar hardcode tidak lolos', () => {
    expect(stripStringLiterals('printf("Hasil: 100");', 'c')).toBe('printf("");');
    expect(stripStringLiterals('print(\'a\', "b")', 'python')).toBe('print("", "")');
  });

  it('preprocessCode menghormati opsi', () => {
    const code = 'x = "for" # for';
    expect(preprocessCode(code, 'python')).toBe('x = "for" ');
    expect(preprocessCode(code, 'python', { removeStrings: true })).toBe('x = "" ');
    expect(preprocessCode(code, 'python', { removeComments: false })).toBe(code);
  });
});

describe('toLessonOutput', () => {
  it('memetakan setiap status runner', () => {
    expect(toLessonOutput(result({ output: 'hi' }))).toEqual({
      output: 'hi',
      error: null,
      images: [],
    });
    expect(toLessonOutput(result({ status: 'compile-error', output: 'oops' })).error).toBe('oops');
    expect(toLessonOutput(result({ status: 'compile-error' })).error).toBe(messages.compileFailed);
    expect(toLessonOutput(result({ status: 'error', output: 'a', error: 'Traceback' }))).toEqual({
      output: 'a',
      error: 'Traceback',
      images: [],
    });
    expect(toLessonOutput(result({ status: 'timeout' })).error).toBe(messages.timeout);
    expect(toLessonOutput(result({ status: 'stopped' })).error).toBe(messages.stopped);
    expect(toLessonOutput(result({ status: 'unavailable', error: 'gagal' })).error).toBe('gagal');
  });

  it('gambar matplotlib dibawa dari images[]', () => {
    expect(toLessonOutput(result({ images: ['data:image/png;base64,AA'] })).images).toEqual([
      'data:image/png;base64,AA',
    ]);
  });
});

describe('checkRules', () => {
  const rule = { pattern: '\\bfor\\b', shouldExist: true, message: 'Gunakan for' };

  it('aturan wajib ada dan larangan', () => {
    expect(checkRules('for i in x: pass', 'python', [rule])).toBeNull();
    expect(checkRules('while True: pass', 'python', [rule])).toBe('Gunakan for');
    expect(
      checkRules('print(1)', 'python', [
        { ...rule, pattern: 'print', shouldExist: false, message: 'dilarang' },
      ]),
    ).toBe('dilarang');
  });

  it('kata di komentar tidak dihitung', () => {
    expect(checkRules('# for\nprint(1)', 'python', [rule])).toBe('Gunakan for');
  });

  it('stripStrings mencegah hardcode di dalam string', () => {
    const strict = { ...rule, pattern: '100', stripStrings: true };
    expect(checkRules('print("100")', 'python', [strict])).toBe('Gunakan for');
    expect(checkRules('print(100)', 'python', [strict])).toBeNull();
  });

  it('regex rusak dan pola kosong dilewati', () => {
    expect(
      checkRules('x', 'python', [
        { ...rule, pattern: '(' },
        { ...rule, pattern: '' },
      ]),
    ).toBeNull();
  });
});

describe('gradeExercise', () => {
  const base = { code: 'print(1)', language: 'python' as const, rules: [] };

  it('lolos bila keluaran sama setelah dinormalisasi', async () => {
    const run = vi.fn().mockResolvedValue(out('Halo  Dunia\n'));
    const outcome = await gradeExercise({
      ...base,
      run,
      testCases: [{ description: '', expectedOutput: 'halo dunia' }],
    });
    expect(outcome.correct).toBe(true);
    expect(run).toHaveBeenCalledWith('print(1)', '');
  });

  it('tanpa kasus uji dianggap benar', async () => {
    const outcome = await gradeExercise({ ...base, run: async () => out('x'), testCases: [] });
    expect(outcome.correct).toBe(true);
  });

  it('galat runtime menghentikan penilaian tanpa vonis', async () => {
    const outcome = await gradeExercise({
      ...base,
      run: async () => out('', 'NameError'),
      testCases: [{ description: '', expectedOutput: 'x' }],
    });
    expect(outcome).toMatchObject({ correct: null, error: 'NameError' });
  });

  it('aturan yang dilanggar menjadi pesan galat dan vonis salah', async () => {
    const outcome = await gradeExercise({
      ...base,
      run: async () => out('1'),
      testCases: [{ description: '', expectedOutput: '1' }],
      rules: [{ pattern: 'for', shouldExist: true, message: 'Gunakan for' }],
    });
    expect(outcome).toMatchObject({ correct: false, error: 'Gunakan for' });
  });

  it('satu kasus dengan input mengirim stdin-nya sekali', async () => {
    const run = vi.fn().mockResolvedValue(out('Halo Budi'));
    const outcome = await gradeExercise({
      ...base,
      run,
      testCases: [{ description: '', input: 'Budi', expectedOutput: 'Halo Budi' }],
    });
    expect(outcome.correct).toBe(true);
    expect(run).toHaveBeenCalledTimes(1);
    expect(run).toHaveBeenCalledWith('print(1)', 'Budi');
  });

  it('beberapa kasus dengan input dijalankan satu per satu', async () => {
    const run = vi.fn(async (_code: string, stdin: string) => out(`Halo ${stdin}`));
    const outcome = await gradeExercise({
      ...base,
      run,
      testCases: [
        { description: '', input: 'A', expectedOutput: 'Halo A' },
        { description: '', input: 'B', expectedOutput: 'Halo B' },
      ],
    });
    expect(outcome.correct).toBe(true);
    expect(run.mock.calls.map((c) => c[1])).toEqual(['A', 'A', 'B']);
  });

  it('kasus kedua yang salah menggagalkan seluruhnya', async () => {
    const run = vi.fn(async (_code: string, stdin: string) =>
      out(stdin === 'B' ? 'salah' : `Halo ${stdin}`),
    );
    const outcome = await gradeExercise({
      ...base,
      run,
      testCases: [
        { description: '', input: 'A', expectedOutput: 'Halo A' },
        { description: '', input: 'B', expectedOutput: 'Halo B' },
      ],
    });
    expect(outcome.correct).toBe(false);
  });
});
