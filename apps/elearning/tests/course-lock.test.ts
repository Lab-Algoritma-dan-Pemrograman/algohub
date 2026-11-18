// @vitest-environment node
import type { Lesson, Level, Module } from '@algohub/api-client/elearning';
import { describe, expect, it } from 'vitest';
import {
  buildCurriculum,
  isLessonLevelLocked,
  isModuleLocked,
  levelCompletionPercent,
  moduleProgressPercent,
  resumeLessonId,
} from '../src/lib/curriculum';

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
const les = (id: string, module_id: string, sort_order = 1) =>
  ({ id, module_id, title: id, sort_order }) as Lesson;

const tree = (l2: Partial<Level> = {}) =>
  buildCurriculum(
    [level('l1', 1), level('l2', 2, l2)],
    [mod('m1', 'l1', 1), mod('m2', 'l1', 2), mod('m3', 'l2', 1)],
    [les('a', 'm1', 1), les('b', 'm1', 2), les('c', 'm2', 1), les('d', 'm3', 1)],
  );
const done = (...ids: string[]) => new Set(ids);

describe('isModuleLocked', () => {
  it('modul pertama level pertama selalu terbuka, berikutnya menunggu modul sebelumnya', () => {
    const t = tree();
    expect(isModuleLocked(t, 0, 0, done(), {})).toBe(false);
    expect(isModuleLocked(t, 0, 1, done('a'), {})).toBe(true);
    expect(isModuleLocked(t, 0, 1, done('a', 'b'), {})).toBe(false);
  });

  it('level berikutnya menunggu seluruh level sebelumnya selesai', () => {
    const t = tree();
    expect(isModuleLocked(t, 1, 0, done('a', 'b'), {})).toBe(true);
    expect(isModuleLocked(t, 1, 0, done('a', 'b', 'c'), {})).toBe(false);
  });

  it('mode level dan override per user mengalahkan urutan', () => {
    expect(isModuleLocked(tree({ access_mode: 'unlocked' }), 1, 0, done(), {})).toBe(false);
    expect(isModuleLocked(tree({ access_mode: 'locked' }), 1, 0, done('a', 'b', 'c'), {})).toBe(
      true,
    );
    expect(isModuleLocked(tree(), 1, 0, done(), { l2: 'unlocked' })).toBe(false);
    expect(isModuleLocked(tree(), 0, 0, done(), { l1: 'locked' })).toBe(true);
  });

  it('level yang tidak ada dianggap terkunci', () => {
    expect(isModuleLocked(tree(), 9, 0, done(), {})).toBe(true);
  });
});

describe('isLessonLevelLocked', () => {
  it('level auto menunggu semua lesson level sebelumnya', () => {
    const t = tree();
    expect(isLessonLevelLocked(t, 0, done(), {})).toBe(false);
    expect(isLessonLevelLocked(t, 1, done('a', 'b'), {})).toBe(true);
    expect(isLessonLevelLocked(t, 1, done('a', 'b', 'c'), {})).toBe(false);
  });

  it('flag locked=true mengunci, mode unlocked dan override membuka', () => {
    expect(isLessonLevelLocked(tree({ locked: true }), 1, done('a', 'b', 'c'), {})).toBe(true);
    expect(isLessonLevelLocked(tree({ access_mode: 'unlocked' }), 1, done(), {})).toBe(false);
    expect(isLessonLevelLocked(tree(), 1, done(), { l2: 'unlocked' })).toBe(false);
    expect(isLessonLevelLocked(tree(), 1, done('a', 'b', 'c'), { l2: 'locked' })).toBe(true);
  });

  it('level sebelumnya yang di-unlock tetap harus selesai (beda dari dashboard)', () => {
    const t = buildCurriculum(
      [level('l1', 1, { access_mode: 'unlocked' }), level('l2', 2)],
      [mod('m1', 'l1'), mod('m2', 'l2')],
      [les('a', 'm1'), les('b', 'm2')],
    );
    expect(isLessonLevelLocked(t, 1, done(), {})).toBe(true);
  });
});

describe('progres dan lanjutkan', () => {
  it('persen modul dan level', () => {
    const t = tree();
    expect(moduleProgressPercent(t[0].modules[0], done('a'))).toBe(50);
    expect(levelCompletionPercent(t[0], done('a', 'b'))).toBe(67);
    expect(levelCompletionPercent({ ...t[0], modules: [] }, done())).toBe(0);
  });

  it('resumeLessonId: pertama yang belum selesai, atau terakhir bila semua selesai', () => {
    const m = tree()[0].modules[0];
    expect(resumeLessonId(m, done('a'))).toBe('b');
    expect(resumeLessonId(m, done('a', 'b'))).toBe('b');
    expect(resumeLessonId({ ...m, lessons: [] }, done())).toBeNull();
  });
});
