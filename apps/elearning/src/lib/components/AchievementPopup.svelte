<script lang="ts">
  import type { Achievement } from '@algohub/api-client/elearning';
  import Sparkles from '@lucide/svelte/icons/sparkles';
  import X from '@lucide/svelte/icons/x';
  import confetti from 'canvas-confetti';
  import { onMount } from 'svelte';
  import { fade, scale } from 'svelte/transition';
  import { themeFor } from '#lib/achievement-theme.ts';
  import { playAchievementSound } from '#lib/soundEffects.ts';
  import { appStore } from '#lib/stores/app.svelte.ts';

  let { achievement, onclose }: { achievement: Achievement; onclose: () => void } = $props();

  const theme = $derived(themeFor(achievement.icon));
  const Icon = $derived(theme.icon);

  onMount(() => {
    if (appStore.soundEnabled) playAchievementSound();
    confetti({
      particleCount: 100,
      spread: 120,
      startVelocity: 50,
      origin: { x: 0.5, y: 0.5 },
      colors: ['#fbbf24', '#f472b6', '#60a5fa', '#34d399', '#a78bfa', '#fb923c'],
      ticks: 160,
      scalar: 1.3,
      zIndex: 300,
    });
    const sides = setTimeout(() => {
      confetti({
        particleCount: 60,
        angle: 60,
        spread: 70,
        origin: { x: 0, y: 0.55 },
        zIndex: 300,
      });
      confetti({
        particleCount: 60,
        angle: 120,
        spread: 70,
        origin: { x: 1, y: 0.55 },
        zIndex: 300,
      });
    }, 300);
    // Hujan confetti berhenti sendiri setelah 3,5 detik.
    const end = Date.now() + 3500;
    const rain = setInterval(() => {
      if (Date.now() > end) {
        clearInterval(rain);
        return;
      }
      confetti({
        particleCount: 10,
        angle: 80,
        spread: 40,
        origin: { x: Math.random() * 0.4, y: -0.05 },
        gravity: 0.7,
        zIndex: 300,
      });
      confetti({
        particleCount: 10,
        angle: 100,
        spread: 40,
        origin: { x: 0.6 + Math.random() * 0.4, y: -0.05 },
        gravity: 0.7,
        zIndex: 300,
      });
    }, 280);
    return () => {
      clearTimeout(sides);
      clearInterval(rain);
      confetti.reset();
    };
  });
</script>

<div
  class="fixed inset-0 z-[200] flex items-center justify-center p-4"
  role="dialog"
  aria-modal="true"
  aria-label="Pencapaian terbuka"
>
  <button
    transition:fade={{ duration: 200 }}
    onclick={onclose}
    class="absolute inset-0 cursor-default bg-black/85 backdrop-blur-xl"
    aria-label="Tutup"
  ></button>
  <div
    transition:scale={{ start: 0.3, duration: 300 }}
    class="relative w-full max-w-sm overflow-hidden rounded-[2.5rem] bg-gradient-to-b from-zinc-900 to-zinc-950 p-8 text-center"
    style:box-shadow="0 0 0 1.5px rgba(255,255,255,0.08), 0 32px 80px rgba(0,0,0,0.7), 0 0 60px {theme.glow}"
  >
    <button
      onclick={onclose}
      class="absolute top-4 right-4 z-10 cursor-pointer rounded-xl p-2 text-zinc-500 transition-colors hover:bg-white/10 hover:text-white"
      aria-label="Tutup"
    >
      <X size={18} />
    </button>
    <div class="relative flex flex-col items-center">
      <div class="relative mt-4 mb-6">
        <div
          class="absolute -inset-4 animate-[spin_10s_linear_infinite] rounded-full border-2 border-dashed border-white/30"
        ></div>
        <div
          class="relative flex h-28 w-28 items-center justify-center rounded-[1.75rem] border-4 border-black bg-gradient-to-br {theme.bg}"
          style:box-shadow="0 8px 40px rgba(0,0,0,0.5), 0 0 30px {theme.glow}"
        >
          <Icon size={48} class="text-white drop-shadow-[0_2px_8px_rgba(0,0,0,0.4)]" />
          <div
            class="absolute -top-3 -right-3 flex h-10 w-10 items-center justify-center rounded-full border-2 border-black bg-gradient-to-br {theme.bg} text-base shadow-lg"
          >
            {theme.emoji}
          </div>
        </div>
      </div>
      <div
        class="mb-3 inline-flex items-center gap-2 rounded-full border border-white/15 bg-white/10 px-4 py-1.5 text-[10px] font-black tracking-[0.3em] text-white uppercase"
      >
        <Sparkles size={10} class="animate-pulse text-yellow-300" />
        Pencapaian Terbuka!
        <Sparkles size={10} class="animate-pulse text-yellow-300" />
      </div>
      <h2
        class="mb-2 bg-gradient-to-r {theme.text} bg-clip-text px-2 text-2xl leading-tight font-black text-transparent"
      >
        {achievement.title}
      </h2>
      <p class="mb-7 px-4 text-sm leading-relaxed font-medium text-zinc-400">
        {achievement.description}
      </p>
      <button
        onclick={onclose}
        class="w-full cursor-pointer rounded-2xl border-b-4 border-black/30 bg-gradient-to-r {theme.bg} py-4 text-sm font-black tracking-widest text-white uppercase shadow-2xl transition-all active:translate-y-1 active:border-b-0"
      >
        Luar Biasa! ✨
      </button>
    </div>
  </div>
</div>
