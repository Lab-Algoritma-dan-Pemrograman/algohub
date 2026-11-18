import { PUBLIC_COMPILER_BASE_URL } from '$app/env/public';
import { getRunner } from '@algohub/runner';
// Dipanggil hanya di browser: Runner membuat Worker dan SharedArrayBuffer saat dibuat.
export const elearningRunner = () => getRunner({ compilerBaseUrl: PUBLIC_COMPILER_BASE_URL });
