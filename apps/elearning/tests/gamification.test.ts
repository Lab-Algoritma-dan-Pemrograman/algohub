import { ApiError } from '@algohub/api-client';
import { describe, expect, it } from 'vitest';
import {
  bugHuntXp,
  isWeeklyLimitError,
  jakartaWeekLabel,
  leaderboardFilter,
  leaderboardParams,
  lessonsToNextLevel,
  prestigeErrorMessage,
  semesterLabelJakarta,
  studentWeeklyParams,
  xpProgress,
} from '../src/lib/gamification';

describe('parameter leaderboard', () => {
  it('filter kosong tidak ikut terkirim', () => {
    expect(leaderboardFilter({ kelas: '', jurusan: '', shift: '' })).toEqual({});
    expect(leaderboardFilter({ kelas: 'A1', jurusan: '', shift: '2' })).toEqual({
      kelas: 'A1',
      shift: '2',
    });
  });

  it('limit dari pilihan Top N diteruskan, 0 berarti semua', () => {
    const view = { kelas: 'A1', jurusan: 'TI', shift: '', limit: 10 };
    expect(leaderboardParams(view)).toEqual({ kelas: 'A1', jurusan: 'TI', limit: 10 });
    expect(leaderboardParams({ ...view, limit: 0 }).limit).toBe(0);
  });

  it('mahasiswa melihat top 3 kelas, jurusan, dan shift miliknya', () => {
    expect(studentWeeklyParams({ kelas: 'A1', jurusan: 'TI', shift: '2' })).toEqual({
      limit: 3,
      kelas: 'A1',
      jurusan: 'TI',
      shift: '2',
    });
    expect(studentWeeklyParams({ kelas: 'A1', jurusan: null, shift: null })).toEqual({
      limit: 3,
      kelas: 'A1',
    });
  });
});

describe('periode Jakarta', () => {
  it('label minggu dimulai Senin WIB walau UTC masih Minggu malam', () => {
    // Minggu 18 Okt 2026 18:00 UTC = Senin 19 Okt 2026 01:00 WIB.
    expect(jakartaWeekLabel(new Date('2026-10-18T18:00:00Z'))).toBe('19–25 Okt 2026 • WIB');
  });

  it('label lintas bulan memuat kedua bulan', () => {
    expect(jakartaWeekLabel(new Date('2026-10-28T05:00:00Z'))).toBe('26 Okt – 1 Nov 2026 • WIB');
  });

  it('label semester mengikuti kalender Jakarta', () => {
    expect(semesterLabelJakarta(new Date('2026-03-10T00:00:00Z'))).toBe('Genap Feb–Jul 2026');
    expect(semesterLabelJakarta(new Date('2026-09-10T00:00:00Z'))).toBe('Ganjil Agu 2026–Jan 2027');
    expect(semesterLabelJakarta(new Date('2027-01-10T00:00:00Z'))).toBe('Ganjil Agu 2026–Jan 2027');
  });
});

describe('Bug Hunt dan prestige', () => {
  it('XP: 20 dasar ditambah 2 per detik tersisa, salah 0', () => {
    expect(bugHuntXp(true, 10)).toBe(40);
    expect(bugHuntXp(true, 0)).toBe(20);
    expect(bugHuntXp(false, 14)).toBe(0);
  });

  it('hanya 400 WEEKLY_LIMIT_REACHED yang dianggap kuota habis', () => {
    expect(isWeeklyLimitError(new ApiError('WEEKLY_LIMIT_REACHED', 400))).toBe(true);
    expect(isWeeklyLimitError(new ApiError('WEEKLY_LIMIT_REACHED', 500))).toBe(false);
    expect(isWeeklyLimitError(new ApiError('lain', 400))).toBe(false);
    expect(isWeeklyLimitError(new Error('WEEKLY_LIMIT_REACHED'))).toBe(false);
  });

  it('pesan prestige dari backend ditampilkan apa adanya', () => {
    expect(prestigeErrorMessage(new ApiError('Akses ditolak.', 403))).toBe('Akses ditolak.');
    expect(prestigeErrorMessage(new ApiError('Cooldown: reset berikutnya 3 hari lagi', 400))).toBe(
      'Cooldown: reset berikutnya 3 hari lagi',
    );
    expect(prestigeErrorMessage(new TypeError('Failed to fetch'))).toBe('Gagal prestige.');
  });

  it('progres level memakai 2000 XP per level', () => {
    expect(xpProgress(4500)).toEqual({ current: 500, percent: 25 });
    expect(lessonsToNextLevel(1950)).toBe(1);
  });
});
