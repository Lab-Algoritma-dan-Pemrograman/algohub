import { createApi } from '../client';
import type {
  AIRequest,
  AIResponse,
  Achievement,
  AchievementCheckRequest,
  ActiveSessionRole,
  ActivityLog,
  ActivityLogRequest,
  BugHuntAccess,
  CheckXPResponse,
  CompleteLessonResponse,
  CurriculumRequest,
  ElearningUser,
  GameHistoryRequest,
  GameHistoryResponse,
  GameQuestion,
  GameQuestionInput,
  GameQuestionUpdate,
  GameSettings,
  GameSettingsUpdate,
  GradeRequest,
  GradeResponse,
  HeartbeatRequest,
  HeartbeatResponse,
  LeaderboardEntry,
  LeaderboardFilter,
  LeaderboardFilters,
  Lesson,
  Level,
  LoginRequest,
  LoginResponse,
  MonitoringProgress,
  Module,
  Page,
  PageParams,
  PlaygroundExample,
  PlaygroundInput,
  PlaygroundUpdate,
  PrestigeEligibility,
  PrestigeResetResponse,
  ProfileSyncResponse,
  RankResponse,
  ReceiveTokenResponse,
  ReportRequest,
  ReportResponse,
  ResetLessonRequest,
  StreakResponse,
  StudentProgress,
  StudyTimeResponse,
  SupabaseTokenResponse,
  TokenRequest,
  UploadResponse,
  ValidateQuizRequest,
  ValidateQuizResponse,
  VerifyResponse,
  WeeklyEntry,
  XPResponse,
} from './types';

export interface ElearningApiOptions {
  baseUrl: string;
  getToken: () => string | null;
  onUnauthorized?: () => void;
}

type QueryValue = string | number | boolean | null | undefined;

// Nilai kosong dibuang supaya filter yang tidak dipilih tidak ikut terkirim.
function qs(params: Record<string, QueryValue>): string {
  const search = new URLSearchParams();
  for (const [key, value] of Object.entries(params)) {
    if (value === undefined || value === null || value === '') continue;
    search.set(key, String(value));
  }
  const text = search.toString();
  return text ? `?${text}` : '';
}

const seg = encodeURIComponent;

export function createElearningApi({ baseUrl, getToken, onUnauthorized }: ElearningApiOptions) {
  const api = createApi({ baseUrl, prefix: '/api/elearning', getToken, onUnauthorized });
  const open = { auth: false } as const;

  return {
    auth: {
      // POST /sso adalah form POST dari portal siakad yang dijawab 303, bukan panggilan fetch.
      ssoUrl: () => `${baseUrl}/api/elearning/sso`,
      receiveToken: (body: TokenRequest) =>
        api.post<ReceiveTokenResponse>('/receive-token', body, open),
      login: (body: LoginRequest) => api.post<LoginResponse>('/login', body, open),
      verify: (body: TokenRequest) => api.post<VerifyResponse>('/verify', body, open),
      supabaseToken: () => api.post<SupabaseTokenResponse>('/auth/supabase-token'),
    },
    curriculum: {
      levels: (p: PageParams = {}) => api.get<Page<Level>>(`/curriculum/levels${qs({ ...p })}`),
      modules: (p: PageParams = {}) => api.get<Page<Module>>(`/curriculum/modules${qs({ ...p })}`),
      lessons: (p: PageParams = {}) => api.get<Page<Lesson>>(`/curriculum/lessons${qs({ ...p })}`),
      save: (body: CurriculumRequest) => api.put<unknown>('/curriculum', body),
      clear: () => api.del<unknown>('/curriculum'),
    },
    progress: {
      myLessons: (p: PageParams = {}) =>
        api.get<Page<StudentProgress>>(`/progress/me/lessons${qs({ ...p })}`),
      recent: (limit?: number) => api.get<StudentProgress[]>(`/progress/me/recent${qs({ limit })}`),
      completeLesson: (lessonId: string, xpReward: number) =>
        api.post<CompleteLessonResponse>(`/progress/lessons/${seg(lessonId)}/complete`, {
          xpReward,
        }),
      grantXp: (amount: number) => api.post<XPResponse>('/progress/xp', { amount }),
      studyTime: (seconds: number) =>
        api.post<StudyTimeResponse>('/progress/study-time', { seconds }),
    },
    streak: {
      touch: () => api.post<StreakResponse>('/streak'),
    },
    profile: {
      sync: () => api.post<ProfileSyncResponse>('/profile/sync'),
    },
    activityLogs: {
      list: (p: PageParams = {}) => api.get<Page<ActivityLog>>(`/activity-logs${qs({ ...p })}`),
      add: (body: ActivityLogRequest) => api.post<unknown>('/activity-logs', body),
      remove: (id: string) => api.del<unknown>(`/activity-logs/${seg(id)}`),
    },
    achievements: {
      catalog: (requirementType?: string) =>
        api.get<Achievement[]>(`/achievements${qs({ requirementType })}`),
      mine: () => api.get<Achievement[]>('/achievements/me'),
      check: (body: AchievementCheckRequest) =>
        api.post<Achievement[]>('/achievements/check', body),
      checkXp: (xp: number) => api.post<CheckXPResponse>('/achievements/check-xp', { xp }),
    },
    admin: {
      users: (p: PageParams & { q?: string } = {}) =>
        api.get<Page<ElearningUser>>(`/admin/users${qs({ ...p })}`),
      userProgress: (nim: string, p: PageParams = {}) =>
        api.get<Page<StudentProgress>>(`/admin/users/${seg(nim)}/progress${qs({ ...p })}`),
      setLevelAccess: (nim: string, levelAccessOverrides: Record<string, unknown>) =>
        api.patch<unknown>(`/admin/users/${seg(nim)}/level-access`, { levelAccessOverrides }),
      setRole: (nim: string, role: string) =>
        api.patch<unknown>(`/admin/users/${seg(nim)}/role`, { role }),
      resetUser: (nim: string) => api.post<unknown>(`/admin/users/${seg(nim)}/reset`),
      resetLevel: (nim: string, levelId: string) =>
        api.post<unknown>(`/admin/users/${seg(nim)}/reset-level`, { levelId }),
      resetLesson: (nim: string, body: ResetLessonRequest) =>
        api.post<unknown>(`/admin/users/${seg(nim)}/reset-lesson`, body),
      adjustXp: (nim: string, xp: number) =>
        api.put<unknown>(`/admin/users/${seg(nim)}/xp`, { xp }),
      deleteUser: (nim: string) => api.del<unknown>(`/admin/users/${seg(nim)}`),
    },
    game: {
      questions: (opts: { language?: string; limit?: number } = {}) =>
        api.get<GameQuestion[]>(`/game/questions${qs({ ...opts })}`),
      addQuestion: (body: GameQuestionInput) => api.post<{ id: string }>('/game/questions', body),
      replaceQuestions: (questions: GameQuestionInput[]) =>
        api.put<unknown>('/game/questions', { questions }),
      seedQuestions: (questions: GameQuestionInput[]) =>
        api.post<{ seeded: boolean }>('/game/questions/seed', { questions }),
      updateQuestion: (id: string, body: GameQuestionUpdate) =>
        api.patch<unknown>(`/game/questions/${seg(id)}`, body),
      deleteQuestion: (id: string) => api.del<unknown>(`/game/questions/${seg(id)}`),
      settings: () => api.get<GameSettings>('/game/settings'),
      updateSettings: (body: GameSettingsUpdate) => api.put<unknown>('/game/settings', body),
      myPlays: () => api.get<BugHuntAccess>('/game/plays/me'),
      saveHistory: (body: GameHistoryRequest) =>
        api.post<GameHistoryResponse>('/game/history', body),
    },
    leaderboard: {
      list: (filter: LeaderboardFilter & PageParams = {}) =>
        api.get<Page<LeaderboardEntry>>(`/leaderboard${qs({ ...filter })}`),
      rank: (opts: LeaderboardFilter & { xp: number }) =>
        api.get<RankResponse>(`/leaderboard/rank${qs({ ...opts })}`),
      filters: () => api.get<LeaderboardFilters>('/leaderboard/filters'),
      weekly: (p: LeaderboardFilter & PageParams = {}) =>
        api.get<Page<WeeklyEntry>>(`/leaderboard/weekly${qs({ ...p })}`),
    },
    prestige: {
      eligibility: (nim?: string) =>
        api.get<PrestigeEligibility>(`/prestige/eligibility${qs({ nim })}`),
      reset: (nim?: string) =>
        api.post<PrestigeResetResponse>('/prestige/reset', nim ? { nim } : {}),
    },
    playground: {
      examples: (opts: PageParams & { language?: string } = {}) =>
        api.get<Page<PlaygroundExample>>(`/playground/examples${qs({ ...opts })}`),
      add: (body: PlaygroundInput) => api.post<{ id: string }>('/playground/examples', body),
      update: (id: string, body: PlaygroundUpdate) =>
        api.patch<unknown>(`/playground/examples/${seg(id)}`, body),
      remove: (id: string) => api.del<unknown>(`/playground/examples/${seg(id)}`),
    },
    sessions: {
      heartbeat: (body: HeartbeatRequest) => api.put<HeartbeatResponse>('/sessions/me', body),
      end: () => api.del<unknown>('/sessions/me'),
    },
    monitoring: {
      sessions: (p: PageParams = {}) =>
        api.get<Page<ActiveSessionRole>>(`/monitoring/sessions${qs({ ...p })}`),
      cleanupStale: (thresholdMinutes?: number) =>
        api.del<unknown>(`/monitoring/sessions/stale${qs({ thresholdMinutes })}`),
      students: (p: PageParams = {}) =>
        api.get<Page<ElearningUser>>(`/monitoring/students${qs({ ...p })}`),
      progress: (nim?: string, p: PageParams = {}) =>
        api.get<Page<MonitoringProgress>>(`/monitoring/progress${qs({ nim, ...p })}`),
    },
    storage: {
      uploadLessonImage: (file: File) => {
        const form = new FormData();
        form.append('file', file);
        return api.upload<UploadResponse>('/storage/lesson-images', form);
      },
    },
    ai: {
      generate: (body: AIRequest) => api.post<AIResponse>('/ai', body),
    },
    grade: {
      run: (body: GradeRequest) => api.post<GradeResponse>('/grade', body),
    },
    report: {
      sync: (body: ReportRequest) => api.post<ReportResponse>('/report', body),
    },
    validateQuiz: (body: ValidateQuizRequest) =>
      api.post<ValidateQuizResponse>('/validate-quiz', body),
  };
}

export type ElearningApi = ReturnType<typeof createElearningApi>;
