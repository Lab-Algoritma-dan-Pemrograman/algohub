<script lang="ts">
  import Download from '@lucide/svelte/icons/download';
  import ExternalLink from '@lucide/svelte/icons/external-link';
  import ImageIcon from '@lucide/svelte/icons/image';

  let { images, class: className = '' }: { images: string[]; class?: string } = $props();

  function download(src: string, idx: number) {
    const link = document.createElement('a');
    link.href = src;
    link.download = `matplotlib_plot_${idx + 1}.png`;
    link.click();
  }

  // Data URL tidak bisa dibuka sebagai navigasi tingkat atas, jadi gambar disisipkan ke tab kosong.
  function openInNewTab(src: string) {
    const win = window.open();
    if (!win) return;
    const img = win.document.createElement('img');
    img.src = src;
    img.style.cssText = 'max-width:100%;height:auto;';
    win.document.body.append(img);
  }
</script>

{#if images.length > 0}
  <div class="space-y-4 {className}">
    {#each images as src, idx (idx)}
      <div
        class="overflow-hidden rounded-2xl border-2 border-zinc-200 bg-white p-4 shadow-sm transition-all hover:shadow-md"
      >
        <div class="mb-3 flex items-center justify-between border-b border-zinc-100 pb-3">
          <div class="flex items-center gap-2 text-xs font-bold text-zinc-700">
            <ImageIcon class="h-4 w-4 text-rose-700" />
            <span>Hasil Grafik Matplotlib {images.length > 1 ? `#${idx + 1}` : ''}</span>
          </div>
          <div class="flex items-center gap-2">
            <button
              type="button"
              onclick={() => openInNewTab(src)}
              class="flex items-center gap-1 rounded-lg bg-zinc-100 px-2.5 py-1 text-[11px] font-bold text-zinc-500 transition-colors hover:bg-zinc-200 hover:text-zinc-800"
              title="Buka Gambar di Tab Baru"
            >
              <ExternalLink class="h-3 w-3" />
              <span>Buka</span>
            </button>
            <button
              type="button"
              onclick={() => download(src, idx)}
              class="flex items-center gap-1 rounded-lg bg-rose-700 px-2.5 py-1 text-[11px] font-bold text-white shadow-sm transition-colors hover:bg-rose-800"
              title="Unduh PNG"
            >
              <Download class="h-3 w-3" />
              <span>Unduh PNG</span>
            </button>
          </div>
        </div>
        <div
          class="flex justify-center overflow-x-auto rounded-xl border border-zinc-100 bg-zinc-50 p-2"
        >
          <img
            {src}
            alt={`Hasil Grafik Matplotlib ${idx + 1}`}
            class="h-auto max-h-[450px] max-w-full rounded-lg object-contain"
          />
        </div>
      </div>
    {/each}
  </div>
{/if}
