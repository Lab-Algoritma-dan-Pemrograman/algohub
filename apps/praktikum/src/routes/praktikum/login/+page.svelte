<script lang="ts">
  import { goto } from '$app/navigation';
  import { User, Lock, AlertCircle, ArrowRight, ArrowLeft, Eye, EyeOff } from '@lucide/svelte';
  import { homeFor } from '#lib/auth.ts';
  import { createLoginMutation } from '#lib/queries/auth.ts';

  const loginMutation = createLoginMutation();

  let identifier = $state('');
  let password = $state('');
  let showPassword = $state(false);
  let err = $state('');
  let loading = $state(false);

  async function doLogin() {
    err = '';
    loading = true;
    try {
      const res = await loginMutation.mutateAsync({
        identifier: identifier.trim(),
        password,
      });
      goto(homeFor(res.user.role));
    } catch (e) {
      err = (e as Error).message;
    } finally {
      loading = false;
    }
  }
</script>

<div
  class="relative flex min-h-screen items-center justify-center p-4 overflow-hidden"
  style="background: url('/bg_login.jpg') no-repeat center center; background-size: cover;"
>
  <div class="absolute inset-0 bg-slate-950/65 backdrop-blur-[2px] z-0"></div>

  <div
    class="relative z-10 w-full max-w-md bg-white border border-rose-100/30 rounded-[2.5rem] p-8 sm:p-10 shadow-[0_25px_60px_-15px_rgba(0,0,0,0.4)] text-center"
  >
    <div
      class="absolute top-0 left-12 right-12 h-1 bg-gradient-to-r from-transparent via-[#8A1538] to-transparent rounded-t-[2.5rem]"
    ></div>

    <div class="mb-8">
      <div
        class="mx-auto mb-4 w-16 h-16 bg-gradient-to-br from-rose-50 to-rose-100 border border-rose-200/40 rounded-2xl flex items-center justify-center shadow-sm group"
      >
        <img
          src="/logo.png"
          alt="Logo Lab AP"
          class="h-10 w-10 object-contain transition-transform duration-500 group-hover:scale-110"
        />
      </div>
      <h1 class="text-2xl font-black tracking-tight text-slate-900">Login Portal</h1>
      <p class="mt-1 text-xs font-bold text-[#8A1538] uppercase tracking-wider">
        Lab Algoritma &amp; Pemrograman
      </p>
    </div>

    <form
      onsubmit={(e) => {
        e.preventDefault();
        doLogin();
      }}
      class="space-y-5 text-left"
    >
      <div>
        <label
          class="block text-xs font-bold text-slate-600 uppercase tracking-wider mb-2"
          for="identifier">Nomor Induk Mahasiswa / Email</label
        >
        <div class="relative">
          <div
            class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-4 text-slate-400"
          >
            <User size={15} />
          </div>
          <input
            id="identifier"
            class="w-full h-12 pl-10 pr-4 bg-slate-50/80 border border-slate-200 rounded-2xl text-sm font-semibold text-slate-800 placeholder-slate-400/50 focus-visible:outline-2 focus-visible:outline-primary focus-visible:outline-offset-2 focus:border-primary/50 focus:bg-white transition-all"
            bind:value={identifier}
            placeholder="Masukkan NIM atau Email"
            required
            aria-describedby={err ? 'login-err' : undefined}
          />
        </div>
        {#if err}
          <div
            id="login-err"
            role="alert"
            class="mt-2 rounded-xl border border-red-100 bg-red-50 p-3 text-xs font-semibold text-red-700 shadow-sm flex items-start gap-2 text-left animate-shake"
          >
            <AlertCircle class="shrink-0 text-red-500 mt-0.5" size={15} />
            <span>{err}</span>
          </div>
        {/if}
      </div>
      <div>
        <label class="block text-xs font-bold text-slate-600 uppercase tracking-wider mb-2" for="pw"
          >Password</label
        >
        <div class="relative">
          <div
            class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-4 text-slate-400"
          >
            <Lock size={15} />
          </div>
          <input
            id="pw"
            type={showPassword ? 'text' : 'password'}
            class="w-full h-12 pl-10 pr-11 bg-slate-50/80 border border-slate-200 rounded-2xl text-sm font-semibold text-slate-800 placeholder-slate-400/50 focus-visible:outline-2 focus-visible:outline-primary focus-visible:outline-offset-2 focus:border-primary/50 focus:bg-white transition-all"
            bind:value={password}
            placeholder="Masukkan password Anda"
            required
          />
          <button
            type="button"
            onclick={() => (showPassword = !showPassword)}
            class="absolute inset-y-0 right-0 flex items-center pr-4 text-slate-400 hover:text-slate-600 transition-colors"
            aria-label={showPassword ? 'Sembunyikan password' : 'Tampilkan password'}
          >
            {#if showPassword}<EyeOff size={15} />{:else}<Eye size={15} />{/if}
          </button>
        </div>
        <div class="mt-2 text-right">
          <a
            href="/praktikum/lupa-password"
            class="text-[11px] font-bold text-[#8A1538] hover:text-[#610a24] transition-colors"
            >Lupa password?</a
          >
        </div>
      </div>
      <button
        class="w-full h-12 bg-[#8A1538] hover:bg-[#730d2d] text-white rounded-2xl text-sm font-black flex items-center justify-center gap-2 shadow-lg shadow-[#8A1538]/10 transition-all active:scale-[0.99]"
        disabled={loading}
      >
        {loading ? 'Autentikasi…' : 'Masuk ke Portal'}
        {#if !loading}<ArrowRight size={14} />{/if}
      </button>
    </form>

    <p class="mt-5 text-xs font-bold text-slate-500">
      Belum punya akun?
      <a href="/praktikum/register" class="text-[#8A1538] hover:text-[#610a24] transition-colors"
        >Daftar di sini</a
      >
    </p>

    <div class="pt-6 border-t border-slate-100 mt-6 text-center">
      <a
        href="/info"
        class="inline-flex items-center gap-1.5 text-xs font-bold text-slate-400 hover:text-primary transition-colors"
      >
        <ArrowLeft size={12} /> Kembali ke portal utama
      </a>
    </div>
  </div>
</div>
