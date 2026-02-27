'use server';

import { deleteSession } from '@/lib/session';
import { redirect } from 'next/navigation';

export async function logout() {
  await deleteSession();

  const loginUrl = process.env.NODE_ENV === 'production'
    ? (process.env.ACCOUNTS_URL || 'https://accounts.storeos.com/login')
    : 'http://accounts.storeos.local:3000/login';

  redirect(loginUrl);
}
