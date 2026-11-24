---
title: Deployment
description: Prosedur deploy backend Go, frontend SvelteKit, dan dokumentasi ke Vercel, termasuk verifikasi pasca-deploy dan rollback.
---

# Deployment

Seluruh layanan produksi berjalan di Vercel dengan satu repositori monorepo. Setiap project memakai Root Directory yang berbeda, jadi perintah deploy harus dijalankan dari root repo.

## 1. Pemetaan Project

| Project Vercel | Root Directory | Alias produksi |
|---|---|---|
| `algohub-api` | `backend` | `api.algohub.web.id` |
| `algohub-praktikum` | `apps/praktikum` | `praktikum.algohub.web.id` |
| `algohub-siakad` | `apps/siakad` | `siakad.algohub.web.id` |
| `algohub-elearning` | `apps/elearning` | `elearning.algohub.web.id` |
| `algohub-portal` | `apps/portal` | `algohub.web.id` |
| `algohub-docs` | `apps/docs` | `docs.algohub.web.id` |

Scope CLI: `lab-algoritma-pemrograman`.

## 2. Deploy Backend

Jalankan dari root repo, **bukan** dari `backend/`:

```bash
vercel deploy --prod --project algohub-api --scope lab-algoritma-pemrograman --yes --logs
```

Project `algohub-api` sudah menetapkan Root Directory `backend`, sehingga menjalankan perintah di dalam `backend/` akan menggandakan path dan gagal.

### Environment wajib produksi
- `DATABASE_URL_PRAKTIKUM`, `DATABASE_URL_SIAKAD`, `DATABASE_URL_ELEARNING`
- `JWT_SECRET`, `SIAKAD_JWT_SECRET`, `AUTH_MODE=dual`
- `SUPABASE_URL`, `SUPABASE_SERVICE_KEY`, `SUPABASE_BUCKET`, `SUPABASE_JWKS_URL`
- `PUBLIC_BLOB_READ_WRITE_TOKEN`, `REPORT_BLOB_READ_WRITE_TOKEN`
- `CRON_SECRET`, `CORS_ORIGINS`

Token Blob hanya boleh berada di environment server; jangan pernah memasukkannya ke bundle frontend.

## 3. Deploy Frontend & Docs

```bash
vercel deploy --prod --project algohub-praktikum --scope lab-algoritma-pemrograman --yes --logs
vercel deploy --prod --project algohub-siakad --scope lab-algoritma-pemrograman --yes --logs
vercel deploy --prod --project algohub-elearning --scope lab-algoritma-pemrograman --yes --logs
vercel deploy --prod --project algohub-portal --scope lab-algoritma-pemrograman --yes --logs
vercel deploy --prod --project algohub-docs --scope lab-algoritma-pemrograman --yes --logs
```

`algohub-docs` juga terhubung ke Git (auto-deploy saat `main` berubah). Deploy manual hanya diperlukan bila ingin mempublikasikan perubahan sebelum push.

## 4. Verifikasi Pasca-Deploy

```bash
curl -s https://api.algohub.web.id/api/health
curl -s -o /dev/null -w "%{http_code}\n" https://praktikum.algohub.web.id/
curl -s -o /dev/null -w "%{http_code}\n" https://siakad.algohub.web.id/
```

Yang harus dicek:
1. Health API mengembalikan `{"service":"algohub-api","status":"ok"}`.
2. Halaman utama tiap app mengembalikan `200`.
3. Header CSP memuat `*.public.blob.vercel-storage.com` (praktikum/siakad).
4. Upload berkas kecil berhasil dan URL hasilnya bisa dibuka.
5. Untuk perubahan yang menyentuh data ujian, jalankan uji tulis `BEGIN/UPDATE/ROLLBACK` (lihat halaman Database & Auth).

## 5. Rollback

Deployment Vercel bersifat immutable. Buka dashboard project, pilih deployment terakhir yang sehat, lalu **Promote to Production**. Alternatif CLI: `vercel rollback` dari root project. Setelah rollback, ulangi verifikasi di atas.

## 6. Urutan Rilis

- Perubahan skema database harus **additive** dan di-deploy sebelum kode yang membacanya.
- Deploy backend lebih dulu, lalu frontend.
- Jangan menjalankan `AutoMigrate` di produksi.
