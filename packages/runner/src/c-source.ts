const MAIN_REGEX = /\bmain\s*\([^)]*\s*\)\s*\{/;
const STDIO_REGEX = /#\s*include\s*[<"]stdio\.h[>"]/i;
// Tanpa setbuf, printf sebelum scanf tertahan di buffer dan prompt baru muncul
// setelah input dikirim.
export function prepareCSource(code: string): string {
  if (!MAIN_REGEX.test(code)) return code;
  const injected = code.replace(
    MAIN_REGEX,
    '$&\n    setbuf(stdout, NULL);\n    setbuf(stdin, NULL);',
  );
  // setbuf butuh deklarasi dari <stdio.h>. Tanpa include itu, suntikan justru
  // bikin error ("setbuf was not declared") pada kode pengguna yang sah.
  return STDIO_REGEX.test(code) ? injected : `#include <stdio.h>\n${injected}`;
}
