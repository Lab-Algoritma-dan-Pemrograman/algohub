export interface TerminalDemoScript {
  name: string;
  code: string;
}

// Skrip Python untuk /terminal-demo, disalin dari TerminalDemo.tsx lama.
export const TERMINAL_DEMO_SCRIPTS: TerminalDemoScript[] = [
  {
    name: '👋 Sapa Pengguna',
    code: `nama = input("Siapa nama kamu? ")
print(f"Halo, {nama}! Selamat datang di E-Learning!")
umur = input("Berapa umur kamu? ")
print(f"Wah, {nama} berusia {umur} tahun. Semangat belajar!")`,
  },
  {
    name: '🔢 Kalkulator',
    code: `print("=== Kalkulator Sederhana ===")
a = input("Masukkan angka pertama: ")
b = input("Masukkan angka kedua: ")
op = input("Operasi (+, -, *, /): ")
x, y = float(a), float(b)
if op == "+":
    print(f"{x} + {y} = {x + y}")
elif op == "-":
    print(f"{x} - {y} = {x - y}")
elif op == "*":
    print(f"{x} × {y} = {x * y}")
elif op == "/":
    print(f"{x} ÷ {y} = {x / y}" if y != 0 else "Error: bagi nol!")
else:
    print("Operasi tidak dikenal.")`,
  },
  {
    name: '🎲 Tebak Angka',
    code: `import random
angka = random.randint(1, 10)
print("Saya memikirkan angka 1-10!")
tebakan = input("Tebakan kamu: ")
if int(tebakan) == angka:
    print("🎉 Benar! Kamu hebat!")
else:
    print(f"❌ Salah! Jawabannya: {angka}")`,
  },
  {
    name: '📊 Grafik Matplotlib',
    code: `import matplotlib.pyplot as plt
print("Membuat grafik penjualan...")
bulan = ["Jan", "Feb", "Mar", "Apr", "Mei"]
penjualan = [12, 19, 15, 25, 30]
plt.plot(bulan, penjualan, marker='o', color='#800000', linewidth=2)
plt.title("Grafik Penjualan Bulanan")
plt.xlabel("Bulan")
plt.ylabel("Penjualan (Juta)")
plt.grid(True)
plt.show()
print("Grafik berhasil dibuat!")`,
  },
];
