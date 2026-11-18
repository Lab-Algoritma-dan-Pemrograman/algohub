<script lang="ts">
  import CircleAlert from '@lucide/svelte/icons/circle-alert';
  let { error }: { error: unknown } = $props();
</script>

<div class="flex min-h-screen flex-col items-center justify-center bg-zinc-50 p-6 text-center">
  <div class="mb-6 flex h-16 w-16 items-center justify-center rounded-2xl bg-red-100 text-red-600">
    <CircleAlert size={32} />
  </div>
  <h1 class="mb-2 text-2xl font-bold text-zinc-900">Ups! Terjadi kesalahan</h1>
  <p class="mb-8 max-w-md text-zinc-500">
    Aplikasi mengalami kendala teknis. Silakan coba muat ulang halaman.
  </p>
  <div
    class="mb-8 w-full max-w-lg overflow-auto rounded-xl border border-zinc-200 bg-white p-4 text-left"
  >
    <p class="font-mono text-xs text-red-500">{String(error)}</p>
  </div>
  <button
    onclick={() => window.location.reload()}
    class="cursor-pointer rounded-xl bg-rose-700 px-8 py-3 font-bold text-white transition-all hover:bg-rose-800"
  >
    Muat Ulang Halaman
  </button>
</div>
