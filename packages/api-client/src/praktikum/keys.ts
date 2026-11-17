export type QueryParams = Record<string, string | number | boolean | null | undefined>;
const all = ['praktikum'] as const;
const info = [...all, 'info'] as const;
const mahasiswa = [...all, 'mahasiswa'] as const;
const admin = [...all, 'admin'] as const;
const sesi = [...admin, 'sesi'] as const;
const aktivasi = [...admin, 'aktivasi'] as const;
const penilaian = [...admin, 'penilaian'] as const;
const rekap = [...admin, 'rekap'] as const;
const laporanBesar = [...admin, 'laporanBesar'] as const;
const superadmin = [...all, 'superadmin'] as const;
export const praktikumKeys = {
  all,
  info: {
    all: info,
    asisten: [...info, 'asisten'] as const,
    jadwal: [...info, 'jadwal'] as const,
    jadwalConfig: [...info, 'jadwalConfig'] as const,
    kelasMahasiswa: (kelasId: number) => [...info, 'kelasMahasiswa', kelasId] as const,
    laporan: [...info, 'laporan'] as const,
    modul: [...info, 'modul'] as const,
    announcements: [...info, 'announcements'] as const,
  },
  mahasiswa: {
    all: mahasiswa,
    dashboard: [...mahasiswa, 'dashboard'] as const,
    sesi: [...mahasiswa, 'sesi'] as const,
    ruang: (aktivasiSesiId: number, courseId: number) =>
      [...mahasiswa, 'ruang', aktivasiSesiId, courseId] as const,
    laporanBesar: [...mahasiswa, 'laporanBesar'] as const,
  },
  profile: [...all, 'profile'] as const,
  admin: {
    all: admin,
    dashboard: [...admin, 'dashboard'] as const,
    users: [...admin, 'users'] as const,
    kelas: [...admin, 'kelas'] as const,
    asisten: [...admin, 'asisten'] as const,
    ampuan: [...admin, 'ampuan'] as const,
    jadwal: [...admin, 'jadwal'] as const,
    konfigurasi: [...admin, 'konfigurasi'] as const,
    aiModel: [...admin, 'aiModel'] as const,
    pedoman: [...admin, 'pedoman'] as const,
    sesi: {
      all: sesi,
      list: [...sesi, 'list'] as const,
      courses: (sesiId: number) => [...sesi, 'courses', sesiId] as const,
    },
    soal: (courseId: number) => [...admin, 'soal', courseId] as const,
    aktivasi: {
      all: aktivasi,
      list: [...aktivasi, 'list'] as const,
      detail: (id: number) => [...aktivasi, 'detail', id] as const,
      susulan: (id: number) => [...aktivasi, 'susulan', id] as const,
      peserta: (id: number) => [...aktivasi, 'peserta', id] as const,
    },
    penilaian: {
      all: penilaian,
      rekap: (aktivasiSesiId: number, courseId: number) =>
        [...penilaian, 'rekap', aktivasiSesiId, courseId] as const,
      aiTargets: (aktivasiSesiId: number, courseId: number) =>
        [...penilaian, 'aiTargets', aktivasiSesiId, courseId] as const,
    },
    rekap: {
      all: rekap,
      kelas: (kelasId: number) => [...rekap, 'kelas', kelasId] as const,
      jawaban: (params: QueryParams) => [...rekap, 'jawaban', params] as const,
    },
    auditLogs: (params: QueryParams) => [...admin, 'auditLogs', params] as const,
    laporanBesar: {
      all: laporanBesar,
      current: [...laporanBesar, 'current'] as const,
      pengumpulan: (id: number) => [...laporanBesar, 'pengumpulan', id] as const,
      berkas: (id: number, mahasiswaId: number, aksi: string, jenis: string) =>
        [...laporanBesar, 'berkas', id, mahasiswaId, aksi, jenis] as const,
    },
  },
  superadmin: {
    all: superadmin,
    permissions: [...superadmin, 'permissions'] as const,
  },
};
