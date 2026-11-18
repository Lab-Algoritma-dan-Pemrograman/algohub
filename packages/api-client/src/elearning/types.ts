// Bentuk mengikuti backend/internal/dto/elearning dan entity/elearning.

export interface Page<T> {
  items: T[];
  total: number;
  page: number;
  limit: number;
}

export interface PageParams {
  page?: number;
  limit?: number;
}

export type JsonValue =
  string | number | boolean | null | JsonValue[] | { [key: string]: JsonValue };

export type ElearningRole = 'admin' | 'kordas' | 'asisten' | 'praktikan' | (string & {});

// Auth

export interface TokenRequest {
  token?: string;
  code?: string;
}

export interface LoginRequest {
  identifier: string;
  password: string;
}

// Nilai claim diteruskan apa adanya dari token lama, jadi bisa string, angka, atau null.
export interface AuthPayload {
  nim: string | number | null;
  nama: string | null;
  kelas: string | null;
  shift: string | null;
  jurusan: string | null;
  email: string | null;
  role: string;
}

export interface VerifyResponse {
  payload: AuthPayload;
  token: string;
}

export interface ReceiveTokenResponse {
  success: boolean;
  token: string;
  redirectUrl: string;
  payload: AuthPayload;
}

export interface LoginUser {
  nim: string;
  nama: string;
  kelas: string;
  role: string;
  nama_kelas: string | null;
  shift: number | null;
}

export interface LoginResponse {
  payload: AuthPayload;
  token: string;
  backendToken: string;
  user: LoginUser;
}

export interface SupabaseTokenResponse {
  token: string;
  expiresIn?: number;
}

// Curriculum

export interface Level {
  id: string;
  title: string;
  description: string | null;
  access_mode: string | null;
  locked: boolean | null;
  sort_order: number;
}

export interface Module {
  id: string;
  level_id: string;
  title: string;
  sort_order: number;
}

export interface Lesson {
  id: string;
  module_id: string;
  title: string;
  explanation: string | null;
  code_example: string | null;
  initial_code: string | null;
  solution: string | null;
  hint: string | null;
  quiz: JsonValue;
  test_cases: JsonValue;
  validation_rules: JsonValue;
  xp_reward: number | null;
  sort_order: number;
  exercise_type: string | null;
  flowchart_distractors: string | null;
}

export interface CurriculumLesson {
  id: string;
  title: string;
  explanation: string | null;
  codeExample: string | null;
  initialCode: string | null;
  solution: string | null;
  hint: string | null;
  quiz: JsonValue;
  testCases: JsonValue;
  validationRules: JsonValue;
}

export interface CurriculumModule {
  id: string;
  title: string;
  lessons: CurriculumLesson[];
}

export interface CurriculumLevel {
  id: string;
  title: string;
  description: string | null;
  accessMode: string;
  locked: boolean;
  modules: CurriculumModule[];
}

export interface CurriculumRequest {
  levels: CurriculumLevel[];
}

// Progress

export interface StudentProgress {
  id: string;
  nim: string;
  lesson_id: string;
  completed: boolean | null;
  completed_at: string | null;
  score: number | null;
}

export interface CompleteLessonResponse {
  xp: number;
  level: number;
  streak: number;
  lastActive: string | null;
  leveledUp: boolean;
  alreadyCompleted: boolean;
  newlyUnlocked: Achievement[];
}

export interface XPResponse {
  xp: number;
  level: number;
  leveledUp: boolean;
  achievement: Achievement | null;
}

export interface StudyTimeResponse {
  studyTime: number | null;
}

export interface StreakResponse {
  streak: number;
  lastActive: string;
}

export interface Profile {
  nim: string;
  nama: string;
  kelas: string;
  jurusan: string | null;
  shift: string | null;
  email: string | null;
  prestigeTier: string | null;
  prestigeMult: number;
  prestigeCount: number;
  xp: number;
  level: number;
  streak: number;
  lastActive: string | null;
  createdAt: string | null;
  role: string;
  assessmentAccess: JsonValue;
  levelAccessOverrides: JsonValue;
  studyTime: number;
}

export interface ProfileSyncResponse {
  profile: Profile;
  newlyUnlocked: Achievement[];
}

// Achievements

export interface Achievement {
  id: string;
  title: string;
  description: string;
  icon: string | null;
  requirementType: string | null;
  requirementValue: number | string | null;
}

export interface AchievementCheckRequest {
  xp?: number;
  gamesPlayed?: number;
  perfectGames?: number;
  lessonCount?: number;
  completedLevelIds?: string[];
}

export interface CheckXPResponse {
  achievement: Achievement | null;
}

// Activity logs

export interface ActivityLogRequest {
  eventType: string;
  details?: string;
  nama?: string;
}

export interface ActivityLog {
  id: string;
  nim: string | null;
  nama: string | null;
  event_type: string;
  timestamp: string | null;
  details: string | null;
  ip_address: string | null;
}

// Admin

export interface ElearningUser {
  nim: string;
  id: string | null;
  nama: string;
  kelas: string;
  jurusan: string | null;
  email: string | null;
  role: string;
  elearning_role: string | null;
  xp: number;
  level: number;
  streak: number;
  study_time: number | null;
  last_active: string | null;
  created_at: string | null;
  assessment_access: JsonValue;
  level_access_overrides: JsonValue;
  shift: string | null;
  prestige_tier: string | null;
  prestige_mult: number;
  prestige_count: number;
  last_prestige_at: string | null;
}

export interface ResetLessonRequest {
  lessonId: string;
  lessonTitle?: string;
}

// Game

export interface GameQuestion {
  id: string;
  language: string;
  difficulty: string | null;
  title: string | null;
  code: string | null;
  bugLine: number | null;
  explanation: string | null;
}

export interface GameQuestionInput {
  language?: string | null;
  difficulty?: string | null;
  title?: string | null;
  code?: string | null;
  bugLine?: number | null;
  explanation?: string | null;
}

export interface GameQuestionUpdate {
  language?: string;
  difficulty?: string;
  title?: string;
  code?: string;
  bugLine?: number | null;
  explanation?: string;
}

export interface GameSettings {
  bugHuntActive: boolean | null;
  bugHuntCActive: boolean | null;
  bugHuntPythonActive: boolean | null;
  bugHuntWeeklyLimit: number | null;
  bugHuntQuestionCount: number;
}

export type GameSettingsUpdate = Partial<
  Pick<
    GameSettings,
    'bugHuntActive' | 'bugHuntCActive' | 'bugHuntPythonActive' | 'bugHuntWeeklyLimit'
  >
>;

export interface BugHuntAccess {
  allowed: boolean;
  playsUsed: number;
  limit: number;
}

export interface GameHistoryRequest {
  gameType: string;
  xpEarned: number;
}

export interface GameHistoryResponse {
  xp: number;
  level: number;
  streak: number;
  lastActive: string;
  leveledUp: boolean;
}

// Leaderboard

export interface LeaderboardFilter {
  kelas?: string;
  jurusan?: string;
  shift?: string;
}

export interface LeaderboardEntry {
  nim: string;
  nama: string | null;
  kelas: string | null;
  jurusan: string | null;
  shift: string | null;
  email: string | null;
  xp: number;
  level: number;
  streak: number;
  lastActive: string;
  createdAt: string;
  role: string;
  assessmentAccess: Record<string, JsonValue>;
}

export interface WeeklyEntry extends LeaderboardEntry {
  weeklyXp: number;
  lessonXp: number;
  bughuntXp: number;
  lessonsCompleted: number;
  isCarry: boolean;
  carryXp: number;
}

export interface RankResponse {
  rank: number | null;
}

export interface LeaderboardFilters {
  kelas: string[];
  jurusan: string[];
  shift: string[];
}

// Prestige

export interface PrestigeEligibility {
  eligible: boolean;
  reason: string;
  levelsDone: number;
  totalLevels: number;
  lessonsDone: number;
  totalLessons: number;
  tier: string;
  mult: number;
  prestigeCount: number;
  lastPrestigeAt: string | null;
  semesterUsed: number;
}

export interface PrestigeResetResponse {
  tier: string;
  mult: number;
  prestigeNo: number;
}

// Playground

export interface PlaygroundExample {
  id: string;
  title: string;
  language: string;
  code: string;
  description: string | null;
  sortOrder: number;
}

export interface PlaygroundInput {
  title?: string | null;
  language?: string | null;
  code?: string | null;
  description?: string | null;
  sortOrder?: number | null;
}

export interface PlaygroundUpdate {
  title?: string;
  language?: string;
  code?: string;
  description?: string | null;
  sortOrder?: number | null;
}

// Sessions dan monitoring

export interface HeartbeatRequest {
  currentActivity?: string | null;
  nama?: string;
  kelas?: string;
  attemptId?: string | null;
  touchStreak?: boolean;
  studySeconds?: number;
}

export interface HeartbeatResponse {
  sessionSaved: boolean;
  streak?: StreakResponse;
  studyTime?: StudyTimeResponse;
}

export interface ActiveSessionRole {
  nim: string;
  last_heartbeat: string | null;
  users: { role: string | null } | null;
}

export interface MonitoringProgress {
  nim: string;
  lessonIds: string[];
}

// AI, grade, report, quiz, storage

export interface AIRequest {
  prompt: string;
  model?: string;
  token?: string;
  responseMimeType?: string;
  responseSchema?: JsonValue;
  fileData?: string;
  fileMimeType?: string;
}

export interface AIResponse {
  text: string;
}

export interface GradeRequest {
  attemptIds: Array<string | number>;
  requestedModel?: string;
  token?: string;
}

export interface GradeResult {
  attemptId: string | number | null;
  success: boolean;
  finalScore?: number | string | null;
  studentName?: string;
  error?: string;
}

export interface GradeResponse {
  success: boolean;
  gradedCount: number;
  results: GradeResult[];
}

export interface ReportRequest {
  token?: string;
  nim?: string | number;
}

export interface ReportResponse {
  success: boolean;
  recalculated: { completed: number; total: number; percentage: number };
}

export interface ValidateQuizRequest {
  lessonId: string;
  selectedOption: number | null;
}

export interface ValidateQuizResponse {
  isCorrect: boolean;
  correctAnswer: number | null;
}

export interface UploadResponse {
  publicUrl: string;
}
