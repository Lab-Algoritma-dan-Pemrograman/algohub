import { svelte } from '@sveltejs/vite-plugin-svelte';
import { defineConfig } from 'vitest/config';
export default defineConfig({
  // @tanstack/svelte-query mengekspor komponen .svelte, jadi perlu dikompilasi saat test.
  plugins: [svelte()],
  test: {
    include: ['src/**/*.test.ts'],
    environment: 'node',
  },
});
