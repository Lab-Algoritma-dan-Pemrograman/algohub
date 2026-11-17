const all = ['siakad'] as const;

export const siakadKeys = {
  all,
  rpcFn: (fn: string) => [...all, 'rpc', fn] as const,
  rpc: (fn: string, params?: object) => [...all, 'rpc', fn, params ?? {}] as const,
  session: () => [...all, 'session'] as const,
  elearningProgress: (nim: string) => [...all, 'sync-elearning', nim] as const,
};
