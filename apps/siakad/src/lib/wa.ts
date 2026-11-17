export const getGreeting = () => {
  const h = new Date().getHours();
  if (h >= 5 && h < 11) return 'Pagi';
  if (h >= 11 && h < 15) return 'Siang';
  if (h >= 15 && h < 19) return 'Sore';
  return 'Malam';
};

// assistant_code P* -> "Kak", L* -> "Bang"; tanpa kode tanpa panggilan supaya
// "Kak undefined" tidak pernah muncul.
export const getHonorific = (assistantCode?: string | null) => {
  const code = (assistantCode || '').toUpperCase();
  if (code.startsWith('P')) return 'Kak';
  if (code.startsWith('L')) return 'Bang';
  return '';
};

export const getWaLink = (phone?: string | null, text = '') => {
  if (!phone) return '#';
  let clean = phone.replace(/\D/g, '');
  if (clean.startsWith('0')) clean = '62' + clean.slice(1);
  return `https://wa.me/${clean}?text=${encodeURIComponent(text)}`;
};

export interface WaVars {
  greeting: string;
  honorific: string;
  assistant?: string;
  student?: string;
  nim?: string;
  major?: string;
  kelas?: string;
}

export const buildWaText = (
  templates: { chat_asisten?: unknown } | null | undefined,
  vars: WaVars,
) => {
  const fallback = `Selamat ${vars.greeting}${vars.honorific ? ' ' + vars.honorific : ''} ${vars.assistant}, Mohon maaf mengganggu waktunya. Saya ${vars.student} dengan NIM ${vars.nim} dari jurusan ${vars.major} kelas ${vars.kelas}.`;
  if (!templates?.chat_asisten) return fallback;
  return String(templates.chat_asisten)
    .replace(/{{waktu}}/g, vars.greeting)
    .replace(/{{panggilan}}/g, vars.honorific)
    .replace(/{{nama_asisten}}/g, vars.assistant || '')
    .replace(/{{nama_praktikan}}/g, vars.student || '')
    .replace(/{{nim}}/g, vars.nim || '')
    .replace(/{{jurusan}}/g, vars.major || '')
    .replace(/{{kelas}}/g, vars.kelas || '');
};
