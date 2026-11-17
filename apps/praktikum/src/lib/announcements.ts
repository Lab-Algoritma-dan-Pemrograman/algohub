export interface RescheduleCard {
  id: string;
  kelas: string;
  hari: string;
  jam: string;
  catatan: string;
}

export interface RecruitCard {
  id: string;
  judul: string;
  deskripsi: string;
  link?: string;
}

export interface SusulanCard {
  id: string;
  kelas: string;
  deskripsi: string;
}

export interface PlagiarismRow {
  nim: string;
  nama: string;
  keterangan: string;
  hari?: string;
  jam?: string;
}

export interface PlagiarismData {
  rows: PlagiarismRow[];
  updatedAt: string;
  catatan?: string;
}

export interface Announcements {
  reschedules: RescheduleCard[];
  recruit: RecruitCard[];
  susulan: SusulanCard[];
  plagiarism: PlagiarismData;
}

export type AnnouncementField = keyof Announcements;

export const ANNOUNCEMENT_FIELDS: readonly AnnouncementField[] = [
  'reschedules',
  'recruit',
  'susulan',
  'plagiarism',
];

// Key localStorage dan key konfigurasi backend memakai nama yang sama.
export const announcementKey = (field: AnnouncementField) => `ann_${field}`;

export function emptyAnnouncements(): Announcements {
  return { reschedules: [], recruit: [], susulan: [], plagiarism: { rows: [], updatedAt: '' } };
}

// Mengisi `target` field demi field: JSON rusak melempar setelah field sebelumnya terisi,
// sama dengan kode lama yang mem-parse dan meng-assign satu per satu.
export function fillAnnouncements(
  target: Announcements,
  read: (field: AnnouncementField) => string | null | undefined,
): Announcements {
  for (const field of ANNOUNCEMENT_FIELDS) {
    const raw = read(field);
    if (raw) target[field] = JSON.parse(raw);
  }
  return target;
}

export function readStoredAnnouncements(): Announcements {
  const result = emptyAnnouncements();
  try {
    fillAnnouncements(result, (field) => localStorage.getItem(announcementKey(field)));
  } catch {
    // Data cadangan rusak tidak boleh menggagalkan halaman; field yang sudah terbaca tetap dipakai.
  }
  return result;
}
