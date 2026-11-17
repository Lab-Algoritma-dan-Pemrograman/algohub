import type { SiakadUser } from '@algohub/api-client/siakad';
import { Session } from '@algohub/auth';

export const session = new Session<SiakadUser>();
