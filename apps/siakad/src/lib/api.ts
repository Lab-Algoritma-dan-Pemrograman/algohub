import { PUBLIC_API_BASE_URL } from '$app/env/public';
import { createSiakadApi } from '@algohub/api-client/siakad';
import { session } from './session.ts';

export const {
  rpc,
  login,
  logout,
  verify,
  generateJwt,
  analyzeSchedule,
  sendPush,
  sendWa,
  syncElearning,
} = createSiakadApi({ baseUrl: PUBLIC_API_BASE_URL, getToken: () => session.token });
