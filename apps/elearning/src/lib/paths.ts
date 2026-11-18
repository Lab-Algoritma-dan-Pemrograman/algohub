import type { RouteId } from '$app/types';

// Rute yang tidak butuh parameter. Rute dinamis (lesson/[id]) dipanggil lewat
// resolve('/(app)/lesson/[id]', { id }) supaya parameternya ikut diperiksa tipe.
export type StaticRouteId = Exclude<RouteId, '/(app)/lesson/[id]'>;
