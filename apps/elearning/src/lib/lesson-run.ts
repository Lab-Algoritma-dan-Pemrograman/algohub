import { messages, type Language, type RunResult } from '@algohub/runner';
import { preprocessCode } from './codePreprocessor.ts';
import { normalizeOutput, type TestCase, type ValidationRule } from './lesson.ts';

export interface LessonOutput {
  output: string;
  error: string | null;
  images: string[];
}

// Panel keluaran lesson menampilkan teks dan galat terpisah, tidak lewat terminal.
export function toLessonOutput(result: RunResult): LessonOutput {
  const { images } = result;
  switch (result.status) {
    case 'ok':
      return { output: result.output, error: null, images };
    case 'compile-error':
      return {
        output: '',
        error: result.output || result.error || messages.compileFailed,
        images,
      };
    case 'error':
      return { output: result.output, error: result.error ?? 'Runtime error', images };
    case 'timeout':
      return { output: result.output, error: messages.timeout, images };
    case 'stopped':
      return { output: result.output, error: messages.stopped, images };
    case 'unavailable':
      return { output: '', error: result.error ?? messages.unavailableC, images };
  }
}

export type RunFn = (code: string, stdin: string) => Promise<LessonOutput>;

// Aturan validasi statis: mengembalikan pesan aturan pertama yang dilanggar.
export function checkRules(
  code: string,
  language: Language,
  rules: ValidationRule[],
): string | null {
  for (const rule of rules) {
    if (!rule.pattern) continue;
    let exists: boolean;
    try {
      const clean = preprocessCode(code, language, { removeStrings: rule.stripStrings });
      exists = new RegExp(rule.pattern, rule.flags || 'im').test(clean);
    } catch {
      // Pola regex rusak dari admin dilewati supaya satu aturan buruk tidak mengunci lesson.
      continue;
    }
    if (rule.shouldExist !== exists) return rule.message;
  }
  return null;
}

export interface GradeOutcome extends LessonOutput {
  // null: program gagal dijalankan, belum ada penilaian.
  correct: boolean | null;
}

export interface GradeInput {
  run: RunFn;
  code: string;
  language: Language;
  testCases: TestCase[];
  rules: ValidationRule[];
}

// Menjalankan kode siswa lalu menilainya. Satu stdin dipakai bersama bila paling banyak satu
// kasus uji punya input; bila beberapa, kode dijalankan ulang per kasus uji.
export async function gradeExercise({
  run,
  code,
  language,
  testCases,
  rules,
}: GradeInput): Promise<GradeOutcome> {
  const firstInput = testCases.find((tc) => tc.input)?.input ?? '';
  const first = await run(code, firstInput);
  if (first.error) return { ...first, correct: null };

  const violation = checkRules(code, language, rules);
  if (violation) return { ...first, error: violation, correct: false };

  const withInput = testCases.filter((tc) => tc.input);
  if (withInput.length <= 1) {
    const actual = normalizeOutput(first.output);
    const correct =
      testCases.length === 0 ||
      testCases.every((tc) => actual === normalizeOutput(tc.expectedOutput));
    return { ...first, correct };
  }

  for (const tc of testCases) {
    const result = tc.input ? await run(code, tc.input) : first;
    if (result.error || normalizeOutput(result.output) !== normalizeOutput(tc.expectedOutput)) {
      return { ...first, error: result.error, correct: false };
    }
  }
  return { ...first, correct: true };
}
