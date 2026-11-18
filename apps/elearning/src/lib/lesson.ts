import type { Lesson } from '@algohub/api-client/elearning';
import type { Language } from '@algohub/runner';
import { isLessonLevelLocked, type AccessOverrides, type TreeLevel } from './curriculum.ts';
import { parseDistractors, type FlowchartSymbol } from './flowchart.ts';
import { getLevelLanguage } from './levelLanguage.ts';

export interface QuizData {
  question: string;
  options: string[];
  // Kosong untuk praktikan: backend membuang kunci jawaban, jadi kuis wajib divalidasi ke API.
  correctAnswer?: number;
}

export interface TestCase {
  description: string;
  input?: string;
  expectedOutput: string;
}

export interface ValidationRule {
  pattern: string;
  flags?: string;
  shouldExist: boolean;
  message: string;
  stripStrings?: boolean;
}

export interface LessonView {
  id: string;
  title: string;
  explanation: string;
  codeExample: string;
  initialCode: string;
  solution: string;
  hint: string;
  quiz: QuizData | null;
  testCases: TestCase[];
  validationRules: ValidationRule[];
  xpReward: number;
  isFlowchart: boolean;
  flowchartDistractors: FlowchartSymbol[];
}

export type LessonStep = 'learn' | 'quiz' | 'code';
export const STEP_ORDER: readonly LessonStep[] = ['learn', 'quiz', 'code'];

const isObject = (value: unknown): value is Record<string, unknown> =>
  typeof value === 'object' && value !== null && !Array.isArray(value);

// Kolom JSON di tabel lesson berisi teks JSON, jadi bisa tiba sebagai string atau sudah terurai.
function parseJsonField(value: unknown, fallback: unknown): unknown {
  if (value === null || value === undefined) return fallback;
  if (typeof value === 'object') return value;
  if (typeof value !== 'string') return fallback;
  try {
    return JSON.parse(value) ?? fallback;
  } catch {
    return fallback;
  }
}

function parseQuiz(raw: unknown): QuizData | null {
  const quiz = parseJsonField(raw, null);
  if (!isObject(quiz)) return null;
  const options = Array.isArray(quiz.options) ? quiz.options.map(String) : [];
  const question = typeof quiz.question === 'string' ? quiz.question : '';
  if (!question && options.length === 0) return null;
  const answer = quiz.correctAnswer ?? quiz.correct_answer;
  return { question, options, correctAnswer: typeof answer === 'number' ? answer : undefined };
}

function parseTestCases(raw: unknown): TestCase[] {
  const list = parseJsonField(raw, []);
  if (!Array.isArray(list)) return [];
  return list.filter(isObject).map((tc) => ({
    expectedOutput: String(tc.expectedOutput ?? tc.expected_output ?? ''),
    description: String(tc.description ?? ''),
    input: tc.input === undefined || tc.input === null ? undefined : String(tc.input),
  }));
}

function parseRules(raw: unknown): ValidationRule[] {
  const list = parseJsonField(raw, []);
  if (!Array.isArray(list)) return [];
  return list.filter(isObject).map((rule) => ({
    pattern: String(rule.pattern ?? ''),
    flags: typeof rule.flags === 'string' ? rule.flags : undefined,
    shouldExist: Boolean(rule.shouldExist),
    message: String(rule.message ?? ''),
    stripStrings: Boolean(rule.stripStrings),
  }));
}

// Port pemetaan di curriculumService.ts lama; nama field backend snake_case diubah ke camelCase.
export function parseLesson(raw: Lesson): LessonView {
  return {
    id: raw.id,
    title: raw.title,
    explanation: raw.explanation ?? '',
    codeExample: raw.code_example ?? '',
    initialCode: raw.initial_code ?? '',
    solution: raw.solution ?? '',
    hint: raw.hint ?? '',
    quiz: parseQuiz(raw.quiz),
    testCases: parseTestCases(raw.test_cases),
    validationRules: parseRules(raw.validation_rules),
    xpReward: raw.xp_reward ?? 60,
    isFlowchart: raw.exercise_type === 'flowchart',
    flowchartDistractors: parseDistractors(raw.flowchart_distractors),
  };
}

// Penjelasan lesson ditulis dengan editor rich text (HTML) atau Markdown polos.
export const looksLikeHtml = (text: string) => /<\/?[a-z][\s\S]*>/i.test(text);

// Port detectLanguage di useCodeRunner.ts lama.
export function detectLanguage(code: string): Language {
  const cPatterns = [
    /#include\s*</,
    /\bprintf\s*\(/,
    /\bscanf\s*\(/,
    /\bint\s+main\s*\(/,
    /\bvoid\s+main\s*\(/,
    /\bstdio\.h\b/,
    /\bstdlib\.h\b/,
    /\breturn\s+0\s*;/,
  ];
  return cPatterns.filter((p) => p.test(code)).length >= 2 ? 'c' : 'python';
}

// Isi kode menang atas judul level, lalu judul level menjadi fallback bila kode masih kosong.
export function lessonLanguage(
  sampleCode: string,
  level: TreeLevel | undefined,
  levelIdx: number,
): Language {
  if (sampleCode.trim().length > 0) return detectLanguage(sampleCode);
  return getLevelLanguage(level, levelIdx);
}

// Perbandingan keluaran longgar: spasi, baris kosong di akhir, dan huruf besar/kecil diabaikan.
export function normalizeOutput(text: string): string {
  return text
    .trim()
    .split('\n')
    .map((line) => line.trimEnd().replace(/\s+/g, ' '))
    .filter((line, idx, arr) => line !== '' || arr.slice(idx).some((l) => l !== ''))
    .join('\n')
    .toLowerCase();
}

export interface LessonPosition {
  levelIdx: number;
  moduleIdx: number;
  lessonIdx: number;
  lesson: Lesson;
}

export function locateLesson(curriculum: TreeLevel[], lessonId: string): LessonPosition | null {
  for (let levelIdx = 0; levelIdx < curriculum.length; levelIdx++) {
    const modules = curriculum[levelIdx].modules;
    for (let moduleIdx = 0; moduleIdx < modules.length; moduleIdx++) {
      const lessonIdx = modules[moduleIdx].lessons.findIndex((l) => l.id === lessonId);
      if (lessonIdx !== -1) {
        return { levelIdx, moduleIdx, lessonIdx, lesson: modules[moduleIdx].lessons[lessonIdx] };
      }
    }
  }
  return null;
}

// Lesson sebelumnya/berikutnya dalam urutan kurikulum. Maju ke level terkunci diblokir.
export function adjacentLessonId(
  curriculum: TreeLevel[],
  lessonId: string,
  direction: 'prev' | 'next',
  completed: ReadonlySet<string>,
  overrides: AccessOverrides,
): string | null {
  const flat = curriculum.flatMap((level, levelIdx) =>
    level.modules.flatMap((m) => m.lessons.map((l) => ({ id: l.id, levelIdx }))),
  );
  const current = flat.findIndex((entry) => entry.id === lessonId);
  if (current === -1) return null;
  if (direction === 'prev') return current > 0 ? flat[current - 1].id : null;
  const next = flat[current + 1];
  if (!next) return null;
  return isLessonLevelLocked(curriculum, next.levelIdx, completed, overrides) ? null : next.id;
}

// Tujuan setelah lesson selesai: lesson berikutnya di modul, modul berikutnya di level yang sama,
// lalu level pertama yang terbuka. null berarti kurikulum habis.
export function nextLessonAfter(
  curriculum: TreeLevel[],
  pos: Pick<LessonPosition, 'levelIdx' | 'moduleIdx' | 'lessonIdx'>,
  completed: ReadonlySet<string>,
  overrides: AccessOverrides,
): string | null {
  const level = curriculum[pos.levelIdx];
  const module = level?.modules[pos.moduleIdx];
  if (module && pos.lessonIdx < module.lessons.length - 1) {
    return module.lessons[pos.lessonIdx + 1].id;
  }
  if (level && pos.moduleIdx < level.modules.length - 1) {
    return level.modules[pos.moduleIdx + 1]?.lessons[0]?.id ?? null;
  }
  for (let i = pos.levelIdx + 1; i < curriculum.length; i++) {
    if (!isLessonLevelLocked(curriculum, i, completed, overrides)) {
      return curriculum[i].modules[0]?.lessons[0]?.id ?? null;
    }
  }
  return null;
}
