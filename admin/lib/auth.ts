import { getSession } from '@/lib/session';
import { userApi } from '@/lib/api';

export interface User {
  id: string;
  email: string;
}

export async function getAuthenticatedUser(): Promise<User | null> {
  const { accessToken } = await getSession();

  if (!accessToken) {
    return null;
  }

  const { data, error } = await userApi.me(accessToken);

  if (error || !data) {
    // TODO: try refresh token before returning null
    return null;
  }

  return {
    id: data.user_id,
    email: data.email,
  };
}
