import { svelte } from '@sveltejs/vite-plugin-svelte';
import { svelteTesting } from '@testing-library/svelte/vite';
import { defineConfig } from 'vitest/config';

export default defineConfig({
  plugins: [svelte(), svelteTesting()],
  test: {
    include: ['src/**/*.test.ts'],
    // Test komponen memakai jsdom lewat anotasi `@vitest-environment` per file.
    environment: 'node',
    server: { deps: { inline: [/katex/] } },
  },
});
