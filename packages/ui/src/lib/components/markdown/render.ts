import DOMPurify from 'dompurify';
import { Marked } from 'marked';

const escapeHtml = (text: string): string =>
  text.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;');

const TERMINAL_LANGS = new Set(['terminal', 'output']);

// Blok terminal/output dibungkus jendela terminal seperti MarkdownRenderer lama.
const terminalWindow = (text: string): string =>
  `<div class="bg-zinc-950 rounded-2xl overflow-hidden border border-zinc-800 shadow-lg my-4 not-prose">` +
  `<div class="flex items-center justify-between px-4 py-2.5 bg-zinc-900 border-b border-zinc-800">` +
  `<div class="flex items-center gap-2">` +
  `<span class="w-3 h-3 rounded-full bg-red-500"></span>` +
  `<span class="w-3 h-3 rounded-full bg-amber-400"></span>` +
  `<span class="w-3 h-3 rounded-full bg-emerald-400"></span></div>` +
  `<span class="text-zinc-500 text-[10px] font-mono tracking-widest font-bold">OUTPUT</span></div>` +
  `<div class="p-4 font-mono text-xs text-zinc-100 leading-relaxed min-h-[60px] whitespace-pre-wrap">${escapeHtml(text.replace(/\n$/, ''))}</div></div>`;

const marked = new Marked({
  gfm: true,
  renderer: {
    code({ text, lang }) {
      const language = (lang ?? '').match(/\S+/)?.[0] ?? '';
      if (TERMINAL_LANGS.has(language)) return terminalWindow(text);
      const cls = language ? ` class="language-${escapeHtml(language)}"` : '';
      return `<pre><code${cls}>${escapeHtml(text.replace(/\n$/, ''))}\n</code></pre>\n`;
    },
  },
});

// Materi lama boleh memuat HTML mentah (rehype-raw), jadi hasil marked wajib disanitasi.
export function renderMarkdown(source: string): string {
  if (!source) return '';
  return DOMPurify.sanitize(marked.parse(source, { async: false }));
}
