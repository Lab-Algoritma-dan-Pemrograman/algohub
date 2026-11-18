import type { Route } from '@playwright/test';
import type { Profile } from '@algohub/api-client/elearning';
import {
  PASSWORD,
  achievements,
  activityLog,
  classmate,
  classmate2,
  levels,
  lessons,
  modules,
  page,
  payloadOf,
  playgroundExamples,
  prestigeEligible,
  prestigeUsed,
  profileOf,
  quizAnswers,
  tokenFor,
  userRow,
  users,
  weeklyEntry,
  type TestUser,
} from './fixtures.ts';

export const API_PREFIX = '/api/elearning';

export interface ApiCall {
  method: string;
  path: string;
  search: URLSearchParams;
  body: unknown;
  authorization: string | undefined;
}

// Union tanpa `unknown` supaya parameter handler fungsi mendapat tipe ApiCall dari konteks.
type Reply = ((call: ApiCall) => unknown) | object | string | number | boolean | null;

interface Handler {
  method: string;
  path: string | RegExp;
  reply: Reply;
  status: number;
}

// Dilempar dari handler untuk menjawab envelope gagal {success:false,message}.
export class ApiFailure {
  constructor(
    readonly status: number,
    readonly message: string,
  ) {}
}

const matches = (path: string | RegExp, actual: string) =>
  typeof path === 'string' ? path === actual : path.test(actual);

// Mock backend Go elearning: envelope {success,message,data} dan page {items,total,page,limit}
// seperti backend/internal/handler/elearning dan dto/elearning.
export class MockApi {
  readonly calls: ApiCall[] = [];
  readonly unmocked: string[] = [];
  readonly completed = new Set<string>();
  private handlers: Handler[] = [];
  private readonly usedCodes = new Set<string>();
  private user: TestUser = users.mahasiswa;
  private profile: Profile = profileOf(users.mahasiswa);
  private prestige = prestigeEligible;

  constructor() {
    this.defaults();
  }

  // Handler yang didaftarkan belakangan menang, jadi test bisa menimpa jawaban default.
  on(method: string, path: string | RegExp, reply: Reply, status = 200) {
    this.handlers.unshift({ method, path, reply, status });
    return this;
  }

  callsTo(method: string, path: string | RegExp) {
    return this.calls.filter((c) => c.method === method && matches(path, c.path));
  }

  async handle(route: Route) {
    const req = route.request();
    const url = new URL(req.url());
    const path = url.pathname.startsWith(API_PREFIX)
      ? url.pathname.slice(API_PREFIX.length)
      : url.pathname;
    const raw = req.postData();
    let body: unknown = raw;
    if (raw && req.headers()['content-type']?.includes('application/json')) body = JSON.parse(raw);
    const call: ApiCall = {
      method: req.method(),
      path,
      search: url.searchParams,
      body,
      authorization: req.headers()['authorization'],
    };
    this.calls.push(call);
    const handler = this.handlers.find((h) => h.method === call.method && matches(h.path, path));
    if (!handler) {
      this.unmocked.push(`${call.method} ${url.pathname}${url.search}`);
      return route.fulfill({ status: 500, json: { success: false, message: 'tidak di-mock' } });
    }
    try {
      const data = typeof handler.reply === 'function' ? handler.reply(call) : handler.reply;
      return route.fulfill({
        status: handler.status,
        json: { success: true, message: 'OK', data },
      });
    } catch (e) {
      if (e instanceof ApiFailure) {
        return route.fulfill({ status: e.status, json: { success: false, message: e.message } });
      }
      throw e;
    }
  }

  private signIn(user: TestUser) {
    this.user = user;
    this.profile = profileOf(user);
    return tokenFor(user.nim);
  }

  private userByToken(token: unknown): TestUser | undefined {
    return Object.values(users).find((u) => tokenFor(u.nim) === token);
  }

  private defaults() {
    // Auth
    this.on('POST', '/login', ({ body }) => {
      const { identifier, password } = body as { identifier: string; password: string };
      const user = Object.values(users).find((u) => u.nim === identifier);
      if (!user || password !== PASSWORD) {
        throw new ApiFailure(401, 'NIM/email atau password salah');
      }
      const token = this.signIn(user);
      return {
        payload: payloadOf(user),
        token,
        backendToken: token,
        user: {
          nim: user.nim,
          nama: user.nama,
          kelas: user.kelas,
          role: user.role,
          nama_kelas: null,
          shift: 1,
        },
      };
    });
    this.on('POST', '/receive-token', ({ body }) => {
      const { code } = body as { code: string };
      if (this.usedCodes.has(code)) {
        throw new ApiFailure(400, 'Kode sudah dipakai atau kedaluwarsa.');
      }
      this.usedCodes.add(code);
      const user = users.mahasiswa;
      const token = this.signIn(user);
      return { success: true, token, redirectUrl: '/dashboard', payload: payloadOf(user) };
    });
    this.on('POST', '/verify', ({ body }) => {
      const { token } = body as { token: string };
      const user = this.userByToken(token);
      if (!user) throw new ApiFailure(401, 'Token tidak valid');
      this.user = user;
      return { payload: payloadOf(user), token };
    });
    this.on('POST', '/auth/supabase-token', { token: 'e2e-realtime-token', expiresIn: 3600 });

    // Kurikulum dan progres
    this.on('GET', '/curriculum/levels', () => page(levels));
    this.on('GET', '/curriculum/modules', () => page(modules));
    this.on('GET', '/curriculum/lessons', () => page(lessons));
    this.on('GET', '/progress/me/lessons', () =>
      page(
        [...this.completed].map((lessonId) => ({
          id: `progress-${lessonId}`,
          nim: this.user.nim,
          lesson_id: lessonId,
          completed: true,
          completed_at: '2026-11-17T08:00:00Z',
          score: null,
        })),
      ),
    );
    this.on('GET', '/progress/me/recent', []);
    this.on('POST', /^\/progress\/lessons\/[^/]+\/complete$/, ({ path }) => {
      const lessonId = decodeURIComponent(path.split('/')[3]);
      const already = this.completed.has(lessonId);
      this.completed.add(lessonId);
      const xp = this.profile.xp + (already ? 0 : 60);
      this.profile = { ...this.profile, xp };
      return {
        xp,
        level: 1,
        streak: 2,
        lastActive: '2026-11-17T08:00:00Z',
        leveledUp: false,
        alreadyCompleted: already,
        newlyUnlocked: [],
      };
    });
    this.on('POST', '/validate-quiz', ({ body }) => {
      const { lessonId, selectedOption } = body as { lessonId: string; selectedOption: number };
      const correctAnswer = quizAnswers[lessonId] ?? null;
      return { isCorrect: selectedOption === correctAnswer, correctAnswer };
    });

    // Profil, sesi, dan aktivitas
    this.on('POST', '/profile/sync', () => ({ profile: this.profile, newlyUnlocked: [] }));
    this.on('PUT', '/sessions/me', {
      sessionSaved: true,
      streak: { streak: 2, lastActive: '2026-11-17T08:00:00Z' },
    });
    this.on('DELETE', '/sessions/me', null);
    this.on('POST', '/activity-logs', null);
    this.on('GET', '/activity-logs', () => page([activityLog()]));
    this.on('DELETE', /^\/activity-logs\/[^/]+$/, null);

    // Achievement, game, prestige, playground
    this.on('GET', '/achievements', achievements);
    this.on('GET', '/achievements/me', []);
    this.on('POST', '/achievements/check-xp', { achievement: null });
    this.on('GET', '/game/settings', {
      bugHuntActive: true,
      bugHuntCActive: true,
      bugHuntPythonActive: true,
      bugHuntWeeklyLimit: 3,
      bugHuntQuestionCount: 5,
    });
    this.on('GET', '/game/plays/me', { allowed: true, playsUsed: 0, limit: 3 });
    this.on('GET', '/game/questions', []);
    this.on('GET', '/prestige/eligibility', () => this.prestige);
    this.on('POST', '/prestige/reset', () => {
      this.prestige = prestigeUsed;
      this.completed.clear();
      this.profile = { ...this.profile, xp: 0, prestigeTier: 'Master', prestigeMult: 1.5 };
      return { tier: 'Master', mult: 1.5, prestigeNo: 1 };
    });
    this.on('GET', '/playground/examples', () => page(playgroundExamples));

    // Leaderboard
    this.on('GET', '/leaderboard/weekly', () =>
      page([
        weeklyEntry(classmate, 480, 1),
        weeklyEntry(users.mahasiswa, 340, 2),
        weeklyEntry(classmate2, 120, 3),
      ]),
    );
    this.on('GET', '/leaderboard', () =>
      page([
        weeklyEntry(classmate, 480, 1),
        weeklyEntry(users.mahasiswa, 340, 2),
        weeklyEntry(classmate2, 120, 3),
      ]),
    );
    this.on('GET', '/leaderboard/rank', { rank: 2 });
    this.on('GET', '/leaderboard/filters', {
      kelas: ['A1', 'B1'],
      jurusan: ['Teknik Elektro'],
      shift: ['1', '2'],
    });

    // Admin dan monitoring
    this.on('GET', '/admin/users', () =>
      page([userRow(users.mahasiswa), userRow(classmate, { xp: 480 })]),
    );
    this.on('GET', /^\/admin\/users\/[^/]+\/progress$/, () => page([]));
    this.on('PUT', /^\/admin\/users\/[^/]+\/xp$/, null);
    this.on('GET', '/monitoring/students', () =>
      page([userRow(users.mahasiswa), userRow(classmate, { xp: 480 })]),
    );
    this.on('GET', '/monitoring/sessions', () => page([]));
    this.on('GET', '/monitoring/progress', () => page([]));
    this.on('DELETE', '/monitoring/sessions/stale', null);
  }
}
