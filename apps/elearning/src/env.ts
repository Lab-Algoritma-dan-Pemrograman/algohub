import process from 'node:process';
import { defineEnvVars } from '@sveltejs/kit/env';
// Default diisi agar dev lokal dan build CI tanpa variabel tetap jalan.
const withDefault = (fallback: string) => (value: string | undefined) => value || fallback;
export const variables = defineEnvVars({
  PUBLIC_API_BASE_URL: {
    public: true,
    static: true,
    description: 'Origin backend Go (tanpa /api)',
    schema: withDefault('http://localhost:8080'),
  },
  PUBLIC_COMPILER_BASE_URL: {
    public: true,
    static: true,
    description: 'Folder berisi clang, lld, memfs, dan sysroot.tar',
    schema: withDefault('/compiler/v1/'),
  },
  PUBLIC_APP_VERSION: {
    public: true,
    static: true,
    description: 'Buster cache Query yang dipersist; berganti tiap deploy',
    schema: withDefault(process.env.VERCEL_GIT_COMMIT_SHA || 'dev'),
  },
  PUBLIC_SUPABASE_URL: {
    public: true,
    static: true,
    description: 'URL project Supabase elearning, hanya untuk Realtime; kosong = Realtime mati',
    schema: withDefault(''),
  },
  PUBLIC_SUPABASE_PUBLISHABLE_KEY: {
    public: true,
    static: true,
    description: 'Anon/publishable key Supabase elearning untuk Realtime',
    schema: withDefault(''),
  },
});
