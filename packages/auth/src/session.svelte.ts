const TOKEN_KEY = 'token';
const USER_KEY = 'user';
const hasStorage = () => typeof localStorage !== 'undefined';
export interface SessionUser {
  id: number;
  role: string;
}
export class Session<U extends SessionUser> {
  // Pembaca template tidak boleh menulis $state (state_unsafe_mutation), jadi hasil baca
  // lazy disimpan di field biasa dan perubahan diumumkan lewat penghitung versi.
  #version = $state(0);
  #user: U | null | undefined = undefined;
  get user(): U | null {
    void this.#version;
    if (this.#user === undefined) {
      const raw = hasStorage() ? localStorage.getItem(USER_KEY) : null;
      this.#user = raw ? (JSON.parse(raw) as U) : null;
    }
    return this.#user;
  }
  get token(): string | null {
    void this.#version;
    return hasStorage() ? localStorage.getItem(TOKEN_KEY) : null;
  }
  get hasToken(): boolean {
    return !!this.token;
  }
  set(token: string, user: U): void {
    if (hasStorage()) {
      localStorage.setItem(TOKEN_KEY, token);
      localStorage.setItem(USER_KEY, JSON.stringify(user));
    }
    this.#user = user;
    this.#version++;
  }
  clear(): void {
    if (hasStorage()) {
      localStorage.removeItem(TOKEN_KEY);
      localStorage.removeItem(USER_KEY);
    }
    this.#user = null;
    this.#version++;
  }
}
