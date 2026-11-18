import { ApiError } from '@algohub/api-client';
import { fireEvent, render, screen, waitFor } from '@testing-library/svelte';
import { describe, expect, it, vi } from 'vitest';
import Quiz from '../src/lib/components/Quiz.svelte';
import { QUIZ_GENERIC, QUIZ_NOT_FOUND, QUIZ_SERVER_MISCONFIGURED } from '../src/lib/quiz';

const options = ['satu', 'dua', 'tiga'];

const setup = (props: Partial<Parameters<typeof render<typeof Quiz>>[1]> = {}) => {
  const validate = vi.fn();
  const oncomplete = vi.fn();
  render(Quiz, { question: 'Berapa 1+1?', options, validate, oncomplete, ...props });
  return { validate, oncomplete };
};

const choose = async (text: string) => {
  await fireEvent.click(screen.getByRole('button', { name: text }));
  await fireEvent.click(screen.getByRole('button', { name: 'Periksa Jawaban' }));
};

describe('Quiz', () => {
  it('tombol periksa mati sampai ada pilihan', () => {
    setup();
    expect(screen.getByRole('button', { name: 'Periksa Jawaban' })).toBeDisabled();
  });

  it('dengan kunci lokal (staf) menilai tanpa memanggil API', async () => {
    const { validate, oncomplete } = setup({ correctAnswer: 1 });
    await choose('dua');
    expect(screen.getByText('Benar! Bagus sekali.')).toBeInTheDocument();
    expect(oncomplete).toHaveBeenCalledWith(true);
    expect(validate).not.toHaveBeenCalled();
  });

  it('tanpa kunci (praktikan) memvalidasi ke API dan menampilkan jawaban benar', async () => {
    const { validate, oncomplete } = setup();
    validate.mockResolvedValue({ isCorrect: false, correctAnswer: 2 });
    await choose('satu');
    await waitFor(() => expect(oncomplete).toHaveBeenCalledWith(false));
    expect(validate).toHaveBeenCalledWith(0);
    expect(screen.getByText('Kurang tepat. Jawaban yang benar adalah: tiga')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Coba Lagi' })).toBeInTheDocument();
  });

  it('Coba Lagi membuka kembali pilihan', async () => {
    const { validate } = setup();
    validate.mockResolvedValue({ isCorrect: false, correctAnswer: 2 });
    await choose('satu');
    await fireEvent.click(await screen.findByRole('button', { name: 'Coba Lagi' }));
    expect(screen.getByRole('button', { name: 'Periksa Jawaban' })).toBeDisabled();
    expect(screen.getByRole('button', { name: 'satu' })).toBeEnabled();
  });

  it.each([
    [new ApiError('Lesson not found', 404), QUIZ_NOT_FOUND],
    [
      new ApiError('Server tidak dapat membaca bank soal. Periksa ...', 500),
      QUIZ_SERVER_MISCONFIGURED,
    ],
    [new TypeError('Failed to fetch'), QUIZ_GENERIC],
  ])('galat validasi %# tampil sebagai pesan dan tidak menyelesaikan kuis', async (error, text) => {
    const { validate, oncomplete } = setup();
    validate.mockRejectedValue(error);
    await choose('dua');
    expect(await screen.findByRole('alert')).toHaveTextContent(text);
    expect(oncomplete).not.toHaveBeenCalled();
    expect(screen.getByRole('button', { name: 'Periksa Jawaban' })).toBeEnabled();
  });
});
