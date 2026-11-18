<script lang="ts">
  import Bug from '@lucide/svelte/icons/bug';
  import CircleAlert from '@lucide/svelte/icons/circle-alert';
  import { createQuery } from '@tanstack/svelte-query';
  import BugHunt from '#lib/components/BugHunt.svelte';
  import { gameSettingsQuery, myPlaysQuery } from '#lib/queries/game.ts';

  let { isAdmin }: { isAdmin: boolean } = $props();

  const settings = createQuery(gameSettingsQuery);
  const plays = createQuery(myPlaysQuery);

  let active = $state<'c' | 'python' | null>(null);

  const weeklyLimit = $derived(settings.data?.bugHuntWeeklyLimit ?? 3);
  const playsUsed = $derived(plays.data?.playsUsed ?? 0);
  const limited = $derived(weeklyLimit > 0 && playsUsed >= weeklyLimit && !isAdmin);
  const disabled = $derived(settings.data?.bugHuntActive === false && !isAdmin);
  const languages = $derived([
    {
      id: 'c' as const,
      label: 'C',
      enabled: settings.data?.bugHuntCActive ?? true,
      button: 'bg-zinc-950 border-zinc-900 shadow-[0_4px_0_#023E8A]',
    },
    {
      id: 'python' as const,
      label: 'Py',
      enabled: settings.data?.bugHuntPythonActive ?? true,
      button: 'bg-rose-700 border-rose-800 shadow-[0_4px_0_#9F1239]',
    },
  ]);

  // Kuota di server bisa berubah dari tab lain, jadi dicek ulang setelah permainan ditutup.
  function close() {
    active = null;
    void plays.refetch();
  }
</script>

<div
  class="shadow-soft card-hover hover:border-maroon/20 group/bughunt relative overflow-hidden rounded-[2rem] border-2 border-gray-100 bg-white p-6 transition-all"
>
  <div
    class="pointer-events-none absolute top-0 right-0 -mt-20 -mr-20 h-40 w-40 rounded-full bg-rose-100/40 blur-[60px]"
  ></div>
  <div class="relative z-10">
    <div class="mb-4 flex items-center justify-between">
      <div class="flex items-center gap-4">
        <div
          class="flex h-14 w-14 rotate-[-5deg] items-center justify-center rounded-2xl bg-gradient-to-br from-rose-600 to-pink-700 text-white shadow-[0_4px_0_#5C0E25] transition-transform group-hover/bughunt:rotate-0"
        >
          <Bug size={22} />
        </div>
        <div>
          <h3 class="text-xl font-black tracking-tight text-zinc-900">Bug Hunt! 🎯</h3>
          <p class="text-xs font-bold text-zinc-500">Cari bug, dapatkan XP!</p>
        </div>
      </div>
      {#if weeklyLimit > 0}
        <div
          class="rotate-[3deg] rounded-xl border border-zinc-200 bg-zinc-50 px-3 py-1.5 text-center shadow-inner"
        >
          <div class="text-[8px] leading-none font-black tracking-widest text-zinc-400 uppercase">
            Sisa
          </div>
          <div
            class="mt-0.5 text-lg font-black {weeklyLimit - playsUsed <= 0
              ? 'text-rose-500'
              : 'text-zinc-900'}"
          >
            {Math.max(0, weeklyLimit - playsUsed)}
          </div>
        </div>
      {/if}
    </div>
    <p class="mb-6 text-sm leading-relaxed text-zinc-500">
      Uji ketelitian mata kamu dengan menemukan bug dalam potongan kode secepat mungkin.
    </p>
    {#if disabled}
      <div
        class="rounded-2xl border border-zinc-200 bg-zinc-50 px-4 py-4 text-center text-xs font-bold text-zinc-400 italic shadow-inner"
      >
        Game sedang dinonaktifkan oleh Admin 🛠️
      </div>
    {:else if limited}
      <div
        class="flex flex-col items-center gap-1.5 rounded-2xl border border-rose-200 bg-rose-50 px-4 py-4 text-center text-xs font-bold text-rose-500"
      >
        <CircleAlert size={16} />
        <span class="font-black">Limit Mingguan Tercapai</span>
        <span class="text-[9px] opacity-60">Kembali lagi minggu depan!</span>
      </div>
    {:else}
      <div class="mt-6 grid grid-cols-2 gap-3">
        {#each languages as lang (lang.id)}
          {#if lang.enabled}
            <button
              onclick={() => (active = lang.id)}
              class="btn-bubbly cursor-pointer rounded-2xl border px-4 py-3 text-sm font-black text-white transition-all active:translate-y-[4px] active:shadow-none {lang.button}"
            >
              <span class="flex items-center justify-center gap-2">
                <span
                  class="flex h-5 w-5 items-center justify-center rounded-lg bg-white/20 text-[10px] font-black"
                  >{lang.label}</span
                >
                Challenge
              </span>
            </button>
          {:else}
            <div
              class="rounded-2xl border border-zinc-200 bg-zinc-50 px-4 py-3 text-center text-xs font-bold text-zinc-400 italic"
            >
              {lang.label} nonaktif
            </div>
          {/if}
        {/each}
      </div>
    {/if}
  </div>
</div>

{#if active}
  <BugHunt language={active} onclose={close} />
{/if}
