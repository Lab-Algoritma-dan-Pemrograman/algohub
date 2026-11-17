import { describe, expect, it } from 'vitest';
import { praktikumKeys } from './keys';
type Key = readonly unknown[];
type Node = { [name: string]: Key | ((...args: never[]) => Key) | Node };
function collect(node: Node, parent: Key, out: { key: Key; parent: Key }[]) {
  const own = (node.all as Key | undefined) ?? parent;
  for (const [name, value] of Object.entries(node)) {
    if (typeof value === 'function') {
      const fn = value as (...args: unknown[]) => Key;
      out.push({ key: fn(...Array.from({ length: fn.length }, (_, i) => i + 1)), parent: own });
    } else if (Array.isArray(value)) {
      out.push({ key: value, parent: name === 'all' ? parent : own });
    } else {
      collect(value as Node, own, out);
    }
  }
  return out;
}
const keys = collect(praktikumKeys as unknown as Node, [], []);
describe('praktikumKeys', () => {
  it('berakar di praktikum', () => {
    expect(praktikumKeys.all).toEqual(['praktikum']);
  });
  it('setiap key unik', () => {
    const serialized = keys.map(({ key }) => JSON.stringify(key));
    expect(new Set(serialized).size).toBe(serialized.length);
  });
  it('setiap key anak diawali key induknya', () => {
    for (const { key, parent } of keys) {
      expect(key.slice(0, parent.length)).toEqual(parent);
      expect(key.length).toBeGreaterThan(parent.length);
    }
  });
});
