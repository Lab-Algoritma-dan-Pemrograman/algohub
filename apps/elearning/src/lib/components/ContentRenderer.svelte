<script lang="ts">
  import { MarkdownRenderer, RichTextRenderer } from '@algohub/ui';
  import { looksLikeHtml } from '#lib/lesson.ts';

  let { content, class: className = '' }: { content: string; class?: string } = $props();
</script>

{#if looksLikeHtml(content)}
  <RichTextRenderer {content} class={className} />
{:else}
  <MarkdownRenderer {content} class={className} />
{/if}
