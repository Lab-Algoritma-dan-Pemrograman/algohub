import type {
  AuthResponse,
  ForgotPasswordResponse,
  LayananLoginResponse,
} from '@algohub/api-client/praktikum';
import { createMutation } from '@tanstack/svelte-query';
import { api } from '#lib/api.ts';
import { queryClient } from '#lib/query.ts';
import { session } from '#lib/session.ts';

// Halaman /layanan dulu memakai fetch mentah: tanpa Bearer dan tanpa pengalihan 401 ke login praktikum.
const layanan = { auth: false };

// Cache memori milik user lain tidak boleh ikut dipersist ke IndexedDB user yang baru masuk.
function startSession({ token, user }: AuthResponse) {
  if (session.user?.id !== user.id) queryClient.clear();
  session.set(token, user);
}

export function createLoginMutation() {
  return createMutation(() => ({
    mutationFn: (body: { identifier: string; password: string }) =>
      api.post<AuthResponse>('/auth/login', body),
    onSuccess: startSession,
  }));
}

export function createRegisterMutation() {
  return createMutation(() => ({
    mutationFn: (body: { nim: string; email: string; password: string }) =>
      api.post<AuthResponse>('/auth/register', body),
    onSuccess: startSession,
  }));
}

export function createForgotPasswordMutation() {
  return createMutation(() => ({
    mutationFn: (body: { email: string }) =>
      api.post<ForgotPasswordResponse>('/auth/forgot-password', body),
  }));
}

export function createResetPasswordMutation() {
  return createMutation(() => ({
    mutationFn: (body: { email: string; token: string; password: string }) =>
      api.post<unknown>('/auth/reset-password', body),
  }));
}

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
