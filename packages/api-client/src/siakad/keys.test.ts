import { describe, expect, it } from 'vitest';
import { siakadKeys } from './keys';

describe('siakadKeys', () => {
  it('semua key berawalan siakad', () => {
    expect(siakadKeys.all).toEqual(['siakad']);
    expect(siakadKeys.session()).toEqual(['siakad', 'session']);
    expect(siakadKeys.elearningProgress('123')).toEqual(['siakad', 'sync-elearning', '123']);
  });

  it('rpcFn adalah prefix dari rpc sehingga invalidasi per fungsi mengenai semua params', () => {
    const prefix = siakadKeys.rpcFn('get_users_secure');
    const full = siakadKeys.rpc('get_users_secure', { p_limit: 10 });
    expect(prefix).toEqual(['siakad', 'rpc', 'get_users_secure']);
    expect(full).toEqual(['siakad', 'rpc', 'get_users_secure', { p_limit: 10 }]);
    expect(full.slice(0, prefix.length)).toEqual(prefix);
  });

  it('params kosong dinormalkan ke {}', () => {
    expect(siakadKeys.rpc('get_public_settings')).toEqual([
      'siakad',
      'rpc',
      'get_public_settings',
      {},
    ]);
  });
});
