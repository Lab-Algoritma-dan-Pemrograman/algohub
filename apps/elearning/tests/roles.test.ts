// @vitest-environment node
import { describe, expect, it } from 'vitest';
import { isAdminOrKordas, isStaff, normalizeRole, numericId } from '../src/lib/roles';
describe('normalizeRole', () => {
  it.each([
    ['admin', 'admin'],
    ['SuperAdmin', 'admin'],
    [' super admin ', 'admin'],
    ['kordas', 'kordas'],
    ['koordinator', 'kordas'],
    ['coordinator', 'kordas'],
    ['asisten', 'asisten'],
    ['laboran', 'asisten'],
    ['mahasiswa', 'praktikan'],
    ['praktikan', 'praktikan'],
    ['authenticated', 'praktikan'],
    [null, 'praktikan'],
    [undefined, 'praktikan'],
    ['', 'praktikan'],
  ])('normalizeRole(%j) = %j', (input, expected) => {
    expect(normalizeRole(input)).toBe(expected);
  });
});
describe('isAdminOrKordas dan isStaff', () => {
  it('admin dan kordas adalah admin, asisten hanya staf', () => {
    expect(isAdminOrKordas('admin')).toBe(true);
    expect(isAdminOrKordas('koordinator')).toBe(true);
    expect(isAdminOrKordas('asisten')).toBe(false);
    expect(isStaff('asisten')).toBe(true);
    expect(isStaff('mahasiswa')).toBe(false);
  });
});
describe('numericId', () => {
  it('stabil per NIM dan berbeda antar NIM', () => {
    expect(numericId('2022110083')).toBe(numericId('2022110083'));
    expect(numericId('2022110083')).not.toBe(numericId('2022110084'));
    expect(Number.isInteger(numericId('admin'))).toBe(true);
  });
});
