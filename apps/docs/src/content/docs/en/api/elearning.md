---
title: E-Learning
description: Summary of the `/api/elearning/*` endpoints.
---

All e-learning endpoints live under `/api/elearning` and use
`Authorization: Bearer <token>`. Most logic lives in services in `backend`
(`curriculumService`, `progressService`, `gameService`, `leaderboardService`,
`monitoringService`, and others).

## Main groups

- **Auth & profile** — `auth/supabase-token`, `profile/sync`, `streak`,
  `sessions/me` (PUT/DELETE), `validate-quiz`.
- **Curriculum** — `curriculum/levels`, `curriculum/modules`,
  `curriculum/lessons`, plus `PUT`/`DELETE /curriculum` for admin/kordas.
- **Progress** — `progress/me/lessons`, `progress/me/recent`,
  `progress/lessons/:id/complete`, `progress/xp`, `progress/study-time`.
- **Achievements & prestige** — `achievements`, `achievements/me`,
  `achievements/check`, `prestige/eligibility`, `prestige/reset`.
- **Bug Hunt (game)** — `game/questions`, `game/settings`, `game/plays/me`,
  `game/history`.
- **Leaderboard** — `leaderboard`, `leaderboard/rank`, `leaderboard/filters`,
  `leaderboard/weekly`.
- **Monitoring** — `monitoring/sessions`, `monitoring/students`,
  `monitoring/progress`, `activity-logs`.
- **Playground & storage** — `playground/examples`, `storage/lesson-images`.

## Notes

- **Rate limit** — list endpoints are cached (for example the leaderboard) and
  some accept a `?limit` query with a default of 10 and a maximum of 100.
- **Roles** — token notation: `A` admin, `K` kordas, `S` asisten, `P` praktikan.
  Write endpoints are generally limited to `A/K`; some allow `S`, some do not.
- **Weekly quota** — Bug Hunt has a weekly limit (`WEEKLY_LIMIT_REACHED`). The
  week starts Monday 00:00 WIB.

Full detail per endpoint is in `docs/api.md` in the repo.
