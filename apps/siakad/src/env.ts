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
  PUBLIC_APP_VERSION: {
    public: true,
    static: true,
    description: 'Buster cache Query yang dipersist; berganti tiap deploy',
    schema: withDefault(process.env.VERCEL_GIT_COMMIT_SHA || 'dev'),
  },
  PUBLIC_SUPABASE_URL: {
    public: true,
    static: true,
    description: 'URL project Supabase siakad, hanya untuk Realtime; kosong = Realtime mati',
    schema: withDefault(''),
  },
  PUBLIC_SUPABASE_PUBLISHABLE_KEY: {
    public: true,
    static: true,
    description: 'Anon/publishable key Supabase siakad untuk Realtime',
    schema: withDefault(''),
  },
  PUBLIC_VAPID_PUBLIC_KEY: {
    public: true,
    static: true,
    // MEDIUM-04: public key hasil rotasi 2026-09-25; aman di-bundle, private key hanya di backend.
    description: 'VAPID public key untuk Web Push',
    schema: withDefault(
      'BKAXOmOt7QXutFWQp9GEPf17gU0BkTAd_xzAeQtIWONnnnUo7XUI3m4hunhgqaY1bbuKMKw3GUq44jD5Xd2Dpc0',
    ),
  },
  PUBLIC_ELEARNING_URL: {
    public: true,
    static: true,
    description: 'URL app elearning tujuan SSO; kosong atau *.vercel.app memakai domain resmi',
    schema: withDefault(''),
  },
});
