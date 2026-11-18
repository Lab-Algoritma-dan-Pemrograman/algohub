<script lang="ts">
  // @ts-expect-error katex tidak menyertakan tipe untuk contrib/auto-render.
  import renderMathInElement from 'katex/contrib/auto-render';
  import 'katex/dist/katex.min.css';
  import { cn } from '../../utils.js';
  import { renderRichText } from './render.js';

  let { content, class: className = '' }: { content: string; class?: string } = $props();

  let container = $state<HTMLDivElement>();
  const html = $derived(renderRichText(content));

  $effect(() => {
    // Baca html agar rumus dirender ulang setiap konten berubah.
    void html;
    if (!container) return;
    renderMathInElement(container, {
      delimiters: [
        { left: '$$', right: '$$', display: true },
        { left: '$', right: '$', display: false },
        { left: '\\(', right: '\\)', display: false },
        { left: '\\[', right: '\\]', display: true },
      ],
      // Rumus yang salah ditulis dosen tampil sebagai teks, bukan membuat halaman error.
      throwOnError: false,
    });
  });
</script>

<div
  bind:this={container}
  class={cn(
    'prose prose-zinc max-w-none prose-headings:font-bold prose-h1:text-3xl prose-h2:text-2xl prose-p:leading-relaxed prose-p:text-zinc-700 prose-code:rounded prose-code:bg-zinc-100 prose-code:px-1.5 prose-code:py-0.5 prose-code:font-mono prose-code:text-sm prose-code:text-rose-700 prose-pre:rounded-xl prose-pre:bg-zinc-900 prose-pre:p-4 prose-pre:text-zinc-100 prose-table:border-collapse prose-th:border prose-th:border-zinc-200 prose-th:bg-zinc-50 prose-th:p-2 prose-td:border prose-td:border-zinc-200 prose-td:p-2',
    className,
  )}
>
  <!-- eslint-disable-next-line svelte/no-at-html-tags -- html disanitasi DOMPurify di renderRichText -->
  {@html html}
</div>
