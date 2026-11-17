<script lang="ts">
  import {
    Button,
    Card,
    CardContent,
    CardDescription,
    CardHeader,
    CardTitle,
    Label,
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
  } from '@algohub/ui';
  import Loader2 from '@lucide/svelte/icons/loader-2';
  import Play from '@lucide/svelte/icons/play';
  import QrCodeIcon from '@lucide/svelte/icons/qr-code';
  import StopCircle from '@lucide/svelte/icons/circle-stop';
  import QRCode from 'qrcode';
  import { createRpcMutation } from '#lib/queries/rpc.ts';
  import { toast } from '#lib/toast.ts';
  import type { PageProps } from './$types';

  const meetingOptions = [
    'Pengarahan',
    'Modul 1',
    'Modul 2',
    'Modul 3&4',
    'Modul 5',
    'Modul 6',
    'Ujian Praktik',
  ];
  const majorOptions = [
    'S1 Teknik Tenaga Listrik',
    'S1 Teknik Elektro',
    'S1 Teknik Sistem Energi',
    'D3 Teknologi Listrik',
  ];
  const classOptions = ['A', 'B', 'C', 'D', 'E'];
  const ROTATE_MS = 10_000;

  function generateSecureToken(length = 12): string {
    const array = new Uint8Array(length);
    window.crypto.getRandomValues(array);
    const chars = 'abcdefghijklmnopqrstuvwxyz0123456789';
    let token = '';
    for (let i = 0; i < length; i++) token += chars[array[i] % chars.length];
    return token;
  }

  let { data }: PageProps = $props();
  const user = $derived(data.auth.user);

  const upsertSession = createRpcMutation<{ id: string }[]>('upsert_qr_session_secure');
  const rotateToken = createRpcMutation('update_qr_session_token_secure');
  const stopSession = createRpcMutation('stop_qr_session_secure');

  let meetingType = $state('');
  let selectedMajor = $state('');
  let selectedClass = $state('');
  let title = $state('');
  let sessionId = $state<string | null>(null);
  let qrToken = $state<string | null>(null);
  let loading = $state(false);
  let canvas = $state<HTMLCanvasElement | null>(null);

  // Token lama disimpan server sebagai previous_token sehingga scan yang telat tetap sah.
  $effect(() => {
    if (!sessionId) return;
    const id = sessionId;
    const callerId = user.id;
    const intervalId = setInterval(async () => {
      const newToken = generateSecureToken(12);
      const { error } = await rotateToken.mutateAsync({
        p_caller_id: callerId,
        p_id: id,
        p_token: newToken,
      });
      if (!error) qrToken = newToken;
    }, ROTATE_MS);
    return () => clearInterval(intervalId);
  });

  // Setara react-qr-code bawaan: 256 px, level L, tanpa margin.
  $effect(() => {
    if (canvas && qrToken) {
      void QRCode.toCanvas(canvas, qrToken, { width: 256, margin: 0, errorCorrectionLevel: 'L' });
    }
  });

  async function handleStartSession() {
    if (!meetingType || !selectedMajor || !selectedClass) {
      return toast.error('Lengkapi semua pilihan (Pertemuan, Jurusan, & Kelas)!');
    }
    loading = true;
    const generatedTitle = `${meetingType} - ${selectedMajor} ${selectedClass}`;
    title = generatedTitle;
    const firstToken = generateSecureToken(12);
    const { data: sessionData, error } = await upsertSession.mutateAsync({
      p_caller_id: user.id,
      p_title: generatedTitle,
      p_token: firstToken,
    });
    if (error) {
      toast.error('Gagal: ' + error.message);
    } else if (sessionData && sessionData.length > 0) {
      sessionId = sessionData[0].id;
      qrToken = firstToken;
      toast.success(`Sesi Dimulai: ${generatedTitle}`);
    }
    loading = false;
  }

  async function handleStopSession() {
    if (!sessionId) return;
    const { error } = await stopSession.mutateAsync({ p_caller_id: user.id, p_id: sessionId });
    if (error) {
      toast.error('Gagal menghentikan: ' + error.message);
    } else {
      sessionId = null;
      qrToken = null;
      title = '';
      toast.info('Sesi Dihentikan.');
    }
  }
</script>

{#snippet picker(
  label: string,
  value: string,
  set: (v: string) => void,
  placeholder: string,
  options: string[],
)}
  <div class="space-y-2 text-left">
    <Label>{label}</Label>
    <Select type="single" {value} onValueChange={set}>
      <SelectTrigger class="w-full" aria-label={label}>
        {#if value}{value}{:else}<span class="text-muted-foreground">{placeholder}</span>{/if}
      </SelectTrigger>
      <SelectContent>
        {#each options as opt (opt)}
          <SelectItem value={opt} label={opt} />
        {/each}
      </SelectContent>
    </Select>
  </div>
{/snippet}

<div class="mx-auto max-w-xl space-y-6 text-center">
  <h1 class="flex items-center justify-center gap-2 text-2xl font-bold">
    <QrCodeIcon class="text-primary" /> Generator Absensi
  </h1>

  {#if !sessionId}
    <Card>
      <CardHeader>
        <CardTitle>Buat Sesi Baru</CardTitle>
        <CardDescription>Pilih detail kelas di bawah ini.</CardDescription>
      </CardHeader>
      <CardContent class="space-y-4">
        {@render picker(
          'Jenis Pertemuan',
          meetingType,
          (v) => (meetingType = v),
          'Pilih Modul...',
          meetingOptions,
        )}
        {@render picker(
          'Program Studi / Jurusan',
          selectedMajor,
          (v) => (selectedMajor = v),
          'Pilih Jurusan...',
          majorOptions,
        )}
        {@render picker(
          'Kelas Paralel',
          selectedClass,
          (v) => (selectedClass = v),
          'Pilih Kelas (A/B/C...',
          classOptions,
        )}
        <Button onclick={handleStartSession} disabled={loading} class="mt-4 w-full">
          {#if loading}<Loader2 class="animate-spin" />{:else}<Play class="mr-2 h-4 w-4" />{/if}
          Mulai Sesi
        </Button>
      </CardContent>
    </Card>
  {:else}
    <Card class="animate-pulse-border border-4 border-green-500">
      <CardHeader class="pb-2">
        <CardTitle class="animate-pulse text-xl text-green-600">{title}</CardTitle>
      </CardHeader>
      <CardContent class="flex flex-col items-center gap-6 pb-8">
        <div class="relative rounded-xl bg-white p-4 shadow-lg">
          {#if qrToken}<canvas bind:this={canvas} data-testid="qr-canvas"></canvas>{/if}
        </div>
        <Button variant="destructive" size="lg" onclick={handleStopSession} class="w-full">
          <StopCircle class="mr-2 h-5 w-5" /> Hentikan Sesi
        </Button>
      </CardContent>
    </Card>
  {/if}
</div>
