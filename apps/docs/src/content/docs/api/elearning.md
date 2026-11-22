---
title: E-Learning
description: Ringkasan endpoint `/api/elearning/*`.
---

Semua endpoint e-learning berada di bawah `/api/elearning` dan memakai
autentikasi `Authorization: Bearer <token>`. Kebanyakan sumber ada pada service
di `backend` (`curriculumService`, `progressService`, `gameService`,
`leaderboardService`, `monitoringService`, dan lainnya).

## Grup utama

- **Autentikasi & profil** — `auth/supabase-token`, `profile/sync`, `streak`,
  `sessions/me` (PUT/DELETE), `validate-quiz`.
- **Kurikulum** — `curriculum/levels`, `curriculum/modules`, `curriculum/lessons`,
  serta `PUT`/`DELETE /curriculum` untuk admin/kordas.
- **Progres** — `progress/me/lessons`, `progress/me/recent`, `progress/lessons/:id/complete`,
  `progress/xp`, `progress/study-time`.
- **Achievement & prestige** — `achievements`, `achievements/me`,
  `achievements/check`, `prestige/eligibility`, `prestige/reset`.
- **Bug Hunt (game)** — `game/questions`, `game/settings`, `game/plays/me`,
  `game/history`.
- **Leaderboard** — `leaderboard`, `leaderboard/rank`, `leaderboard/filters`,
  `leaderboard/weekly`.
- **Monitoring** — `monitoring/sessions`, `monitoring/students`,
  `monitoring/progress`, `activity-logs`.
- **Playground & storage** — `playground/examples`, `storage/lesson-images`.

## Catatan penting

- **Rate limit** — endpoint ber-daftar punya cache (misal leaderboard) dan
  beberapa menerima query `?limit` dengan default 10 dan maksimum 100.
- **Role** — notasi token: `A` admin, `K` kordas, `S` asisten, `P` praktikan.
  Endpoint tulis umumnya dibatasi `A/K`; sebagian `S` boleh, sebagian tidak.
- **Kuota mingguan** — Bug Hunt punya batas mingguan (`WEEKLY_LIMIT_REACHED`).
  Minggu dimulai Senin 00:00 WIB.

Detail lengkap per endpoint ada di `docs/api.md` di repo.
