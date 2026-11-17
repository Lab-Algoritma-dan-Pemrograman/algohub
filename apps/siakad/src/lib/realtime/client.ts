import { PUBLIC_SUPABASE_PUBLISHABLE_KEY, PUBLIC_SUPABASE_URL } from '$app/env/public';
import {
  createClient,
  type RealtimePostgresChangesPayload,
  type SupabaseClient,
} from '@supabase/supabase-js';

export type ChangePayload = RealtimePostgresChangesPayload<Record<string, unknown>>;

export interface ChangeFilter {
  event?: '*' | 'INSERT' | 'UPDATE' | 'DELETE';
  table: string;
  filter?: string;
}

let client: SupabaseClient | null = null;

// supabase-js di frontend hanya untuk Realtime; data, auth, dan storage lewat backend Go.
function realtime(): SupabaseClient {
  client ??= createClient(PUBLIC_SUPABASE_URL, PUBLIC_SUPABASE_PUBLISHABLE_KEY, {
    auth: { persistSession: false, autoRefreshToken: false },
  });
  return client;
}

export function subscribe(
  channelName: string,
  { event = '*', table, filter }: ChangeFilter,
  onChange: (payload: ChangePayload) => void,
): () => void {
  // Dev lokal dan CI tanpa env Supabase tetap jalan; tampilan hanya tidak diperbarui otomatis.
  if (!PUBLIC_SUPABASE_URL) return () => {};
  const supabase = realtime();
  const channel = supabase
    .channel(channelName)
    .on(
      'postgres_changes',
      { event, schema: 'public', table, ...(filter ? { filter } : {}) },
      onChange,
    )
    .subscribe();
  return () => {
    void supabase.removeChannel(channel);
  };
}
