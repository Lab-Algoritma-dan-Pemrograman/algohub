import { describe, expect, it } from 'vitest';
import { prepareCSource } from './c-source';
const unbuffer = '\n    setbuf(stdout, NULL);\n    setbuf(stdin, NULL);';
describe('prepareCSource', () => {
  it('menyisipkan setbuf setelah pembuka main bila stdio.h sudah ada', () => {
    const code = '#include <stdio.h>\nint main(void) {\n  return 0;\n}';
    expect(prepareCSource(code)).toBe(
      `#include <stdio.h>\nint main(void) {${unbuffer}\n  return 0;\n}`,
    );
  });
  it('menambah #include <stdio.h> bila belum ada', () => {
    const code = 'int main() {\n  return 0;\n}';
    expect(prepareCSource(code)).toBe(
      `#include <stdio.h>\nint main() {${unbuffer}\n  return 0;\n}`,
    );
  });
  it('tidak mengubah kode tanpa main', () => {
    const code = 'int tambah(int a, int b) { return a + b; }';
    expect(prepareCSource(code)).toBe(code);
  });
});
