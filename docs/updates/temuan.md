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