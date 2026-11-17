export interface RunTimer {
  start(): void;
  pause(): void;
  clear(): void;
}
export function createRunTimer(ms: number, onTimeout: () => void): RunTimer {
  let handle: ReturnType<typeof setTimeout> | undefined;
  const stop = () => {
    clearTimeout(handle);
    handle = undefined;
  };
  return {
    start() {
      stop();
      handle = setTimeout(() => {
        handle = undefined;
        onTimeout();
      }, ms);
    },
    pause: stop,
    clear: stop,
  };
}
