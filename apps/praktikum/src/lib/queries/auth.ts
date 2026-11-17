import type { LayananLoginResponse } from '@algohub/api-client/praktikum';
import { createMutation } from '@tanstack/svelte-query';
import { api } from '#lib/api.ts';

// Halaman /layanan dulu memakai fetch mentah: tanpa Bearer dan tanpa pengalihan 401 ke login praktikum.
const layanan = { auth: false };

export function createLogoutMutation() {
  return createMutation(() => ({
    mutationFn: () => api.post<unknown>('/auth/logout'),
  }));
}

export function createLayananLoginMutation() {
  return createMutation(() => ({
    mutationFn: (body: { username: string; password: string }) =>
      api.post<LayananLoginResponse>('/auth/login', body, layanan),
  }));
}

export function createLayananRegisterMutation() {
  return createMutation(() => ({
    mutationFn: (body: {
      username: string;
      password: string;
      full_name: string;
      role: string;
      phone_number: string;
    }) => api.post<unknown>('/auth/register', body, layanan),
  }));
}
