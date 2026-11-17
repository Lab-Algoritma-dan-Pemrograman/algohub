import { messages } from './messages';
import type { RunResult } from './runtime';

export function formatResult(result: RunResult): string {
  switch (result.status) {
    case 'compile-error':
      return `❌ \x1b[1;31mCompilation Error:\x1b[0m\n\x1b[31m${result.output || result.error || messages.compileFailed}\x1b[0m\r\n`;
    case 'ok': {
      const images = result.images.length
        ? `\x1b[2m[${result.images.length} grafik dirender]\x1b[0m\n`
        : '';
      return `${images}\r\n\x1b[1;32m✓ ${messages.done}\x1b[0m\n`;
    }
    case 'error':
      return `\r\n\x1b[33m⚠ ${result.error}\x1b[0m\r\n`;
    case 'timeout':
      return `\r\n\x1b[1;31m❌ ${messages.timeout}\x1b[0m\n`;
    case 'stopped':
      return `\r\n\x1b[1;31m⚠ ${messages.stopped}\x1b[0m\n`;
    case 'unavailable':
      return `\x1b[33m⚠ ${result.error}\x1b[0m\n`;
  }
}
