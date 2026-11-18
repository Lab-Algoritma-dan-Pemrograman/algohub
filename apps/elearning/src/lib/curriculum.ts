import type { Lesson, Level, Module } from '@algohub/api-client/elearning';
export type AccessMode = 'auto' | 'unlocked' | 'locked';
export type AccessOverrides = Record<string, AccessMode>;
export interface TreeLevel {
  id: string;
  title: string;
  description: string | null;
  accessMode: string | null;
  locked: boolean | null;
  modules: { id: string; title: string; lessons: Lesson[] }[];
}
const bySortOrder = <T extends { sort_order: number }>(a: T, b: T) => a.sort_order - b.sort_order;
// Menyusun pohon level > modul > lesson dari tiga daftar datar yang dikembalikan backend.
export function buildCurriculum(
  levels: Level[],
  modules: Module[],
  lessons: Lesson[],
): TreeLevel[] {
  return [...levels].sort(bySortOrder).map((level) => ({
    id: level.id,
    title: level.title,
    description: level.description,
    accessMode: level.access_mode,
    locked: level.locked,
    modules: modules
      .filter((m) => m.level_id === level.id)
      .sort(bySortOrder)
      .map((m) => ({
        id: m.id,
        title: m.title,
        lessons: lessons.filter((l) => l.module_id === m.id).sort(bySortOrder),
      })),
  }));
}
export function lessonIds(level: TreeLevel): string[] {
  return level.modules.flatMap((m) => m.lessons.map((l) => l.id));
}
export function overridesOf(raw: unknown): AccessOverrides {
  return raw && typeof raw === 'object' ? (raw as AccessOverrides) : {};
}
// Mode akses efektif sebuah level: override per user menang, lalu mode level, lalu flag lama.
export function effectiveAccessMode(
  curriculum: TreeLevel[],
  levelId: string,
  overrides: AccessOverrides,
): AccessMode {
  const override = overrides[levelId];
  if (override && override !== 'auto') return override;
  const level = curriculum.find((l) => l.id === levelId);
  if (!level) return 'locked';
  if (level.accessMode === 'locked' || level.accessMode === 'unlocked') return level.accessMode;
  if (level.locked === true) return 'locked';
  return 'auto';
}
// Port isLevelLockedDisplay di Dashboard.tsx lama.
export function isLevelLocked(
  curriculum: TreeLevel[],
  idx: number,
  completed: ReadonlySet<string>,
  overrides: AccessOverrides,
): boolean {
  const level = curriculum[idx];
  if (!level) return true;
  const override = overrides[level.id];
  if (override && override !== 'auto') return override === 'locked';
  if (level.accessMode === 'locked') return true;
  if (level.accessMode === 'unlocked') return false;
  // locked=false adalah default DB, jadi hanya true yang dianggap kunci.
  if (level.locked === true) return true;
  if (idx === 0) return false;
  // Mode auto: semua level sebelumnya harus selesai, kecuali yang di-unlock eksplisit.
  for (let i = 0; i < idx; i++) {
    const prev = curriculum[i];
    if (effectiveAccessMode(curriculum, prev.id, overrides) === 'unlocked') continue;
    if (lessonIds(prev).some((id) => !completed.has(id))) return true;
  }
  return false;
}
export interface Checkpoint {
  lessonId: string;
  step?: string;
  timestamp?: number;
}
export interface ContinueTarget {
  lessonId: string;
  // Lesson yang dibuka dari awal; LessonPage membaca langkah per lesson dari localStorage.
  resetStep: boolean;
}
// Port handleContinue di Dashboard.tsx lama. Mengembalikan lesson yang harus dibuka atau null
// bila tidak ada level terbuka.
export function pickContinueLesson(
  curriculum: TreeLevel[],
  completed: ReadonlySet<string>,
  overrides: AccessOverrides,
  checkpoint: Checkpoint | null,
): ContinueTarget | null {
  const all = curriculum.flatMap((level, i) =>
    isLevelLocked(curriculum, i, completed, overrides) ? [] : lessonIds(level),
  );
  if (all.length === 0) return null;
  const firstUncompleted = all.find((id) => !completed.has(id));
  if (checkpoint?.lessonId && all.includes(checkpoint.lessonId)) {
    const cpIdx = all.indexOf(checkpoint.lessonId);
    const firstIdx = firstUncompleted ? all.indexOf(firstUncompleted) : -1;
    // Ada lesson belum selesai sebelum checkpoint (mis. setelah reset admin): kembali ke sana.
    if (firstUncompleted && firstIdx !== -1 && firstIdx < cpIdx) {
      return { lessonId: firstUncompleted, resetStep: true };
    }
    if (!completed.has(checkpoint.lessonId)) {
      return { lessonId: checkpoint.lessonId, resetStep: false };
    }
    const next = all.slice(cpIdx + 1).find((id) => !completed.has(id));
    if (next) return { lessonId: next, resetStep: false };
  }
  if (firstUncompleted) return { lessonId: firstUncompleted, resetStep: true };
  // Semua selesai: buka lesson terakhir untuk ditinjau.
  return { lessonId: all[all.length - 1], resetStep: false };
}
export function levelProgressPercent(
  levels: TreeLevel[],
  curriculum: TreeLevel[],
  completed: ReadonlySet<string>,
  overrides: AccessOverrides,
): number {
  let total = 0;
  let done = 0;
  for (const level of levels) {
    const ids = lessonIds(level);
    done += ids.filter((id) => completed.has(id)).length;
    const idx = curriculum.findIndex((l) => l.id === level.id);
    if (!isLevelLocked(curriculum, idx, completed, overrides)) total += ids.length;
  }
  return total > 0 ? Math.round((done / total) * 100) : 0;
}

// Port isLevelLocked di LessonPage.tsx lama. Berbeda dari tampilan dashboard: level sebelumnya
// yang di-unlock eksplisit tetap harus selesai sebelum level auto berikutnya terbuka.
export function isLessonLevelLocked(
  curriculum: TreeLevel[],
  idx: number,
  completed: ReadonlySet<string>,
  overrides: AccessOverrides,
): boolean {
  const level = curriculum[idx];
  if (!level) return true;
  const override = overrides[level.id];
  if (override && override !== 'auto') return override === 'locked';
  if (level.accessMode === 'locked') return true;
  if (level.accessMode === 'unlocked') return false;
  if (level.locked === true) return true;
  if (idx === 0) return false;
  for (let i = 0; i < idx; i++) {
    if (lessonIds(curriculum[i]).some((id) => !completed.has(id))) return true;
  }
  return false;
}

// Port isModuleLocked di CourseExplorer.tsx lama.
export function isModuleLocked(
  curriculum: TreeLevel[],
  levelIdx: number,
  moduleIdx: number,
  completed: ReadonlySet<string>,
  overrides: AccessOverrides,
): boolean {
  const level = curriculum[levelIdx];
  if (!level) return true;
  const mode = effectiveAccessMode(curriculum, level.id, overrides);
  if (mode === 'locked') return true;
  if (mode === 'unlocked') return false;
  if (levelIdx === 0 && moduleIdx === 0) return false;
  for (let i = 0; i < levelIdx; i++) {
    const prev = curriculum[i];
    if (effectiveAccessMode(curriculum, prev.id, overrides) === 'unlocked') continue;
    if (lessonIds(prev).some((id) => !completed.has(id))) return true;
  }
  if (moduleIdx > 0) {
    const prevModule = level.modules[moduleIdx - 1];
    if (!prevModule) return false;
    return !prevModule.lessons.every((lesson) => completed.has(lesson.id));
  }
  return false;
}

const percentOf = (done: number, total: number) =>
  total > 0 ? Math.round((done / total) * 100) : 0;

export function moduleProgressPercent(
  module: TreeLevel['modules'][number],
  completed: ReadonlySet<string>,
): number {
  return percentOf(module.lessons.filter((l) => completed.has(l.id)).length, module.lessons.length);
}

// Progres level di halaman kursus menghitung semua lesson, termasuk yang levelnya terkunci.
export function levelCompletionPercent(level: TreeLevel, completed: ReadonlySet<string>): number {
  const ids = lessonIds(level);
  return percentOf(ids.filter((id) => completed.has(id)).length, ids.length);
}

// Lesson pertama yang belum selesai di sebuah modul, atau lesson terakhir bila semuanya selesai.
export function resumeLessonId(
  module: TreeLevel['modules'][number],
  completed: ReadonlySet<string>,
): string | null {
  if (module.lessons.length === 0) return null;
  const firstUncompleted = module.lessons.find((l) => !completed.has(l.id));
  return (firstUncompleted ?? module.lessons[module.lessons.length - 1]).id;
}
