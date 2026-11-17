<script lang="ts">
  import {
    Button,
    Card,
    CardContent,
    CardDescription,
    CardHeader,
    CardTitle,
    Input,
    Label,
  } from '@algohub/ui';
  import Eye from '@lucide/svelte/icons/eye';
  import EyeOff from '@lucide/svelte/icons/eye-off';
  import Loader2 from '@lucide/svelte/icons/loader-2';
  import Save from '@lucide/svelte/icons/save';
  import UserCircle from '@lucide/svelte/icons/circle-user';
  import { updateSessionUser } from '#lib/auth.ts';
  import { createRpcMutation } from '#lib/queries/rpc.ts';
  import { toast } from '#lib/toast.ts';
  import type { PageProps } from './$types';

  let { data }: PageProps = $props();
  const currentUser = $derived(data.auth.user);

  const updateProfile = createRpcMutation('update_user_profile_secure', [
    'get_dashboard_stats_secure',
  ]);
  const updatePassword = createRpcMutation('update_password');

  let loading = $state(false);
  let showPassword = $state(false);
  let fullName = $state('');
  let username = $state('');
  let password = $state('');

  $effect(() => {
    fullName = currentUser.full_name;
    username = currentUser.username;
  });

  async function handleUpdate(e: SubmitEvent) {
    e.preventDefault();
    loading = true;
    try {
      // p_caller_id tetap dikirim walau gateway menimpanya dari JWT.
      const { data: result, error } = await updateProfile.mutateAsync({
        p_caller_id: currentUser.id,
        p_full_name: fullName,
      });
      if (error) throw error;
      if (password && password.length > 0) {
        if (password.length < 3) throw new Error('Password minimal 3 karakter');
        const { error: pwdError } = await updatePassword.mutateAsync({
          p_caller_id: currentUser.id,
          p_target_id: currentUser.id,
          p_new_password: password,
        });
        if (pwdError) throw pwdError;
      }
      const updatedData = Array.isArray(result) ? result[0] : result;
      if (updatedData && typeof updatedData === 'object') await updateSessionUser(updatedData);
      toast.success('Profil berhasil diperbarui!');
      password = '';
    } catch (error) {
      toast.error('Gagal update profil: ' + (error as { message?: string }).message);
    } finally {
      loading = false;
    }
  }
</script>

<div class="mx-auto max-w-2xl space-y-6">
  <div>
    <h1 class="text-2xl font-bold">Profil Saya</h1>
    <p class="text-muted-foreground">Kelola informasi akun Anda.</p>
  </div>
  <Card>
    <CardHeader>
      <div class="flex items-center gap-4">
        <div
          class="bg-primary/10 text-primary flex h-16 w-16 items-center justify-center rounded-full"
        >
          <UserCircle class="h-10 w-10" />
        </div>
        <div>
          <CardTitle>{currentUser.full_name}</CardTitle>
          <CardDescription class="capitalize">
            Role: <span class="text-primary font-bold">{currentUser.role}</span>
            {currentUser.nim ? ` | NIM: ${currentUser.nim}` : ''}
          </CardDescription>
        </div>
      </div>
    </CardHeader>
    <CardContent>
      <form onsubmit={handleUpdate} class="space-y-4">
        <div class="space-y-2">
          <Label for="profil-username">Username / NIM</Label>
          <Input id="profil-username" value={username} disabled class="bg-muted" />
          <p class="text-muted-foreground text-[10px]">Username tidak dapat diubah.</p>
        </div>
        <div class="space-y-2">
          <Label for="profil-name">Nama Lengkap</Label>
          <Input id="profil-name" bind:value={fullName} placeholder="Nama Lengkap Anda" />
        </div>
        <div class="space-y-2">
          <Label for="profil-password">Password Baru</Label>
          <div class="relative">
            <Input
              id="profil-password"
              type={showPassword ? 'text' : 'password'}
              bind:value={password}
              placeholder="Kosongkan jika tidak ingin mengganti password"
            />
            <button
              type="button"
              aria-label="Tampilkan password"
              onclick={() => (showPassword = !showPassword)}
              class="text-muted-foreground hover:text-foreground absolute top-2.5 right-3"
            >
              {#if showPassword}
                <EyeOff class="h-4 w-4" />
              {:else}
                <Eye class="h-4 w-4" />
              {/if}
            </button>
          </div>
        </div>
        <div class="flex justify-end pt-4">
          <Button type="submit" disabled={loading}>
            {#if loading}<Loader2 class="mr-2 h-4 w-4 animate-spin" />{/if}
            <Save class="mr-2 h-4 w-4" />
            Simpan Perubahan
          </Button>
        </div>
      </form>
    </CardContent>
  </Card>
</div>
