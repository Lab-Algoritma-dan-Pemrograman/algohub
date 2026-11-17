export const VIRTUAL_THRESHOLD = 100;
export function shouldVirtualize(count: number, threshold = VIRTUAL_THRESHOLD): boolean {
  return count > threshold;
}
export function virtualPadding(
  items: readonly { start: number; end: number }[],
  totalSize: number,
): { top: number; bottom: number } {
  if (items.length === 0) return { top: 0, bottom: 0 };
  return { top: items[0].start, bottom: totalSize - items[items.length - 1].end };
}
