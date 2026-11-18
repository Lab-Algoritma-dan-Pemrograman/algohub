// @vitest-environment node
import type { Lesson, Level, Module } from '@algohub/api-client/elearning';
import { describe, expect, it } from 'vitest';
import {
  buildCurriculum,
  isLevelLocked,
  pickContinueLesson,
  type TreeLevel,
} from '../src/lib/curriculum';
import { getLevelLanguage } from '../src/lib/levelLanguage';
const level = (id: string, sort_order: number, extra: Partial<Level> = {}): Level => ({
  id,
  title: id,
  description: null,
  access_mode: null,
  locked: null,
  sort_order,
  ...extra,
});
const mod = (id: string, level_id: string, sort_order = 1): Module => ({
  id,
  level_id,
  title: id,
  sort_order,
});
const lesson = (id: string, module_id: string, sort_order = 1) =>
  ({ id, module_id, title: id, sort_order }) as Lesson;
const tree = (): TreeLevel[] =>
  buildCurriculum(
    [level('l2', 2), level('l1', 1), level('l3', 3, { access_mode: 'locked' })],
    [mod('m1', 'l1'), mod('m2', 'l2'), mod('m3', 'l3')],
    [lesson('b', 'm1', 2), lesson('a', 'm1', 1), lesson('c', 'm2'), lesson('d', 'm3')],
  );
describe('buildCurriculum', () => {
  it('mengurutkan level dan lesson menurut sort_order', () => {
    const curriculum = tree();
    expect(curriculum.map((l) => l.id)).toEqual(['l1', 'l2', 'l3']);
    expect(curriculum[0].modules[0].lessons.map((l) => l.id)).toEqual(['a', 'b']);
  });
});
describe('isLevelLocked', () => {
  it('level pertama selalu terbuka, berikutnya menunggu level sebelumnya selesai', () => {
    const curriculum = tree();
    expect(isLevelLocked(curriculum, 0, new Set(), {})).toBe(false);
    expect(isLevelLocked(curriculum, 1, new Set(['a']), {})).toBe(true);
    expect(isLevelLocked(curriculum, 1, new Set(['a', 'b']), {})).toBe(false);
  });
  it('mode level locked mengunci walau level sebelumnya selesai', () => {
    expect(isLevelLocked(tree(), 2, new Set(['a', 'b', 'c']), {})).toBe(true);
  });
  it('override per user mengalahkan mode level', () => {
    const curriculum = tree();
    expect(isLevelLocked(curriculum, 2, new Set(), { l3: 'unlocked' })).toBe(false);
    expect(isLevelLocked(curriculum, 0, new Set(), { l1: 'locked' })).toBe(true);
  });
  it('level yang di-unlock eksplisit tidak menahan level sesudahnya', () => {
    expect(isLevelLocked(tree(), 1, new Set(), { l1: 'unlocked' })).toBe(false);
  });
});
describe('pickContinueLesson', () => {
  it('tanpa checkpoint membuka lesson pertama yang belum selesai dari awal', () => {
    expect(pickContinueLesson(tree(), new Set(['a']), {}, null)).toEqual({
      lessonId: 'b',
      resetStep: true,
    });
  });
  it('checkpoint yang belum selesai dilanjutkan di langkah terakhir', () => {
    expect(pickContinueLesson(tree(), new Set(['a']), {}, { lessonId: 'b' })).toEqual({
      lessonId: 'b',
      resetStep: false,
    });
  });
  it('lesson belum selesai sebelum checkpoint (setelah reset admin) didahulukan', () => {
    expect(pickContinueLesson(tree(), new Set(['b']), {}, { lessonId: 'b' })).toEqual({
      lessonId: 'a',
      resetStep: true,
    });
  });
  it('checkpoint basi pindah ke lesson sesudahnya', () => {
    expect(pickContinueLesson(tree(), new Set(['a', 'b']), {}, { lessonId: 'b' })).toEqual({
      lessonId: 'c',
      resetStep: false,
    });
  });
  it('semua selesai membuka lesson terakhir untuk ditinjau', () => {
    expect(pickContinueLesson(tree(), new Set(['a', 'b', 'c']), {}, null)).toEqual({
      lessonId: 'c',
      resetStep: false,
    });
  });
  it('tanpa level terbuka mengembalikan null', () => {
    expect(pickContinueLesson([], new Set(), {}, null)).toBeNull();
  });
});
describe('getLevelLanguage', () => {
  it('judul menentukan bahasa, prefix ID dan urutan hanya fallback', () => {
    expect(getLevelLanguage({ title: 'Dasar PYTHON' })).toBe('python');
    expect(getLevelLanguage({ title: 'Pengantar C' })).toBe('c');
    expect(getLevelLanguage({ id: 'c-level-1', title: 'Level' })).toBe('c');
    expect(getLevelLanguage({ id: 'level-1', title: 'Level' }, 0)).toBe('c');
    expect(getLevelLanguage({ id: 'level-5', title: 'Level' }, 4)).toBe('python');
  });
});
