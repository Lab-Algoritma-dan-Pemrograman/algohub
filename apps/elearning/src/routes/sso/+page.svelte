<script lang="ts">
  import { goto } from '$app/navigation';
  import { resolve } from '$app/paths';
  import { page } from '$app/state';
  import Loader2 from '@lucide/svelte/icons/loader-circle';
  import { onMount } from 'svelte';
  import { exchangeSsoCode } from '#lib/auth.ts';
  import { authErrorMessage } from '#lib/auth-messages.ts';

  onMount(async () => {
    const code = page.url.searchParams.get('code');
    if (!code) {
      await goto(resolve('/login'), {
        replaceState: true,
        state: { ssoError: 'Kode masuk tidak ditemukan. Silakan masuk lagi dari portal.' },
      });
      return;
    }
    try {
      await exchangeSsoCode(code);
      await goto(resolve('/(app)/dashboard'), { replaceState: true });
    } catch (e) {
      await goto(resolve('/login'), {
        replaceState: true,
        state: { ssoError: authErrorMessage(e, 'Gagal masuk lewat portal. Silakan coba lagi.') },
      });
    }
  });
</script>

<div class="flex min-h-screen flex-col items-center justify-center gap-4 text-zinc-500">
  <Loader2 size={32} class="text-maroon animate-spin" />
  <p class="text-sm font-bold">Memeriksa akun…</p>
</div>
