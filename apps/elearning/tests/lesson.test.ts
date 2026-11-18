// @vitest-environment node
import type { Lesson, Level, Module } from '@algohub/api-client/elearning';
import { describe, expect, it } from 'vitest';
import { buildCurriculum } from '../src/lib/curriculum';
import {
  adjacentLessonId,
  detectLanguage,
  lessonLanguage,
  locateLesson,
  looksLikeHtml,
  nextLessonAfter,
  normalizeOutput,
  parseLesson,
} from '../src/lib/lesson';

const raw = (extra: Partial<Lesson> = {}): Lesson => ({
  id: 'les1',
  module_id: 'm1',
  title: 'Judul',
  explanation: null,
  code_example: null,
  initial_code: null,
  solution: null,
  hint: null,
  quiz: null,
  test_cases: null,
  validation_rules: null,
  xp_reward: null,
  sort_order: 1,
  exercise_type: null,
  flowchart_distractors: null,
  ...extra,
});

describe('parseLesson', () => {
  it('memberi nilai bawaan untuk kolom kosong', () => {
    expect(parseLesson(raw())).toMatchObject({
      explanation: '',
      codeExample: '',
      initialCode: '',
      hint: '',
      quiz: null,
      testCases: [],
      validationRules: [],
      xpReward: 60,
      isFlowchart: false,
      flowchartDistractors: [],
    });
  });

  it('mengurai kolom JSON yang tiba sebagai teks', () => {
    const lesson = parseLesson(
      raw({
        quiz: JSON.stringify({ question: 'Q?', options: ['a', 'b'], correctAnswer: 1 }),
        test_cases: JSON.stringify([{ expected_output: 'hi', description: 'cetak', input: 5 }]),
        validation_rules: JSON.stringify([
          { pattern: 'for', shouldExist: true, message: 'pakai for' },
        ]),
        xp_reward: 80,
      }),
    );
    expect(lesson.quiz).toEqual({ question: 'Q?', options: ['a', 'b'], correctAnswer: 1 });
    expect(lesson.testCases).toEqual([{ expectedOutput: 'hi', description: 'cetak', input: '5' }]);
    expect(lesson.validationRules).toMatchObject([{ pattern: 'for', shouldExist: true }]);
    expect(lesson.xpReward).toBe(80);
  });

  it('kuis tanpa kunci jawaban (praktikan) tidak diberi jawaban bawaan 0', () => {
    const lesson = parseLesson(raw({ quiz: { question: 'Q?', options: ['a', 'b'] } }));
    expect(lesson.quiz).toEqual({ question: 'Q?', options: ['a', 'b'], correctAnswer: undefined });
  });

  it('kuis kosong dan JSON rusak menjadi null atau daftar kosong', () => {
    expect(parseLesson(raw({ quiz: {} })).quiz).toBeNull();
    expect(parseLesson(raw({ quiz: 'bukan json' })).quiz).toBeNull();
    expect(parseLesson(raw({ test_cases: '{rusak' })).testCases).toEqual([]);
  });

  it('latihan flowchart membawa pengecoh', () => {
    const lesson = parseLesson(
      raw({
        exercise_type: 'flowchart',
        flowchart_distractors: JSON.stringify([
          { shape: 'process', label: 'x' },
          { shape: 'aneh' },
        ]),
      }),
    );
    expect(lesson.isFlowchart).toBe(true);
    expect(lesson.flowchartDistractors).toEqual([{ shape: 'process', label: 'x' }]);
  });
});

describe('normalizeOutput', () => {
  it('mengabaikan spasi berlebih, baris kosong akhir, dan huruf besar', () => {
    expect(normalizeOutput('  Halo   Dunia \n\n')).toBe('halo dunia');
    expect(normalizeOutput('A  \nB\n\n\n')).toBe('a\nb');
  });

  it('mempertahankan baris kosong di tengah', () => {
    expect(normalizeOutput('a\n\nb')).toBe('a\n\nb');
  });
});

describe('bahasa lesson', () => {
  it('detectLanguage butuh dua pola C', () => {
    expect(detectLanguage('#include <stdio.h>\nint main() { return 0; }')).toBe('c');
    expect(detectLanguage('print("halo")')).toBe('python');
    expect(detectLanguage('printf("x")')).toBe('python');
  });

  it('kode kosong jatuh ke judul level', () => {
    expect(
      lessonLanguage('  ', { id: 'a', title: 'Bahasa C', description: null } as never, 0),
    ).toBe('c');
    expect(lessonLanguage('', undefined, 3)).toBe('python');
    expect(lessonLanguage('#include <stdio.h>\nint main(){}', undefined, 3)).toBe('c');
  });

  it('looksLikeHtml membedakan HTML dari Markdown', () => {
    expect(looksLikeHtml('<p>halo</p>')).toBe(true);
    expect(looksLikeHtml('## Judul\n\nteks')).toBe(false);
  });
});

const level = (id: string, sort_order: number, extra: Partial<Level> = {}): Level => ({
  id,
  title: id,
  description: null,
  access_mode: null,
  locked: null,
  sort_order,
  ...extra,
});
const mod = (id: string, level_id: string, sort_order = 1): Module => ({
  id,
  level_id,
  title: id,
  sort_order,
});
const les = (id: string, module_id: string, sort_order = 1) => raw({ id, module_id, sort_order });

const tree = () =>
  buildCurriculum(
    [level('l1', 1), level('l2', 2)],
    [mod('m1', 'l1', 1), mod('m2', 'l1', 2), mod('m3', 'l2', 1)],
    [les('a', 'm1', 1), les('b', 'm1', 2), les('c', 'm2', 1), les('d', 'm3', 1)],
  );

describe('navigasi lesson', () => {
  it('locateLesson mengembalikan posisi atau null untuk id tak dikenal', () => {
    expect(locateLesson(tree(), 'c')).toMatchObject({ levelIdx: 0, moduleIdx: 1, lessonIdx: 0 });
    expect(locateLesson(tree(), 'zzz')).toBeNull();
  });

  it('adjacentLessonId melintasi modul dan memblokir level terkunci', () => {
    const t = tree();
    expect(adjacentLessonId(t, 'b', 'next', new Set(), {})).toBe('c');
    expect(adjacentLessonId(t, 'b', 'prev', new Set(), {})).toBe('a');
    expect(adjacentLessonId(t, 'a', 'prev', new Set(), {})).toBeNull();
    expect(adjacentLessonId(t, 'c', 'next', new Set(['a', 'b']), {})).toBeNull();
    expect(adjacentLessonId(t, 'c', 'next', new Set(['a', 'b', 'c']), {})).toBe('d');
    expect(adjacentLessonId(t, 'zzz', 'next', new Set(), {})).toBeNull();
  });

  it('nextLessonAfter: modul sama, modul berikutnya, level terbuka, lalu habis', () => {
    const t = tree();
    const at = (id: string) => locateLesson(t, id)!;
    expect(nextLessonAfter(t, at('a'), new Set(['a']), {})).toBe('b');
    expect(nextLessonAfter(t, at('b'), new Set(['a', 'b']), {})).toBe('c');
    expect(nextLessonAfter(t, at('c'), new Set(['a', 'b', 'c']), {})).toBe('d');
    expect(nextLessonAfter(t, at('c'), new Set(['c']), {})).toBeNull();
    expect(nextLessonAfter(t, at('d'), new Set(['a', 'b', 'c', 'd']), {})).toBeNull();
  });

  it('override akses per user membuka level berikutnya', () => {
    const t = tree();
    expect(nextLessonAfter(t, locateLesson(t, 'c')!, new Set(['c']), { l2: 'unlocked' })).toBe('d');
  });
});
