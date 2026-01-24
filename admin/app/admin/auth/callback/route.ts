import { NextRequest, NextResponse } from 'next/server';
import { authApi } from '@/lib/api';
import { createSession } from '@/lib/session';

export async function GET(request: NextRequest) {
  console.log('[auth/callback] Request URL:', request.url);
  console.log('[auth/callback] Host:', request.headers.get('host'));

  const searchParams = request.nextUrl.searchParams;
  const code = searchParams.get('code');

  if (!code) {
    return NextResponse.redirect(new URL('/login', request.url));
  }

  const { data, error } = await authApi.exchangeCode(code);

  if (error || !data?.access_token || !data?.refresh_token) {
    // Redirect to login with error
    const loginUrl = new URL('/login', request.url);
    loginUrl.searchParams.set('error', error || 'Authentication failed');
    return NextResponse.redirect(loginUrl);
  }

  // Store both tokens in session
  await createSession({
    accessToken: data.access_token,
    refreshToken: data.refresh_token,
  });

  // Redirect to admin dashboard
  return NextResponse.redirect(new URL('/admin', request.url));
}
