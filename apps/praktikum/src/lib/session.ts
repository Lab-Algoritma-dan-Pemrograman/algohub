import type { User } from '@algohub/api-client/praktikum';
import { Session } from '@algohub/auth';

export const session = new Session<User>();
