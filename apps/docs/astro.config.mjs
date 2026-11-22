import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';

export default defineConfig({
  site: 'https://docs.algohub.web.id',
  integrations: [
    starlight({
      title: 'AlgoHub',
      description: 'Dokumentasi Resmi Lab Algoritma dan Pemrograman',
      defaultLocale: 'root',
      locales: {
        root: { label: 'Bahasa Indonesia', lang: 'id' },
        en: { label: 'English', lang: 'en' },
      },
      sidebar: [
        {
          label: 'Pengenalan',
          translations: { en: 'Introduction' },
          items: [
            {
              label: 'Pengenalan',
              translations: { en: 'Introduction' },
              link: '/introduction',
            },
          ],
        },
        {
          label: 'Panduan Aplikasi',
          translations: { en: 'Application Guides' },
          items: [
            {
              label: 'Portal',
              translations: { en: 'Portal' },
              link: '/apps/portal',
            },
            {
              label: 'SIAKAD',
              translations: { en: 'SIAKAD' },
              link: '/apps/siakad',
            },
            {
              label: 'E-Learning',
              translations: { en: 'E-Learning' },
              link: '/apps/elearning',
            },
            {
              label: 'Praktikum',
              translations: { en: 'Praktikum' },
              link: '/apps/praktikum',
            },
          ],
        },
        {
          label: 'Panduan Pengguna',
          translations: { en: 'User Guides' },
          items: [
            {
              label: 'Mahasiswa',
              translations: { en: 'Students' },
              link: '/guides/mahasiswa',
            },
            {
              label: 'Asisten & Koordinator',
              translations: { en: 'Assistants & Coordinators' },
              link: '/guides/asisten',
            },
          ],
        },
        {
          label: 'Arsitektur & Sistem',
          translations: { en: 'Architecture & System' },
          items: [
            {
              label: 'Monorepo',
              translations: { en: 'Monorepo' },
              link: '/architecture/monorepo',
            },
            {
              label: 'Database & Auth',
              translations: { en: 'Database & Auth' },
              link: '/architecture/database-auth',
            },
            {
              label: 'Deployment',
              translations: { en: 'Deployment' },
              link: '/architecture/deployment',
            },
          ],
        },
        {
          label: 'API Reference',
          translations: { en: 'API Reference' },
          items: [
            {
              label: 'Overview',
              translations: { en: 'Overview' },
              link: '/api/overview',
            },
            {
              label: 'Umum & CORS',
              translations: { en: 'General & CORS' },
              link: '/api/general',
            },
            {
              label: 'Auth & Cron',
              translations: { en: 'Auth & Cron' },
              link: '/api/cron-auth',
            },
            {
              label: 'Praktikum',
              translations: { en: 'Praktikum' },
              link: '/api/praktikum',
            },
            {
              label: 'E-Learning',
              translations: { en: 'E-Learning' },
              link: '/api/elearning',
            },
            {
              label: 'SIAKAD',
              translations: { en: 'SIAKAD' },
              link: '/api/siakad',
            },
          ],
        },
        {
          label: 'Riwayat & Pembaruan',
          translations: { en: 'History & Updates' },
          items: [
            {
              label: 'Changelog',
              translations: { en: 'Changelog' },
              link: '/changelog',
            },
            {
              label: 'Index Updates',
              translations: { en: 'Updates Index' },
              link: '/updates/index',
            },
            {
              label: 'E-Learning RLS',
              translations: { en: 'E-Learning RLS' },
              link: '/updates/elearning-rls',
            },
            {
              label: 'Temuan Audit',
              translations: { en: 'Audit Findings' },
              link: '/updates/temuan',
            },
          ],
        },
      ],
    }),
  ],
});
