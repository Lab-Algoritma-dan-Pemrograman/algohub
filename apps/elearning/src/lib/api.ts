import { PUBLIC_API_BASE_URL } from '$app/env/public';
import { createElearningApi } from '@algohub/api-client/elearning';
import { session } from './session.ts';
// Token ditolak server (kedaluwarsa atau dicabut): sesi lokal dihapus dan user dikembalikan ke
// landing. Cache query ikut bersih karena halaman dimuat ulang dari awal.
function onUnauthorized() {
  session.clear();
  if (location.pathname !== '/') location.assign('/');
}
export const api = createElearningApi({
  baseUrl: PUBLIC_API_BASE_URL,
  getToken: () => session.token,
  onUnauthorized,
});
