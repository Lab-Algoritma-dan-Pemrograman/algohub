// Port lib/soundEffects.ts lama: efek suara disintesis dengan Web Audio API (tanpa berkas audio),
// jadi tidak ada aset yang perlu dicache atau diunduh.
let audioContext: AudioContext | null = null;

function ctx(): AudioContext {
  if (!audioContext || audioContext.state === 'closed') audioContext = new AudioContext();
  if (audioContext.state === 'suspended') void audioContext.resume();
  return audioContext;
}

// Gema buatan (impulse response noise yang meluruh) supaya ledakan terdengar di ruang terbuka.
function makeReverb(ac: AudioContext, duration = 1.2, decay = 2.5): ConvolverNode {
  const rate = ac.sampleRate;
  const length = Math.floor(rate * duration);
  const impulse = ac.createBuffer(2, length, rate);
  for (let ch = 0; ch < 2; ch++) {
    const data = impulse.getChannelData(ch);
    for (let i = 0; i < length; i++)
      data[i] = (Math.random() * 2 - 1) * Math.pow(1 - i / length, decay);
  }
  const conv = ac.createConvolver();
  conv.buffer = impulse;
  return conv;
}

function note(
  ac: AudioContext,
  type: OscillatorType,
  freq: number,
  amp: number,
  t: number,
  dur: number,
) {
  const o = ac.createOscillator();
  const g = ac.createGain();
  o.type = type;
  o.frequency.setValueAtTime(freq, t);
  g.gain.setValueAtTime(amp, t);
  g.gain.exponentialRampToValueAtTime(0.0001, t + dur);
  o.connect(g).connect(ac.destination);
  o.start(t);
  o.stop(t + dur + 0.02);
}

function glide(
  ac: AudioContext,
  type: OscillatorType,
  f0: number,
  f1: number,
  amp: number,
  t: number,
  dur: number,
) {
  const o = ac.createOscillator();
  const g = ac.createGain();
  o.type = type;
  o.frequency.setValueAtTime(f0, t);
  o.frequency.exponentialRampToValueAtTime(f1, t + dur);
  g.gain.setValueAtTime(amp, t);
  g.gain.exponentialRampToValueAtTime(0.0001, t + dur);
  o.connect(g).connect(ac.destination);
  o.start(t);
  o.stop(t + dur + 0.02);
}

function noiseBurst(
  ac: AudioContext,
  seconds: number,
  filter: { type: BiquadFilterType; from: number; to?: number; q?: number },
  peak: number,
  attack: number,
  decay: number,
  t: number,
  outputs: AudioNode[],
) {
  const length = Math.floor(ac.sampleRate * seconds);
  const buffer = ac.createBuffer(1, length, ac.sampleRate);
  const data = buffer.getChannelData(0);
  for (let i = 0; i < length; i++) data[i] = Math.random() * 2 - 1;
  const src = ac.createBufferSource();
  src.buffer = buffer;
  const biquad = ac.createBiquadFilter();
  biquad.type = filter.type;
  biquad.frequency.setValueAtTime(filter.from, t);
  if (filter.to) biquad.frequency.exponentialRampToValueAtTime(filter.to, t + decay - 0.05);
  if (filter.q) biquad.Q.setValueAtTime(filter.q, t);
  const g = ac.createGain();
  g.gain.setValueAtTime(0, t);
  g.gain.linearRampToValueAtTime(peak, t + attack);
  g.gain.exponentialRampToValueAtTime(0.001, t + decay);
  src.connect(biquad).connect(g);
  for (const out of outputs) g.connect(out);
  src.start(t);
  src.stop(t + seconds);
}

// Satu ledakan petasan: dentuman sub, boom rendah, retakan menengah dan tinggi, lalu serpihan
// yang berkilau. Dry dan reverb dicampur supaya ada kesan gema.
function boom(ac: AudioContext, t: number, power = 1, pitch: 'low' | 'mid' | 'high' = 'mid') {
  const reverb = makeReverb(ac, 1.4, 2.8);
  const dry = ac.createGain();
  dry.gain.setValueAtTime(0.7, t);
  dry.connect(ac.destination);
  const wet = ac.createGain();
  wet.gain.setValueAtTime(0.45, t);
  wet.connect(reverb);
  reverb.connect(ac.destination);

  const subFreq = pitch === 'low' ? 55 : pitch === 'high' ? 110 : 75;
  const sub = ac.createOscillator();
  const subGain = ac.createGain();
  sub.type = 'sine';
  sub.frequency.setValueAtTime(subFreq * 1.8, t);
  sub.frequency.exponentialRampToValueAtTime(subFreq * 0.2, t + 0.22);
  subGain.gain.setValueAtTime(0, t);
  subGain.gain.linearRampToValueAtTime(0.9 * power, t + 0.003);
  subGain.gain.exponentialRampToValueAtTime(0.001, t + 0.35);
  sub.connect(subGain);
  subGain.connect(dry);
  subGain.connect(wet);
  sub.start(t);
  sub.stop(t + 0.4);

  noiseBurst(ac, 0.5, { type: 'lowpass', from: 400, to: 80 }, 0.8 * power, 0.004, 0.45, t, [
    dry,
    wet,
  ]);
  noiseBurst(ac, 0.3, { type: 'bandpass', from: 2200, q: 0.7 }, 0.55 * power, 0.002, 0.28, t, [
    dry,
    wet,
  ]);
  noiseBurst(ac, 0.15, { type: 'highpass', from: 6000 }, 0.35 * power, 0.001, 0.12, t, [dry]);

  const shimmerBase = pitch === 'low' ? 900 : pitch === 'high' ? 1800 : 1300;
  for (let i = 0; i < Math.round(8 * power); i++) {
    const f = shimmerBase * (0.6 + Math.random() * 1.4);
    const offset = 0.05 + Math.random() * 0.55;
    const dur = 0.2 + Math.random() * 0.5;
    const amp = 0.04 * power * (0.5 + Math.random());
    const o = ac.createOscillator();
    const g = ac.createGain();
    o.type = 'sine';
    o.frequency.setValueAtTime(f, t + offset);
    o.frequency.exponentialRampToValueAtTime(f * 0.3, t + offset + dur);
    g.gain.setValueAtTime(amp, t + offset);
    g.gain.exponentialRampToValueAtTime(0.0001, t + offset + dur);
    o.connect(g);
    g.connect(wet);
    g.connect(dry);
    o.start(t + offset);
    o.stop(t + offset + dur + 0.02);
  }
}

// Audio hanya hiasan: browser yang menolak AudioContext (belum ada interaksi, mode hemat)
// tidak boleh membuat alur belajar gagal.
function safely(play: (ac: AudioContext, t: number) => void): void {
  try {
    const ac = ctx();
    play(ac, ac.currentTime);
  } catch {
    return;
  }
}

export function playAchievementSound(): void {
  safely((ac, t) => {
    boom(ac, t, 1.1, 'mid');
    [523.25, 659.25, 783.99, 1046.5].forEach((f, i) => {
      note(ac, 'sine', f, 0.16, t + 0.22 + i * 0.09, 0.28);
      note(ac, 'triangle', f * 2, 0.05, t + 0.22 + i * 0.09, 0.22);
    });
    boom(ac, t + 0.65, 0.65, 'high');
  });
}

export function playLevelUpSound(): void {
  safely((ac, t) => {
    boom(ac, t, 1.3, 'low');
    [196.0, 246.94, 293.66, 392.0, 493.88].forEach((f) => {
      note(ac, 'triangle', f, 0.13, t + 0.32, 0.8);
      note(ac, 'sine', f * 2, 0.04, t + 0.32, 0.6);
    });
    boom(ac, t + 0.7, 1.0, 'mid');
    glide(ac, 'sine', 300, 1400, 0.06, t + 1.05, 0.22);
    boom(ac, t + 1.28, 0.9, 'high');
    [1047, 1319, 1568, 1760, 2093, 2637].forEach((f, i) => {
      note(ac, 'sine', f, 0.06, t + 1.45 + i * 0.07, 0.22);
    });
  });
}

export function playQuizCorrectSound(): void {
  safely((ac, t) => {
    note(ac, 'sine', 659.25, 0.2, t, 0.18);
    note(ac, 'triangle', 659.25, 0.07, t, 0.15);
    note(ac, 'sine', 880.0, 0.22, t + 0.16, 0.3);
    note(ac, 'triangle', 880.0, 0.07, t + 0.16, 0.26);
    note(ac, 'sine', 1760.0, 0.05, t + 0.3, 0.14);
  });
}

export function playQuizWrongSound(): void {
  safely((ac, t) => {
    glide(ac, 'triangle', 466, 392, 0.14, t, 0.18);
    glide(ac, 'triangle', 392, 330, 0.1, t + 0.15, 0.22);
    glide(ac, 'sine', 260, 180, 0.07, t + 0.05, 0.2);
  });
}

export function playCodeCorrectSound(): void {
  safely((ac, t) => {
    boom(ac, t, 0.6, 'mid');
    [523.25, 659.25, 783.99].forEach((f, i) => {
      note(ac, 'sine', f, 0.15, t + 0.12 + i * 0.08, 0.22);
      note(ac, 'triangle', f * 2, 0.04, t + 0.12 + i * 0.08, 0.18);
    });
    [523.25, 659.25, 783.99, 1046.5].forEach((f) => note(ac, 'triangle', f, 0.09, t + 0.38, 0.45));
  });
}

export function playCodeWrongSound(): void {
  safely((ac, t) => {
    glide(ac, 'triangle', 330, 220, 0.12, t, 0.25);
    glide(ac, 'sine', 260, 180, 0.08, t + 0.05, 0.2);
  });
}

export function playLessonCompleteSound(): void {
  safely((ac, t) => {
    boom(ac, t, 0.85, 'mid');
    [523.25, 659.25, 783.99, 1046.5].forEach((f, i) => {
      note(ac, 'sine', f, 0.14, t + 0.18 + i * 0.09, 0.28);
      note(ac, 'triangle', f * 2, 0.04, t + 0.18 + i * 0.09, 0.22);
    });
    [523.25, 659.25, 783.99, 1046.5].forEach((f) => note(ac, 'triangle', f, 0.08, t + 0.58, 0.65));
    [1568, 2093, 2637].forEach((f, i) => note(ac, 'sine', f, 0.05, t + 0.65 + i * 0.07, 0.22));
  });
}
