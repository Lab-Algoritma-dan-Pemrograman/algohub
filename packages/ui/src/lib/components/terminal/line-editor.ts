export interface KeyModifiers {
  keyCode: number;
  ctrlKey: boolean;
  altKey: boolean;
  metaKey: boolean;
}

export interface LineEditResult {
  echo: string;
  line?: string;
}

export interface LineEditor {
  handle(key: string, ev: KeyModifiers): LineEditResult;
  reset(): void;
}

const ENTER = 13;
const BACKSPACE = 8;

export function createLineEditor(): LineEditor {
  let buffer = '';
  return {
    handle(key, ev) {
      if (ev.keyCode === ENTER) {
        const line = buffer;
        buffer = '';
        return { echo: '\r\n', line };
      }
      if (ev.keyCode === BACKSPACE) {
        if (buffer.length === 0) return { echo: '' };
        buffer = buffer.slice(0, -1);
        return { echo: '\b \b' };
      }
      if (ev.ctrlKey || ev.altKey || ev.metaKey) return { echo: '' };
      buffer += key;
      return { echo: key };
    },
    reset() {
      buffer = '';
    },
  };
}
