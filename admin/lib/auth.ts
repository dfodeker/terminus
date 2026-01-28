import { redirect } from 'next/navigation';
import { getSession } from '@/lib/session';
import { userApi } from '@/lib/api';

export interface User {
  id: string;
  email: string;
}

export async function getAuthenticatedUser(): Promise<User> {
  const { accessToken } = await getSession();

  if (!accessToken) {
    redirect('/login');
  }

  const { data, error } = await userApi.me(accessToken);

  if (error || !data) {
    // TODO: try refresh token before redirecting
    redirect('/login');
  }

  return {
    id: data.user_id,
    email: data.email,
  };
}
