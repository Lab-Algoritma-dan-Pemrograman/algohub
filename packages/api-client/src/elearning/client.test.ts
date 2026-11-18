import { afterEach, describe, expect, it, vi } from 'vitest';
import { ApiError } from '../client';
import { createElearningApi } from './client';

function jsonResponse(status: number, body: unknown): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { 'Content-Type': 'application/json' },
  });
}

function setup(response: Response, token: string | null = 'tok') {
  const fetchMock = vi.fn(async (_url: string, _init: RequestInit) => response.clone());
  vi.stubGlobal('fetch', fetchMock);
  const onUnauthorized = vi.fn();
  const api = createElearningApi({
    baseUrl: 'http://api.test',
    getToken: () => token,
    onUnauthorized,
  });
  const call = () => {
    const [url, init] = fetchMock.mock.calls[0];
    return { url, init, headers: init.headers as Record<string, string> };
  };
  return { api, fetchMock, onUnauthorized, call };
}

const ok = jsonResponse(200, { success: true, message: 'ok', data: { hello: 'dunia' } });

afterEach(() => {
  vi.unstubAllGlobals();
});

describe('createElearningApi', () => {
  it('memakai prefix /api/elearning, Bearer, dan membuka envelope', async () => {
    const { api, call } = setup(ok);
    await expect(api.game.settings()).resolves.toEqual({ hello: 'dunia' });
    const { url, init, headers } = call();
    expect(url).toBe('http://api.test/api/elearning/game/settings');
    expect(init.method).toBe('GET');
    expect(headers['Authorization']).toBe('Bearer tok');
  });

  it('401 memanggil onUnauthorized dan melempar ApiError', async () => {
    const { api, onUnauthorized } = setup(
      jsonResponse(401, { success: false, message: 'Token tidak valid' }),
    );
    await expect(api.profile.sync()).rejects.toMatchObject({
      name: 'Error',
      message: 'Token tidak valid',
      status: 401,
    });
    await expect(api.profile.sync()).rejects.toBeInstanceOf(ApiError);
    expect(onUnauthorized).toHaveBeenCalled();
  });

  it('panggilan tanpa auth tidak mengirim Authorization dan tidak memanggil onUnauthorized', async () => {
    const { api, call, onUnauthorized } = setup(
      jsonResponse(401, { success: false, message: 'NIM/email atau password salah.' }),
    );
    await expect(api.auth.login({ identifier: 'a', password: 'b' })).rejects.toMatchObject({
      status: 401,
    });
    expect(call().headers['Authorization']).toBeUndefined();
    expect(onUnauthorized).not.toHaveBeenCalled();
  });

  it('filter kosong dibuang dari query string dan segmen path di-encode', async () => {
    const { api, fetchMock } = setup(ok);
    await api.leaderboard.list({ kelas: 'A', jurusan: '', page: 2 });
    await api.progress.completeLesson('a/b', 10);
    expect(fetchMock.mock.calls[0][0]).toBe(
      'http://api.test/api/elearning/leaderboard?kelas=A&page=2',
    );
    expect(fetchMock.mock.calls[1][0]).toBe(
      'http://api.test/api/elearning/progress/lessons/a%2Fb/complete',
    );
  });

  it('upload mengirim FormData berisi file', async () => {
    const { api, call } = setup(ok);
    await api.storage.uploadLessonImage(new File(['x'], 'a.png', { type: 'image/png' }));
    const { init, headers } = call();
    expect(init.method).toBe('POST');
    expect(init.body).toBeInstanceOf(FormData);
    expect((init.body as FormData).get('file')).toBeInstanceOf(File);
    expect(headers['Content-Type']).toBeUndefined();
  });

  it('ssoUrl menunjuk ke endpoint form POST', () => {
    const { api } = setup(ok);
    expect(api.auth.ssoUrl()).toBe('http://api.test/api/elearning/sso');
  });
});

describe('cakupan route docs/api.md', () => {
  type Row = [string, string, (api: ReturnType<typeof createElearningApi>) => Promise<unknown>];
  const routes: Row[] = [
    ['POST', '/receive-token', (a) => a.auth.receiveToken({ code: 'c' })],
    ['POST', '/login', (a) => a.auth.login({ identifier: 'i', password: 'p' })],
    ['POST', '/verify', (a) => a.auth.verify({ token: 't' })],
    ['POST', '/ai', (a) => a.ai.generate({ prompt: 'p' })],
    ['POST', '/grade', (a) => a.grade.run({ attemptIds: [1] })],
    ['POST', '/report', (a) => a.report.sync({ nim: '1' })],
    ['POST', '/validate-quiz', (a) => a.validateQuiz({ lessonId: 'l', selectedOption: 0 })],
    ['POST', '/auth/supabase-token', (a) => a.auth.supabaseToken()],
    ['GET', '/curriculum/levels', (a) => a.curriculum.levels()],
    ['GET', '/curriculum/modules', (a) => a.curriculum.modules()],
    ['GET', '/curriculum/lessons', (a) => a.curriculum.lessons()],
    ['PUT', '/curriculum', (a) => a.curriculum.save({ levels: [] })],
    ['DELETE', '/curriculum', (a) => a.curriculum.clear()],
    ['GET', '/progress/me/lessons', (a) => a.progress.myLessons()],
    ['GET', '/progress/me/recent', (a) => a.progress.recent()],
    ['POST', '/progress/lessons/l1/complete', (a) => a.progress.completeLesson('l1', 5)],
    ['POST', '/progress/xp', (a) => a.progress.grantXp(5)],
    ['POST', '/progress/study-time', (a) => a.progress.studyTime(30)],
    ['POST', '/streak', (a) => a.streak.touch()],
    ['POST', '/profile/sync', (a) => a.profile.sync()],
    ['POST', '/activity-logs', (a) => a.activityLogs.add({ eventType: 'X' })],
    ['GET', '/activity-logs', (a) => a.activityLogs.list()],
    ['DELETE', '/activity-logs/7', (a) => a.activityLogs.remove('7')],
    ['GET', '/achievements', (a) => a.achievements.catalog()],
    ['GET', '/achievements/me', (a) => a.achievements.mine()],
    ['POST', '/achievements/check', (a) => a.achievements.check({ xp: 1 })],
    ['POST', '/achievements/check-xp', (a) => a.achievements.checkXp(1)],
    ['GET', '/admin/users', (a) => a.admin.users({ q: 'x' })],
    ['GET', '/admin/users/1/progress', (a) => a.admin.userProgress('1')],
    ['PATCH', '/admin/users/1/level-access', (a) => a.admin.setLevelAccess('1', {})],
    ['PATCH', '/admin/users/1/role', (a) => a.admin.setRole('1', 'asisten')],
    ['POST', '/admin/users/1/reset', (a) => a.admin.resetUser('1')],
    ['POST', '/admin/users/1/reset-level', (a) => a.admin.resetLevel('1', 'lv')],
    ['POST', '/admin/users/1/reset-lesson', (a) => a.admin.resetLesson('1', { lessonId: 'l' })],
    ['PUT', '/admin/users/1/xp', (a) => a.admin.adjustXp('1', 10)],
    ['DELETE', '/admin/users/1', (a) => a.admin.deleteUser('1')],
    ['GET', '/game/questions', (a) => a.game.questions()],
    ['POST', '/game/questions', (a) => a.game.addQuestion({ title: 't' })],
    ['PUT', '/game/questions', (a) => a.game.replaceQuestions([])],
    ['POST', '/game/questions/seed', (a) => a.game.seedQuestions([])],
    ['PATCH', '/game/questions/q1', (a) => a.game.updateQuestion('q1', { title: 't' })],
    ['DELETE', '/game/questions/q1', (a) => a.game.deleteQuestion('q1')],
    ['GET', '/game/settings', (a) => a.game.settings()],
    ['PUT', '/game/settings', (a) => a.game.updateSettings({ bugHuntActive: true })],
    ['GET', '/game/plays/me', (a) => a.game.myPlays()],
    ['POST', '/game/history', (a) => a.game.saveHistory({ gameType: 'bug', xpEarned: 1 })],
    ['GET', '/leaderboard', (a) => a.leaderboard.list()],
    ['GET', '/leaderboard/rank', (a) => a.leaderboard.rank({ xp: 5 })],
    ['GET', '/leaderboard/filters', (a) => a.leaderboard.filters()],
    ['GET', '/leaderboard/weekly', (a) => a.leaderboard.weekly()],
    ['GET', '/prestige/eligibility', (a) => a.prestige.eligibility()],
    ['POST', '/prestige/reset', (a) => a.prestige.reset()],
    ['GET', '/playground/examples', (a) => a.playground.examples()],
    ['POST', '/playground/examples', (a) => a.playground.add({ title: 't' })],
    ['PATCH', '/playground/examples/p1', (a) => a.playground.update('p1', { title: 't' })],
    ['DELETE', '/playground/examples/p1', (a) => a.playground.remove('p1')],
    ['PUT', '/sessions/me', (a) => a.sessions.heartbeat({ currentActivity: 'x' })],
    ['DELETE', '/sessions/me', (a) => a.sessions.end()],
    ['GET', '/monitoring/sessions', (a) => a.monitoring.sessions()],
    ['DELETE', '/monitoring/sessions/stale', (a) => a.monitoring.cleanupStale()],
    ['GET', '/monitoring/students', (a) => a.monitoring.students()],
    ['GET', '/monitoring/progress', (a) => a.monitoring.progress()],
    [
      'POST',
      '/storage/lesson-images',
      (a) => a.storage.uploadLessonImage(new File(['x'], 'a.png')),
    ],
  ];

  it.each(routes)('%s %s', async (method, path, run) => {
    const { api, call } = setup(ok);
    await run(api);
    const { url, init } = call();
    expect(init.method).toBe(method);
    expect(new URL(url).pathname).toBe(`/api/elearning${path}`);
  });

  it('63 route fetch + 1 form POST /sso = 64 route docs', () => {
    expect(routes).toHaveLength(63);
  });
});
