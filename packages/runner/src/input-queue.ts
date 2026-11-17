import type { InputChannel } from './input-channel';
type Channel = Pick<InputChannel, 'fits' | 'send'>;
// SAB hanya punya satu slot. Baris yang diketik sebelum program meminta input
// diantre supaya ketikan cepat tidak saling menimpa.
export class InputQueue {
  readonly #channel: Channel;
  #pending: string[] = [];
  #waiting = false;
  constructor(channel: Channel) {
    this.#channel = channel;
  }
  get waiting(): boolean {
    return this.#waiting;
  }
  push(line: string): boolean {
    if (this.#waiting) {
      if (!this.#channel.send(line)) return false;
      this.#waiting = false;
      return true;
    }
    if (!this.#channel.fits(line)) return false;
    this.#pending.push(line);
    return true;
  }
  needInput(): void {
    const next = this.#pending.shift();
    if (next === undefined) {
      this.#waiting = true;
      return;
    }
    this.#channel.send(next);
    this.#waiting = false;
  }
  clear(): void {
    this.#pending = [];
    this.#waiting = false;
  }
}
