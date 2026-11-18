import type {
  Achievement,
  ActivityLog,
  AuthPayload,
  ElearningUser,
  Lesson,
  Level,
  Module,
  Page,
  PlaygroundExample,
  PrestigeEligibility,
  Profile,
  WeeklyEntry,
} from '@algohub/api-client/elearning';

// Bentuk mengikuti backend/internal/dto/elearning/*.go; field snake_case milik entity tabel.
export type Role = 'mahasiswa' | 'asisten' | 'kordas' | 'admin';

export const PASSWORD = 'rahasia-uji';

export interface TestUser {
  nim: string;
  nama: string;
  kelas: string;
  jurusan: string;
  role: Role;
}

export const users: Record<Role, TestUser> = {
  mahasiswa: {
    nim: '2022110001',
    nama: 'Budi Santoso',
    kelas: 'A1',
    jurusan: 'Teknik Elektro',
    role: 'mahasiswa',
  },
  asisten: {
    nim: '2020110010',
    nama: 'Asisten Uji',
    kelas: 'AS',
    jurusan: 'Teknik Elektro',
    role: 'asisten',
  },
  kordas: {
    nim: '2020110020',
    nama: 'Kordas Uji',
    kelas: 'KD',
    jurusan: 'Teknik Elektro',
    role: 'kordas',
  },
  admin: {
    nim: '2019110030',
    nama: 'Admin Uji',
    kelas: 'AD',
    jurusan: 'Teknik Elektro',
    role: 'admin',
  },
};

export const tokenFor = (nim: string) => `token-${nim}`;

export const payloadOf = (user: TestUser): AuthPayload => ({
  nim: user.nim,
  nama: user.nama,
  kelas: user.kelas,
  shift: '1',
  jurusan: user.jurusan,
  email: `${user.nim}@mhs.test`,
  role: user.role,
});

export const page = <T>(items: T[]): Page<T> => ({
  items,
  total: items.length,
  page: 1,
  limit: 100,
});

export const profileOf = (user: TestUser, patch: Partial<Profile> = {}): Profile => ({
  nim: user.nim,
  nama: user.nama,
  kelas: user.kelas,
  jurusan: user.jurusan,
  shift: '1',
  email: `${user.nim}@mhs.test`,
  prestigeTier: null,
  prestigeMult: 1,
  prestigeCount: 0,
  xp: 120,
  level: 1,
  streak: 2,
  lastActive: '2026-11-17T08:00:00Z',
  createdAt: '2025-09-01T00:00:00Z',
  role: user.role,
  assessmentAccess: {},
  levelAccessOverrides: {},
  studyTime: 600,
  ...patch,
});

// Kode contoh lesson dipakai langsung oleh panel sandbox, jadi tidak perlu mengetik di editor.
export const C_SCANF = `#include <stdio.h>

int main() {
    int a, b;
    scanf("%d %d", &a, &b);
    printf("Hasil: %d\\n", a + b);
    return 0;
}
`;

export const PY_INPUT_EXAMPLE = `nama = input("Nama: ")
print("Halo " + nama)
`;

export const PY_INPUT_EXERCISE = `nama = input()
print("Halo " + nama)
`;

export const levels: Level[] = [
  {
    id: 'level-1',
    title: 'Level 1: Dasar Logika (Bahasa C)',
    description: 'Input dan output di Bahasa C.',
    access_mode: 'auto',
    locked: false,
    sort_order: 1,
  },
  {
    id: 'level-2',
    title: 'Level 2: Pengenalan Python',
    description: 'Sintaks dasar Python.',
    access_mode: 'auto',
    locked: false,
    sort_order: 2,
  },
];

export const modules: Module[] = [
  { id: 'mod-c', level_id: 'level-1', title: 'Input Output C', sort_order: 1 },
  { id: 'mod-py', level_id: 'level-2', title: 'Input Python', sort_order: 1 },
];

// Backend membuang correctAnswer dari kuis untuk peran praktikan, jadi fixture tidak membawanya.
const quiz = (question: string, options: string[]) => ({ question, options });

export const lessons: Lesson[] = [
  {
    id: 'les-c',
    module_id: 'mod-c',
    title: 'Menjumlahkan dengan scanf',
    explanation: '<p>Gunakan scanf untuk membaca dua bilangan bulat.</p>',
    code_example: C_SCANF,
    initial_code: C_SCANF,
    solution: C_SCANF,
    hint: 'Pakai format %d.',
    quiz: quiz('Format specifier untuk int?', ['%s', '%d', '%f']),
    test_cases: [{ description: 'Cetak hasil', input: '3 4', expectedOutput: 'Hasil: 7' }],
    validation_rules: [],
    xp_reward: 60,
    sort_order: 1,
    exercise_type: null,
    flowchart_distractors: null,
  },
  {
    id: 'les-py',
    module_id: 'mod-py',
    title: 'Membaca teks dengan input()',
    explanation: '<p>Fungsi input() membaca satu baris teks dari pengguna.</p>',
    code_example: PY_INPUT_EXAMPLE,
    initial_code: PY_INPUT_EXERCISE,
    solution: PY_INPUT_EXERCISE,
    hint: 'Gabungkan sapaan dan nama.',
    quiz: quiz('Fungsi untuk membaca teks?', ['input()', 'read()', 'scan()']),
    test_cases: [{ description: 'Sapa Budi', input: 'Budi', expectedOutput: 'Halo Budi' }],
    validation_rules: [],
    xp_reward: 60,
    sort_order: 1,
    exercise_type: null,
    flowchart_distractors: null,
  },
];

// Jawaban benar tiap lesson, dipakai mock /validate-quiz yang di backend membaca bank soal.
export const quizAnswers: Record<string, number> = { 'les-c': 1, 'les-py': 0 };

export const achievements: Achievement[] = [
  {
    id: 'first-lesson',
    title: 'Langkah Pertama',
    description: 'Selesaikan lesson pertama',
    icon: 'star',
    requirementType: 'lessons',
    requirementValue: 1,
  },
];

export const prestigeEligible: PrestigeEligibility = {
  eligible: true,
  reason: 'Semua level sudah tamat.',
  levelsDone: 6,
  totalLevels: 6,
  lessonsDone: 30,
  totalLessons: 30,
  tier: 'Master',
  mult: 1.5,
  prestigeCount: 0,
  lastPrestigeAt: null,
  semesterUsed: 0,
};

export const prestigeUsed: PrestigeEligibility = {
  ...prestigeEligible,
  eligible: false,
  reason: 'Cooldown prestige 14 hari belum selesai.',
  levelsDone: 0,
  lessonsDone: 0,
  prestigeCount: 1,
  semesterUsed: 1,
};

export const weeklyEntry = (user: TestUser, weeklyXp: number, rank: number): WeeklyEntry => ({
  nim: user.nim,
  nama: user.nama,
  kelas: user.kelas,
  jurusan: user.jurusan,
  shift: '1',
  email: null,
  xp: weeklyXp + 100,
  level: 1,
  streak: rank,
  lastActive: '2026-11-17T08:00:00Z',
  createdAt: '2025-09-01T00:00:00Z',
  role: 'praktikan',
  assessmentAccess: {},
  weeklyXp,
  lessonXp: weeklyXp,
  bughuntXp: 0,
  lessonsCompleted: rank,
  isCarry: false,
  carryXp: 0,
});

export const classmate: TestUser = {
  nim: '2022110002',
  nama: 'Citra Lestari',
  kelas: 'A1',
  jurusan: 'Teknik Elektro',
  role: 'mahasiswa',
};

export const classmate2: TestUser = {
  nim: '2022110003',
  nama: 'Dewi Anggraini',
  kelas: 'A1',
  jurusan: 'Teknik Elektro',
  role: 'mahasiswa',
};

export const userRow = (user: TestUser, patch: Partial<ElearningUser> = {}): ElearningUser => ({
  nim: user.nim,
  id: `uuid-${user.nim}`,
  nama: user.nama,
  kelas: user.kelas,
  jurusan: user.jurusan,
  email: `${user.nim}@mhs.test`,
  role: 'praktikan',
  elearning_role: null,
  xp: 340,
  level: 1,
  streak: 3,
  study_time: 1800,
  last_active: '2026-11-17T08:00:00Z',
  created_at: '2025-09-01T00:00:00Z',
  assessment_access: {},
  level_access_overrides: {},
  shift: '1',
  prestige_tier: null,
  prestige_mult: 1,
  prestige_count: 0,
  last_prestige_at: null,
  ...patch,
});

export const playgroundExamples: PlaygroundExample[] = [
  {
    id: 'ex-1',
    title: 'Halo Dunia Python',
    language: 'python',
    code: 'print("Halo Dunia")\n',
    description: 'Cetak teks pertama.',
    sortOrder: 1,
  },
];

export const activityLog = (patch: Partial<ActivityLog> = {}): ActivityLog => ({
  id: 'log-1',
  nim: users.mahasiswa.nim,
  nama: users.mahasiswa.nama,
  event_type: 'login',
  timestamp: '2026-11-17T08:00:00Z',
  details: 'Login ke sistem',
  ip_address: '127.0.0.1',
  ...patch,
});
