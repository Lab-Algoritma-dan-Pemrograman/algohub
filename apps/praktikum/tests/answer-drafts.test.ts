// @vitest-environment jsdom
import { beforeEach, describe, expect, it, vi } from 'vitest';
import {
  clearAllAnswerDrafts,
  clearAnswerDraft,
  loadAnswerDraft,
  saveAnswerDraft,
} from '../src/lib/answer-drafts';

describe('draft jawaban lokal', () => {
  beforeEach(() => {
    localStorage.clear();
    vi.useRealTimers();
  });

  it('menyimpan draft terpisah per user dan course', () => {
    saveAnswerDraft(7, 10, 2, { 101: 'jawaban A' });
    saveAnswerDraft(8, 10, 2, { 101: 'jawaban B' });

    expect(loadAnswerDraft(7, 10, 2)).toEqual({ 101: 'jawaban A' });
    expect(loadAnswerDraft(8, 10, 2)).toEqual({ 101: 'jawaban B' });
    expect(loadAnswerDraft(7, 10, 3)).toEqual({});
  });

  it('menghapus draft setelah semua jawaban terkonfirmasi server', () => {
    saveAnswerDraft(7, 10, 2, { 101: 'jawaban' });
    clearAnswerDraft(7, 10, 2);
    expect(loadAnswerDraft(7, 10, 2)).toEqual({});
  });

  it('membersihkan seluruh draft saat logout', () => {
    saveAnswerDraft(7, 10, 2, { 101: 'A' });
    saveAnswerDraft(7, 10, 3, { 102: 'B' });
    localStorage.setItem('token', 'tetap-diuji-terpisah');

    clearAllAnswerDrafts();

    expect(loadAnswerDraft(7, 10, 2)).toEqual({});
    expect(loadAnswerDraft(7, 10, 3)).toEqual({});
    expect(localStorage.getItem('token')).toBe('tetap-diuji-terpisah');
  });

  it('membuang draft kedaluwarsa', () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date('2026-11-01T00:00:00Z'));
    saveAnswerDraft(7, 10, 2, { 101: 'lama' });
    vi.setSystemTime(new Date('2026-11-09T00:00:01Z'));
    expect(loadAnswerDraft(7, 10, 2)).toEqual({});
  });
});
