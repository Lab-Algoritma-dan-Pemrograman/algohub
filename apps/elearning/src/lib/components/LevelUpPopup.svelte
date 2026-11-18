<script lang="ts">
  import ArrowUp from '@lucide/svelte/icons/arrow-up';
  import Crown from '@lucide/svelte/icons/crown';
  import Sparkles from '@lucide/svelte/icons/sparkles';
  import Star from '@lucide/svelte/icons/star';
  import Trophy from '@lucide/svelte/icons/trophy';
  import X from '@lucide/svelte/icons/x';
  import Zap from '@lucide/svelte/icons/zap';
  import confetti from 'canvas-confetti';
  import { onMount } from 'svelte';
  import { fade, scale } from 'svelte/transition';
  import { playLevelUpSound } from '#lib/soundEffects.ts';
  import { appStore } from '#lib/stores/app.svelte.ts';

  const BACKDROP =
    'background: radial-gradient(ellipse at 50% 45%, rgba(120,53,15,0.35) 0%, rgba(0,0,0,0.9) 100%)';
  const CARD =
    'background: linear-gradient(160deg, #1c1008 0%, #0f0a04 40%, #0a0a12 100%); box-shadow: 0 0 0 1.5px rgba(251,191,36,0.25), 0 40px 100px rgba(0,0,0,0.7), 0 0 80px rgba(251,191,36,0.12)';
  const GLOW =
    'background: radial-gradient(ellipse at 50% 0%, rgba(251,191,36,0.25) 0%, transparent 70%)';
  const MEDAL =
    'background: linear-gradient(135deg, #f59e0b 0%, #fbbf24 35%, #d97706 65%, #fbbf24 100%); box-shadow: 0 0 0 4px rgba(251,191,36,0.25), 0 12px 50px rgba(245,158,11,0.45)';

  let { level, onclose }: { level: number; onclose: () => void } = $props();

  onMount(() => {
    if (appStore.soundEnabled) playLevelUpSound();
    confetti({
      particleCount: 160,
      spread: 360,
      startVelocity: 55,
      origin: { x: 0.5, y: 0.45 },
      colors: ['#fbbf24', '#f59e0b', '#e11d48', '#3b82f6', '#10b981', '#8b5cf6', '#f472b6', '#fff'],
      scalar: 1.4,
      ticks: 200,
      zIndex: 310,
    });
    const cannons = setTimeout(() => {
      confetti({
        particleCount: 90,
        angle: 55,
        spread: 65,
        startVelocity: 65,
        origin: { x: 0, y: 0.5 },
        zIndex: 310,
      });
      confetti({
        particleCount: 90,
        angle: 125,
        spread: 65,
        startVelocity: 65,
        origin: { x: 1, y: 0.5 },
        zIndex: 310,
      });
    }, 200);
    const stars = setTimeout(() => {
      confetti({
        particleCount: 60,
        spread: 100,
        origin: { x: 0.5, y: 0.3 },
        shapes: ['star'],
        colors: ['#fbbf24', '#fff', '#fde68a'],
        scalar: 1.5,
        ticks: 150,
        zIndex: 310,
      });
    }, 800);
    // Aliran confetti bergantian kiri dan kanan selama 5 detik.
    const end = Date.now() + 5000;
    let left = false;
    const stream = setInterval(() => {
      if (Date.now() > end) {
        clearInterval(stream);
        return;
      }
      left = !left;
      confetti({
        particleCount: 8,
        angle: left ? 70 : 110,
        spread: 50,
        origin: { x: left ? 0.1 : 0.9, y: 0.35 },
        gravity: 0.8,
        ticks: 120,
        zIndex: 310,
      });
    }, 150);
    const shower = setTimeout(() => {
      confetti({
        particleCount: 120,
        spread: 120,
        startVelocity: 40,
        origin: { x: 0.5, y: 0.2 },
        shapes: ['star', 'circle'],
        colors: ['#fbbf24', '#f59e0b', '#fff'],
        scalar: 1.2,
        ticks: 160,
        zIndex: 310,
      });
    }, 3000);
    return () => {
      clearTimeout(cannons);
      clearTimeout(stars);
      clearTimeout(shower);
      clearInterval(stream);
      confetti.reset();
    };
  });
</script>

<div
  class="fixed inset-0 z-[210] flex items-center justify-center p-4"
  role="dialog"
  aria-modal="true"
  aria-label="Naik level"
>
  <button
    transition:fade={{ duration: 300 }}
    onclick={onclose}
    class="absolute inset-0 cursor-default backdrop-blur-xl"
    style={BACKDROP}
    aria-label="Tutup"
  ></button>
  <div
    transition:scale={{ start: 0.2, duration: 350 }}
    class="relative w-full max-w-sm overflow-hidden rounded-[3.5rem] border-2 border-amber-400/40 p-10 text-center text-white"
    style={CARD}
  >
    <button
      onclick={onclose}
      class="absolute top-4 right-4 z-10 cursor-pointer rounded-xl p-2 text-zinc-500 transition-colors hover:bg-white/10 hover:text-zinc-200"
      aria-label="Tutup"
    >
      <X size={18} />
    </button>
    <div
      class="pointer-events-none absolute top-0 left-0 h-36 w-full opacity-80"
      style={GLOW}
    ></div>
    <div class="relative flex flex-col items-center">
      <div class="relative mt-2 mb-6">
        <div
          class="absolute -inset-5 animate-[spin_8s_linear_infinite] rounded-full border-2 border-amber-400/50"
        ></div>
        <div
          class="relative flex h-32 w-32 items-center justify-center rounded-full border-4 border-black"
          style={MEDAL}
        >
          <Trophy size={50} class="text-amber-900" />
          <div
            class="absolute -top-3 -right-2 flex h-10 w-10 items-center justify-center rounded-full border-2 border-black bg-violet-600"
          >
            <Crown size={16} class="text-white" />
          </div>
          <div
            class="absolute -bottom-2 -left-2 flex h-8 w-8 animate-pulse items-center justify-center rounded-full border-2 border-black bg-blue-600"
          >
            <Zap size={14} class="fill-white text-white" />
          </div>
        </div>
      </div>
      <div
        class="mb-2 inline-flex items-center gap-2 rounded-full border border-amber-400/40 bg-amber-400/10 px-5 py-1.5 text-[10px] font-black tracking-[0.3em] text-amber-300 uppercase"
      >
        <Star size={10} class="fill-amber-300 text-amber-300" />
        Level Up!
        <Star size={10} class="fill-amber-300 text-amber-300" />
      </div>
      <div
        class="mb-1 bg-gradient-to-br from-amber-200 via-amber-400 to-amber-500 bg-clip-text text-7xl leading-none font-black text-transparent"
      >
        {level}
      </div>
      <h2 class="mb-2 text-xl font-black tracking-tight text-white">Mencapai Level {level}</h2>
      <p class="mx-auto mb-8 max-w-[270px] px-4 text-sm leading-relaxed text-zinc-400">
        Keren! Kemampuanmu terus berkembang. Setiap langkah kecil membawamu lebih dekat ke puncak!
      </p>
      <button
        onclick={onclose}
        class="w-full cursor-pointer rounded-2xl border-b-4 border-amber-800 bg-gradient-to-br from-amber-400 via-amber-500 to-amber-400 py-4 text-sm font-black tracking-widest text-amber-900 uppercase transition-all active:translate-y-1 active:border-b-0"
      >
        <span class="flex items-center justify-center gap-2">
          <ArrowUp size={14} />
          Lanjutkan Belajar
          <Sparkles size={14} />
        </span>
      </button>
    </div>
  </div>
</div>
