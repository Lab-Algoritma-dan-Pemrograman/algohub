import { redirect } from '@sveltejs/kit';
import { loadSession } from '#lib/auth.ts';
import type { PageLoad } from './$types';
export const load: PageLoad = async () => {
  if (await loadSession()) redirect(307, '/dashboard');
};
