import type {
  AdminStatistik,
  AktivasiSesi,
  AuditLogsResponse,
  DashboardResponse,
  Kelas,
  PesertaAktivasi,
  Profile,
  RekapKelasResponse,
  RekapMahasiswa,
  RolePermissions,
  RuangCourse,
  SesiUserItem,
  User,
} from '../../../packages/api-client/src/praktikum/types';

export const SOAL_MARKER = 'SOAL-RAHASIA-E2E';
export const OTHER_USER_NAME = 'Mahasiswa Lain E2E';
export const NILAI_MARKER = 9137;
export const PASSWORD = 'rahasia-e2e';

export type Role = User['role'];

export const users: Record<Role, User> = {
  mahasiswa: {
    id: 101,
    role: 'mahasiswa',
    nim: '2027100001',
    nama: 'Mahasiswa Uji',
    is_registered: true,
    kelas_id: 1,
    shift: 1,
    gelombang: 1,
    kelompok: 'A',
  },
  asisten: { id: 201, role: 'asisten', nim: 'AST001', nama: 'Asisten Uji', is_registered: true },
  koordinator: {
    id: 301,
    role: 'koordinator',
    nim: 'KOR001',
    nama: 'Koordinator Uji',
    is_registered: true,
  },
};

export const tokenFor = (role: Role) => `tok-e2e-${role}`;

export const otherUser: User = {
  id: 102,
  role: 'mahasiswa',
  nim: '2027100002',
  nama: OTHER_USER_NAME,
  is_registered: true,
  kelas_id: 1,
  shift: 1,
  gelombang: 1,
  kelompok: 'A',
};

export const kelas: Kelas[] = [
  { id: 1, nama_kelas: 'TTL A' },
  { id: 2, nama_kelas: 'TTL B' },
];

export const profile = (role: Role): Profile => ({ ...users[role], email: `${role}@e2e.test` });

export const dashboard: DashboardResponse = {
  profil: users.mahasiswa,
  jadwal: null,
  sesi_aktif: [],
  riwayat_nilai: [
    { sesi_judul: 'Sesi 1', jenis: 'pretest', status: 'selesai', total_nilai: NILAI_MARKER },
  ],
};

export const JENIS = ['pretest', 'posttest', 'keterampilan', 'ujian'] as const;
export type JenisRoute = (typeof JENIS)[number];

export const sesiUser: SesiUserItem[] = [
  {
    sesi_id: 1,
    judul: 'Sesi 1',
    deskripsi: '',
    urutan: 1,
    is_ujian_praktik: false,
    aktif: true,
    susulan: false,
    aktivasi_sesi_id: 11,
    courses: JENIS.map((jenis, i) => ({
      course_id: 31 + i,
      aktivasi_course_id: 21 + i,
      jenis: jenis === 'ujian' ? 'ujian_praktik' : jenis,
      judul: `Course ${jenis}`,
      durasi_menit: 30,
      is_open: true,
      status: 'belum',
      total_nilai: null,
    })),
  },
];

export function ruang(jenis: JenisRoute, over: Partial<RuangCourse> = {}): RuangCourse {
  return {
    aktivasi_sesi_id: 11,
    course_id: 31 + JENIS.indexOf(jenis),
    jenis: jenis === 'ujian' ? 'ujian_praktik' : jenis,
    durasi_menit: 30,
    waktu_mulai: null,
    deadline: null,
    status: 'belum',
    is_open: true,
    require_token: jenis === 'ujian',
    soal: [
      {
        soal_terpilih_id: 501,
        urutan: 1,
        jenis_soal: 'essay',
        kategori_ujian: null,
        teks_soal: `<p>Jelaskan algoritma ${SOAL_MARKER}</p>`,
        gambar_url: null,
        poin: 100,
        jawaban_teks: '',
        is_submitted: false,
      },
    ],
    ...over,
  };
}

export function startedRuang(base: RuangCourse, deadlineMs = 30 * 60_000): RuangCourse {
  const now = Date.now();
  return {
    ...base,
    status: 'sedang',
    waktu_mulai: new Date(now).toISOString(),
    deadline: new Date(now + deadlineMs).toISOString(),
  };
}

export const statistik: AdminStatistik = {
  total_mahasiswa: 2,
  total_asisten: 1,
  sudah_register: 2,
  belum_register: 0,
  per_kelas_shift: [{ nama_kelas: 'TTL A', shift: 1, jumlah: 2 }],
  sesi_aktif: [],
};

export const aktivasi: AktivasiSesi = {
  id: 11,
  sesi_praktikum_id: 1,
  kelas_id: 1,
  shift: 1,
  gelombang: 1,
  is_active: true,
  activated_at: '2026-11-17T08:00:00Z',
  token: 'ABC123',
  sesi: { judul_sesi: 'Sesi 1' },
  kelas: { nama_kelas: 'TTL A' },
  aktivasi_courses: [
    {
      id: 21,
      course_id: 31,
      is_open: true,
      urutan: 1,
      course: { jenis: 'pretest', judul: 'Pretest 1' },
    },
  ],
};

export const peserta: PesertaAktivasi[] = [
  {
    mahasiswa_id: otherUser.id,
    nim: otherUser.nim,
    nama: OTHER_USER_NAME,
    course_id: 31,
    judul_course: 'Pretest 1',
    status: 'sedang',
    waktu_mulai: '2026-11-17T08:05:00Z',
    waktu_selesai: null,
  },
];

export const auditLogs: AuditLogsResponse = {
  logs: [
    {
      id: 1,
      user_id: otherUser.id,
      nim: otherUser.nim,
      nama: OTHER_USER_NAME,
      role: 'mahasiswa',
      action: 'login',
      description: 'Login berhasil',
      ip_address: '127.0.0.1',
      user_agent: 'e2e',
      created_at: '2026-11-17T08:00:00Z',
    },
  ],
  total: 1,
  page: 1,
  limit: 20,
};

export const permissions: RolePermissions = {
  asisten: { 'users:write': false, 'sesi:write': true },
};

export function rekapKelas(count: number): RekapKelasResponse {
  const columns = [
    { key: 's1_pretest', label: 'S1 Pretest' },
    { key: 's1_posttest', label: 'S1 Posttest' },
    { key: 's1_keterampilan', label: 'S1 Keterampilan' },
  ];
  const data: RekapMahasiswa[] = Array.from({ length: count }, (_, i) => {
    const n = i + 1;
    return {
      nim: `2027${String(n).padStart(6, '0')}`,
      nama: n === 1 ? OTHER_USER_NAME : `Mahasiswa ${n}`,
      scores: {
        s1_pretest: {
          pengerjaan_id: 10_000 + n,
          murni: 70,
          keaktifan: null,
          total: 70,
          edit_keaktifan: true,
          nilai_akhir: 70,
        },
        s1_posttest: {
          pengerjaan_id: 20_000 + n,
          murni: 80 + (n % 20),
          keaktifan: null,
          total: 80 + (n % 20),
          edit_keaktifan: true,
          nilai_akhir: 80 + (n % 20),
        },
      },
      total_score: 150 + (n % 20),
    };
  });
  return { columns, data };
}
