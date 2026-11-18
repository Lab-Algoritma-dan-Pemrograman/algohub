<script lang="ts">
  import type { ValidateQuizResponse } from '@algohub/api-client/elearning';
  import CheckCircle from '@lucide/svelte/icons/circle-check';
  import Loader from '@lucide/svelte/icons/loader-circle';
  import XCircle from '@lucide/svelte/icons/circle-x';
  import ContentRenderer from '#lib/components/ContentRenderer.svelte';
  import { looksLikeHtml } from '#lib/lesson.ts';
  import { quizErrorMessage } from '#lib/quiz.ts';

  let {
    question,
    options,
    correctAnswer,
    validate,
    oncomplete,
  }: {
    question: string;
    options: string[];
    // Hanya ada untuk staf; praktikan selalu divalidasi lewat `validate`.
    correctAnswer?: number;
    validate: (selected: number) => Promise<ValidateQuizResponse>;
    oncomplete: (correct: boolean) => void;
  } = $props();

  let selected = $state<number | null>(null);
  let submitted = $state(false);
  let serverAnswer = $state<number | null>(null);
  let loading = $state(false);
  let errorMessage = $state('');

  const answer = $derived(serverAnswer ?? correctAnswer);
  const isCorrect = $derived(selected !== null && selected === answer);

  async function submit() {
    if (selected === null || loading) return;
    errorMessage = '';
    if (typeof correctAnswer === 'number' && correctAnswer >= 0) {
      submitted = true;
      oncomplete(selected === correctAnswer);
      return;
    }
    loading = true;
    try {
      const res = await validate(selected);
      serverAnswer = Number(res.correctAnswer);
      submitted = true;
      oncomplete(res.isCorrect);
    } catch (error) {
      errorMessage = quizErrorMessage(error);
    } finally {
      loading = false;
    }
  }

  function retry() {
    submitted = false;
    selected = null;
  }
</script>

<div class="space-y-8 rounded-3xl border border-zinc-200 bg-white p-8 shadow-sm">
  <div class="space-y-2">
    <div class="text-xs font-black tracking-widest text-rose-800 uppercase">Kuis Singkat</div>
    {#if looksLikeHtml(question)}
      <ContentRenderer
        content={question}
        class="text-2xl font-bold text-zinc-900 prose-p:text-2xl prose-p:font-bold prose-p:text-zinc-900"
      />
    {:else}
      <h3 class="text-2xl font-bold text-zinc-900">{question}</h3>
    {/if}
  </div>

  <div class="space-y-3">
    {#each options as option, idx (idx)}
      {@const isAnswer = submitted && idx === answer}
      {@const isWrong = submitted && selected === idx && idx !== answer}
      <button
        type="button"
        disabled={submitted}
        onclick={() => (selected = idx)}
        class="group flex w-full items-center justify-between rounded-2xl border-2 p-4 text-left transition-all {isWrong
          ? 'border-red-500 bg-red-50'
          : isAnswer || selected === idx
            ? 'border-rose-700 bg-rose-50'
            : 'border-zinc-100 bg-zinc-50 hover:border-zinc-200'}"
      >
        <span class="font-medium {selected === idx ? 'text-rose-900' : 'text-zinc-600'}">
          {option}
        </span>
        {#if isAnswer}
          <CheckCircle class="text-rose-700" size={20} />
        {:else if isWrong}
          <XCircle class="text-red-500" size={20} />
        {/if}
      </button>
    {/each}
  </div>

  {#if errorMessage}
    <p role="alert" class="rounded-2xl bg-red-50 p-4 text-center text-sm font-bold text-red-700">
      {errorMessage}
    </p>
  {/if}

  {#if !submitted}
    <button
      type="button"
      onclick={submit}
      disabled={selected === null || loading}
      class="flex w-full items-center justify-center gap-2 rounded-2xl bg-zinc-900 py-4 font-bold text-white transition-all hover:bg-zinc-800 active:scale-95 disabled:opacity-50"
    >
      {#if loading}
        <Loader class="h-5 w-5 animate-spin" />
        Memvalidasi...
      {:else}
        Periksa Jawaban
      {/if}
    </button>
  {:else}
    <div class="space-y-4">
      <div
        class="rounded-2xl p-4 text-center font-bold {isCorrect
          ? 'bg-rose-100 text-rose-900'
          : 'bg-red-100 text-red-700'}"
      >
        {isCorrect
          ? 'Benar! Bagus sekali.'
          : `Kurang tepat. Jawaban yang benar adalah: ${typeof answer === 'number' ? (options[answer] ?? '') : ''}`}
      </div>
      {#if !isCorrect}
        <button
          type="button"
          onclick={retry}
          class="w-full rounded-2xl bg-zinc-100 py-3 font-bold text-zinc-900 transition-all hover:bg-zinc-200 active:scale-95"
        >
          Coba Lagi
        </button>
      {/if}
    </div>
  {/if}
</div>
