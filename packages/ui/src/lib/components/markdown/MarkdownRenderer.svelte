<script lang="ts">
  import { cn } from '../../utils.js';
  import { renderMarkdown } from './render.js';

  let { content, class: className = '' }: { content: string; class?: string } = $props();

  const html = $derived(renderMarkdown(content));
</script>

<div
  class={cn(
    'prose prose-zinc max-w-none prose-headings:font-bold prose-h1:text-3xl prose-h2:text-2xl prose-p:text-zinc-600 prose-code:rounded prose-code:bg-zinc-100 prose-code:px-1 prose-pre:bg-zinc-900 prose-pre:text-zinc-100',
    className,
  )}
>
  <!-- eslint-disable-next-line svelte/no-at-html-tags -- html disanitasi DOMPurify di renderMarkdown -->
  {@html html}
</div>
