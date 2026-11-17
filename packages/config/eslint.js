import js from '@eslint/js';
import { defineConfig, globalIgnores } from 'eslint/config';
import svelte from 'eslint-plugin-svelte';
import globals from 'globals';
import tseslint from 'typescript-eslint';
export default defineConfig(
  globalIgnores([
    '**/node_modules/',
    '.worktrees/',
    '.agents/',
    '**/build/',
    '**/dist/',
    '**/.svelte-kit/',
    '**/.vercel/',
    '**/test-results/',
    '**/playwright-report/',
    'backend/',
    'packages/runner/static/vendor/',
  ]),
  js.configs.recommended,
  tseslint.configs.recommended,
  svelte.configs.recommended,
  {
    languageOptions: {
      globals: { ...globals.browser, ...globals.node },
    },
    rules: {
      'no-console': ['error', { allow: ['error'] }],
      'no-debugger': 'error',
      eqeqeq: ['error', 'always'],
      '@typescript-eslint/no-unused-vars': [
        'error',
        { argsIgnorePattern: '^_', varsIgnorePattern: '^_' },
      ],
    },
  },
  {
    files: ['**/*.svelte', '**/*.svelte.ts'],
    languageOptions: {
      parserOptions: { parser: tseslint.parser },
    },
  },
  {
    files: ['packages/runner/static/**/*.js'],
    languageOptions: {
      sourceType: 'script',
      globals: { ...globals.worker },
    },
    // Aturan ini crash (scope.upper null) pada file sourceType script.
    rules: { 'svelte/no-inner-declarations': 'off' },
  },
  {
    // App praktikum dilayani di root domain tanpa paths.base, jadi resolve() tidak mengubah URL;
    // href dan goto dibiarkan sama dengan app lama.
    files: ['apps/praktikum/**'],
    rules: { 'svelte/no-navigation-without-resolve': 'off' },
  },
  {
    // Kode edra di-vendor dari CLI edra; dibiarkan apa adanya agar mudah dibandingkan dengan upstream.
    files: ['packages/ui/src/lib/components/rich-text/edra/**'],
    rules: { eqeqeq: 'off', 'no-useless-assignment': 'off' },
  },
);
