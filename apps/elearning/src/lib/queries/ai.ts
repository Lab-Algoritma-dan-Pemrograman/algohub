import { api } from '#lib/api.ts';

const HINT_FALLBACK =
  'Maaf, saya tidak bisa memberikan petunjuk saat ini. Coba periksa kembali sintaks Anda.';
const HINT_RATE_LIMITED =
  'Anda meminta petunjuk terlalu cepat. Silakan tunggu sesaat sebelum mencoba lagi.';
const HINT_LIMIT = 10;
const HINT_WINDOW_MS = 60_000;

export interface HintInput {
  lessonTitle: string;
  explanation: string;
  userCode: string;
  error: string;
  expectedOutput?: string;
}

export function buildHintPrompt({
  lessonTitle,
  explanation,
  userCode,
  error,
  expectedOutput,
}: HintInput): string {
  return `You are a Python tutor. A student is stuck on a lesson.
Lesson: ${lessonTitle}
Context: ${explanation}
Expected Output: ${expectedOutput || 'N/A'}
User's Code:
\`\`\`python
${userCode}
\`\`\`
Error/Output:
${error}
Provide a short, encouraging hint in Indonesian to help the student fix their code. 
Do NOT give the full solution immediately, but point out the logical or syntax error.
Keep it under 3 sentences.`;
}

let requestTimes: number[] = [];

// Batas klien 10 permintaan per menit; backend tetap memeriksa token dan kuotanya sendiri.
export function allowHintRequest(now = Date.now()): boolean {
  requestTimes = requestTimes.filter((t) => now - t < HINT_WINDOW_MS);
  if (requestTimes.length >= HINT_LIMIT) return false;
  requestTimes.push(now);
  return true;
}

export async function fetchCodeHint(input: HintInput): Promise<string> {
  if (!allowHintRequest()) return HINT_RATE_LIMITED;
  try {
    const { text } = await api.ai.generate({ prompt: buildHintPrompt(input) });
    return text;
  } catch {
    return HINT_FALLBACK;
  }
}
