# Temuan

Bug dan kejanggalan dari kode lama yang **sengaja tidak diperbaiki** selama porting, karena aturan Update 1 adalah perilaku harus sama dengan prod. Setiap temuan diberi nomor supaya bisa dirujuk dari `TODO(T-n)`.

| No | Asal | Temuan | Dampak | Usulan |
|---|---|---|---|---|
| T-1 | praktikum `database/connection.go` | Komentar menyebut koneksi lazy tanpa Ping, padahal `gorm.Open` melakukan ping otomatis | menyesatkan, tidak ada bug perilaku | komentar sudah tidak dibawa |
| T-2 | praktikum `config` | `ONLINE_TTL_SECONDS` dibaca ke `Config.OnlineTTL` tapi tidak dipakai | env mati | hapus atau pakai di update berikutnya |
| T-3 | praktikum cron handler | `X-Cron-Secret` dibandingkan dengan `!=`, bukan constant-time | risiko timing attack kecil | `subtle.ConstantTimeCompare` |
| T-4 | praktikum `cmd/migrate` | Masih memakai `AutoMigrate` untuk `schema_migrations` dan membaca `database/migration`, yang tidak dibawa | tool tidak bisa dipakai | tulis ulang kalau migrasi dibutuhkan |
| T-5 | praktikum swag | Anotasi preview/download laporan besar menunjuk `dto.BerkasURLResponse` yang tidak ada | dokumentasi salah | sudah diganti `map[string]string` |
| T-6 | praktikum `pkg/ollama/client.go` | `RawChat` dan `chat` hampir identik | duplikasi | satukan |
| T-7 | backend baru | Rantai middleware alias hanya diuji lewat nama handler, belum lewat respons 403 | jaminan lewat struktur `mountPraktikum`, bukan test | tambah test 403 untuk satu route `write(...)` dan satu route admin di kedua path |
| T-8 | runner | Stop Python pada program CPU-bound (`while True: pass`) tidak menghentikan program, sama dengan kedua app lama | run berikutnya ditolak sampai reload | terminate + respawn setelah masa tenggang |
| T-9 | runner | `Runner.stop()` dan `Runner.sendInput()` belum dibedakan per bahasa | aman selama satu run per halaman | tambah argumen `Language` |
| T-10 | backend siakad (desain Fase 5) | Satu `CRON_SECRET` membuka cron praktikum, cron siakad, dan bypass `send-push`/`send-wa`. Push tanpa `targetUserId` dikirim ke semua subscription. | kalau secret cron-job.org bocor, pelaku bisa broadcast ke semua user siakad | pisahkan secret untuk bypass notifikasi |
| T-11 | siakad `send-wa.ts` | WA bisa diarahkan ke URL gateway kustom | penyalahgunaan oleh pemegang token staf | batasi URL ke daftar gateway yang dikenal |
| T-12 | siakad metadata RPC | Cache `pg_proc` per instance tidak pernah diinvalidasi | perubahan signature fungsi baru terlihat setelah instance didaur ulang | TTL cache |
| T-10 | elearning `api/index.ts` | Cabang `!tokenPayload` → `/?error=invalid_token` tidak pernah jalan karena `verifyToken` melempar error, jadi token rusak berakhir di `/?error=server_error`. `POST /api/elearning/sso` mengikuti plan: token rusak → `invalid_token` | SPA lama tidak membaca `?error=`, jadi tidak terlihat pengguna | biarkan; halaman `/sso` baru boleh menampilkan pesan per kode |
| T-11 | elearning `api/report.ts` | Upsert `elearning_progress` menulis `lessons_completed`, `completed_levels`, `current_level`, padahal skema siakad `20260410060000_create_elearning_progress.sql` memakai `completed_lessons` dan tidak punya dua kolom lainnya. Tipe `completed_levels` di DB elearning tidak diketahui | upsert gagal bila tabel di DB elearning sama dengan skema siakad | pemilik cek `\d public.elearning_progress` di DB elearning |
| T-12 | elearning `api/validate-quiz.ts` | Token yang gagal diverifikasi tidak dijawab 401: `verifyToken` melempar error dan jatuh ke `catch` umum, jadi responsnya 500 `Internal server error`. `Quiz.tsx` mengirim `Bearer ` kosong bila sesi tidak ada, dan itu juga berakhir 500. `POST /api/elearning/validate-quiz` meniru perilaku ini | kuis menampilkan "Gagal memvalidasi kuis" alih-alih meminta login ulang | jawab 401 `Unauthorized: Invalid token` setelah frontend baru siap |
