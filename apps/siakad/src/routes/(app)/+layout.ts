import { guardRoute } from '#lib/auth.ts';
import type { LayoutLoad } from './$types';

export const load: LayoutLoad = async ({ url }) => ({ auth: await guardRoute(url.pathname) });
