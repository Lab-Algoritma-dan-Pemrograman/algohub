<script lang="ts">
  import { goto } from '$app/navigation';
  import { resolve } from '$app/paths';
  import { page } from '$app/state';
  import AlertCircle from '@lucide/svelte/icons/circle-alert';
  import ArrowLeft from '@lucide/svelte/icons/arrow-left';
  import ArrowRight from '@lucide/svelte/icons/arrow-right';
  import Eye from '@lucide/svelte/icons/eye';
  import EyeOff from '@lucide/svelte/icons/eye-off';
  import Loader2 from '@lucide/svelte/icons/loader-circle';
  import Lock from '@lucide/svelte/icons/lock';
  import User from '@lucide/svelte/icons/user';
  import { loginDirect } from '#lib/auth.ts';
  import { authErrorMessage, LOGIN_REQUIRED } from '#lib/auth-messages.ts';

  let identifier = $state('');
  let password = $state('');
  let showPassword = $state(false);
  let loading = $state(false);
  let failure = $state('');
  const error = $derived(failure || page.state.ssoError || '');

  async function submit(event: SubmitEvent) {
    event.preventDefault();
    failure = '';
    if (!identifier.trim() || !password) {
      failure = LOGIN_REQUIRED;
      return;
    }
    loading = true;
    try {
      await loginDirect(identifier.trim(), password);
      await goto(resolve('/(app)/dashboard'));
    } catch (e) {
      failure = authErrorMessage(e, 'Login gagal. Coba lagi.');
    } finally {
      loading = false;
    }
  }
</script>

<div class="flex min-h-screen flex-col font-sans">
  <nav
    class="border-maroon/5 sticky top-0 z-50 flex h-20 items-center justify-between border-b bg-[#FDFBF7]/80 px-6 backdrop-blur-md md:px-12"
  >
    <a href={resolve('/')} class="flex items-center gap-3 text-2xl font-black tracking-tighter">
      <div
        class="bg-maroon shadow-bubbly-maroon flex h-11 w-11 rotate-[-6deg] items-center justify-center rounded-xl text-2xl font-black text-white"
      >
        E
      </div>
      <span class="text-maroon ml-1 font-black tracking-tight">E-Learning</span>
    </a>
    <a
      href={resolve('/')}
      class="hover:text-maroon inline-flex items-center gap-1.5 text-sm font-bold text-zinc-500 transition-colors"
    >
      <ArrowLeft size={15} /> Kembali ke beranda
    </a>
  </nav>

  <div class="flex flex-1 items-center justify-center p-4">
    <div
      class="border-maroon/10 relative w-full max-w-md rounded-[2.5rem] border bg-white p-8 text-center shadow-[0_25px_60px_-15px_rgba(138,21,56,0.25)] sm:p-10"
    >
      <div
        class="via-maroon absolute top-0 right-12 left-12 h-1 rounded-t-[2.5rem] bg-gradient-to-r from-transparent to-transparent"
      ></div>
      <div class="mb-8">
        <div
          class="bg-maroon shadow-bubbly-maroon mx-auto mb-4 flex h-16 w-16 rotate-[-6deg] items-center justify-center rounded-2xl text-3xl font-black text-white"
        >
          E
        </div>
        <h1 class="text-dark text-2xl font-black tracking-tight">Masuk E-Learning</h1>
        <p class="text-maroon mt-1 text-xs font-bold tracking-wider uppercase">
          Lab Algoritma &amp; Pemrograman
        </p>
      </div>

      {#if error}
        <div
          role="alert"
          class="mb-5 flex items-start gap-2.5 rounded-2xl border border-red-100 bg-red-50 p-4 text-left text-xs font-semibold text-red-700 shadow-sm"
        >
          <AlertCircle size={15} class="mt-0.5 shrink-0 text-red-500" />
          <span>{error}</span>
        </div>
      {/if}

      <form onsubmit={submit} class="space-y-5 text-left">
        <div>
          <label
            class="mb-2 block text-xs font-bold tracking-wider text-zinc-600 uppercase"
            for="identifier">Nomor Induk Mahasiswa / Email</label
          >
          <div class="relative">
            <div
              class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-4 text-zinc-400"
            >
              <User size={15} />
            </div>
            <input
              id="identifier"
              class="focus:border-maroon/50 h-12 w-full rounded-2xl border border-zinc-200 bg-zinc-50/80 pr-4 pl-10 text-sm font-semibold text-zinc-800 placeholder-zinc-400/50 transition-all focus:bg-white focus:outline-none"
              bind:value={identifier}
              placeholder="Masukkan NIM atau Email"
              autocomplete="username"
              required
            />
          </div>
        </div>
        <div>
          <label
            class="mb-2 block text-xs font-bold tracking-wider text-zinc-600 uppercase"
            for="pw">Password</label
          >
          <div class="relative">
            <div
              class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-4 text-zinc-400"
            >
              <Lock size={15} />
            </div>
            <input
              id="pw"
              type={showPassword ? 'text' : 'password'}
              class="focus:border-maroon/50 h-12 w-full rounded-2xl border border-zinc-200 bg-zinc-50/80 pr-11 pl-10 text-sm font-semibold text-zinc-800 placeholder-zinc-400/50 transition-all focus:bg-white focus:outline-none"
              bind:value={password}
              placeholder="Masukkan password Anda"
              autocomplete="current-password"
              required
            />
            <button
              type="button"
              onclick={() => (showPassword = !showPassword)}
              class="absolute inset-y-0 right-0 flex cursor-pointer items-center pr-4 text-zinc-400 transition-colors hover:text-zinc-600"
              aria-label={showPassword ? 'Sembunyikan password' : 'Tampilkan password'}
            >
              {#if showPassword}<EyeOff size={15} />{:else}<Eye size={15} />{/if}
            </button>
          </div>
        </div>
        <button
          type="submit"
          disabled={loading}
          class="bg-maroon hover:bg-maroon-light shadow-maroon/20 flex h-12 w-full cursor-pointer items-center justify-center gap-2 rounded-2xl text-sm font-black text-white shadow-lg transition-all active:scale-[0.99] disabled:opacity-60"
        >
          {#if loading}
            <Loader2 size={15} class="animate-spin" /> Memeriksa akun…
          {:else}
            Masuk ke E-Learning <ArrowRight size={15} />
          {/if}
        </button>
      </form>
      <p class="mt-5 text-xs font-semibold text-zinc-500">
        Akun yang sama dengan portal praktikum. Belum punya password? Daftar dulu di portal
        praktikum.
      </p>
    </div>
  </div>
</div>
