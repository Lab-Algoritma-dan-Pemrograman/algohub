import { describe, expect, it } from 'vitest';
import { transition, type RuntimeEvent, type RuntimeStatus } from './state';
const statuses: RuntimeStatus[] = ['idle', 'loading', 'ready', 'busy', 'waiting-input', 'failed'];
const events: Record<string, RuntimeEvent> = {
  init: { type: 'init' },
  ready: { type: 'ready' },
  retry: { type: 'init-failed', retry: true },
  fail: { type: 'init-failed', retry: false },
  start: { type: 'start' },
  needInput: { type: 'need-input' },
  input: { type: 'input' },
  finish: { type: 'finish' },
  reset: { type: 'reset' },
};
const allowed: Partial<Record<RuntimeStatus, Partial<Record<keyof typeof events, RuntimeStatus>>>> =
  {
    idle: { init: 'loading' },
    failed: { init: 'loading' },
    loading: { ready: 'ready', retry: 'loading', fail: 'failed' },
    ready: { start: 'busy' },
    busy: { needInput: 'waiting-input', finish: 'ready' },
    'waiting-input': { input: 'busy', finish: 'ready' },
  };
describe('transition', () => {
  for (const status of statuses) {
    for (const [name, event] of Object.entries(events)) {
      const expected = name === 'reset' ? 'idle' : (allowed[status]?.[name] ?? status);
      it(`${status} + ${name} -> ${expected}`, () => {
        expect(transition(status, event)).toBe(expected);
      });
    }
  }
});
