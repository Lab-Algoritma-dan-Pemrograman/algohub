import { ApiError } from '@algohub/api-client';

export const QUIZ_NOT_FOUND = 'Soal kuis untuk pelajaran ini tidak ditemukan.';
export const QUIZ_NOT_CONFIGURED = 'Pelajaran ini belum memiliki kuis yang dikonfigurasi.';
export const QUIZ_SERVER_MISCONFIGURED =
  'Server tidak dapat membaca bank soal. Periksa konfigurasi Supabase di server.';
export const QUIZ_GENERIC = 'Gagal memvalidasi jawaban kuis. Harap periksa koneksi internet Anda.';

// Backend membedakan "soal tidak ada" (404) dari "server salah konfigurasi" (500 bank soal);
// keduanya dulu sama-sama tampil sebagai kegagalan umum.
export function quizErrorMessage(error: unknown): string {
  if (!(error instanceof ApiError)) return QUIZ_GENERIC;
  if (error.status === 404) return QUIZ_NOT_FOUND;
  if (error.status === 400) return QUIZ_NOT_CONFIGURED;
  if (error.status === 500 && error.message.includes('bank soal')) return QUIZ_SERVER_MISCONFIGURED;
  return QUIZ_GENERIC;
}
