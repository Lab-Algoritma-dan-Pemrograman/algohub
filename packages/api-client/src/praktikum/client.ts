import { createApi } from '../client';
export interface PraktikumApiOptions {
  baseUrl: string;
  onUnauthorized?: () => void;
}
export function createPraktikumApi({ baseUrl, onUnauthorized }: PraktikumApiOptions) {
  return createApi({
    baseUrl,
    prefix: '/api/praktikum',
    getToken: () => (typeof localStorage === 'undefined' ? null : localStorage.getItem('token')),
    onUnauthorized,
  });
}
