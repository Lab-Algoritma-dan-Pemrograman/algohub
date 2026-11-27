import {
  praktikumKeys,
  type AnnouncementsResponse,
  type Jadwal,
  type JadwalConfig,
  type KelasMahasiswaInfo,
  type ModulInfo,
  type Pedoman,
  type User,
} from '@algohub/api-client/praktikum';
import { queryOptions } from '@tanstack/svelte-query';
import {
  emptyAnnouncements,
  fillAnnouncements,
  readStoredAnnouncements,
} from '#lib/announcements.ts';
import { api } from '#lib/api.ts';

// Pengganti cache-control 300 detik dari +page.ts lama halaman /info dan /info/jadwal.
export const INFO_PAGE_STALE_TIME = 5 * 60_000;

const persist = { persist: true };

export const infoQueries = {
  asisten: () =>
    queryOptions({
      queryKey: praktikumKeys.info.asisten,
      queryFn: () => api.get<User[]>('/info/asisten'),
      meta: persist,
    }),
  jadwal: () =>
    queryOptions({
      queryKey: praktikumKeys.info.jadwal,
      queryFn: () => api.get<Jadwal[]>('/info/jadwal'),
      meta: persist,
    }),
  jadwalConfig: () =>
    queryOptions({
      queryKey: praktikumKeys.info.jadwalConfig,
      queryFn: () => api.get<JadwalConfig>('/info/jadwal/config'),
      meta: persist,
    }),
  kelasMahasiswa: (kelasId: number) =>
    queryOptions({
      queryKey: praktikumKeys.info.kelasMahasiswa(kelasId),
      queryFn: () => api.get<KelasMahasiswaInfo>(`/info/kelas/${kelasId}/mahasiswa`),
      meta: persist,
    }),
  laporan: () =>
    queryOptions({
      queryKey: praktikumKeys.info.laporan,
      queryFn: () => api.get<Pedoman[]>('/info/laporan'),
      meta: persist,
    }),
  modul: () =>
    queryOptions({
      queryKey: praktikumKeys.info.modul,
      queryFn: () => api.get<ModulInfo>('/info/modul'),
      meta: persist,
    }),
  announcements: () =>
    queryOptions({
      queryKey: praktikumKeys.info.announcements,
      // ?t= ada di URL, bukan di key, agar cache HTTP tertembus seperti app lama tanpa
      // memecah cache Query. API gagal jatuh ke salinan localStorage yang disimpan admin.
      queryFn: async () => {
        try {
          const raw = await api.get<AnnouncementsResponse | null>(
            `/info/announcements?t=${Date.now()}`,
          );
          return fillAnnouncements(emptyAnnouncements(), (field) => raw?.[field]);
        } catch {
          return readStoredAnnouncements();
        }
      },
      meta: persist,
    }),
};
