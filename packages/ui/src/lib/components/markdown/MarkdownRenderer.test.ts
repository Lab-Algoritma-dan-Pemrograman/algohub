// @vitest-environment jsdom
import { render } from '@testing-library/svelte';
import { describe, expect, it } from 'vitest';
import MarkdownRenderer from './MarkdownRenderer.svelte';

describe('MarkdownRenderer', () => {
  it('merender heading dan blok kode', () => {
    const { container } = render(MarkdownRenderer, {
      content: '# Judul\n\n```c\nint main() {}\n```',
    });
    expect(container.querySelector('h1')?.textContent).toBe('Judul');
    expect(container.querySelector('pre code.language-c')?.textContent).toContain('int main()');
  });

  it('membungkus blok output seperti jendela terminal', () => {
    const { container } = render(MarkdownRenderer, { content: '```output\nHalo <dunia>\n```' });
    expect(container.querySelector('.not-prose')).not.toBeNull();
    expect(container.textContent).toContain('OUTPUT');
    expect(container.textContent).toContain('Halo <dunia>');
    expect(container.querySelector('dunia')).toBeNull();
  });

  it('membuang script, handler event, dan URL javascript:', () => {
    const { container } = render(MarkdownRenderer, {
      content:
        '<script>window.hacked = 1</script>\n\n<img src="x" onerror="window.hacked = 1">\n\n[klik](javascript:alert(1))',
    });
    expect(container.querySelector('script')).toBeNull();
    expect(container.querySelector('img')?.getAttribute('onerror')).toBeNull();
    const href = container.querySelector('a')?.getAttribute('href') ?? '';
    expect(href).not.toMatch(/javascript:/i);
  });
});
