<script lang="ts">
  import { goto } from '$app/navigation';
  import { resolve } from '$app/paths';
  import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogHeader,
    DialogTitle,
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
  } from '@algohub/ui';
  import Loader2 from '@lucide/svelte/icons/loader-2';
  import { createQuery } from '@tanstack/svelte-query';
  import { login } from '#lib/api.ts';
  import { startSession } from '#lib/auth.ts';
  import FloatingIcons from '#lib/components/FloatingIcons.svelte';
  import { createRpcMutation, fetchRpc, rpcQueryOptions } from '#lib/queries/rpc.ts';
  import { toast } from '#lib/toast.ts';

  interface PublicSettings {
    is_recruitment_open?: boolean;
    recruitment_link?: string;
    active_shift?: string;
    announcement?: string;
    login_guide_text?: string;
    procedure_text?: { step: string; desc: string }[];
    reschedule_steps?: string;
  }

  interface Step {
    title: string;
    desc: string;
  }

  const settingsQuery = createQuery(() => rpcQueryOptions<PublicSettings[]>('get_public_settings'));
  const sysSettings = $derived(
    settingsQuery.data && settingsQuery.data.length > 0 ? settingsQuery.data[0] : null,
  );
  const register = createRpcMutation('register_user');

  let isLoading = $state(false);
  let showLoginPassword = $state(false);
  let showSignupPassword = $state(false);
  let activeTab = $state<'masuk' | 'daftar'>('masuk');
  let openPanduan = $state(false);
  let openProsedur = $state(false);
  let openReschedule = $state(false);
  let loginUsername = $state('');
  let loginPassword = $state('');
  let signupName = $state('');
  let signupUsername = $state('');
  let signupPassword = $state('');
  let signupPhone = $state('');
  let signupRole = $state<'praktikan' | 'penyewa'>('praktikan');

  const roleLabels = { praktikan: 'Praktikan (Mahasiswa)', penyewa: 'Penyewa (Umum)' };

  async function handleLogin(e: SubmitEvent) {
    e.preventDefault();
    isLoading = true;
    try {
      const res = await login({ username: loginUsername, password: loginPassword });
      const resData = (res.body ?? {}) as {
        user?: { id: number; role: string; full_name?: string };
        token?: string;
        error?: string;
        details?: string;
      };
      if (!res.ok) {
        const errMsg =
          resData.error ||
          (res.status === 404
            ? 'Endpoint API tidak ditemukan (/api/auth/login). Pastikan Vercel Dev berjalan.'
            : 'Gagal melakukan login');
        const errDetails = resData.details || '';
        toast.error(errMsg, errDetails ? { description: errDetails } : undefined);
        return;
      }
      const { user: userData, token } = resData;
      if (!userData || !token) {
        toast.error('Gagal Masuk', { description: 'Respon server tidak valid.' });
        return;
      }
      await startSession(token, userData);
      toast.success('Login Berhasil!', { description: `Selamat datang, ${userData.full_name}` });
      await goto(resolve('/(app)/beranda'));
    } catch (err) {
      console.error(err);
      toast.error('Terjadi Kesalahan', {
        description: (err as Error).message || 'Gagal menghubungi server',
      });
    } finally {
      isLoading = false;
    }
  }

  async function handleSignup(e: SubmitEvent) {
    e.preventDefault();
    isLoading = true;
    if (signupPassword.length < 6) {
      toast.error('Password terlalu pendek', { description: 'Minimal 6 karakter.' });
      isLoading = false;
      return;
    }
    if (signupRole === 'penyewa' && !signupPhone) {
      toast.error('Mohon Isi WhatsApp!', {
        description:
          'Nomor WA wajib diisi agar asisten bisa menghubungi Anda untuk verifikasi identitas.',
      });
      isLoading = false;
      return;
    }
    try {
      const usernameExists = await fetchRpc<boolean>('check_username_exists', {
        p_username: signupUsername,
      });
      if (usernameExists) {
        toast.error('Gagal Daftar', { description: 'ID / NIM sudah terdaftar.' });
        return;
      }
      const { error } = await register.mutateAsync({
        p_username: signupUsername,
        p_password: signupPassword,
        p_full_name: signupName,
        p_role: signupRole,
        p_nim: signupRole === 'praktikan' ? signupUsername : null,
        p_assistant_code: null,
        p_phone_number: signupPhone || null,
        p_is_active: true,
      });
      if (error) throw error;
      toast.success('Pendaftaran Berhasil!', {
        description: 'Silakan login dengan akun baru Anda.',
      });
      signupUsername = '';
      signupPassword = '';
      signupName = '';
      signupPhone = '';
      activeTab = 'masuk';
    } catch (err) {
      console.error(err);
      toast.error('Gagal Daftar', { description: (err as { message?: string }).message });
    } finally {
      isLoading = false;
    }
  }

  type Announcement = { type: 'oprec' | 'reschedule' | 'info'; text: string; kelas?: string };

  const announcementsList = $derived.by(() => {
    const list: Announcement[] = [];
    if (sysSettings?.is_recruitment_open) {
      list.push({ type: 'oprec', text: 'OPEN RECRUITMENT ASISTEN SEDANG DIBUKA! DAFTAR SEKARANG' });
    }
    if (
      sysSettings?.active_shift &&
      sysSettings.active_shift !== 'all' &&
      sysSettings.active_shift !== 'none'
    ) {
      list.push({
        type: 'info',
        text: `AKSES LOGIN SHIFT ${sysSettings.active_shift} SEDANG DIBUKA`,
      });
    }
    if (sysSettings?.announcement) {
      const isReschedule =
        /ganti\s*jadwal|reschedule|pindah\s*jadwal|pindah\s*shift|pengganti/i.test(
          sysSettings.announcement,
        );
      list.push({
        type: isReschedule ? 'reschedule' : 'info',
        kelas: isReschedule ? 'GANTI JADWAL' : undefined,
        text: sysSettings.announcement,
      });
    }
    if (list.length === 0) {
      const hour = new Date().getHours();
      list.push({
        type: 'info',
        text:
          hour < 11
            ? '🌅 Selamat Pagi! Semangat Praktikum Hari Ini!'
            : hour < 15
              ? '☀️ Selamat Siang! Jangan Lupa Istirahat ya~'
              : hour < 19
                ? '🌆 Selamat Sore! Tetap Semangat Belajarnya!'
                : '🌙 Selamat Malam! Tetap Produktif & Jaga Kesehatan!',
      });
    }
    return list;
  });

  const shiftActive = $derived(
    !!sysSettings?.active_shift &&
      sysSettings.active_shift !== 'all' &&
      sysSettings.active_shift !== 'none',
  );

  function parseLines(text: string): Step[] {
    return text
      .split('\n')
      .filter((l) => l.trim())
      .map((line, idx) => {
        const colonIdx = line.indexOf(':');
        return {
          title: colonIdx > -1 ? line.substring(0, colonIdx).trim() : `Langkah ${idx + 1}`,
          desc: colonIdx > -1 ? line.substring(colonIdx + 1).trim() : line.trim(),
        };
      });
  }

  const inputClass =
    'w-full pl-11 py-3.5 rounded-2xl font-bold transition-all text-sm border-2 bg-slate-50 border-slate-200 text-slate-800 placeholder-slate-300 focus:outline-none focus:border-maroon focus:bg-white focus:ring-4 focus:ring-maroon/5 focus:shadow-sm disabled:opacity-50';
  const tabClass = (active: boolean) =>
    `flex-1 py-2 rounded-xl text-sm transition-all font-black focus:outline-none ${
      active
        ? 'bg-white text-maroon border-2 border-slate-200 shadow-[0_4px_0_rgba(203,213,225,0.4)]'
        : 'text-slate-400 font-bold hover:text-slate-700'
    }`;
  const submitClass =
    'btn-bubbly-action w-full mt-4 bg-maroon hover:bg-maroon-light text-white py-4 rounded-2xl font-black text-base border-2 border-maroon-dark shadow-bubbly-maroon flex items-center justify-center gap-2 disabled:opacity-50 disabled:translate-y-0 disabled:shadow-none';
  const labelClass = 'block text-xs font-black text-slate-400 uppercase tracking-widest mb-2 ml-1';
  const iconClass =
    'absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none text-slate-300';
  const procButtonClass =
    'bg-white border-2 border-slate-100 rounded-2xl p-3 shadow-[0_4px_0_rgba(203,213,225,0.4)] text-center flex flex-col items-center justify-center gap-2 hover:-translate-y-1 btn-bubbly-action group focus:outline-none';
</script>

{#snippet star()}
  <svg class="h-3 w-3 text-pink-400" fill="currentColor" viewBox="0 0 20 20"
    ><path
      d="M9.049 2.927c.3-.921 1.603-.921 1.902 0l1.07 3.292a1 1 0 00.95.69h3.462c.969 0 1.371 1.24.588 1.81l-2.8 2.034a1 1 0 00-.364 1.118l1.07 3.292c.3.921-.755 1.688-1.54 1.118l-2.8-2.034a1 1 0 00-1.175 0l-2.8 2.034c-.784.57-1.838-.197-1.539-1.118l1.07-3.292a1 1 0 00-.364-1.118L2.98 8.72c-.783-.57-.38-1.81.588-1.81h3.461a1 1 0 00.951-.69l1.07-3.292z"
    /></svg
  >
{/snippet}

{#snippet warnIcon(cls: string)}
  <svg class={cls} fill="currentColor" viewBox="0 0 20 20"
    ><path
      fill-rule="evenodd"
      d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z"
      clip-rule="evenodd"
    /></svg
  >
{/snippet}

{#snippet announcements()}
  {#each announcementsList as item, idx (idx)}
    <span class="inline-flex items-center">
      {#if item.type === 'oprec'}
        <span class="mx-3 flex shrink-0 items-center gap-1 font-black text-pink-600">
          {@render star()}
          {item.text}
          {@render star()}
        </span>
      {:else if item.type === 'reschedule'}
        <span class="mx-3 flex shrink-0 items-center gap-1.5 text-amber-600">
          <span
            class="rounded-md border border-amber-200 bg-amber-100 px-2 py-0.5 text-[10px] text-amber-600"
          >
            {item.kelas}
          </span>
          <span>{item.text}</span>
        </span>
      {:else if item.text.includes('SHIFT') && sysSettings?.active_shift}
        <span class="mx-3 flex shrink-0 items-center gap-1.5">
          {@render warnIcon('w-3 h-3 text-amber-500 shrink-0')}
          <span
            class="rounded-lg border border-amber-300 bg-amber-100 px-2 py-0.5 text-[10px] font-black text-amber-700"
            >SHIFT {sysSettings?.active_shift} AKTIF</span
          >
          <span class="font-bold text-amber-600"
            >{item.text.replace(
              `AKSES LOGIN SHIFT ${sysSettings?.active_shift} SEDANG DIBUKA`,
              'Login hanya untuk shift ini',
            )}</span
          >
        </span>
      {:else}
        <span class="mx-3 flex shrink-0 items-center gap-1.5 text-slate-500">
          <svg class="h-3 w-3 shrink-0 text-blue-400" fill="currentColor" viewBox="0 0 20 20"
            ><path
              fill-rule="evenodd"
              d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z"
              clip-rule="evenodd"
            /></svg
          >
          <span>{item.text}</span>
        </span>
      {/if}
      {#if idx < announcementsList.length - 1}
        <span class="mx-2 text-slate-300">|</span>
      {/if}
    </span>
  {/each}
{/snippet}

{#snippet steps(list: Step[], badge: string)}
  {#each list as step, idx (idx)}
    <div
      class="flex items-start gap-3 {idx < list.length - 1 ? 'border-b border-slate-100 pb-3' : ''}"
    >
      <span
        class="flex h-6 w-6 shrink-0 items-center justify-center rounded-full text-xs font-black {badge}"
        >{idx + 1}</span
      >
      <div>
        <strong class="block font-bold text-slate-900">{step.title}</strong>
        {step.desc}
      </div>
    </div>
  {/each}
{/snippet}

<div
  class="relative flex min-h-screen flex-col items-center justify-center overflow-x-hidden bg-[#F8FAFC] p-4 antialiased sm:p-8"
>
  <FloatingIcons />

  <div class="animate-fade-in relative z-10 mt-8 flex w-full max-w-[480px] flex-col gap-5 md:mt-0">
    <div class="mb-2 text-center">
      <img src="/logo.png" alt="Logo Lab AP" class="mx-auto mb-4 h-16 w-16 object-contain" />
      <h1 class="mb-1 text-2xl leading-tight font-black text-slate-800">Laboratorium Algoritma</h1>
      <p class="text-[11px] font-bold tracking-widest text-slate-400 uppercase">
        Sistem Manajemen Terpadu
      </p>
    </div>

    <div class="flex h-[46px] w-full gap-2">
      <div
        class="relative flex flex-1 items-center overflow-hidden rounded-2xl border-2 border-pink-200 bg-white shadow-[0_4px_0_#FBCFE8]"
      >
        <div
          class="text-maroon relative z-20 flex h-full shrink-0 items-center gap-1.5 border-r-2 border-pink-100 bg-white px-3 font-black"
        >
          <svg
            class="h-4 w-4 shrink-0 text-pink-500"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            viewBox="0 0 24 24"
            ><path
              stroke-linecap="round"
              stroke-linejoin="round"
              d="M11 5.882V19.24a1.76 1.76 0 01-3.417.592l-2.147-6.15M18 13a3 3 0 100-6M5.436 13.683A4.001 4.001 0 017 6h1.832c4.1 0 7.625-1.234 9.168-3v14c-1.543-1.766-5.067-3-9.168-3H7a3.988 3.988 0 01-1.564-.317z"
            /></svg
          >
          Info:
          <div
            class="absolute top-0 -right-3 h-full w-3 bg-linear-to-r from-white to-transparent"
          ></div>
        </div>
        <div class="mask-edges relative flex h-full flex-1 items-center overflow-hidden">
          <div
            class="animate-marquee-half flex w-max items-center text-[11px] font-bold tracking-widest whitespace-nowrap uppercase"
          >
            {@render announcements()}
            <span class="mx-2 text-slate-300">|</span>
            {@render announcements()}
          </div>
        </div>
      </div>
      {#if sysSettings?.is_recruitment_open}
        <a
          href={sysSettings.recruitment_link || 'https://bit.ly/OprecAsistenLabAP'}
          target="_blank"
          rel="noopener noreferrer external"
          class="btn-bubbly-action flex h-full shrink-0 items-center justify-center gap-1 rounded-xl border-2 border-pink-600 bg-pink-500 px-4 text-[11px] font-black text-white shadow-[0_4px_0_#BE185D] hover:bg-pink-600"
        >
          Daftar <i class="fa-solid fa-bolt"></i>
        </a>
      {/if}
    </div>

    <div
      class="rounded-[2rem] border-2 border-slate-100 bg-white p-7 shadow-[0_8px_0_rgba(203,213,225,0.4)]"
    >
      <div class="mb-6 flex rounded-2xl border-2 border-slate-100 bg-slate-50 p-1.5">
        <button
          type="button"
          onclick={() => (activeTab = 'masuk')}
          class={tabClass(activeTab === 'masuk')}
        >
          Masuk
        </button>
        <button
          type="button"
          onclick={() => (activeTab = 'daftar')}
          class={tabClass(activeTab === 'daftar')}
        >
          Daftar
        </button>
      </div>

      {#if activeTab === 'masuk'}
        <form onsubmit={handleLogin} class="space-y-4">
          <div>
            <label for="login-username" class={labelClass}>NIM / Username</label>
            <div class="relative">
              <div class={iconClass}><i class="fa-regular fa-user"></i></div>
              <input
                id="login-username"
                type="text"
                placeholder="Masukkan NIM atau Username"
                bind:value={loginUsername}
                required
                disabled={isLoading}
                class="{inputClass} pr-4"
              />
            </div>
          </div>
          <div>
            <label for="login-password" class={labelClass}>Password</label>
            <div class="relative">
              <div class={iconClass}><i class="fa-solid fa-lock"></i></div>
              <input
                id="login-password"
                type={showLoginPassword ? 'text' : 'password'}
                placeholder="••••••••"
                bind:value={loginPassword}
                required
                disabled={isLoading}
                class="{inputClass} pr-12"
              />
              <button
                type="button"
                aria-label="Tampilkan password"
                onclick={() => (showLoginPassword = !showLoginPassword)}
                class="hover:text-maroon absolute inset-y-0 right-0 flex items-center pr-4 text-slate-400 transition-colors focus:outline-none"
              >
                <i class="fa-regular {showLoginPassword ? 'fa-eye' : 'fa-eye-slash'}"></i>
              </button>
            </div>
          </div>
          <button type="submit" disabled={isLoading} class={submitClass}>
            {#if isLoading}
              <Loader2 class="h-5 w-5 animate-spin" />
              Memproses...
            {:else}
              Masuk Sekarang <i class="fa-solid fa-arrow-right-to-bracket"></i>
            {/if}
          </button>
        </form>
      {:else}
        <form onsubmit={handleSignup} class="space-y-4">
          <div>
            <span class={labelClass}>Daftar Sebagai</span>
            <div class="relative">
              <div
                class="pointer-events-none absolute inset-y-0 left-0 z-10 flex items-center pl-4 text-slate-400"
              >
                <i class="fa-solid fa-users"></i>
              </div>
              <Select type="single" bind:value={signupRole} disabled={isLoading}>
                <SelectTrigger
                  class="focus:border-maroon focus:ring-maroon/5 h-auto w-full rounded-2xl border-2 border-slate-200 bg-slate-50 py-3.5 pl-11 text-sm font-bold text-slate-800 transition-all focus:bg-white focus:shadow-sm focus:ring-4 focus:outline-none"
                >
                  {roleLabels[signupRole]}
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="praktikan" label="Praktikan (Mahasiswa)" />
                  <SelectItem value="penyewa" label="Penyewa (Umum)" />
                </SelectContent>
              </Select>
            </div>
          </div>
          <div>
            <label for="signup-name" class={labelClass}>Nama Lengkap</label>
            <div class="relative">
              <div class={iconClass}><i class="fa-regular fa-id-card"></i></div>
              <input
                id="signup-name"
                type="text"
                placeholder="Nama Lengkap Anda"
                bind:value={signupName}
                required
                disabled={isLoading}
                class="{inputClass} pr-4"
              />
            </div>
          </div>
          <div>
            <label for="signup-phone" class={labelClass}>
              Nomor WhatsApp {#if signupRole === 'penyewa'}<span class="text-red-500">*</span>{/if}
            </label>
            <div class="relative">
              <div class={iconClass}><i class="fa-brands fa-whatsapp text-lg"></i></div>
              <input
                id="signup-phone"
                type="text"
                placeholder="Contoh: 081234567890"
                bind:value={signupPhone}
                required={signupRole === 'penyewa'}
                disabled={isLoading}
                class="{inputClass} pr-4"
              />
            </div>
          </div>
          <div>
            <label for="signup-username" class={labelClass}>
              {signupRole === 'praktikan' ? 'NIM' : 'Username'}
            </label>
            <div class="relative">
              <div class={iconClass}><i class="fa-regular fa-user"></i></div>
              <input
                id="signup-username"
                type="text"
                placeholder={signupRole === 'praktikan' ? 'Contoh: 202511090' : 'Masukkan Username'}
                bind:value={signupUsername}
                required
                disabled={isLoading}
                class="{inputClass} pr-4"
              />
            </div>
          </div>
          <div>
            <label for="signup-password" class={labelClass}>Password</label>
            <div class="relative">
              <div class={iconClass}><i class="fa-solid fa-lock"></i></div>
              <input
                id="signup-password"
                type={showSignupPassword ? 'text' : 'password'}
                placeholder="Minimal 6 karakter"
                bind:value={signupPassword}
                required
                disabled={isLoading}
                class="{inputClass} pr-12"
              />
              <button
                type="button"
                aria-label="Tampilkan password"
                onclick={() => (showSignupPassword = !showSignupPassword)}
                class="hover:text-maroon absolute inset-y-0 right-0 flex items-center pr-4 text-slate-400 transition-colors focus:outline-none"
              >
                <i class="fa-regular {showSignupPassword ? 'fa-eye' : 'fa-eye-slash'}"></i>
              </button>
            </div>
          </div>
          <button type="submit" disabled={isLoading} class={submitClass}>
            {#if isLoading}
              <Loader2 class="h-5 w-5 animate-spin" />
              Mendaftar...
            {:else}
              Daftar Sekarang <i class="fa-solid fa-user-plus"></i>
            {/if}
          </button>
        </form>
      {/if}
    </div>

    <div class="mt-2 grid grid-cols-3 gap-3">
      <button type="button" onclick={() => (openPanduan = true)} class={procButtonClass}>
        <div
          class="flex h-8 w-8 items-center justify-center rounded-full bg-slate-50 transition-transform group-hover:scale-110"
        >
          <i class="fa-solid fa-right-to-bracket text-lg text-blue-500"></i>
        </div>
        <span class="text-[10px] leading-tight font-black tracking-wide text-slate-500 uppercase">
          Panduan<br />Login
        </span>
      </button>
      <button type="button" onclick={() => (openProsedur = true)} class={procButtonClass}>
        <div
          class="flex h-8 w-8 items-center justify-center rounded-full bg-slate-50 transition-transform group-hover:scale-110"
        >
          <i class="fa-solid fa-list-check text-lg text-emerald-500"></i>
        </div>
        <span class="text-[10px] leading-tight font-black tracking-wide text-slate-500 uppercase">
          Alur<br />Prosedur
        </span>
      </button>
      <button type="button" onclick={() => (openReschedule = true)} class={procButtonClass}>
        <div
          class="flex h-8 w-8 items-center justify-center rounded-full bg-slate-50 transition-transform group-hover:scale-110"
        >
          <i class="fa-solid fa-calendar-days text-lg text-orange-500"></i>
        </div>
        <span class="text-[10px] leading-tight font-black tracking-wide text-slate-500 uppercase">
          Langkah<br />Reschedule
        </span>
      </button>
    </div>
  </div>

  <div class="fixed bottom-4 z-0 hidden w-full text-center md:block">
    <p class="text-[11px] font-bold text-slate-400">
      &copy; 2026 Laboratorium Algoritma Pemrograman.
    </p>
  </div>

  <Dialog bind:open={openPanduan}>
    <DialogContent class="max-w-md bg-white">
      <DialogHeader>
        <DialogTitle class="font-black text-slate-800">Detail Ketentuan Shift Login</DialogTitle>
        <DialogDescription class="text-xs font-bold tracking-wider text-slate-400 uppercase">
          Panduan Akses Sistem Terpadu
        </DialogDescription>
      </DialogHeader>
      <div class="mt-2 space-y-4 text-sm font-medium text-slate-700">
        {@render steps(
          sysSettings?.login_guide_text
            ? parseLines(sysSettings.login_guide_text)
            : [
                {
                  title: 'Cek Shift Aktif',
                  desc: 'Lihat pengumuman di atas. Login hanya dibuka sesuai shift yang sedang aktif.',
                },
                {
                  title: 'Gunakan NIM/Username',
                  desc: 'Masukkan NIM (untuk praktikan) atau username yang sudah terdaftar.',
                },
                {
                  title: 'Password Benar',
                  desc: 'Pastikan password minimal 6 karakter dan sesuai yang didaftarkan.',
                },
                {
                  title: 'Hubungi Asisten',
                  desc: 'Jika tetap tidak bisa login di shift aktif Anda, segera hubungi Asisten PJ kelas.',
                },
              ],
          'bg-blue-100 text-blue-600',
        )}
        {#if shiftActive}
          <div
            class="rounded-xl border border-amber-100 bg-amber-50 p-4 text-sm leading-relaxed font-medium text-amber-900/80"
          >
            <span class="mb-1 block font-black text-amber-700">⚠️ SHIFT AKTIF SAAT INI</span>
            Sistem login hanya dibuka untuk praktikan yang terdaftar di
            <strong>SHIFT {sysSettings?.active_shift}</strong>.
          </div>
        {/if}
      </div>
    </DialogContent>
  </Dialog>

  <Dialog bind:open={openProsedur}>
    <DialogContent class="max-w-md bg-white">
      <DialogHeader>
        <DialogTitle class="font-black text-slate-800">Alur Prosedur Praktikan</DialogTitle>
        <DialogDescription class="text-xs font-bold tracking-wider text-slate-400 uppercase">
          Siklus Aktivitas Praktikum
        </DialogDescription>
      </DialogHeader>
      <div class="mt-2 space-y-4 text-sm font-medium text-slate-700">
        {@render steps(
          sysSettings?.procedure_text && Array.isArray(sysSettings.procedure_text)
            ? sysSettings.procedure_text.map((item) => ({ title: item.step, desc: item.desc }))
            : [
                {
                  title: 'Pendaftaran Akun',
                  desc: 'Daftarkan akun menggunakan NIM/ID yang valid. Pilih peran "Praktikan" untuk mahasiswa kelas atau "Penyewa" untuk umum.',
                },
                {
                  title: 'Masuk Sistem',
                  desc: 'Login ke portal dengan NIM/username dan password yang telah didaftarkan pada jadwal shift aktif kelas Anda.',
                },
                {
                  title: 'Presensi Kehadiran',
                  desc: 'Lakukan presensi absensi melalui menu kehadiran tepat waktu saat praktikum dimulai di laboratorium.',
                },
                {
                  title: 'Modul & Peminjaman',
                  desc: 'Unduh modul praktikum melalui menu E-Learning atau ajukan sewa inventaris barang lab langsung melalui halaman khusus.',
                },
              ],
          'bg-slate-100 text-slate-600',
        )}
      </div>
    </DialogContent>
  </Dialog>

  <Dialog bind:open={openReschedule}>
    <DialogContent class="max-w-md bg-white">
      <DialogHeader>
        <DialogTitle class="font-black text-slate-800"
          >Langkah Reschedule (Ganti Jadwal)</DialogTitle
        >
        <DialogDescription class="text-xs font-bold tracking-wider text-slate-400 uppercase">
          Prosedur Pengajuan Jadwal Pengganti
        </DialogDescription>
      </DialogHeader>
      <div class="mt-2 space-y-4 text-sm font-medium text-slate-700">
        {#if sysSettings?.reschedule_steps}
          {@render steps(parseLines(sysSettings.reschedule_steps), 'bg-orange-100 text-orange-600')}
        {:else}
          {#each [{ n: 1, t: 'Ajukan Izin' }, { n: 2, t: 'Kirim Bukti WA' }, { n: 3, t: 'Tunggu Verifikasi' }, { n: 4, t: 'Pilih Jadwal Baru' }, { n: 5, t: 'Tunggu ACC' }] as s (s.n)}
            <div class="flex items-start gap-3 {s.n < 5 ? 'border-b border-slate-100 pb-3' : ''}">
              <span
                class="flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-orange-100 text-xs font-black text-orange-600"
                >{s.n}</span
              >
              <div>
                <strong class="block font-bold text-slate-900">{s.t}</strong>
                {#if s.n === 1}
                  Buka menu <strong>Absensi</strong> setelah login, lalu isi form Izin/Sakit di tab
                  <strong>Izin</strong>.
                {:else if s.n === 2}
                  Klik tombol <strong>Kirim Bukti WA</strong> pada riwayat untuk kirim berkas ke Asisten.
                {:else if s.n === 3}
                  Status izin akan berubah jika Asisten sudah memverifikasi.
                {:else if s.n === 4}
                  Klik <strong>Pilih Jadwal</strong> pada riwayat izin untuk memilih slot ganti yang tersedia.
                {:else}
                  Tunggu persetujuan Asisten agar presensi jadwal pengganti dianggap sah.
                {/if}
              </div>
            </div>
          {/each}
        {/if}
        <div
          class="flex gap-2 rounded-xl border border-amber-100 bg-amber-50 p-3 text-xs text-amber-800"
        >
          {@render warnIcon('w-4 h-4 text-amber-500 mt-0.5 shrink-0')}
          <p>Izin dan reschedule wajib diajukan sebelum praktikum kelas pengganti dimulai.</p>
        </div>
      </div>
    </DialogContent>
  </Dialog>
</div>
