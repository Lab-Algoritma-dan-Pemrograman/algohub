// @vitest-environment node
import { ApiError } from '@algohub/api-client';
import { describe, expect, it } from 'vitest';
import {
  QUIZ_GENERIC,
  QUIZ_NOT_CONFIGURED,
  QUIZ_NOT_FOUND,
  QUIZ_SERVER_MISCONFIGURED,
  quizErrorMessage,
} from '../src/lib/quiz';

describe('quizErrorMessage', () => {
  it('404 berarti soal tidak ada', () => {
    expect(quizErrorMessage(new ApiError('Lesson not found', 404))).toBe(QUIZ_NOT_FOUND);
  });

  it('500 bank soal berarti server salah konfigurasi', () => {
    expect(
      quizErrorMessage(
        new ApiError(
          'Server tidak dapat membaca bank soal. Periksa konfigurasi Supabase di server.',
          500,
        ),
      ),
    ).toBe(QUIZ_SERVER_MISCONFIGURED);
  });

  it('400 berarti kuis belum dikonfigurasi', () => {
    expect(quizErrorMessage(new ApiError('Lesson does not have a quiz configured', 400))).toBe(
      QUIZ_NOT_CONFIGURED,
    );
  });

  it('galat lain, termasuk 500 umum dan jaringan, memakai pesan umum', () => {
    expect(quizErrorMessage(new ApiError('Internal server error', 500))).toBe(QUIZ_GENERIC);
    expect(quizErrorMessage(new TypeError('Failed to fetch'))).toBe(QUIZ_GENERIC);
  });

  it('pesan 404 dan salah konfigurasi berbeda', () => {
    expect(QUIZ_NOT_FOUND).not.toBe(QUIZ_SERVER_MISCONFIGURED);
  });
});
