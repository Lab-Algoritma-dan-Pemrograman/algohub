// @vitest-environment jsdom
import { render } from '@testing-library/svelte';
import { describe, expect, it } from 'vitest';
import RichTextRenderer from './RichTextRenderer.svelte';

describe('RichTextRenderer', () => {
  it('mengembalikan tag ter-escape dan merender blok kode', () => {
    const { container } = render(RichTextRenderer, {
      content:
        '&lt;p&gt;Halo&lt;/p&gt;&lt;pre&gt;&lt;code&gt;printf("x");&lt;/code&gt;&lt;/pre&gt;',
    });
    expect(container.querySelector('p')?.textContent).toBe('Halo');
    expect(container.querySelector('pre code')?.textContent).toBe('printf("x");');
  });

  it('membuang script, handler event, dan URL javascript:', () => {
    const { container } = render(RichTextRenderer, {
      content:
        '<p>aman</p><script>window.hacked = 1</script><a href="javascript:alert(1)" onclick="x()">link</a><iframe src="https://evil.test"></iframe>',
    });
    expect(container.querySelector('script')).toBeNull();
    expect(container.querySelector('iframe')).toBeNull();
    const a = container.querySelector('a');
    expect(a?.getAttribute('onclick')).toBeNull();
    expect(a?.getAttribute('href') ?? '').not.toMatch(/javascript:/i);
    expect(container.textContent).toContain('aman');
  });

  it('merender rumus LaTeX dengan KaTeX', () => {
    const { container } = render(RichTextRenderer, { content: '<p>Nilai $x^2$ ya</p>' });
    expect(container.querySelector('.katex')).not.toBeNull();
  });
});
