import js from '@eslint/js';
import { defineConfig, globalIgnores } from 'eslint/config';
import svelte from 'eslint-plugin-svelte';
import globals from 'globals';
import tseslint from 'typescript-eslint';
export default defineConfig(
  globalIgnores([
    '**/node_modules/',
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
  },
);
