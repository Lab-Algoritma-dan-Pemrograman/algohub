import { requireToken } from '@algohub/auth';
import { LOGIN_PATH, PUBLIC_PATHS } from '#lib/auth.ts';
import { session } from '#lib/session.ts';
import type { LayoutLoad } from './$types';

export const load: LayoutLoad = ({ url }) => {
  if (!PUBLIC_PATHS.includes(url.pathname)) requireToken(session, LOGIN_PATH);
};
