import { elearningKeys } from '@algohub/api-client/elearning';
import { api } from '#lib/api.ts';

// Jawaban kuis tidak boleh tersimpan di cache persist: ini mutasi dengan kunci sendiri, dan
// hasilnya hanya hidup di state komponen Quiz.
export const validateQuizOptions = (lessonId: string) => ({
  mutationKey: [...elearningKeys.all, 'quiz', 'validate', lessonId] as const,
  mutationFn: (selectedOption: number) => api.validateQuiz({ lessonId, selectedOption }),
});
