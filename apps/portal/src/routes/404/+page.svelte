<script lang="ts">
  import { resolve } from '$app/paths';
  import { Button } from '@algohub/ui';
</script>

<svelte:head>
  <title>Halaman Tidak Ditemukan · Lab Algoritma & Pemrograman</title>
</svelte:head>

<section class="flex flex-1 flex-col items-center justify-center gap-6 px-4 py-16 text-center">
  <p class="text-6xl font-extrabold text-primary">404</p>
  <h1 class="text-2xl font-bold text-ink-heading">Halaman tidak ditemukan</h1>
  <Button href={resolve('/')}>Kembali ke Beranda</Button>
</section>
