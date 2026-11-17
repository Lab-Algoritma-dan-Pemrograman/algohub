export type RuntimeStatus = 'idle' | 'loading' | 'ready' | 'busy' | 'waiting-input' | 'failed';
export type RuntimeEvent =
  | { type: 'init' }
  | { type: 'ready' }
  | { type: 'init-failed'; retry: boolean }
  | { type: 'start' }
  | { type: 'need-input' }
  | { type: 'input' }
  | { type: 'finish' }
  | { type: 'reset' };
export function transition(status: RuntimeStatus, event: RuntimeEvent): RuntimeStatus {
  switch (event.type) {
    case 'reset':
      return 'idle';
    case 'init':
      return status === 'idle' || status === 'failed' ? 'loading' : status;
    case 'ready':
      return status === 'loading' ? 'ready' : status;
    case 'init-failed':
      if (status !== 'loading') return status;
      return event.retry ? 'loading' : 'failed';
    case 'start':
      return status === 'ready' ? 'busy' : status;
    case 'need-input':
      return status === 'busy' ? 'waiting-input' : status;
    case 'input':
      return status === 'waiting-input' ? 'busy' : status;
    case 'finish':
      return status === 'busy' || status === 'waiting-input' ? 'ready' : status;
  }
}
