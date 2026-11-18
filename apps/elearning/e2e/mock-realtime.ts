import type { Page, WebSocketRoute } from '@playwright/test';

// Format frame Phoenix VSN 2.0.0 (serializer @supabase/realtime-js): [join_ref, ref, topic, event, payload].
type Frame = [string | null, string | null, string, string, Record<string, unknown>];

interface Binding {
  id: number;
  joinRef: string | null;
  topic: string;
  event: string;
  table?: string;
  filter?: string;
}

// Host .test tidak pernah ada di jaringan; WebSocket ke sana dijawab mock ini.
export const SUPABASE_HOST = 'supabase.elearning.test';

export class MockRealtime {
  private sockets: WebSocketRoute[] = [];
  private bindings: Binding[] = [];
  private nextId = 1;

  async install(page: Page): Promise<void> {
    await page.routeWebSocket(
      new RegExp(`${SUPABASE_HOST.replace(/\./g, '\\.')}/realtime`),
      (ws) => {
        this.sockets.push(ws);
        ws.onMessage((raw) => this.onMessage(ws, String(raw)));
      },
    );
  }

  channels(): string[] {
    return [...new Set(this.bindings.map((b) => b.topic.replace(/^realtime:/, '')))];
  }

  private send(ws: WebSocketRoute, frame: Frame) {
    ws.send(JSON.stringify(frame));
  }

  private onMessage(ws: WebSocketRoute, raw: string) {
    const [joinRef, ref, topic, event, payload] = JSON.parse(raw) as Frame;
    if (event === 'heartbeat' || event === 'access_token') {
      this.send(ws, [null, ref, topic, 'phx_reply', { status: 'ok', response: {} }]);
      return;
    }
    if (event === 'phx_leave') {
      this.bindings = this.bindings.filter((b) => b.topic !== topic);
      this.send(ws, [joinRef, ref, topic, 'phx_reply', { status: 'ok', response: {} }]);
      return;
    }
    if (event !== 'phx_join') return;
    const config = (payload.config ?? {}) as { postgres_changes?: Record<string, string>[] };
    const changes = (config.postgres_changes ?? []).map((filter) => {
      const binding: Binding = {
        id: this.nextId++,
        joinRef,
        topic,
        event: filter.event,
        table: filter.table,
        filter: filter.filter,
      };
      this.bindings.push(binding);
      return { ...filter, id: binding.id };
    });
    this.send(ws, [
      joinRef,
      ref,
      topic,
      'phx_reply',
      { status: 'ok', response: { postgres_changes: changes } },
    ]);
  }

  async waitForChannel(page: Page, name: string): Promise<void> {
    for (let i = 0; i < 100 && !this.channels().includes(name); i++) {
      await page.waitForTimeout(100);
    }
    if (!this.channels().includes(name)) throw new Error(`channel ${name} tidak pernah join`);
  }

  // Mengirim frame postgres_changes ke setiap channel yang berlangganan tabel tersebut.
  emit(
    table: string,
    type: 'INSERT' | 'UPDATE' | 'DELETE',
    record: Record<string, unknown>,
    oldRecord: Record<string, unknown> = {},
  ): number {
    const matched = this.bindings.filter(
      (b) => b.table === table && (b.event === '*' || b.event === type),
    );
    for (const ws of this.sockets) {
      for (const b of matched) {
        this.send(ws, [
          b.joinRef,
          null,
          b.topic,
          'postgres_changes',
          {
            ids: [b.id],
            data: {
              type,
              table,
              schema: 'public',
              record,
              old_record: oldRecord,
              columns: [],
              commit_timestamp: new Date().toISOString(),
              errors: null,
            },
          },
        ]);
      }
    }
    return matched.length;
  }
}
