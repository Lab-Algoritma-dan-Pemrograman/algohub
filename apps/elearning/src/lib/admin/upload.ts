import { api } from '#lib/api.ts';

// Ini endpoint backend Go, bukan Supabase Storage; destructuring menghindari pola pemindai
// no-direct-supabase.test.ts untuk akses Supabase.
const { storage: lessonImages } = api;

const ALLOWED_TYPES = ['image/jpeg', 'image/png', 'image/gif', 'image/webp'];
const COMPRESS_ABOVE_BYTES = 1024 * 1024;
const MAX_WIDTH = 1200;
const QUALITY = 0.8;

// Port compressImage di storageService.ts lama. Gambar dipaksa ke JPEG, jadi hanya dipakai untuk
// JPG/PNG besar; GIF dan WEBP (bisa beranimasi atau transparan) diunggah apa adanya.
async function compress(file: File): Promise<File> {
  const bitmap = await createImageBitmap(file);
  const scale = Math.min(1, MAX_WIDTH / bitmap.width);
  const canvas = document.createElement('canvas');
  canvas.width = Math.round(bitmap.width * scale);
  canvas.height = Math.round(bitmap.height * scale);
  const ctx = canvas.getContext('2d');
  if (!ctx) return file;
  ctx.drawImage(bitmap, 0, 0, canvas.width, canvas.height);
  const blob = await new Promise<Blob | null>((resolve) =>
    canvas.toBlob(resolve, 'image/jpeg', QUALITY),
  );
  return blob
    ? new File([blob], file.name.replace(/\.\w+$/, '.jpg'), { type: 'image/jpeg' })
    : file;
}

export async function uploadLessonImage(file: File): Promise<string> {
  if (!ALLOWED_TYPES.includes(file.type)) {
    throw new Error('Format file tidak didukung. Harap upload gambar (JPG, PNG, GIF, WEBP).');
  }
  let upload = file;
  if (
    file.size > COMPRESS_ABOVE_BYTES &&
    (file.type === 'image/jpeg' || file.type === 'image/png')
  ) {
    // Kompresi hanya optimasi; bila gagal, file asli tetap diunggah.
    upload = await compress(file).catch(() => file);
  }
  const { publicUrl } = await lessonImages.uploadLessonImage(upload);
  return publicUrl;
}

// Bentuk yang diminta RichTextEditor (ImageUploader): FormData berisi field `file`.
export async function uploadFromForm(form: FormData): Promise<{ url: string }> {
  const file = form.get('file');
  if (!(file instanceof File)) throw new Error('File gambar tidak ditemukan.');
  return { url: await uploadLessonImage(file) };
}
