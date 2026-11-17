import { requireRole } from '@algohub/auth';
import { session } from '#lib/session.ts';
import type { LayoutLoad } from './$types';

export const load: LayoutLoad = async ({ parent }) => {
  await parent();
  requireRole(session, ['koordinator'], '/praktikum/dashboard');
};
