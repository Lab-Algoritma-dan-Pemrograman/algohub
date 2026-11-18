<script lang="ts">
  import { resolve } from '$app/paths';
  import { page } from '$app/state';
  import BarChart3 from '@lucide/svelte/icons/chart-column';
  import BookOpen from '@lucide/svelte/icons/book-open';
  import ChevronRight from '@lucide/svelte/icons/chevron-right';
  import Code2 from '@lucide/svelte/icons/code-xml';
  import LogIn from '@lucide/svelte/icons/log-in';
  import Play from '@lucide/svelte/icons/play';
  import Terminal from '@lucide/svelte/icons/terminal';
  import Zap from '@lucide/svelte/icons/zap';
  import { fade, fly } from 'svelte/transition';
  import { ssoRedirectMessage } from '#lib/auth-messages.ts';
  import LandingDemo from '#lib/components/LandingDemo.svelte';

  const ssoError = $derived(ssoRedirectMessage(page.url.searchParams.get('error')));

  const cards = [
    {
      icon: Terminal,
      title: 'Dasar Logika Pemrograman',
      desc: 'Pengenalan dasar algoritma, tipe data dasar, operator, input-output, serta pemecahan masalah sederhana dalam Bahasa C.',
      level: 'Bahasa C',
      stats: 'Level 1 • 6 Lesson',
    },
    {
      icon: Code2,
      title: 'Struktur Kontrol & Kondisional',
      desc: 'Menguasai percabangan (if-else, switch-case) dan perulangan (for, while, do-while) untuk mengontrol alur eksekusi program C.',
      level: 'Bahasa C',
      stats: 'Level 2 • 5 Lesson',
    },
    {
      icon: BookOpen,
      title: 'Array, Struct, & Operasi File',
      desc: 'Penyimpanan data majemuk (Array multi-dimensi), pembuatan tipe data kustom (Struct), dan manipulasi file external di C.',
      level: 'Bahasa C',
      stats: 'Level 3 • 7 Lesson',
    },
    {
      icon: Terminal,
      title: 'Pengenalan Sintaks Python',
      desc: 'Memahami filosofi Python, perbedaan sintaksis dengan C, tipe data dinamis, dan operasi dasar input-output.',
      level: 'Python',
      stats: 'Level 4 • 5 Lesson',
    },
    {
      icon: Code2,
      title: 'Kontrol Alur Python',
      desc: 'Implementasi percabangan if-elif-else dan perulangan for-in/while dengan blok indentasi khas Python.',
      level: 'Python',
      stats: 'Level 5 • 6 Lesson',
    },
    {
      icon: BarChart3,
      title: 'List, Dict, & File Handling',
      desc: 'Menguasai koleksi data Python (List, Tuple, Dictionary, Set) serta pembacaan dan penulisan berkas teks eksternal.',
      level: 'Python',
      stats: 'Level 6 • 8 Lesson',
    },
  ] as const;

  function scrollToPlayground() {
    document.getElementById('interactive-playground')?.scrollIntoView({ behavior: 'smooth' });
  }
</script>

<div class="text-dark selection:bg-maroon-bg selection:text-maroon min-h-screen font-sans">
  <nav
    class="border-maroon/5 sticky top-0 z-50 flex h-20 items-center justify-between border-b bg-[#FDFBF7]/80 px-6 backdrop-blur-md md:px-12"
  >
    <a href={resolve('/')} class="flex items-center gap-3 text-2xl font-black tracking-tighter">
      <div
        class="bg-maroon shadow-bubbly-maroon flex h-11 w-11 rotate-[-6deg] items-center justify-center rounded-xl text-2xl font-black text-white transition-transform hover:rotate-0"
      >
        E
      </div>
      <span class="text-maroon ml-1 font-black tracking-tight">E-Learning</span>
    </a>
    <a
      href={resolve('/login')}
      class="bg-maroon hover:bg-maroon-light shadow-bubbly-maroon btn-bubbly flex items-center gap-2 rounded-2xl px-8 py-3.5 text-sm font-black text-white transition-all active:scale-95"
    >
      <LogIn size={16} />
      Mulai Sekarang
    </a>
  </nav>

  {#if ssoError}
    <div
      role="alert"
      class="mx-auto mt-6 max-w-3xl rounded-2xl border border-red-100 bg-red-50 p-4 text-sm font-semibold text-red-700"
    >
      {ssoError}
    </div>
  {/if}

  <section
    class="relative mx-auto grid max-w-7xl grid-cols-1 items-center gap-16 px-6 py-20 md:px-12 md:py-32 lg:grid-cols-2"
  >
    <div
      class="bg-fun-yellow/10 absolute top-0 -left-20 -z-10 h-96 w-96 rounded-full blur-[120px]"
    ></div>
    <div
      class="bg-maroon/5 absolute -right-20 bottom-0 -z-10 h-96 w-96 rounded-full blur-[120px]"
    ></div>

    <div in:fly={{ y: 30, duration: 800 }} class="space-y-10">
      <div
        class="bg-maroon-bg text-maroon border-maroon/10 inline-flex items-center gap-2 rounded-2xl border px-5 py-2 text-xs font-black tracking-widest uppercase"
      >
        <Zap size={14} fill="currentColor" class="text-fun-yellow" />
        Join the pro league!
      </div>
      <h1 class="text-dark text-6xl leading-[0.95] font-black tracking-tight md:text-8xl">
        Jago
        <span
          class="from-maroon via-maroon-light bg-gradient-to-r to-rose-600 bg-clip-text text-transparent"
          >Coding</span
        >
        <br />
        Tanpa
        <span class="relative font-serif text-zinc-400 italic">
          Pusing.
          <span class="bg-fun-yellow/30 absolute bottom-1.5 left-0 -z-10 h-3 w-full rounded-full"
          ></span>
        </span>
      </h1>
      <p class="max-w-lg text-xl leading-relaxed font-bold text-zinc-500">
        Siap buat naik level? Yuk pelajari
        <span
          class="text-dark decoration-maroon font-black underline decoration-2 underline-offset-4"
          >Python</span
        >
        &amp;
        <span
          class="text-dark decoration-fun-blue font-black underline decoration-2 underline-offset-4"
          >Bahasa C</span
        > dengan cara yang lebih seru dan interaktif.
      </p>
      <div class="flex flex-col items-center gap-4 pt-4 sm:flex-row">
        <a
          href={resolve('/login')}
          class="group from-maroon to-maroon-light flex w-full items-center justify-center gap-3 rounded-2xl bg-gradient-to-r px-10 py-5 text-lg font-black text-white shadow-[0_8px_30px_rgb(138,21,56,0.2)] transition-all hover:-translate-y-0.5 hover:shadow-[0_8px_35px_rgb(138,21,56,0.35)] active:translate-y-1 sm:w-auto"
        >
          Let's Go!
          <ChevronRight size={22} class="transition-transform group-hover:translate-x-1.5" />
        </a>
        <button
          onclick={scrollToPlayground}
          class="group text-maroon border-maroon/10 hover:border-maroon/35 shadow-soft flex w-full cursor-pointer items-center justify-center gap-3 rounded-2xl border-2 bg-white/80 px-10 py-5 text-lg font-black transition-all hover:-translate-y-0.5 hover:shadow-md active:translate-y-1 sm:w-auto"
        >
          Coba Playground
          <Terminal size={22} class="transition-transform group-hover:rotate-12" />
        </button>
      </div>
      <div
        class="shadow-soft border-maroon/10 relative max-w-md overflow-hidden rounded-[2rem] border bg-white/60 p-5 backdrop-blur-md"
      >
        <div class="relative z-10 flex gap-4">
          <div
            class="bg-maroon-bg text-maroon border-maroon/10 flex h-12 w-12 shrink-0 items-center justify-center rounded-2xl border shadow-sm"
          >
            <Play size={20} fill="currentColor" class="ml-0.5" />
          </div>
          <div>
            <h4 class="text-dark mb-1 text-sm font-black tracking-wide uppercase">
              Gimana caranya masuk?
            </h4>
            <p class="text-xs leading-relaxed font-semibold text-zinc-500">
              Masuk pakai NIM dan password akun portal praktikum kamu, langsung di halaman ini.
              Gampang kan?
            </p>
          </div>
        </div>
      </div>
    </div>

    <div in:fade={{ duration: 1000, delay: 200 }} class="relative">
      <div
        class="relative z-10 overflow-hidden rounded-[3rem] border border-white/5 bg-zinc-950 p-8 shadow-[0_50px_100px_-20px_rgba(138,21,56,0.3)]"
      >
        <div class="mb-8 flex items-center justify-between px-2">
          <div class="flex gap-1.5">
            <div class="h-3 w-3 rounded-full bg-red-500/80"></div>
            <div class="h-3 w-3 rounded-full bg-amber-500/80"></div>
            <div class="h-3 w-3 rounded-full bg-green-500/80"></div>
          </div>
          <div class="font-mono text-[10px] font-bold tracking-widest text-zinc-600 uppercase">
            coding_adventure.v1
          </div>
        </div>
        <pre class="p-4 font-mono text-sm leading-relaxed text-zinc-300 md:text-lg"><span
            class="text-zinc-600 italic">// Bahasa C</span
          >
<span class="text-rose-400">#include</span> <span class="text-amber-300">&lt;stdio.h&gt;</span>
<span class="text-rose-400">int</span> <span class="text-blue-400">main</span>() &#123;
  <span class="text-blue-400">printf</span>(<span class="text-amber-300">"Waktunya Jadi Pro!"</span
          >);
  <span class="text-rose-400">return</span> <span class="text-amber-300">0</span>;
&#125;

<span class="text-zinc-600 italic"># Python</span>
<span class="text-rose-400">def</span> <span class="text-blue-400">semangat</span>():
  <span class="text-blue-400">print</span>(<span class="text-amber-300">"Kamu Pasti Bisa!"</span
          >)</pre>
      </div>
    </div>
  </section>

  <section id="interactive-playground" class="relative mx-auto max-w-7xl px-6 py-20 md:px-12">
    <div class="mb-10 space-y-4">
      <div
        class="bg-maroon-bg border-maroon/10 text-maroon inline-flex items-center gap-2 rounded-2xl border px-4 py-1.5 text-xs font-black tracking-wider uppercase"
      >
        <Terminal size={14} class="animate-pulse" />
        Live WebAssembly Playground
      </div>
      <h2 class="text-dark text-4xl font-black tracking-tight md:text-6xl">
        Cobain Editor Interaktif
        <span
          class="from-maroon via-maroon-light bg-gradient-to-r to-rose-600 bg-clip-text text-transparent"
          >Tanpa Batas</span
        >
      </h2>
      <p class="max-w-2xl text-base leading-relaxed font-bold text-zinc-500">
        Tidak perlu login, registrasi, atau instalasi. Tulis kode Python atau C Anda langsung di
        browser secara offline dan lihat eksekusinya secara instan!
      </p>
    </div>
    <LandingDemo />
  </section>

  <section class="border-maroon/5 bg-maroon-bg/25 border-y px-6 py-32 md:px-12">
    <div class="mx-auto max-w-7xl space-y-20">
      <div class="space-y-6 text-center">
        <h2 class="text-dark text-5xl font-black tracking-tighter md:text-7xl">
          Level Up Perjalananmu
        </h2>
        <p class="mx-auto max-w-2xl text-xl font-bold text-zinc-500">
          Gak perlu bingung mulai dari mana. Kita udah susun jalan ninja kamu biar makin jago
          step-by-step.
        </p>
      </div>
      <div class="grid grid-cols-1 gap-8 md:grid-cols-2 lg:grid-cols-3">
        {#each cards as card (card.title)}
          {@const isC = card.level === 'Bahasa C'}
          {@const Icon = card.icon}
          <div
            class="group relative overflow-hidden rounded-[2.5rem] border border-zinc-200 bg-white p-8 transition-all duration-300 hover:-translate-y-2 hover:shadow-lg {isC
              ? 'hover:border-blue-300'
              : 'hover:border-amber-300'}"
          >
            <div class="relative z-10 mb-8 flex items-start justify-between">
              <div
                class="flex h-14 w-14 items-center justify-center rounded-2xl shadow-sm transition-all group-hover:scale-110 {isC
                  ? 'bg-blue-50/50 text-blue-600'
                  : 'bg-amber-50/50 text-amber-600'}"
              >
                <Icon size={24} />
              </div>
              <span
                class="rounded-full border px-2.5 py-1 text-[10px] font-black tracking-wider uppercase {isC
                  ? 'border-blue-100 bg-blue-50 text-blue-600'
                  : 'border-amber-100 bg-amber-50 text-amber-700'}">{card.level}</span
              >
            </div>
            <div class="mb-1 text-[11px] font-bold tracking-wider text-zinc-400 uppercase">
              {card.stats}
            </div>
            <h3
              class="text-dark group-hover:text-maroon mb-3 text-2xl font-black transition-colors"
            >
              {card.title}
            </h3>
            <p class="min-h-[48px] text-sm leading-relaxed font-semibold text-zinc-500 italic">
              "{card.desc}"
            </p>
          </div>
        {/each}
      </div>
    </div>
  </section>

  <section class="px-6 py-32 md:px-12">
    <div
      class="to-zinc-900 via-maroon-dark/95 relative mx-auto max-w-6xl space-y-10 overflow-hidden rounded-[3.5rem] border border-white/5 bg-gradient-to-br from-zinc-950 p-12 text-center shadow-[0_50px_100px_-20px_rgba(138,21,56,0.35)] md:p-24"
    >
      <h2
        class="relative z-10 text-5xl leading-tight font-black tracking-tighter text-white md:text-8xl"
      >
        Siap Jadi <br />Game Changer?
      </h2>
      <p class="relative z-10 mx-auto max-w-2xl text-xl leading-relaxed font-bold text-zinc-400">
        Buruan masuk dan tunjukin skill kamu. Ratusan tantangan seru udah nungguin buat kamu
        taklukin!
      </p>
      <a
        href={resolve('/login')}
        class="text-dark hover:bg-maroon-bg hover:text-maroon relative z-10 mx-auto flex w-fit items-center gap-3 rounded-[2rem] bg-white px-16 py-6 text-xl font-black shadow-2xl transition-all duration-300 hover:scale-105 active:scale-95"
      >
        Mulai Sekarang!
        <ChevronRight size={22} />
      </a>
    </div>
  </section>

  <footer class="border-maroon/5 space-y-4 border-t px-6 py-16 text-center md:px-12">
    <div class="text-maroon/30 text-xl font-black tracking-tighter uppercase">
      Lab-AP E-Learning
    </div>
    <div class="text-sm font-bold text-zinc-500">Dibuat oleh Aslab AP untuk praktikan tercinta</div>
  </footer>
</div>
