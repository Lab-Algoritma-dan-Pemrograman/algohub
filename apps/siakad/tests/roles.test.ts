// @vitest-environment node
import { describe, expect, it } from 'vitest';
import { canonRole, isStaffRole, roleAllowed } from '../src/lib/roles';

describe('roles', () => {
  it.each([
    ['mahasiswa', 'praktikan'],
    ['MHS', 'praktikan'],
    [' peminjam ', 'penyewa'],
    ['kordas', 'koordinator'],
    ['admin', 'asisten'],
    ['asisten', 'asisten'],
    ['superadmin', 'superadmin'],
    [null, ''],
  ])('canonRole(%j) = %j', (input, expected) => {
    expect(canonRole(input)).toBe(expected);
  });

  it('koordinator selalu lolos, role lain dicek whitelist', () => {
    expect(roleAllowed('koordinator', [])).toBe(true);
    expect(roleAllowed('mahasiswa', ['praktikan'])).toBe(true);
    expect(roleAllowed('penyewa', ['praktikan', 'koordinator'])).toBe(false);
  });

  it('staff hanya koordinator dan asisten', () => {
    expect(isStaffRole('admin')).toBe(true);
    expect(isStaffRole('kordas')).toBe(true);
    expect(isStaffRole('mahasiswa')).toBe(false);
  });
});
