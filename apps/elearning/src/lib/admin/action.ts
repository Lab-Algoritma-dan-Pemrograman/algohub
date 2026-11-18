import { toast } from '@algohub/ui';

// Pengganti showModal alert di AdminDashboard lama: hasil aksi admin ditampilkan sebagai toast.
export async function perform(
  task: () => Promise<unknown>,
  success: string,
  failure: string,
): Promise<boolean> {
  try {
    await task();
    toast.success(success);
    return true;
  } catch (error) {
    toast.error(error instanceof Error && error.message ? `${failure}: ${error.message}` : failure);
    return false;
  }
}
