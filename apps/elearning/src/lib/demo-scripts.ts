import type { Language } from '@algohub/runner';
export interface DemoScript {
  name: string;
  code: string;
}
export const DEMO_SCRIPTS: Record<Language, DemoScript[]> = {
  python: [
    {
      name: 'Sapa',
      code: 'nama = input("Siapa nama kamu? ")\nprint(f"Halo, {nama}! Selamat datang di E-Learning!")\numur = input("Berapa umur kamu? ")\nprint(f"Keren, umur {umur} tahun! Semangat belajar Python!")',
    },
    {
      name: 'Ganjil Genap',
      code: 'angka = input("Masukkan angka: ")\nn = int(angka)\nif n % 2 == 0:\n    print(f"{n} adalah bilangan Genap!")\nelse:\n    print(f"{n} adalah bilangan Ganjil!")',
    },
    {
      name: 'Bintang',
      code: 'baris = input("Masukkan jumlah baris: ")\nfor i in range(1, int(baris) + 1):\n    print("*" * i)',
    },
    {
      name: 'Grafik',
      code: 'import matplotlib.pyplot as plt\n\nbulan = ["Jan", "Feb", "Mar", "Apr", "Mei"]\npenjualan = [10, 25, 18, 30, 42]\n\nplt.plot(bulan, penjualan, marker="o", color="#800000", linewidth=2)\nplt.title("Grafik Penjualan Bulanan")\nplt.xlabel("Bulan")\nplt.ylabel("Penjualan (Juta)")\nplt.grid(True)\n\nplt.show()\nprint("Grafik berhasil ditampilkan!")',
    },
  ],
  c: [
    {
      name: 'Sapa',
      code: '#include <stdio.h>\n\nint main() {\n    char nama[50];\n    int umur;\n    printf("Siapa nama kamu? ");\n    scanf("%49s", nama);\n    printf("Halo, %s!\\n", nama);\n    printf("Berapa umur kamu? ");\n    scanf("%d", &umur);\n    printf("Umur kamu %d tahun!\\n", umur);\n    return 0;\n}',
    },
    {
      name: 'Ganjil Genap',
      code: '#include <stdio.h>\n\nint main() {\n    int n;\n    printf("Masukkan angka: ");\n    scanf("%d", &n);\n    if (n % 2 == 0) {\n        printf("%d adalah Genap!\\n", n);\n    } else {\n        printf("%d adalah Ganjil!\\n", n);\n    }\n    return 0;\n}',
    },
    {
      name: 'Bintang',
      code: '#include <stdio.h>\n\nint main() {\n    int baris, i, j;\n    printf("Masukkan jumlah baris: ");\n    scanf("%d", &baris);\n    for (i = 1; i <= baris; i++) {\n        for (j = 1; j <= i; j++) {\n            printf("*");\n        }\n        printf("\\n");\n    }\n    return 0;\n}',
    },
  ],
};
export const DEFAULT_CODE: Record<Language, string> = {
  python:
    '# Tulis kode Python Anda di sini\nnama = input("Siapa nama kamu? ")\nprint(f"Halo, {nama}! Selamat datang!")',
  c: '#include <stdio.h>\n\nint main() {\n    char nama[100];\n    printf("Siapa nama kamu? ");\n    scanf("%[^\\n]", nama);\n    printf("Halo, %s! Selamat datang!\\n", nama);\n    return 0;\n}',
};
