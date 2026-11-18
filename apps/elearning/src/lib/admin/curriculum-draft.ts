import type {
  CurriculumLevel,
  CurriculumRequest,
  JsonValue,
  Lesson as ApiLesson,
} from '@algohub/api-client/elearning';
import type { TreeLevel } from '#lib/curriculum.ts';
import type { AccessMode, Level, Lesson, Module, Quiz, TestCase, ValidationRule } from './types.ts';

const emptyQuiz = (): Quiz => ({ question: '', options: ['', '', '', ''], correctAnswer: 0 });

function asQuiz(raw: JsonValue): Quiz {
  if (!raw || typeof raw !== 'object' || Array.isArray(raw)) return emptyQuiz();
  const options = Array.isArray(raw.options) ? raw.options.map(String) : [];
  return {
    question: typeof raw.question === 'string' ? raw.question : '',
    options,
    correctAnswer: typeof raw.correctAnswer === 'number' ? raw.correctAnswer : 0,
  };
}

const asList = <T>(raw: JsonValue): T[] => (Array.isArray(raw) ? (raw as T[]) : []);

export function lessonFromApi(lesson: ApiLesson): Lesson {
  return {
    id: lesson.id,
    title: lesson.title,
    explanation: lesson.explanation ?? '',
    codeExample: lesson.code_example ?? '',
    initialCode: lesson.initial_code ?? '',
    solution: lesson.solution ?? '',
    hint: lesson.hint ?? '',
    quiz: asQuiz(lesson.quiz),
    testCases: asList<TestCase>(lesson.test_cases),
    validationRules: asList<ValidationRule>(lesson.validation_rules),
  };
}

const asAccessMode = (mode: string | null, locked: boolean | null): AccessMode => {
  if (mode === 'auto' || mode === 'unlocked' || mode === 'locked') return mode;
  return locked ? 'locked' : 'auto';
};

// Pohon hasil buildCurriculum menjadi draf yang bebas diubah tanpa menyentuh cache query.
export function draftFromTree(tree: TreeLevel[]): Level[] {
  return tree.map((level) => ({
    id: level.id,
    title: level.title,
    description: level.description ?? '',
    accessMode: asAccessMode(level.accessMode, level.locked),
    locked: level.locked ?? false,
    modules: level.modules.map((m) => ({
      id: m.id,
      title: m.title,
      lessons: m.lessons.map(lessonFromApi),
    })),
  }));
}

// Urutan array menjadi sort_order di backend, jadi tidak ada field urutan terpisah.
export function toRequest(draft: readonly Level[]): CurriculumRequest {
  const levels: CurriculumLevel[] = draft.map((level) => {
    const accessMode = level.accessMode ?? (level.locked ? 'locked' : 'auto');
    return {
      id: level.id,
      title: level.title,
      description: level.description,
      accessMode,
      locked: accessMode === 'locked',
      modules: level.modules.map((m) => ({
        id: m.id,
        title: m.title,
        lessons: m.lessons.map((l) => ({
          id: l.id,
          title: l.title,
          explanation: l.explanation,
          codeExample: l.codeExample,
          initialCode: l.initialCode,
          solution: l.solution,
          hint: l.hint,
          quiz: l.quiz as unknown as JsonValue,
          testCases: l.testCases as unknown as JsonValue,
          validationRules: (l.validationRules ?? []) as unknown as JsonValue,
        })),
      })),
    };
  });
  return { levels };
}

// Siklus tombol akses: auto, lalu unlocked, lalu locked, lalu kembali ke auto.
export function nextAccessMode(current: AccessMode): AccessMode {
  if (current === 'auto') return 'unlocked';
  if (current === 'unlocked') return 'locked';
  return 'auto';
}

export function effectiveMode(level: Pick<Level, 'accessMode' | 'locked'>): AccessMode {
  return level.accessMode ?? (level.locked ? 'locked' : 'auto');
}

export function cycleLevelAccess(level: Level): Level {
  const next = nextAccessMode(effectiveMode(level));
  return { ...level, accessMode: next, locked: next === 'locked' };
}

// Menukar posisi dengan tetangga; di tepi daftar tidak ada yang berubah.
export function moveItem<T>(items: readonly T[], index: number, direction: -1 | 1): T[] {
  const target = index + direction;
  if (target < 0 || target >= items.length) return [...items];
  const next = [...items];
  [next[index], next[target]] = [next[target], next[index]];
  return next;
}

export const newLevel = (count: number, now = Date.now()): Level => ({
  id: `level-${now}`,
  title: `Level Baru ${count + 1}`,
  description: 'Deskripsi level baru...',
  accessMode: 'auto',
  locked: false,
  modules: [],
});

export const newModule = (now = Date.now()): Module => ({
  id: `m-${now}`,
  title: 'Modul Baru',
  lessons: [],
});

export const newLesson = (now = Date.now()): Lesson => ({
  id: `l-${now}`,
  title: 'Pelajaran Baru',
  explanation: 'Isi penjelasan...',
  codeExample: '',
  initialCode: '',
  solution: '',
  hint: '',
  quiz: { question: 'Soal?', options: ['A', 'B', 'C', 'D'], correctAnswer: 0 },
  testCases: [{ description: 'Test', expectedOutput: 'Output' }],
});

export function replaceLevel(draft: readonly Level[], level: Level): Level[] {
  return draft.map((l) => (l.id === level.id ? level : l));
}

export function updateModule(level: Level, modIdx: number, patch: Partial<Module>): Level {
  const modules = level.modules.map((m, i) => (i === modIdx ? { ...m, ...patch } : m));
  return { ...level, modules };
}

export function updateLesson(level: Level, modIdx: number, lessonIdx: number, next: Lesson): Level {
  const lessons = level.modules[modIdx].lessons.map((l, i) => (i === lessonIdx ? next : l));
  return updateModule(level, modIdx, { lessons });
}

export interface LessonRef {
  levelId: string;
  modIdx: number;
  lessonIdx: number;
}

// Port handleApplyImportDocument untuk target satu modul: lesson hasil impor yang judul atau
// ID-nya sama menimpa isi lesson lama, sisanya ditambahkan di akhir modul.
export function importLessonsIntoModule(
  draft: readonly Level[],
  levelId: string,
  modIdx: number,
  imported: readonly Level[],
): Level[] {
  const level = draft.find((l) => l.id === levelId);
  const current = level?.modules[modIdx];
  if (!level || !current) return [...draft];
  const incoming = imported.flatMap((l) => l.modules.flatMap((m) => m.lessons));
  if (incoming.length === 0) {
    throw new Error('Tidak ada materi pelajaran yang terdeteksi di dalam dokumen.');
  }
  const lessons = [...current.lessons];
  for (const imp of incoming) {
    const title = imp.title.trim().toLowerCase();
    const at = lessons.findIndex((l) => l.title.trim().toLowerCase() === title || l.id === imp.id);
    if (at === -1) {
      lessons.push({ ...imp, id: imp.id || `${current.id}-l${lessons.length + 1}` });
      continue;
    }
    const target = { ...lessons[at] };
    if (imp.explanation) target.explanation = imp.explanation;
    target.codeExample = imp.codeExample;
    target.initialCode = imp.initialCode;
    target.solution = imp.solution;
    target.hint = imp.hint;
    if (imp.quiz?.question) target.quiz = imp.quiz;
    if (imp.testCases?.length) target.testCases = imp.testCases;
    if (imp.validationRules?.length) target.validationRules = imp.validationRules;
    lessons[at] = target;
  }
  return replaceLevel(draft, updateModule(level, modIdx, { lessons }));
}

// Memindahkan lesson ke modul lain (boleh di level lain); lesson dilepas dari sumbernya dulu.
export function moveLessonTo(
  draft: readonly Level[],
  from: LessonRef,
  targetLevelId: string,
  targetModId: string,
): Level[] | null {
  const lesson = draft.find((l) => l.id === from.levelId)?.modules[from.modIdx]?.lessons[
    from.lessonIdx
  ];
  if (!lesson) return null;
  return draft.map((level) => ({
    ...level,
    modules: level.modules.map((m, mIdx) => {
      let lessons = m.lessons;
      if (level.id === from.levelId && mIdx === from.modIdx) {
        lessons = lessons.filter((_, i) => i !== from.lessonIdx);
      }
      if (level.id === targetLevelId && m.id === targetModId) lessons = [...lessons, lesson];
      return lessons === m.lessons ? m : { ...m, lessons };
    }),
  }));
}
