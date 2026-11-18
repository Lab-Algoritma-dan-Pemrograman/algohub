import type { LeaderboardFilter } from './types';

export const elearningKeys = {
  all: ['elearning'] as const,
  session: () => ['elearning', 'session'] as const,
  curriculum: {
    levels: () => ['elearning', 'curriculum', 'levels'] as const,
    modules: () => ['elearning', 'curriculum', 'modules'] as const,
    lessons: () => ['elearning', 'curriculum', 'lessons'] as const,
  },
  progress: {
    lessons: () => ['elearning', 'progress', 'lessons'] as const,
    recent: () => ['elearning', 'progress', 'recent'] as const,
  },
  achievements: {
    catalog: () => ['elearning', 'achievements', 'catalog'] as const,
    mine: () => ['elearning', 'achievements', 'mine'] as const,
  },
  leaderboard: {
    list: (filter: LeaderboardFilter = {}) => ['elearning', 'leaderboard', 'list', filter] as const,
    rank: (xp: number, filter: LeaderboardFilter = {}) =>
      ['elearning', 'leaderboard', 'rank', xp, filter] as const,
    filters: () => ['elearning', 'leaderboard', 'filters'] as const,
    weekly: () => ['elearning', 'leaderboard', 'weekly'] as const,
  },
  game: {
    questions: (language?: string) => ['elearning', 'game', 'questions', language ?? null] as const,
    settings: () => ['elearning', 'game', 'settings'] as const,
    plays: () => ['elearning', 'game', 'plays'] as const,
  },
  playground: {
    examples: (language?: string) =>
      ['elearning', 'playground', 'examples', language ?? null] as const,
  },
  monitoring: {
    sessions: () => ['elearning', 'monitoring', 'sessions'] as const,
    students: () => ['elearning', 'monitoring', 'students'] as const,
    progress: (nim?: string) => ['elearning', 'monitoring', 'progress', nim ?? null] as const,
  },
  activityLogs: {
    list: (page = 1) => ['elearning', 'activity-logs', page] as const,
  },
  admin: {
    users: (q = '', page = 1) => ['elearning', 'admin', 'users', q, page] as const,
    userProgress: (nim: string, page = 1) =>
      ['elearning', 'admin', 'user-progress', nim, page] as const,
  },
  prestige: {
    eligibility: (nim?: string) => ['elearning', 'prestige', 'eligibility', nim ?? null] as const,
  },
};
