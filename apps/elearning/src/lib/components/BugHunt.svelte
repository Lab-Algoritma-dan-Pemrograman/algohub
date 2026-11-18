<script lang="ts">
  import type { GameQuestion } from '@algohub/api-client/elearning';
  import ArrowRight from '@lucide/svelte/icons/arrow-right';
  import Brain from '@lucide/svelte/icons/brain';
  import Bug from '@lucide/svelte/icons/bug';
  import CircleCheck from '@lucide/svelte/icons/circle-check';
  import CircleX from '@lucide/svelte/icons/circle-x';
  import LoaderCircle from '@lucide/svelte/icons/loader-circle';
  import Terminal from '@lucide/svelte/icons/terminal';
  import Timer from '@lucide/svelte/icons/timer';
  import Trophy from '@lucide/svelte/icons/trophy';
  import X from '@lucide/svelte/icons/x';
  import { MutationObserver } from '@tanstack/svelte-query';
  import { onMount } from 'svelte';
  import { api } from '#lib/api.ts';
  import { BUG_HUNT_SECONDS, bugHuntXp, isWeeklyLimitError } from '#lib/gamification.ts';
  import { announceAchievements } from '#lib/queries/achievements.ts';
  import {
    gameQuestionsQuery,
    gameSettingsQuery,
    myPlaysQuery,
    saveGameHistoryOptions,
  } from '#lib/queries/game.ts';
  import { queryClient } from '#lib/query.ts';

  type Status = 'loading' | 'blocked' | 'empty' | 'playing' | 'result' | 'finished';

  let { language, onclose }: { language: 'c' | 'python'; onclose: () => void } = $props();

  let status = $state<Status>('loading');
  let questions = $state<GameQuestion[]>([]);
  let index = $state(0);
  let timeLeft = $state(BUG_HUNT_SECONDS);
  let score = $state(0);
  let totalXp = $state(0);
  let selectedLine = $state<number | null>(null);
  let lastResult = $state<{ correct: boolean; xp: number } | null>(null);
  let saving = $state(false);
  let limitError = $state(false);
  let saveFailed = $state(false);

  const question = $derived(questions[index]);
  const lines = $derived((question?.code ?? '').replace(/\\n/g, '\n').split('\n'));

  let timer: ReturnType<typeof setInterval> | undefined;
  const stopTimer = () => clearInterval(timer);

  function startTimer() {
    stopTimer();
    timeLeft = BUG_HUNT_SECONDS;
    timer = setInterval(() => {
      timeLeft = Math.max(0, timeLeft - 0.1);
      if (timeLeft <= 0) answer(-1);
    }, 100);
  }

  onMount(() => {
    void load();
    return stopTimer;
  });

  // Kuota dicek ulang ke server tepat sebelum main (staleTime 0), bukan dari cache dasbor.
  async function load() {
    try {
      const [settings, plays] = await Promise.all([
        queryClient.fetchQuery(gameSettingsQuery()),
        queryClient.fetchQuery({ ...myPlaysQuery(), staleTime: 0 }),
      ]);
      if (!plays.allowed) {
        status = 'blocked';
        return;
      }
      questions = await queryClient.fetchQuery(
        gameQuestionsQuery(language, settings.bugHuntQuestionCount || 5),
      );
    } catch {
      questions = [];
    }
    if (questions.length === 0) {
      status = 'empty';
      return;
    }
    status = 'playing';
    startTimer();
  }

  function answer(line: number) {
    if (status !== 'playing') return;
    stopTimer();
    const correct = line === question.bugLine;
    const xp = bugHuntXp(correct, timeLeft);
    selectedLine = line;
    lastResult = { correct, xp };
    totalXp += xp;
    if (correct) score += 1;
    status = 'result';
  }

  function next() {
    if (index < questions.length - 1) {
      index += 1;
      selectedLine = null;
      lastResult = null;
      status = 'playing';
      startTimer();
    } else {
      void finish();
    }
  }

  async function finish() {
    saving = true;
    status = 'finished';
    try {
      const res = await new MutationObserver(queryClient, saveGameHistoryOptions()).mutate({
        gameType: 'bug_hunt',
        xpEarned: totalXp,
      });
      const [unlocked, byXp] = await Promise.all([
        api.achievements.check({
          xp: res.xp,
          gamesPlayed: 1,
          perfectGames: score === questions.length ? 1 : 0,
        }),
        api.achievements.checkXp(res.xp),
      ]);
      announceAchievements([...unlocked, ...(byXp.achievement ? [byXp.achievement] : [])]);
    } catch (error) {
      // Kuota habis: XP tidak disimpan, jadi total XP tidak boleh tampil sebagai hasil.
      if (isWeeklyLimitError(error)) {
        limitError = true;
        totalXp = 0;
      } else {
        saveFailed = true;
      }
    } finally {
      saving = false;
    }
  }
</script>

{#snippet message(icon: 'blocked' | 'empty', title: string, text: string)}
  <div
    class="fixed inset-0 z-[100] flex items-center justify-center bg-zinc-900/90 backdrop-blur-sm"
  >
    <div class="max-w-sm space-y-6 rounded-3xl border border-zinc-800 bg-zinc-900 p-8 text-center">
      <div
        class="mx-auto flex h-16 w-16 items-center justify-center rounded-2xl {icon === 'blocked'
          ? 'bg-rose-900/30 text-rose-500'
          : 'bg-zinc-800 text-zinc-400'}"
      >
        <CircleX size={32} />
      </div>
      <div>
        <h3 class="text-xl font-bold text-white">{title}</h3>
        <p class="mt-2 text-sm text-zinc-500">{text}</p>
      </div>
      <button
        onclick={onclose}
        class="w-full cursor-pointer rounded-xl bg-zinc-800 py-3 font-bold text-white transition-all hover:bg-zinc-700"
        >Tutup</button
      >
    </div>
  </div>
{/snippet}

{#if status === 'loading'}
  <div
    class="fixed inset-0 z-[100] flex items-center justify-center bg-zinc-900/90 backdrop-blur-sm"
  >
    <div class="space-y-4 text-center">
      <div
        class="mx-auto h-16 w-16 animate-spin rounded-full border-4 border-rose-700 border-t-transparent"
      ></div>
      <p class="animate-pulse font-bold text-white">Menyiapkan Perburuan Bug...</p>
    </div>
  </div>
{:else if status === 'blocked'}
  {@render message(
    'blocked',
    'Limit Tercapai!',
    'Kamu sudah mencapai batas main Bug Hunt minggu ini. Kembali lagi minggu depan! 🎯',
  )}
{:else if status === 'empty'}
  {@render message(
    'empty',
    'Tidak Ada Soal',
    'Maaf, saat ini tidak ada soal yang tersedia untuk bahasa ini. Silakan hubungi admin.',
  )}
{:else}
  <div class="fixed inset-0 z-[100] flex flex-col overflow-hidden bg-zinc-950">
    <div
      class="flex h-16 items-center justify-between border-b border-zinc-800 bg-zinc-900/50 px-6"
    >
      <div class="flex items-center gap-3">
        <div
          class="flex h-10 w-10 items-center justify-center rounded-xl bg-rose-700 text-white shadow-lg shadow-rose-700/20"
        >
          <Bug size={24} />
        </div>
        <div>
          <h2 class="leading-none font-bold text-white">
            Bug Hunt: {language === 'c' ? 'Bahasa C' : 'Python'}
          </h2>
          <p class="mt-1 text-[10px] font-bold tracking-widest text-zinc-500 uppercase">
            Sesi {index + 1} / {questions.length}
          </p>
        </div>
      </div>
      <div class="flex items-center gap-6">
        <div class="flex items-center gap-2">
          <Timer
            class="transition-colors {timeLeft < 5
              ? 'animate-pulse text-rose-500'
              : 'text-zinc-400'}"
            size={18}
          />
          <div class="h-2 w-32 overflow-hidden rounded-full bg-zinc-800">
            <div
              class="h-full transition-[width] duration-100 {timeLeft < 5
                ? 'bg-rose-500'
                : 'bg-emerald-500'}"
              style:width="{(timeLeft / BUG_HUNT_SECONDS) * 100}%"
            ></div>
          </div>
        </div>
        <div
          class="flex items-center gap-2 rounded-lg border border-zinc-700 bg-zinc-800 px-3 py-1.5"
        >
          <Trophy class="text-amber-500" size={16} />
          <span class="font-mono text-sm font-bold text-white">+{totalXp} XP</span>
        </div>
        <button
          onclick={onclose}
          class="cursor-pointer text-zinc-500 transition-colors hover:text-white"
          aria-label="Tutup permainan"
        >
          <X size={24} />
        </button>
      </div>
    </div>

    <div class="flex flex-1 flex-col gap-6 overflow-hidden p-6 lg:flex-row">
      <div class="flex flex-1 flex-col gap-4 overflow-hidden">
        <div
          class="flex h-full flex-col overflow-hidden rounded-2xl border border-zinc-800 bg-zinc-900 shadow-2xl"
        >
          <div
            class="flex items-center justify-between border-b border-zinc-800 bg-zinc-800/50 px-4 py-3"
          >
            <div
              class="flex items-center gap-2 text-xs font-bold tracking-widest text-zinc-400 uppercase"
            >
              <Terminal size={14} />
              Temukan baris yang mengandung bug!
            </div>
            <div class="font-mono text-xs text-zinc-500">{question?.title}</div>
          </div>
          <div
            class="custom-scrollbar relative flex-1 overflow-y-auto font-mono text-sm leading-relaxed"
          >
            {#each lines as line, i (i)}
              <button
                type="button"
                onclick={() => answer(i)}
                disabled={status !== 'playing'}
                class="group relative flex w-full items-start px-4 py-1.5 text-left transition-colors {status ===
                'playing'
                  ? 'cursor-pointer hover:bg-zinc-800/50'
                  : 'cursor-default'} {selectedLine === i && !lastResult?.correct
                  ? 'bg-rose-500/20'
                  : ''} {status === 'result' && i === question.bugLine ? 'bg-emerald-500/20' : ''}"
              >
                <span class="w-8 pr-4 text-right text-zinc-600 select-none">{i + 1}</span>
                <span
                  class="flex-1 whitespace-pre {status === 'playing'
                    ? 'text-zinc-100'
                    : 'text-zinc-400'} {status === 'result' && i === question.bugLine
                    ? 'font-bold text-emerald-400'
                    : ''}">{line || ' '}</span
                >
                {#if selectedLine === i && !lastResult?.correct}
                  <CircleX class="absolute right-4 text-rose-500" size={18} />
                {/if}
                {#if status === 'result' && i === question.bugLine}
                  <CircleCheck class="absolute right-4 text-emerald-500" size={18} />
                {/if}
              </button>
            {/each}
          </div>
        </div>
      </div>

      <div class="flex w-full flex-col gap-4 lg:w-80">
        {#if status === 'playing'}
          <div class="space-y-4 rounded-2xl border border-zinc-800 bg-zinc-900 p-6">
            <div
              class="flex h-12 w-12 items-center justify-center rounded-xl bg-blue-500/10 text-blue-500"
            >
              <Brain size={24} />
            </div>
            <h3 class="text-lg font-bold text-white">Misi Kamu</h3>
            <p class="text-sm leading-relaxed text-zinc-400">
              Periksa kode di sebelah kiri dengan cermat. Salah satu baris memiliki kesalahan
              sintaks atau logika yang fatal. Klik baris tersebut sesegera mungkin!
            </p>
            <div class="border-t border-zinc-800 pt-4">
              <p class="text-xs text-zinc-500 italic">
                "Gunakan insting programmer-mu untuk menemukan bug-nya!"
              </p>
            </div>
          </div>
        {:else if status === 'result'}
          <div
            class="space-y-6 rounded-2xl p-8 text-center shadow-2xl {lastResult?.correct
              ? 'bg-emerald-50'
              : 'bg-rose-50'}"
          >
            <div
              class="mx-auto mb-4 flex h-20 w-20 items-center justify-center rounded-full {lastResult?.correct
                ? 'bg-emerald-100 text-emerald-700'
                : 'bg-rose-100 text-rose-700'}"
            >
              {#if lastResult?.correct}
                <CircleCheck size={48} />
              {:else}
                <CircleX size={48} />
              {/if}
            </div>
            <div>
              <h3
                class="mb-2 text-2xl font-black {lastResult?.correct
                  ? 'text-emerald-900'
                  : 'text-rose-900'}"
              >
                {lastResult?.correct ? 'BUG DITEMUKAN!' : 'YAH, SALAH KLIK!'}
              </h3>
              <p class="mb-6 font-medium text-zinc-600">
                {lastResult?.correct
                  ? `Luar biasa! Kamu mendapatkan ${lastResult.xp} XP.`
                  : 'Jangan menyerah, perhatikan baik-baik baris hijau yang benar.'}
              </p>
              <div class="mb-6 rounded-xl border border-zinc-200 bg-white/50 p-4 text-left">
                <p class="mb-1 text-xs font-bold tracking-widest text-zinc-400 uppercase">
                  Penjelasan:
                </p>
                <p class="text-sm leading-relaxed font-medium text-zinc-700">
                  {question?.explanation}
                </p>
              </div>
            </div>
            <button
              onclick={next}
              class="flex w-full cursor-pointer items-center justify-center gap-2 rounded-xl py-4 font-bold text-white shadow-lg transition-all active:scale-95 {lastResult?.correct
                ? 'bg-emerald-600 shadow-emerald-600/20 hover:bg-emerald-700'
                : 'bg-rose-600 shadow-rose-600/20 hover:bg-rose-700'}"
            >
              {index < questions.length - 1 ? 'Soal Berikutnya' : 'Lihat Hasil Akhir'}
              <ArrowRight size={18} />
            </button>
          </div>
        {:else}
          <div class="space-y-6 rounded-3xl border border-zinc-800 bg-zinc-900 p-8 text-center">
            <div
              class="mx-auto flex h-24 w-24 items-center justify-center rounded-full border-4 border-amber-500/20 bg-amber-500/10 text-amber-500"
            >
              <Trophy size={48} />
            </div>
            <div>
              <h2 class="mb-2 text-3xl font-black text-white">MISI SELESAI!</h2>
              <p class="text-zinc-400">
                Kamu telah menyelesaikan tantangan perburuan bug hari ini.
              </p>
            </div>
            <div class="grid grid-cols-2 gap-4">
              <div class="rounded-2xl bg-zinc-800 p-4">
                <div class="mb-1 text-[10px] font-bold tracking-widest text-zinc-500 uppercase">
                  Skor
                </div>
                <div class="text-2xl font-black text-white">{score} / {questions.length}</div>
              </div>
              <div class="rounded-2xl bg-rose-700 p-4">
                <div class="mb-1 text-[10px] font-bold tracking-widest text-rose-100 uppercase">
                  Total XP
                </div>
                <div class="text-2xl font-black text-white">{limitError ? '—' : `+${totalXp}`}</div>
              </div>
            </div>
            {#if saveFailed}
              <p class="text-xs font-bold text-rose-400">
                Hasil belum tersimpan karena gangguan koneksi.
              </p>
            {/if}
            <button
              onclick={onclose}
              disabled={saving}
              class="flex w-full items-center justify-center gap-2 rounded-2xl py-4 font-bold transition-all active:scale-95 {saving
                ? 'cursor-not-allowed bg-zinc-800 text-zinc-500'
                : 'cursor-pointer bg-white text-black hover:bg-zinc-200'}"
            >
              {#if saving}
                <LoaderCircle size={20} class="animate-spin" />
                Menyimpan...
              {:else if limitError}
                ⚠️ Limit tercapai — XP tidak disimpan
              {:else}
                Kembali ke Dashboard
              {/if}
            </button>
          </div>
        {/if}
      </div>
    </div>
  </div>
{/if}
