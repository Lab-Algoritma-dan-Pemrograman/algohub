import { render, screen } from '@testing-library/svelte';
import { describe, expect, it, vi } from 'vitest';
import Countdown from './Countdown.svelte';

describe('Countdown Component', () => {
  it('renders correctly with default values', () => {
    const futureDate = new Date(Date.now() + 120000).toISOString();
    render(Countdown, { deadline: futureDate });
    const span = screen.getByText(/\d{2}:\d{2}/);
    expect(span).toBeInTheDocument();
  });

  it('fires onExpire when deadline is reached', () => {
    vi.useFakeTimers();
    const onExpire = vi.fn();
    const pastDate = new Date(Date.now() - 1000).toISOString();
    render(Countdown, { deadline: pastDate, onExpire });
    vi.advanceTimersByTime(1000);
    expect(onExpire).toHaveBeenCalled();
    vi.useRealTimers();
  });
});
