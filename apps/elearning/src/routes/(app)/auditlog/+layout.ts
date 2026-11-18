import { guardRoute } from '#lib/auth.ts';
import type { LayoutLoad } from './$types';

// Pagar per route; (app)/+layout.ts sudah memastikan sesi, di sini peran dicek terhadap ROUTE_ACCESS
// dan yang tidak berhak dialihkan ke /dashboard.
export const load: LayoutLoad = async ({ url }) => {
  await guardRoute(url.pathname);
};
