import { Session, type SessionUser } from '@algohub/auth';
import type { AppRole } from './roles.ts';
// Hanya identitas ringan yang disimpan di localStorage; XP, level, dan sisanya hidup di memori
// (stores/app.svelte.ts) dan diambil ulang dari /profile/sync.
export interface ElearningSessionUser extends SessionUser {
  nim: string;
  nama: string;
  kelas: string;
  role: AppRole;
  jurusan: string | null;
  email: string | null;
  shift: string | null;
}
export const session = new Session<ElearningSessionUser>();
