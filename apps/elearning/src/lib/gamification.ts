import { ApiError } from '@algohub/api-client';
import type { LeaderboardFilter, PageParams } from '@algohub/api-client/elearning';

export const WEEKLY_LIMIT_MESSAGE = 'WEEKLY_LIMIT_REACHED';
export const BUG_HUNT_SECONDS = 15;
export const XP_PER_LEVEL = 2000;
// Batas atas daftar "Semua" di backend: limit=0 dijawab 100 baris.
export const LEADERBOARD_ALL = 0;

const MONTHS_ID = [
  'Jan',
  'Feb',
  'Mar',
  'Apr',
  'Mei',
  'Jun',
  'Jul',
  'Agu',
  'Sep',
  'Okt',
  'Nov',
  'Des',
];
const JAKARTA_OFFSET_MS = 7 * 60 * 60 * 1000;

// Port getJakartaWeek: hanya label periode untuk header; angka mingguan dihitung backend.
export function jakartaWeekLabel(now: Date = new Date()): string {
  const jakarta = new Date(now.getTime() + JAKARTA_OFFSET_MS);
  const daysSinceMonday = (jakarta.getUTCDay() + 6) % 7;
  const y = jakarta.getUTCFullYear();
  const m = jakarta.getUTCMonth();
  const monday = jakarta.getUTCDate() - daysSinceMonday;
  const start = new Date(Date.UTC(y, m, monday));
  const end = new Date(Date.UTC(y, m, monday + 6));
  const fmt = (d: Date) => `${d.getUTCDate()} ${MONTHS_ID[d.getUTCMonth()]}`;
  const range =
    start.getUTCMonth() === end.getUTCMonth()
      ? `${start.getUTCDate()}–${end.getUTCDate()} ${MONTHS_ID[start.getUTCMonth()]} ${start.getUTCFullYear()}`
      : `${fmt(start)} – ${fmt(end)} ${end.getUTCFullYear()}`;
  return `${range} • WIB`;
}

// Port semesterLabelJakarta; backend memakai pembagian yang sama untuk kuota prestige.
export function semesterLabelJakarta(now: Date = new Date()): string {
  const jakarta = new Date(now.getTime() + JAKARTA_OFFSET_MS);
  const m = jakarta.getUTCMonth();
  const y = jakarta.getUTCFullYear();
  if (m >= 1 && m <= 6) return `Genap Feb–Jul ${y}`;
  if (m >= 7) return `Ganjil Agu ${y}–Jan ${y + 1}`;
  return `Ganjil Agu ${y - 1}–Jan ${y}`;
}

export interface LeaderboardView {
  kelas: string;
  jurusan: string;
  shift: string;
  limit: number;
}

export const DEFAULT_LEADERBOARD_LIMIT = 20;

// Filter kosong tidak dikirim (qs() membuangnya), dan limit mengikuti pilihan Top 3/10/20/Semua.
export function leaderboardFilter(view: Pick<LeaderboardView, 'kelas' | 'jurusan' | 'shift'>) {
  const filter: LeaderboardFilter = {};
  if (view.kelas) filter.kelas = view.kelas;
  if (view.jurusan) filter.jurusan = view.jurusan;
  if (view.shift) filter.shift = view.shift;
  return filter;
}

export function leaderboardParams(view: LeaderboardView): LeaderboardFilter & PageParams {
  return { ...leaderboardFilter(view), limit: view.limit };
}

// Mahasiswa selalu melihat papan mingguan kelasnya, top 3; staf boleh memilih periode dan filter.
export function studentWeeklyParams(user: {
  kelas: string;
  jurusan: string | null;
  shift: string | null;
}): PageParams & LeaderboardFilter {
  return {
    limit: 3,
    ...leaderboardFilter({
      kelas: user.kelas,
      jurusan: user.jurusan ?? '',
      shift: user.shift ? String(user.shift) : '',
    }),
  };
}

export const isWeeklyLimitError = (error: unknown): boolean =>
  error instanceof ApiError && error.status === 400 && error.message === WEEKLY_LIMIT_MESSAGE;

// Pesan 403 prestige dari backend adalah "Akses ditolak."; pesan lain (syarat, cooldown, kuota)
// sudah berbahasa Indonesia dan ditampilkan apa adanya.
export function prestigeErrorMessage(error: unknown): string {
  if (error instanceof ApiError && error.message) return error.message;
  return 'Gagal prestige.';
}

export function bugHuntXp(correct: boolean, secondsLeft: number): number {
  return correct ? 20 + Math.round(secondsLeft * 2) : 0;
}

export function xpProgress(xp: number): { current: number; percent: number } {
  const current = xp % XP_PER_LEVEL;
  return { current, percent: Math.round((current / XP_PER_LEVEL) * 100) };
}

export function lessonsToNextLevel(xp: number): number {
  return Math.ceil((XP_PER_LEVEL - (xp % XP_PER_LEVEL)) / 50);
}
