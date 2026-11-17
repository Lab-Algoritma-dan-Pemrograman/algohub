export type Role = 'asisten' | 'mahasiswa' | 'koordinator';
export interface User {
  id: number;
  role: Role;
  nim: string;
  nama: string;
  is_registered?: boolean;
  kelas_id?: number | null;
  nama_kelas?: string;
  shift?: number | null;
  gelombang?: number | null;
  foto_url?: string | null;
  nomor_hp?: string | null;
  medsos_link?: string | null;
  kelompok?: string | null;
  kelas?: Kelas;
}

export interface AmpuanKelompok {
  id: number;
  asisten_id: number;
  kelas_id: number;
  kelompok: string;
  asisten?: User;
  kelas?: Kelas;
}

export interface AuthResponse {
  token: string;
  user: User;
}

export interface CekNIMResponse {
  nim: string;
  ditemukan: boolean;
  is_registered: boolean;
  nama?: string;
  pesan: string;
}

export interface Kelas {
  id: number;
  nama_kelas: string;
}

export interface Jadwal {
  id: number;
  kelas_id: number;
  shift: number;
  hari: string;
  jam_mulai: string;
  jam_selesai: string;
  keterangan: string;
  kelas?: Kelas;
}

export interface Sesi {
  id: number;
  judul_sesi: string;
  deskripsi: string;
  urutan: number;
  is_ujian_praktik: boolean;
  courses?: Course[];
}

export interface Course {
  id: number;
  sesi_praktikum_id: number;
  jenis: string;
  judul: string;
  deskripsi: string;
  durasi_menit: number;
  kuota_easy?: number | null;
  kuota_medium?: number | null;
  kuota_hard?: number | null;
}

export interface Soal {
  id: number;
  course_id: number;
  jenis_soal: string;
  difficulty?: string | null;
  kategori_ujian?: string | null;
  teks_soal: string;
  gambar_url?: string | null;
  poin: number;
  kunci_jawaban?: string | null;
  kode_rumpang?: string | null;
}

export interface CourseUserItem {
  course_id: number;
  aktivasi_course_id: number;
  jenis: string;
  judul: string;
  durasi_menit: number;
  is_open: boolean;
  status: string;
  total_nilai?: number | null;
}

export interface SesiUserItem {
  sesi_id: number;
  judul: string;
  deskripsi: string;
  urutan: number;
  is_ujian_praktik: boolean;
  aktif: boolean;
  susulan: boolean;
  aktivasi_sesi_id?: number | null;
  courses: CourseUserItem[];
}

export interface SoalTampil {
  soal_terpilih_id: number;
  urutan: number;
  jenis_soal: string;
  kategori_ujian?: string | null;
  teks_soal: string;
  gambar_url?: string | null;
  poin: number;
  jawaban_teks: string;
  is_submitted: boolean;
}

export interface RuangCourse {
  aktivasi_sesi_id: number;
  course_id: number;
  jenis: string;
  durasi_menit: number;
  waktu_mulai?: string | null;
  deadline?: string | null;
  status: string;
  is_open: boolean;
  require_token: boolean;
  soal: SoalTampil[];
}

export interface JadwalConfig {
  mode: string;
  gdrive_url: string;
}

export interface KelasMahasiswaInfo {
  mahasiswa: User[];
  ampuan: AmpuanKelompok[];
}

export interface Pedoman {
  id: number;
  nama_dokumen: string;
  file_url: string;
}

export interface ModulInfo {
  file_url: string;
}

export interface AnnouncementsResponse {
  reschedules?: string;
  recruit?: string;
  susulan?: string;
  plagiarism?: string;
}

export interface RunResponse {
  stdout: string;
  stderr: string;
  error: string;
}

export interface ForgotPasswordResponse {
  sent: boolean;
  reason?: string;
}

export interface LayananLoginResponse {
  token: string;
  user: { full_name?: string; username?: string };
}

export interface JadwalInfo {
  hari: string;
  jam_mulai: string;
  jam_selesai: string;
  keterangan: string;
}

export interface NilaiCourseItem {
  sesi_judul: string;
  jenis: string;
  status: string;
  total_nilai?: number | null;
}

export interface DashboardResponse {
  profil: User;
  jadwal?: JadwalInfo | null;
  sesi_aktif: SesiUserItem[];
  riwayat_nilai: NilaiCourseItem[];
}

export interface Profile extends User {
  email?: string | null;
}

export interface UploadResponse {
  url: string;
}

export interface LaporanBesarStatus {
  laporan_besar_id: number;
  judul: string;
  deskripsi: string;
  butuh_laporan: boolean;
  butuh_analisa: boolean;
  kelas_id: number | null;
  buka_at: string | null;
  deadline_at: string | null;
  late_until: string | null;
  fase: 'belum_dibuka' | 'dibuka' | 'telat' | 'ditutup';
  terlambat: boolean;
  sisa_detik: number;
  bisa_kumpul: boolean;
  status: 'belum' | 'terkumpul';
  sumber: string | null;
  laporan_path: string | null;
  analisa_path: string | null;
  submitted_at: string | null;
  override_pada: string | null;
  override_alasan: string | null;
}

export interface UploadUrlResponse {
  upload_url: string;
  path: string;
}

export interface KelasShiftCount {
  nama_kelas: string;
  shift: number;
  jumlah: number;
}

export interface CourseProgressInfo {
  course_id: number;
  jenis: string;
  is_open: boolean;
  selesai: number;
  sedang: number;
  belum: number;
}

export interface SesiAktifInfo {
  aktivasi_sesi_id: number;
  judul_sesi: string;
  nama_kelas: string;
  shift: number;
  courses: CourseProgressInfo[];
}

export interface AdminStatistik {
  total_mahasiswa: number;
  total_asisten: number;
  sudah_register: number;
  belum_register: number;
  per_kelas_shift: KelasShiftCount[];
  sesi_aktif: SesiAktifInfo[];
}

export interface Konfigurasi {
  key: string;
  value: string;
}

export interface AiModelConfig {
  effective: string;
  override: string;
  default: string;
}

export interface BulkResponse {
  total_processed: number;
  total_inserted: number;
  total_updated: number;
}

export interface AktivasiCourse {
  id: number;
  course_id: number;
  is_open: boolean;
  urutan: number;
  course?: { jenis: string; judul: string };
}

export interface AktivasiSesi {
  id: number;
  sesi_praktikum_id: number;
  kelas_id: number;
  shift: number;
  gelombang?: number | null;
  is_active: boolean;
  activated_at: string;
  token?: string;
  sesi?: { judul_sesi: string };
  kelas?: { nama_kelas: string };
  aktivasi_courses?: AktivasiCourse[];
}

export interface Susulan {
  id: number;
  mahasiswa_id: number;
  alasan: string;
  mahasiswa?: { nama: string; nim: string };
}

export interface PesertaAktivasi {
  mahasiswa_id: number;
  nim: string;
  nama: string;
  course_id: number;
  judul_course: string;
  status: string;
  waktu_mulai: string | null;
  waktu_selesai: string | null;
}

export interface PenilaianRekapItem {
  jawaban_id: number;
  mahasiswa_id: number;
  soal_terpilih_id: number;
  nama_mahasiswa: string;
  nim: string;
  soal_id: number;
  teks_soal: string;
  jenis_soal: string;
  poin: number;
  urutan: number;
  jawaban_teks: string;
  is_submitted: boolean;
  nilai: number | null;
  feedback: string | null;
}

export interface PenilaianRekapResponse {
  aktivasi_sesi_id: number;
  course_id: number;
  items: PenilaianRekapItem[];
}

export interface AiGradeTargets {
  jawaban_ids: number[];
  total: number;
}

export interface RekapJawabanItem {
  jawaban_id: number;
  nim: string;
  nama_mahasiswa: string;
  kelas_id: number;
  nama_kelas: string;
  sesi_praktikum_id: number;
  judul_sesi: string;
  course_id: number;
  judul_course: string;
  jenis_course: string;
  jenis_soal: string;
  teks_soal: string;
  poin_maksimal: number;
  jawaban_teks: string;
  is_submitted: boolean;
  waktu_submit: string | null;
  nilai: number | null;
  feedback: string | null;
}

export interface RekapJawabanResponse {
  items: RekapJawabanItem[];
  total: number;
  limited?: boolean;
}

export interface RekapKolom {
  key: string;
  label: string;
}

export interface RekapSel {
  pengerjaan_id: number;
  murni: number;
  keaktifan: number | null;
  total: number;
  edit_keaktifan: boolean;
  nilai_akhir: number | null;
}

export interface RekapMahasiswa {
  nim: string;
  nama: string;
  scores: Record<string, RekapSel>;
  total_score: number;
}

export interface RekapKelasResponse {
  columns: RekapKolom[];
  data: RekapMahasiswa[];
}

export interface AuditLog {
  id: number;
  user_id: number | null;
  nim: string;
  nama: string;
  role: string;
  action: string;
  description: string;
  ip_address: string;
  user_agent: string;
  created_at: string;
}

export interface AuditLogsResponse {
  logs: AuditLog[];
  total: number;
  page: number;
  limit: number;
}

export interface LaporanBesarTarget {
  id?: number;
  kelas_id: number;
  shift: number | null;
  buka_at: string | null;
  deadline_at: string | null;
  late_until: string | null;
  kelas?: { id: number; nama_kelas: string };
}

export interface LaporanBesar {
  id: number;
  judul: string;
  deskripsi: string;
  is_aktif: boolean;
  butuh_laporan: boolean;
  butuh_analisa: boolean;
  targets: LaporanBesarTarget[];
}

export interface PengumpulanLaporanBesar {
  mahasiswa_id: number;
  nim: string;
  nama: string;
  kelas_id: number;
  nama_kelas: string;
  status: 'belum' | 'terkumpul';
  sumber: 'web' | 'manual' | null;
  submitted_at: string | null;
  is_late: boolean;
  deadline_at: string | null;
  manual_catatan: string | null;
  override_alasan: string | null;
  override_oleh: string | null;
  laporan_url: string | null;
  analisa_url: string | null;
}

export interface BerkasUrl {
  url: string;
  filename: string;
}

export interface DownloadBatchResult {
  mhs_id: number;
  jenis: string;
  filename: string;
  url: string;
}

export type RolePermissions = Record<string, Record<string, boolean>>;
