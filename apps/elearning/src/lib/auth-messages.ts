import { ApiError } from '@algohub/api-client';
export const NETWORK_ERROR = 'Tidak bisa menghubungi server. Periksa koneksi lalu coba lagi.';
export const LOGIN_REQUIRED = 'NIM/email dan password wajib diisi.';
// Pesan dari API dipakai apa adanya (mis. kode SSO sudah dipakai); selain itu berarti request
// tidak sampai ke server.
export function authErrorMessage(error: unknown, fallback: string): string {
  if (error instanceof ApiError) return error.message || fallback;
  return NETWORK_ERROR;
}
// POST /api/elearning/sso mengalihkan kegagalan ke /?error=<kode>.
const SSO_REDIRECT_ERRORS: Record<string, string> = {
  invalid_token: 'Token dari portal tidak valid atau sudah kedaluwarsa. Silakan masuk lagi.',
  missing_fields: 'Data akun dari portal tidak lengkap. Silakan masuk lagi.',
  server_error: 'Terjadi kesalahan server saat masuk lewat portal. Coba lagi.',
};
export function ssoRedirectMessage(code: string | null): string | null {
  return code ? (SSO_REDIRECT_ERRORS[code] ?? null) : null;
}
