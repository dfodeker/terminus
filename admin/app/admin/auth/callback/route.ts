import { NextRequest, NextResponse } from 'next/server';
import { authApi, shopApi } from '@/lib/api';
import { createSession, getAdminUrl } from '@/lib/session';

export async function GET(request: NextRequest) {
  const searchParams = request.nextUrl.searchParams;
  const code = searchParams.get('code');
  const adminUrl = getAdminUrl();

  if (!code) {
    return NextResponse.redirect(new URL('/login', request.url));
  }

  const { data, error } = await authApi.exchangeCode(code);

  if (error || !data?.access_token || !data?.refresh_token) {
    const loginUrl = new URL('/login', request.url);
    loginUrl.searchParams.set('error', error || 'Authentication failed');
    return NextResponse.redirect(loginUrl);
  }

  await createSession({
    accessToken: data.access_token,
    refreshToken: data.refresh_token,
  });

  // Check if user has any stores — if not, redirect to onboarding
  const { data: shopsData } = await shopApi.list(data.access_token);
  const shops = shopsData?.shops ?? [];

  if (shops.length === 0) {
    return NextResponse.redirect(`${adminUrl}/onboarding`);
  }

  // Redirect to the last-used (or first) store
  const storeHandle = shops[0].handle;
  return NextResponse.redirect(`${adminUrl}/stores/${storeHandle}`);
}
