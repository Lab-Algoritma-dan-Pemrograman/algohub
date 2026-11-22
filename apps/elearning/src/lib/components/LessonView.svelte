<script lang="ts">
  import { goto } from '$app/navigation';
  import { resolve } from '$app/paths';
  import { messages } from '@algohub/runner';
  import { toast } from '@algohub/ui';
  import ArrowLeft from '@lucide/svelte/icons/arrow-left';
  import BookOpen from '@lucide/svelte/icons/book-open';
  import CheckCircle from '@lucide/svelte/icons/circle-check';
  import ChevronLeft from '@lucide/svelte/icons/chevron-left';
  import ChevronRight from '@lucide/svelte/icons/chevron-right';
  import Lightbulb from '@lucide/svelte/icons/lightbulb';
  import Loader from '@lucide/svelte/icons/loader-circle';
  import Menu from '@lucide/svelte/icons/menu';
  import TriangleAlert from '@lucide/svelte/icons/triangle-alert';
  import Trophy from '@lucide/svelte/icons/trophy';
  import { createMutation } from '@tanstack/svelte-query';
  import confetti from 'canvas-confetti';
  import { onDestroy, untrack } from 'svelte';
  import { fade, fly, scale, slide } from 'svelte/transition';
  import ContentRenderer from '#lib/components/ContentRenderer.svelte';
  import FlowchartPuzzle from '#lib/components/FlowchartPuzzle.svelte';
  import LessonDrawer from '#lib/components/LessonDrawer.svelte';
  import LessonEditor from '#lib/components/LessonEditor.svelte';
  import PlotDisplay from '#lib/components/PlotDisplay.svelte';
  import Quiz from '#lib/components/Quiz.svelte';
  import type { AccessOverrides, TreeLevel } from '#lib/curriculum.ts';
  import {
    STEP_ORDER,
    adjacentLessonId,
    lessonLanguage,
    looksLikeHtml,
    nextLessonAfter,
    parseLesson,
    type LessonPosition,
    type LessonStep,
  } from '#lib/lesson.ts';
  import { gradeExercise, toLessonOutput, type LessonOutput } from '#lib/lesson-run.ts';
  import { completeLessonOptions } from '#lib/queries/progress.ts';
  import { validateQuizOptions } from '#lib/queries/quiz.ts';
  import { elearningRunner } from '#lib/runner.ts';
  import { appStore } from '#lib/stores/app.svelte.ts';

  // Satu instance per lesson: halaman induk membungkusnya dengan {#key} sehingga seluruh state
  // lokal di bawah ini (langkah, kode, hasil run) otomatis mulai dari awal tiap pindah lesson.
  let {
    curriculum,
    pos,
    completed,
    overrides,
  }: {
    curriculum: TreeLevel[];
    pos: LessonPosition;
    completed: ReadonlySet<string>;
    overrides: AccessOverrides;
  } = $props();

  const lessonId = $derived(pos.lesson.id);
  const lesson = $derived(parseLesson(pos.lesson));
  const level = $derived(curriculum[pos.levelIdx]);
  const moduleTitle = $derived(level?.modules[pos.moduleIdx]?.title ?? 'Pelajaran');
  const done = $derived(completed.has(lessonId));
  const hasQuiz = $derived(lesson.quiz !== null);

  const runner = elearningRunner();

  const read = (key: string): string | null => {
    try {
      return localStorage.getItem(key);
    } catch {
      return null;
    }
  };
  const write = (key: string, value: string) => {
    try {
      localStorage.setItem(key, value);
    } catch {
      // Penyimpanan penuh atau diblokir: lesson tetap jalan, hanya tanpa fitur lanjutkan.
    }
  };
  const forget = (...keys: string[]) => {
    try {
      for (const key of keys) localStorage.removeItem(key);
    } catch {
      // Sama seperti write: tidak fatal.
    }
  };

  // Lesson yang sudah selesai dibuka dalam mode tinjau dari awal; yang belum, dilanjutkan dari
  // langkah dan kode terakhir yang tersimpan di perangkat.
  const initial = untrack(() => {
    const view = parseLesson(pos.lesson);
    const id = view.id;
    const template = view.initialCode || view.codeExample;
    const base = {
      sandboxCode: view.codeExample,
      quizPassed: read(`lesson-quiz:${id}`) === '1',
    };
    if (completed.has(id)) return { ...base, step: 'learn' as LessonStep, code: template };
    const saved = read(`lesson-step:${id}`);
    return {
      ...base,
      step: (STEP_ORDER.find((s) => s === saved) ?? 'learn') as LessonStep,
      code: read(`lesson-code:${id}`) ?? template,
    };
  });

  let step = $state<LessonStep>(initial.step);
  let code = $state(initial.code);
  let isCorrect = $state<boolean | null>(null);
  let showHint = $state(false);
  let showSuccess = $state(false);
  let showDrawer = $state(false);
  let showResetConfirm = $state(false);
  let completing = $state(false);
  // Kuis yang sudah lulus tidak boleh hilang saat muat ulang, kalau tidak tombol Selesaikan
  // terkunci lagi tanpa penjelasan.
  let quizPassed = $state(initial.quizPassed);
  let running = $state(false);
  let exercise = $state<LessonOutput | null>(null);
  let sandboxCode = $state(initial.sandboxCode);
  let sandboxInput = $state('');
  let sandbox = $state<LessonOutput | null>(null);
  let quizTimer: ReturnType<typeof setTimeout> | null = null;

  const isQuizPassed = $derived(done || quizPassed || !hasQuiz);
  const language = $derived(
    lessonLanguage(code || lesson.initialCode || lesson.codeExample, level, pos.levelIdx),
  );
  const runnerStatus = $derived(runner.status[language]);
  const runnerLoading = $derived(runnerStatus === 'idle' || runnerStatus === 'loading');
  const stepIdx = $derived(STEP_ORDER.indexOf(step));
  const prevLessonId = $derived(
    adjacentLessonId(curriculum, lessonId, 'prev', completed, overrides),
  );
  const nextLessonId = $derived(
    adjacentLessonId(curriculum, lessonId, 'next', completed, overrides),
  );
  const canGoPrev = $derived(stepIdx > 0 || prevLessonId !== null);
  const canGoNext = $derived(stepIdx < STEP_ORDER.length - 1 || nextLessonId !== null);

  const completeMutation = createMutation(completeLessonOptions);
  const quizMutation = createMutation(() => validateQuizOptions(lessonId));

  // Compiler dimuat begitu lesson dibuka supaya Jalankan pertama tidak menunggu lama.
  $effect(() => {
    void runner.ensure(language);
  });

  // Titik lanjut untuk kartu "Lanjutkan Belajar" di dashboard.
  $effect(() => {
    if (done) return;
    write(`lesson-step:${lessonId}`, step);
    write('last-lesson-checkpoint', JSON.stringify({ lessonId, step, timestamp: Date.now() }));
  });

  $effect(() => {
    if (!done) write(`lesson-code:${lessonId}`, code);
  });

  onDestroy(() => cancelQuizAdvance());

  // Timer pindah otomatis dari kuis ke latihan harus mati saat user menekan Selanjutnya atau
  // berpindah lesson; kalau tidak, timer menyala di lesson berikutnya dan melompatkannya.
  function cancelQuizAdvance() {
    if (quizTimer !== null) clearTimeout(quizTimer);
    quizTimer = null;
  }

  function goToStep(next: LessonStep) {
    cancelQuizAdvance();
    step = next;
  }

  function openLesson(id: string) {
    cancelQuizAdvance();
    showDrawer = false;
    void goto(resolve('/(app)/lesson/[id]', { id }));
  }

  function handlePrev() {
    cancelQuizAdvance();
    if (stepIdx > 0) step = STEP_ORDER[stepIdx - 1];
    else if (prevLessonId) openLesson(prevLessonId);
  }

  function handleNext() {
    cancelQuizAdvance();
    if (stepIdx < STEP_ORDER.length - 1) step = STEP_ORDER[stepIdx + 1];
    else if (nextLessonId) openLesson(nextLessonId);
  }

  function onQuizComplete(correct: boolean) {
    if (!correct) return;
    if (!quizPassed && !done) {
      quizPassed = true;
      write(`lesson-quiz:${lessonId}`, '1');
    }
    cancelQuizAdvance();
    quizTimer = setTimeout(() => {
      quizTimer = null;
      step = 'code';
    }, 1500);
  }

  const celebrate = () =>
    void confetti({
      particleCount: 100,
      spread: 70,
      origin: { y: 0.6 },
      colors: ['#9f1239', '#e11d48', '#fb7185'],
    });

  function onFlowchartResult(correct: boolean) {
    isCorrect = correct;
    if (correct) celebrate();
  }

  // Run yang gagal dimulai (mis. program lain masih berjalan) tampil sebagai galat di panel
  // keluaran, bukan exception yang membuat halaman error.
  async function exec(source: string, stdin: string): Promise<LessonOutput> {
    try {
      return toLessonOutput(await runner.run(language, source, { stdin }));
    } catch (error) {
      return { output: '', error: (error as Error).message, images: [] };
    }
  }

  async function runSandbox() {
    if (running) return;
    running = true;
    sandbox = null;
    try {
      sandbox = await exec(sandboxCode, sandboxInput);
    } finally {
      running = false;
    }
  }

  async function runExercise() {
    if (running) return;
    running = true;
    exercise = null;
    try {
      const outcome = await gradeExercise({
        run: exec,
        code,
        language,
        testCases: lesson.testCases,
        rules: lesson.validationRules,
      });
      exercise = outcome;
      if (outcome.correct === null) return;
      isCorrect = outcome.correct;
      if (outcome.correct) celebrate();
    } finally {
      running = false;
    }
  }

  function resetExercise() {
    code = lesson.initialCode || lesson.codeExample;
    forget(`lesson-code:${lessonId}`);
    exercise = null;
    isCorrect = null;
    showResetConfirm = false;
  }

  async function finishLesson() {
    if (!isQuizPassed) {
      toast.error('Harap selesaikan kuis terlebih dahulu!');
      return;
    }
    if (completing) return;
    completing = true;
    try {
      // Lesson yang sudah selesai tidak memberi XP lagi, tetapi tetap boleh lanjut.
      if (!done) {
        const result = await completeMutation.mutateAsync({ lessonId, xpReward: lesson.xpReward });
        if (!result.alreadyCompleted) {
          appStore.patchProfile({
            xp: result.xp,
            level: result.level,
            streak: result.streak,
            lastActive: result.lastActive,
          });
          if (result.leveledUp) appStore.setLevelUp(result.level);
          appStore.pushAchievements(result.newlyUnlocked ?? []);
        }
      }
      const nextId = nextLessonAfter(curriculum, pos, new Set([...completed, lessonId]), overrides);
      forget(`lesson-step:${lessonId}`, `lesson-code:${lessonId}`, `lesson-quiz:${lessonId}`);
      if (nextId) {
        write(
          'last-lesson-checkpoint',
          JSON.stringify({ lessonId: nextId, step: 'learn', timestamp: Date.now() }),
        );
        openLesson(nextId);
      } else {
        forget('last-lesson-checkpoint');
        showSuccess = true;
      }
    } catch {
      toast.error('Gagal menyimpan progres pelajaran. Periksa koneksi lalu coba lagi.');
    } finally {
      completing = false;
    }
  }

  const stepLabels: { id: LessonStep; label: string }[] = [
    { id: 'learn', label: 'Pelajari' },
    { id: 'quiz', label: 'Kuis' },
    { id: 'code', label: 'Latihan' },
  ];
</script>

{#snippet stepIndicator(label: string, active: boolean, finished: boolean)}
  <div class="flex items-center gap-2">
    <div
      class="flex h-8 w-8 items-center justify-center rounded-full text-xs font-bold transition-all {active
        ? 'scale-110 bg-rose-700 text-white shadow-lg shadow-rose-700/20'
        : finished
          ? 'bg-rose-100 text-rose-700'
          : 'bg-zinc-100 text-zinc-400'}"
    >
      {#if finished}<CheckCircle size={16} />{:else}{label[0]}{/if}
    </div>
    <span
      class="text-xs font-bold tracking-widest uppercase {active
        ? 'text-zinc-900'
        : 'text-zinc-400'}"
    >
      {label}
    </span>
  </div>
{/snippet}

{#snippet outputPanel(result: LessonOutput | null, okClass: string, idle: string)}
  <div
    class="custom-scrollbar flex-1 overflow-y-auto font-mono text-sm whitespace-pre-wrap text-zinc-100"
  >
    {#if result?.error}
      <span class="text-red-400">{result.error}</span>
    {:else if result && (result.output || result.images.length > 0)}
      <div class="space-y-3">
        {#if result.output}<span class={okClass}>{result.output}</span>{/if}
        {#if result.images.length > 0}<PlotDisplay images={result.images} />{/if}
      </div>
    {:else}
      <span class="text-zinc-600 italic">{idle}</span>
    {/if}
  </div>
{/snippet}

<div class="mx-auto max-w-6xl space-y-8">
  {#if showSuccess}
    <div
      transition:fade={{ duration: 150 }}
      class="fixed inset-0 z-[100] flex items-center justify-center bg-black/50 p-4 backdrop-blur-sm"
    >
      <div
        in:scale={{ start: 0.95, duration: 200 }}
        role="dialog"
        aria-modal="true"
        aria-label="Kurikulum selesai"
        class="relative w-full max-w-lg space-y-8 overflow-hidden rounded-[2.5rem] bg-white p-12 text-center shadow-2xl"
      >
        <div class="absolute top-0 left-0 h-2 w-full bg-rose-700"></div>
        <div
          class="mx-auto mb-6 flex h-24 w-24 items-center justify-center rounded-3xl bg-rose-100 text-rose-700"
        >
          <Trophy size={48} />
        </div>
        <div class="space-y-4">
          <h3 class="text-4xl font-black tracking-tight">Luar Biasa!</h3>
          <p class="text-lg leading-relaxed text-zinc-500">
            Selamat! Anda telah menyelesaikan semua pelajaran yang tersedia di kurikulum ini. Anda
            telah mengambil langkah besar dalam menguasai pemrograman!
          </p>
        </div>
        <a
          href={resolve('/(app)/dashboard')}
          class="block w-full rounded-2xl bg-zinc-900 py-5 font-bold text-white shadow-xl shadow-zinc-900/10 transition-all hover:bg-zinc-800 active:scale-95"
        >
          Kembali ke Dashboard
        </a>
      </div>
    </div>
  {/if}

  {#if showResetConfirm}
    <div
      transition:fade={{ duration: 150 }}
      class="fixed inset-0 z-[100] flex items-center justify-center bg-zinc-950/60 p-4 backdrop-blur-sm"
    >
      <div
        in:scale={{ start: 0.95, duration: 200 }}
        role="alertdialog"
        aria-modal="true"
        aria-labelledby="reset-title"
        class="relative mt-8 flex w-full max-w-sm flex-col items-center rounded-[2.5rem] border border-zinc-100 bg-white p-8 text-center shadow-2xl"
      >
        <div
          class="absolute -top-10 left-1/2 flex h-20 w-20 -translate-x-1/2 items-center justify-center rounded-full border-4 border-white bg-rose-50 shadow-lg"
        >
          <TriangleAlert class="h-10 w-10 text-amber-500" />
        </div>
        <div class="mt-8 w-full space-y-3">
          <h3 id="reset-title" class="text-2xl font-black tracking-tight text-zinc-900">
            Reset Kode?
          </h3>
          <p class="text-sm leading-relaxed font-semibold text-zinc-500">
            Apakah kamu yakin ingin mereset kode ke pengaturan awal?
          </p>
          <p class="text-sm font-extrabold text-red-500">
            Semua perubahanmu di pelajaran ini akan hilang.
          </p>
        </div>
        <div class="mt-8 flex w-full gap-3">
          <button
            type="button"
            onclick={() => (showResetConfirm = false)}
            class="flex-1 rounded-2xl border border-zinc-200 bg-zinc-50 py-3.5 text-sm font-black text-zinc-600 transition-all hover:bg-zinc-100"
          >
            Batal
          </button>
          <button
            type="button"
            onclick={resetExercise}
            class="flex-1 rounded-2xl border-b-4 border-red-700 bg-red-600 py-3.5 text-sm font-black text-white shadow-md shadow-red-500/10 transition-all hover:bg-red-700 active:translate-y-[4px] active:border-b-0"
          >
            Ya, Reset!
          </button>
        </div>
      </div>
    </div>
  {/if}

  {#if showDrawer}
    <LessonDrawer
      {curriculum}
      currentLessonId={lessonId}
      currentLevelIdx={pos.levelIdx}
      {completed}
      {overrides}
      onselect={openLesson}
      onclose={() => (showDrawer = false)}
    />
  {/if}

  <div class="flex items-center justify-between">
    <div class="flex items-center gap-4">
      <a
        href={resolve('/(app)/courses')}
        class="flex h-10 w-10 items-center justify-center rounded-xl bg-rose-50 text-rose-700 transition-colors hover:bg-rose-100"
        title="Kembali ke Belajar"
        aria-label="Kembali ke Belajar"
      >
        <ArrowLeft size={20} />
      </a>
      <div>
        <div class="text-xs font-bold tracking-widest text-zinc-400 uppercase">
          Level {pos.levelIdx + 1} â€¢ {moduleTitle}
        </div>
        <h1 class="text-xl font-bold">{lesson.title}</h1>
      </div>
    </div>
    <button
      type="button"
      onclick={() => (showDrawer = true)}
      class="group flex cursor-pointer items-center gap-2 rounded-2xl border border-rose-200/60 bg-gradient-to-r from-rose-50 to-pink-50 px-4 py-2 text-xs font-black text-rose-700 shadow-sm transition-all hover:from-rose-100/70 hover:to-pink-100/70 hover:shadow-md hover:shadow-rose-500/5 active:scale-95"
      title="Buka Daftar Pelajaran"
    >
      <Menu
        size={15}
        class="text-rose-700 transition-transform duration-300 group-hover:rotate-90"
      />
      <span>Daftar Menu</span>
    </button>
  </div>

  <div class="flex items-center justify-center gap-4">
    {#each stepLabels as item, i (item.id)}
      {#if i > 0}<div class="h-px w-12 bg-zinc-200"></div>{/if}
      {@render stepIndicator(
        item.label,
        step === item.id,
        item.id === 'learn' ? step !== 'learn' : item.id === 'quiz' ? step === 'code' : false,
      )}
    {/each}
  </div>

  {#if step === 'learn'}
    <div
      in:fly={{ x: 20, duration: 200 }}
      class="grid h-auto grid-cols-1 items-stretch gap-8 lg:h-[calc(100dvh-240px)] lg:grid-cols-12"
    >
      <div
        class="custom-scrollbar flex h-full min-h-[300px] flex-col justify-between gap-6 overflow-y-auto pr-4 pb-4 lg:col-span-5"
      >
        <div class="space-y-6">
          <div
            class="flex w-fit items-center gap-2 rounded-full bg-rose-50 px-3 py-1 text-sm font-bold text-rose-700"
          >
            <BookOpen size={16} />
            Level {pos.levelIdx + 1} â€¢ Pelajaran {pos.lessonIdx + 1}
          </div>
          <h2 class="text-4xl leading-tight font-black tracking-tight text-zinc-900">
            {lesson.title}
          </h2>
          <ContentRenderer content={lesson.explanation} />
        </div>
        <div class="shrink-0 pt-4">
          <button
            type="button"
            onclick={() => goToStep('quiz')}
            class="flex w-full items-center justify-center gap-2 rounded-2xl bg-zinc-900 py-4 font-bold text-white shadow-lg shadow-zinc-900/10 transition-all hover:bg-zinc-800 active:scale-95"
          >
            Ikuti Kuis
            <ChevronRight size={20} />
          </button>
        </div>
      </div>

      <div class="flex h-full min-h-[500px] flex-col gap-4 lg:col-span-7">
        <div class="min-h-[300px] flex-1">
          <LessonEditor
            bind:value={sandboxCode}
            {language}
            running={running || runnerLoading}
            onrun={runSandbox}
          />
        </div>
        {#if runnerLoading}
          <p class="text-xs font-bold text-zinc-400" role="status">
            {language === 'c' ? messages.loadingC : messages.loadingPython}
          </p>
        {/if}
        <div class="flex shrink-0 items-center gap-2">
          <label
            for="sandbox-stdin"
            class="text-[10px] font-black tracking-widest whitespace-nowrap text-zinc-400 uppercase"
          >
            Input (stdin)
          </label>
          <textarea
            id="sandbox-stdin"
            rows={1}
            bind:value={sandboxInput}
            placeholder={language === 'c'
              ? 'Masukkan input untuk scanf...\n(Gunakan Enter untuk baris baru jika input banyak, Ctrl+Enter untuk menjalankan)'
              : 'Masukkan input untuk input()...\n(Gunakan Enter untuk baris baru, Ctrl+Enter untuk menjalankan)'}
            onkeydown={(e) => {
              if (e.key === 'Enter' && (e.ctrlKey || e.metaKey)) {
                e.preventDefault();
                void runSandbox();
              }
            }}
            class="max-h-[120px] min-h-[38px] flex-1 resize-y rounded-xl border border-zinc-200 bg-zinc-50 px-3 py-1.5 font-mono text-sm text-zinc-700 transition-all placeholder:text-zinc-300 focus:border-rose-300 focus:ring-2 focus:ring-rose-200 focus:outline-none"
          ></textarea>
        </div>
        <div
          class="flex h-44 shrink-0 flex-col rounded-2xl border border-zinc-800 bg-zinc-950 p-4 shadow-inner"
        >
          <div
            class="mb-2 flex items-center justify-between text-xs font-bold tracking-widest text-zinc-500 uppercase"
          >
            <div class="flex items-center gap-1.5">
              <span class="h-2 w-2 rounded-full bg-[#ff5f56]"></span>
              <span class="h-2 w-2 rounded-full bg-[#ffbd2e]"></span>
              <span class="h-2 w-2 rounded-full bg-[#27c93f]"></span>
              <span class="ml-1 text-[10px] font-bold tracking-widest text-zinc-500 uppercase">
                Output Sandbox
              </span>
            </div>
            <button
              type="button"
              onclick={() => {
                sandboxCode = lesson.codeExample;
                sandbox = null;
              }}
              class="rounded bg-zinc-800 px-2 py-1 font-sans text-[10px] font-bold tracking-wider text-zinc-400 uppercase transition-colors hover:bg-zinc-700 hover:text-white"
            >
              Reset Kode
            </button>
          </div>
          {@render outputPanel(
            sandbox,
            'text-emerald-400',
            "// Klik 'Jalankan' atau Ctrl+Enter untuk melihat output",
          )}
        </div>
      </div>
    </div>
  {:else if step === 'quiz'}
    <div in:fly={{ x: 20, duration: 200 }} class="mx-auto max-w-2xl">
      {#if lesson.quiz}
        <Quiz
          question={lesson.quiz.question}
          options={lesson.quiz.options}
          correctAnswer={lesson.quiz.correctAnswer}
          validate={(selected) => quizMutation.mutateAsync(selected)}
          oncomplete={onQuizComplete}
        />
      {:else}
        <div class="space-y-4 rounded-3xl border border-zinc-200 bg-white p-8 text-center">
          <p class="font-medium text-zinc-600">Tidak ada kuis untuk pelajaran ini.</p>
          <button
            type="button"
            onclick={() => goToStep('code')}
            class="rounded-2xl bg-zinc-900 px-6 py-3 font-bold text-white transition-all hover:bg-zinc-800"
          >
            Lanjut ke Latihan
          </button>
        </div>
      {/if}
    </div>
  {:else}
    <div
      in:fly={{ x: 20, duration: 200 }}
      class="grid grid-cols-1 gap-8 lg:grid-cols-2 {lesson.isFlowchart
        ? 'h-auto'
        : 'h-auto lg:h-[calc(100dvh-280px)]'}"
    >
      <div class="custom-scrollbar flex h-auto flex-col gap-6 pr-4 lg:h-full lg:overflow-y-auto">
        <h2 class="text-3xl font-bold tracking-tight">
          Latihan {lesson.isFlowchart ? 'Flowchart' : 'Pemrograman'}
        </h2>
        <div class="rounded-2xl border border-zinc-200 bg-white p-6 shadow-sm">
          <h3 class="mb-4 text-lg font-bold">Tugas Anda</h3>
          <div class="mb-4 space-y-4">
            {#each lesson.testCases as tc, idx (idx)}
              <div class="flex items-start gap-3 text-zinc-600">
                <div class="mt-1.5 h-1.5 w-1.5 shrink-0 rounded-full bg-rose-700"></div>
                <div class="min-w-0 flex-1">
                  {#if looksLikeHtml(tc.description)}
                    <ContentRenderer
                      content={tc.description}
                      class="prose-sm max-w-none text-zinc-600 prose-ol:my-0 prose-p:my-0 prose-p:text-zinc-600 prose-ul:my-0"
                    />
                  {:else}
                    <p class="my-0 text-sm text-zinc-600">{tc.description}</p>
                  {/if}
                </div>
              </div>
            {/each}
          </div>
          <button
            type="button"
            onclick={() => (showHint = !showHint)}
            class="flex items-center gap-1 text-sm font-medium text-zinc-500 transition-colors hover:text-zinc-900"
          >
            <Lightbulb size={14} />
            {showHint ? 'Sembunyikan Petunjuk' : 'Butuh petunjuk?'}
          </button>
          {#if showHint}
            <div transition:slide={{ duration: 200 }} class="space-y-3 overflow-hidden">
              <div
                class="mt-4 rounded-xl border border-amber-100 bg-amber-50 p-4 text-sm text-amber-800 italic"
              >
                {#if looksLikeHtml(lesson.hint)}
                  <ContentRenderer
                    content={lesson.hint}
                    class="text-sm prose-p:text-amber-800 prose-p:italic"
                  />
                {:else}
                  {lesson.hint || 'Tidak ada petunjuk.'}
                {/if}
              </div>
            </div>
          {/if}

          {#if lesson.testCases.length > 0}
            <div class="mt-4 overflow-hidden rounded-xl border border-zinc-700 shadow-lg">
              <div
                class="flex items-center justify-between border-b border-zinc-700 bg-zinc-800 px-4 py-2"
              >
                <div class="flex items-center gap-2">
                  <div class="flex gap-1.5">
                    <span class="inline-block h-2.5 w-2.5 rounded-full bg-rose-500"></span>
                    <span class="inline-block h-2.5 w-2.5 rounded-full bg-amber-500"></span>
                    <span class="inline-block h-2.5 w-2.5 rounded-full bg-emerald-500"></span>
                  </div>
                  <span class="ml-2 text-[10px] font-bold tracking-widest text-zinc-400 uppercase">
                    Output
                  </span>
                </div>
              </div>
              <div class="min-h-[60px] space-y-3 bg-zinc-950 p-4 font-mono text-xs text-zinc-100">
                {#each lesson.testCases as tc, idx (idx)}
                  <div class="space-y-1">
                    {#if lesson.testCases.length > 1}
                      <div class="text-[10px] tracking-wider text-zinc-500 uppercase">
                        Kasus Uji #{idx + 1}
                      </div>
                    {/if}
                    {#if tc.input}
                      <div class="text-zinc-400">
                        Input: <span class="text-emerald-400">{tc.input}</span>
                      </div>
                    {/if}
                    <pre
                      class="font-mono whitespace-pre-wrap text-zinc-100 select-all">{tc.expectedOutput}</pre>
                  </div>
                {/each}
              </div>
            </div>
          {/if}
        </div>
      </div>

      <div class="flex h-auto flex-col gap-4 lg:h-full">
        {#if lesson.isFlowchart}
          <div class="min-h-[620px] rounded-2xl border border-zinc-200 bg-white p-4 shadow-sm">
            <FlowchartPuzzle
              code={lesson.initialCode || lesson.codeExample}
              solutionJson={lesson.solution || '[]'}
              distractors={lesson.flowchartDistractors}
              disabled={done}
              onresult={onFlowchartResult}
            />
          </div>
        {:else}
          <div class="h-[400px] lg:h-full lg:flex-1">
            <LessonEditor
              bind:value={code}
              {language}
              running={running || runnerLoading}
              onrun={runExercise}
              onreset={() => (showResetConfirm = true)}
              oninput={() => (isCorrect = null)}
            />
          </div>
        {/if}

        {#if lesson.testCases.some((tc) => tc.input)}
          <div class="flex shrink-0 items-center gap-2">
            <span
              class="text-[10px] font-black tracking-widest whitespace-nowrap text-zinc-400 uppercase"
            >
              Input Uji
            </span>
            <div
              class="flex-1 rounded-xl border border-zinc-200 bg-zinc-50 px-3 py-2 font-mono text-sm text-zinc-600"
            >
              {#each lesson.testCases.filter((tc) => tc.input) as tc, i (i)}
                <span class="inline-block">
                  {#if i > 0}<span class="mx-1 text-zinc-300">â€¢</span>{/if}
                  <span class="rounded bg-zinc-100 px-2 py-0.5 font-mono text-xs text-zinc-700">
                    {tc.input}
                  </span>
                </span>
              {/each}
            </div>
          </div>
        {/if}

        {#if !lesson.isFlowchart}
          {#if runnerLoading}
            <p class="text-xs font-bold text-zinc-400" role="status">
              {language === 'c' ? messages.loadingC : messages.loadingPython}
            </p>
          {/if}
          <div
            class="flex h-40 flex-col rounded-2xl border border-zinc-800 bg-zinc-950 p-4 font-mono text-sm shadow-inner"
          >
            <div
              class="mb-2 flex items-center justify-between text-xs font-bold tracking-widest text-zinc-500 uppercase"
            >
              <div class="flex items-center gap-1.5">
                <span class="h-2 w-2 rounded-full bg-[#ff5f56]"></span>
                <span class="h-2 w-2 rounded-full bg-[#ffbd2e]"></span>
                <span class="h-2 w-2 rounded-full bg-[#27c93f]"></span>
                <span class="ml-1 text-[10px] font-bold tracking-widest text-zinc-500 uppercase">
                  Output Terminal
                </span>
              </div>
              {#if isCorrect !== null}
                <span class={isCorrect ? 'text-rose-400' : 'text-red-400'}>
                  {isCorrect ? 'Berhasil!' : 'Coba lagi'}
                </span>
              {/if}
            </div>
            {@render outputPanel(exercise, 'text-zinc-100', '')}
          </div>
        {/if}

        {#if (!isCorrect || !isQuizPassed) && !completing}
          <p class="text-center text-xs leading-relaxed font-medium text-zinc-500">
            {!isQuizPassed
              ? 'Kembali ke tahap Kuis dan jawab dengan benar untuk membuka tombol ini.'
              : 'Tekan Jalankan sampai muncul Berhasil! untuk membuka tombol ini.'}
          </p>
        {/if}
        <button
          type="button"
          disabled={!isCorrect || !isQuizPassed || completing}
          onclick={finishLesson}
          class="flex w-full items-center justify-center gap-2 rounded-2xl bg-rose-700 py-4 font-bold text-white shadow-lg shadow-rose-700/20 transition-all hover:bg-rose-600 active:scale-95 disabled:opacity-50"
        >
          {#if completing}
            <Loader size={20} class="animate-spin" />
            Memproses...
          {:else}
            Selesaikan Pelajaran
          {/if}
        </button>
      </div>
    </div>
  {/if}

  <div class="mt-2 flex items-center justify-between border-t border-zinc-100 pt-4 pb-8">
    <button
      type="button"
      onclick={handlePrev}
      disabled={!canGoPrev}
      class="flex items-center gap-2 rounded-xl px-5 py-3 text-sm font-bold transition-all {canGoPrev
        ? 'border border-zinc-200 bg-zinc-50 text-zinc-600 shadow-sm hover:border-zinc-300 hover:bg-zinc-100 active:scale-95'
        : 'cursor-not-allowed text-zinc-300'}"
    >
      <ChevronLeft size={18} />
      Sebelumnya
    </button>
    <button
      type="button"
      onclick={handleNext}
      disabled={!canGoNext}
      class="flex items-center gap-2 rounded-xl px-5 py-3 text-sm font-bold transition-all {canGoNext
        ? 'border border-zinc-200 bg-zinc-50 text-zinc-600 shadow-sm hover:border-zinc-300 hover:bg-zinc-100 active:scale-95'
        : 'cursor-not-allowed text-zinc-300'}"
    >
      Selanjutnya
      <ChevronRight size={18} />
    </button>
  </div>
</div>
