// Bentuk draf kurikulum yang diedit admin; sama dengan data/curriculum.ts lama. Dikirim ke
// PUT /curriculum sebagai CurriculumRequest (lihat curriculum-draft.ts).
export interface Quiz {
  question: string;
  options: string[];
  correctAnswer: number;
}

export interface TestCase {
  input?: string;
  expectedOutput: string;
  description: string;
}

export interface ValidationRule {
  pattern: string;
  message: string;
  shouldExist: boolean;
}

export interface Lesson {
  id: string;
  title: string;
  explanation: string;
  codeExample: string;
  initialCode: string;
  solution: string;
  hint: string;
  quiz: Quiz;
  testCases: TestCase[];
  validationRules?: ValidationRule[];
}

export interface Module {
  id: string;
  title: string;
  lessons: Lesson[];
}

export type AccessMode = 'auto' | 'unlocked' | 'locked';

export interface Level {
  id: string;
  title: string;
  description: string;
  accessMode?: AccessMode;
  locked?: boolean;
  modules: Module[];
}

// Tujuan impor dokumen: seluruh draf, atau satu modul tertentu.
export type ImportTarget = 'global' | { levelId: string; modIdx: number };
