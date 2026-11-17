import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { RUN_TIMEOUT_MS } from './constants';
import { createRunTimer } from './run-timer';
describe('createRunTimer', () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });
  afterEach(() => {
    vi.useRealTimers();
  });
  it('habis tepat setelah 40000 ms', () => {
    const onTimeout = vi.fn();
    createRunTimer(RUN_TIMEOUT_MS, onTimeout).start();
    vi.advanceTimersByTime(RUN_TIMEOUT_MS - 1);
    expect(onTimeout).not.toHaveBeenCalled();
    vi.advanceTimersByTime(1);
    expect(onTimeout).toHaveBeenCalledOnce();
  });
  it('pause tidak memicu timeout', () => {
    const onTimeout = vi.fn();
    const timer = createRunTimer(RUN_TIMEOUT_MS, onTimeout);
    timer.start();
    vi.advanceTimersByTime(30_000);
    timer.pause();
    vi.advanceTimersByTime(RUN_TIMEOUT_MS * 2);
    expect(onTimeout).not.toHaveBeenCalled();
  });
  it('start setelah pause mulai lagi dari 0', () => {
    const onTimeout = vi.fn();
    const timer = createRunTimer(RUN_TIMEOUT_MS, onTimeout);
    timer.start();
    vi.advanceTimersByTime(30_000);
    timer.pause();
    timer.start();
    vi.advanceTimersByTime(RUN_TIMEOUT_MS - 1);
    expect(onTimeout).not.toHaveBeenCalled();
    vi.advanceTimersByTime(1);
    expect(onTimeout).toHaveBeenCalledOnce();
  });
  it('clear membatalkan timer', () => {
    const onTimeout = vi.fn();
    const timer = createRunTimer(RUN_TIMEOUT_MS, onTimeout);
    timer.start();
    timer.clear();
    vi.advanceTimersByTime(RUN_TIMEOUT_MS);
    expect(onTimeout).not.toHaveBeenCalled();
  });
});
