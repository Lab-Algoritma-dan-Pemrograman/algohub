import { PUBLIC_API_BASE_URL } from '$app/env/public';
import { createPraktikumApi } from '@algohub/api-client/praktikum';
import { expireSession } from './auth.ts';

export const api = createPraktikumApi({
  baseUrl: PUBLIC_API_BASE_URL,
  onUnauthorized: expireSession,
});
