import type { Language } from '../../src';
export interface Sample {
  id: string;
  label: string;
  language: Language;
  code: string;
}
export const samples: Sample[] = [
  {
    id: 'c-two-scanf',
    label: 'C: dua scanf',
    language: 'c',
    code: `#include <stdio.h>

int main() {
    int a, b;
    printf("Masukkan a: ");
    scanf("%d", &a);
    printf("Masukkan b: ");
    scanf("%d", &b);
    printf("Hasil: %d\\n", a + b);
    return 0;
}
`,
  },
  {
    id: 'c-compile-error',
    label: 'C: error kompilasi',
    language: 'c',
    code: `#include <stdio.h>

int main() {
    int x = ;
    return 0;
}
`,
  },
  {
    id: 'c-infinite-loop',
    label: 'C: infinite loop',
    language: 'c',
    code: `#include <stdio.h>

int main() {
    volatile unsigned long n = 0;
    printf("Mulai loop\\n");
    while (1) {
        n++;
    }
    return 0;
}
`,
  },
  {
    id: 'python-input',
    label: 'Python: input()',
    language: 'python',
    code: `def tanya(pesan):
    return input(pesan)

nama = input("Nama: ")
kota = tanya("Kota: ")
print(f"Halo {nama} dari {kota}")
`,
  },
];
