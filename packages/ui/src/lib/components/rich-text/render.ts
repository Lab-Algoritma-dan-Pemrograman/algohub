import DOMPurify from 'dompurify';

// Isi lama tersimpan sebagai HTML yang ter-escape (&lt;div&gt;), jadi tag yang diizinkan dikembalikan dulu.
export function cleanContent(html: string): string {
  if (!html) return '';

  let cleaned = html
    .replace(/&amp;lt;/gi, '&lt;')
    .replace(/&amp;gt;/gi, '&gt;')
    .replace(/&amp;quot;/gi, '&quot;')
    .replace(/&amp;amp;/gi, '&amp;');

  cleaned = cleaned.replace(/&quot;/g, '"').replace(/&#39;/g, "'");

  const tagRegex =
    /&lt;(\/?(div|pre|code|span|p|br|h1|h2|h3|h4|h5|h6|ul|ol|li|strong|em|table|thead|tbody|tr|td|th|a|img|blockquote|svg|path|hr)[^>]*?)&gt;/gi;
  cleaned = cleaned.replace(tagRegex, '<$1>');

  cleaned = cleaned.replace(/\s+on[a-zA-Z]+\s*=\s*(?:'[^']*'|"[^"]*"|[^\s>]+)/gi, '');
  cleaned = cleaned.replace(
    /\s+(href|src|action)\s*=\s*(?:'[^']*(?:javascript|data|vbscript):[^']*'|"[^"]*(?:javascript|data|vbscript):[^"]*"|[^\s>]+(?:javascript|data|vbscript):[^\s>]+)/gi,
    ' $1="#"',
  );
  cleaned = cleaned.replace(/<script\b[^<]*(?:(?!<\/script>)<[^<]*)*<\/script>/gi, '');

  cleaned = cleaned
    .replace(/Output Terminal\s*\(Mac\):/gi, 'Output Terminal:')
    .replace(/macbook-pro\s*—\s*~user\/workspace/gi, 'Terminal')
    .replace(/\s*<span class="text-zinc-500">\s*\$\s*[^<]*<\/span>\s*/gi, '');

  const preRegex =
    /<pre class="bg-zinc-950 text-zinc-100 p-4 rounded-xl font-mono text-xs overflow-x-auto border border-zinc-800">\s*<code>([\s\S]*?)<\/code>\s*<\/pre>/gi;
  cleaned = cleaned.replace(preRegex, (_match, codeContent: string) =>
    `
      <div class="bg-zinc-950 p-4 rounded-2xl font-mono text-xs shadow-xl border border-zinc-800 max-w-xl my-4">
        <div class="flex items-center gap-1.5 mb-2.5 border-b border-zinc-800 pb-2 text-zinc-500">
          <span class="w-2.5 h-2.5 rounded-full bg-[#ff5f56]"></span>
          <span class="w-2.5 h-2.5 rounded-full bg-[#ffbd2e]"></span>
          <span class="w-2.5 h-2.5 rounded-full bg-[#27c93f]"></span>
          <span class="ml-2 text-[10px] font-bold text-zinc-400">Source Code</span>
        </div>
        <pre class="overflow-x-auto bg-transparent p-0 border-0 text-zinc-100 font-mono text-xs"><code>${codeContent}</code></pre>
      </div>
    `.trim(),
  );

  return cleaned;
}

// Regex di cleanContent hanya menormalkan tag lama; DOMPurify dengan allowlist adalah sanitasi yang menentukan.
export function sanitizeHtml(html: string): string {
  if (!html) return '';
  return DOMPurify.sanitize(html, {
    ALLOWED_TAGS: [
      'div',
      'pre',
      'code',
      'span',
      'p',
      'br',
      'h1',
      'h2',
      'h3',
      'h4',
      'h5',
      'h6',
      'ul',
      'ol',
      'li',
      'strong',
      'em',
      'table',
      'thead',
      'tbody',
      'tr',
      'td',
      'th',
      'a',
      'img',
      'blockquote',
      'svg',
      'path',
      'hr',
      'sub',
      'sup',
    ],
    ALLOWED_ATTR: [
      'class',
      'href',
      'src',
      'alt',
      'title',
      'target',
      'rel',
      'width',
      'height',
      'viewBox',
      'd',
      'fill',
      'stroke',
      'stroke-width',
      'colspan',
      'rowspan',
      'style',
    ],
    ALLOW_DATA_ATTR: false,
    FORBID_TAGS: ['script', 'style', 'iframe', 'object', 'embed', 'form', 'input'],
  });
}

export const renderRichText = (content: string): string => sanitizeHtml(cleanContent(content));
