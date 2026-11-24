const PREFIX = 'algohub-praktikum-answer-draft:';
const MAX_AGE_MS = 7 * 24 * 60 * 60 * 1000;

interface StoredDraft {
  answers: Record<string, string>;
  updatedAt: number;
}

const hasStorage = () => typeof localStorage !== 'undefined';

function key(userId: number, aktivasiSesiId: number, courseId: number): string {
  return `${PREFIX}${userId}:${aktivasiSesiId}:${courseId}`;
}

// Draft hanya berisi jawaban yang belum dikonfirmasi server. Penulisan ke
// localStorage sengaja sinkron agar refresh/crash tab tidak menghilangkan ketikan.
export function saveAnswerDraft(
  userId: number,
  aktivasiSesiId: number,
  courseId: number,
  answers: Record<number, string>,
): void {
  if (!hasStorage()) return;
  const storageKey = key(userId, aktivasiSesiId, courseId);
  if (Object.keys(answers).length === 0) {
    localStorage.removeItem(storageKey);
    return;
  }
  try {
    localStorage.setItem(storageKey, JSON.stringify({ answers, updatedAt: Date.now() }));
  } catch {
    // Quota browser tidak boleh merusak editor; autosave server tetap berjalan.
  }
}

export function loadAnswerDraft(
  userId: number,
  aktivasiSesiId: number,
  courseId: number,
): Record<number, string> {
  if (!hasStorage()) return {};
  const storageKey = key(userId, aktivasiSesiId, courseId);
  try {
    const raw = localStorage.getItem(storageKey);
    if (!raw) return {};
    const parsed = JSON.parse(raw) as StoredDraft;
    if (
      !parsed ||
      typeof parsed.updatedAt !== 'number' ||
      Date.now() - parsed.updatedAt > MAX_AGE_MS ||
      !parsed.answers ||
      typeof parsed.answers !== 'object'
    ) {
      localStorage.removeItem(storageKey);
      return {};
    }
    const answers: Record<number, string> = {};
    for (const [id, value] of Object.entries(parsed.answers)) {
      const numericID = Number(id);
      if (Number.isInteger(numericID) && typeof value === 'string') answers[numericID] = value;
    }
    return answers;
  } catch {
    localStorage.removeItem(storageKey);
    return {};
  }
}

export function clearAnswerDraft(userId: number, aktivasiSesiId: number, courseId: number): void {
  if (hasStorage()) localStorage.removeItem(key(userId, aktivasiSesiId, courseId));
}

export function clearAllAnswerDrafts(): void {
  if (!hasStorage()) return;
  for (let index = localStorage.length - 1; index >= 0; index--) {
    const storageKey = localStorage.key(index);
    if (storageKey?.startsWith(PREFIX)) localStorage.removeItem(storageKey);
  }
}
