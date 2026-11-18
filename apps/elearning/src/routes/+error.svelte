<script lang="ts">
  import { resolve } from '$app/paths';
  import { page } from '$app/state';
</script>

<div class="flex min-h-screen items-center justify-center">
  <div class="text-center">
    <h1 class="text-maroon mb-4 text-5xl font-black">{page.status}</h1>
    <p class="mb-4 text-xl font-bold text-zinc-500">
      {page.status === 404 ? 'Halaman tidak ditemukan' : 'Terjadi kesalahan'}
    </p>
    <a href={resolve('/')} data-sveltekit-reload class="text-maroon font-bold underline"
      >Kembali ke beranda</a
    >
  </div>
</div>
