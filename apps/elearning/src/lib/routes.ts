import type { AppRole } from './roles.ts';
export interface RouteAccess {
  roles: readonly AppRole[];
}
// Disalin dari App.tsx lama: admin dan auditlog hanya admin/kordas, monitoring juga asisten.
// Route lain cukup login; role yang ditolak jatuh ke /dashboard seperti fallback App.tsx.
export const ROUTE_ACCESS: Record<string, RouteAccess> = {
  '/admin': { roles: ['admin', 'kordas'] },
  '/auditlog': { roles: ['admin', 'kordas'] },
  '/monitoring': { roles: ['admin', 'kordas', 'asisten'] },
};
export const FALLBACK_ROUTE = '/dashboard';
export function accessFor(pathname: string): RouteAccess | undefined {
  const key = Object.keys(ROUTE_ACCESS).find(
    (path) => pathname === path || pathname.startsWith(`${path}/`),
  );
  return key ? ROUTE_ACCESS[key] : undefined;
}
