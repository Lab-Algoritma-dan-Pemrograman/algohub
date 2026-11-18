import { describe, expect, it } from 'vitest';
import {
  cycleLevelAccess,
  draftFromTree,
  importLessonsIntoModule,
  moveItem,
  moveLessonTo,
  toRequest,
} from '../src/lib/admin/curriculum-draft';
import type { Level, Lesson } from '../src/lib/admin/types';

const lesson = (id: string, title = id): Lesson => ({
  id,
  title,
  explanation: '',
  codeExample: '',
  initialCode: '',
  solution: '',
  hint: '',
  quiz: { question: '', options: [], correctAnswer: 0 },
  testCases: [],
});

const draft = (): Level[] => [
  {
    id: 'lv1',
    title: 'L1',
    description: '',
    modules: [
      { id: 'm1', title: 'M1', lessons: [lesson('a'), lesson('b')] },
      { id: 'm2', title: 'M2', lessons: [] },
    ],
  },
  { id: 'lv2', title: 'L2', description: '', modules: [{ id: 'm3', title: 'M3', lessons: [] }] },
];

describe('moveItem', () => {
  it('menukar dengan tetangga dan diam di tepi daftar', () => {
    expect(moveItem([1, 2, 3], 1, -1)).toEqual([2, 1, 3]);
    expect(moveItem([1, 2, 3], 2, 1)).toEqual([1, 2, 3]);
    expect(moveItem([1, 2, 3], 0, -1)).toEqual([1, 2, 3]);
  });
});

describe('cycleLevelAccess', () => {
  it('auto, unlocked, locked, lalu kembali ke auto dengan flag locked ikut', () => {
    const base = draft()[0];
    const unlocked = cycleLevelAccess(base);
    expect(unlocked).toMatchObject({ accessMode: 'unlocked', locked: false });
    const locked = cycleLevelAccess(unlocked);
    expect(locked).toMatchObject({ accessMode: 'locked', locked: true });
    expect(cycleLevelAccess(locked)).toMatchObject({ accessMode: 'auto', locked: false });
  });
});

describe('moveLessonTo', () => {
  it('memindahkan lesson antar level dan modul', () => {
    const next = moveLessonTo(draft(), { levelId: 'lv1', modIdx: 0, lessonIdx: 0 }, 'lv2', 'm3');
    expect(next?.[0].modules[0].lessons.map((l) => l.id)).toEqual(['b']);
    expect(next?.[1].modules[0].lessons.map((l) => l.id)).toEqual(['a']);
  });

  it('lesson yang tak ada tidak mengubah apa pun', () => {
    expect(
      moveLessonTo(draft(), { levelId: 'lv1', modIdx: 0, lessonIdx: 9 }, 'lv2', 'm3'),
    ).toBeNull();
  });
});

describe('importLessonsIntoModule', () => {
  const imported: Level[] = [
    {
      id: 'x',
      title: 'X',
      description: '',
      modules: [
        {
          id: 'xm',
          title: 'XM',
          lessons: [{ ...lesson('n1', 'a'), explanation: 'baru' }, lesson('n2', 'Lain')],
        },
      ],
    },
  ];

  it('judul yang sama menimpa isi, judul baru ditambahkan di akhir', () => {
    const next = importLessonsIntoModule(draft(), 'lv1', 0, imported);
    const lessons = next[0].modules[0].lessons;
    expect(lessons.map((l) => l.id)).toEqual(['a', 'b', 'n2']);
    expect(lessons[0].explanation).toBe('baru');
  });

  it('dokumen tanpa pelajaran ditolak', () => {
    expect(() => importLessonsIntoModule(draft(), 'lv1', 0, [])).toThrow(
      'Tidak ada materi pelajaran',
    );
  });
});

describe('toRequest dan draftFromTree', () => {
  it('draf menjadi payload PUT /curriculum dengan mode akses dan aturan validasi bawaan', () => {
    const req = toRequest(draft());
    expect(req.levels[0]).toMatchObject({ id: 'lv1', accessMode: 'auto', locked: false });
    expect(req.levels[0].modules[0].lessons[0].validationRules).toEqual([]);
  });

  it('lesson API dengan quiz kosong menjadi draf yang aman diedit', () => {
    const [level] = draftFromTree([
      {
        id: 'lv',
        title: 'L',
        description: null,
        accessMode: null,
        locked: true,
        modules: [
          {
            id: 'm',
            title: 'M',
            lessons: [
              {
                id: 'l',
                module_id: 'm',
                title: 'T',
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
              },
            ],
          },
        ],
      },
    ]);
    expect(level.accessMode).toBe('locked');
    expect(level.modules[0].lessons[0]).toMatchObject({
      explanation: '',
      quiz: { question: '', options: ['', '', '', ''], correctAnswer: 0 },
      testCases: [],
    });
  });
});
