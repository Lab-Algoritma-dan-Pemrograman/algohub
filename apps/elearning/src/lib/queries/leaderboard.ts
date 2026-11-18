import {
  elearningKeys,
  type LeaderboardFilter,
  type PageParams,
} from '@algohub/api-client/elearning';
import { queryOptions } from '@tanstack/svelte-query';
import { api } from '#lib/api.ts';

type Params = LeaderboardFilter & PageParams;
// Papan total memuat data publik antar-mahasiswa yang sama untuk semua pembuka, jadi boleh
// dipersist (03-frontend.md). Mingguan dan peringkat pribadi cepat basi, jadi tidak.
export const leaderboardListQuery = (params: Params) =>
  queryOptions({
    queryKey: elearningKeys.leaderboard.list(params),
    queryFn: () => api.leaderboard.list(params),
    meta: { persist: true },
  });
export const weeklyLeaderboardQuery = (params: Params) =>
  queryOptions({
    queryKey: elearningKeys.leaderboard.weekly(params),
    queryFn: () => api.leaderboard.weekly(params),
  });
export const leaderboardRankQuery = (xp: number, filter: LeaderboardFilter) =>
  queryOptions({
    queryKey: elearningKeys.leaderboard.rank(xp, filter),
    queryFn: () => api.leaderboard.rank({ ...filter, xp }),
  });
export const leaderboardFiltersQuery = () =>
  queryOptions({
    queryKey: elearningKeys.leaderboard.filters(),
    queryFn: () => api.leaderboard.filters(),
  });
