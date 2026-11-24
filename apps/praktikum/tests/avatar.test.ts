// @vitest-environment jsdom
import { describe, expect, it } from 'vitest';
import { avatarUrl, useAvatarFallback } from '../src/lib/avatar';

describe('fallback foto asisten', () => {
  it('membuat URL avatar yang aman', () => {
    const url = new URL(avatarUrl('Budi Santoso'));
    expect(url.hostname).toBe('ui-avatars.com');
    expect(url.searchParams.get('name')).toBe('Budi Santoso');
  });

  it('mengganti gambar rusak satu kali', () => {
    const image = document.createElement('img');
    useAvatarFallback({ currentTarget: image } as unknown as Event, 'Budi');
    expect(image.src).toContain('ui-avatars.com/api/');
    expect(image.dataset.fallbackApplied).toBe('true');
  });
});
