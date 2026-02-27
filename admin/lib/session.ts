import { cookies } from 'next/headers';

const ACCESS_TOKEN_NAME = 'access_token';
const REFRESH_TOKEN_NAME = 'refresh_token';
const ACCESS_TOKEN_MAX_AGE = 60 * 60; // 1 hour (should match backend)
const REFRESH_TOKEN_MAX_AGE = 60 * 60 * 24 * 7; // 7 days

// For local development, use .storeos.local for subdomain cookie sharing
// Make sure to add these to /etc/hosts:
//   127.0.0.1 storeos.local
//   127.0.0.1 admin.storeos.local
//   127.0.0.1 accounts.storeos.local
// In production, use .storeos.com to share cookies across subdomains
function getCookieDomain(): string | undefined {
  if (process.env.NODE_ENV === 'production') {
    return process.env.COOKIE_DOMAIN || '.storeos.org';
  }
  // Use .storeos.local for local development with subdomain cookie sharing
  return '.storeos.local';
}

interface SessionTokens {
  accessToken: string;
  refreshToken: string;
}

export async function createSession(tokens: SessionTokens): Promise<void> {
  const cookieStore = await cookies();
  const domain = getCookieDomain();
  const isProduction = process.env.NODE_ENV === 'production';

  // Store access token
  cookieStore.set(ACCESS_TOKEN_NAME, tokens.accessToken, {
    httpOnly: true,
    secure: isProduction,
    sameSite: 'lax',
    maxAge: ACCESS_TOKEN_MAX_AGE,
    path: '/',
    ...(domain && { domain }),
  });

  // Store refresh token
  cookieStore.set(REFRESH_TOKEN_NAME, tokens.refreshToken, {
    httpOnly: true,
    secure: isProduction,
    sameSite: 'lax',
    maxAge: REFRESH_TOKEN_MAX_AGE,
    path: '/',
    ...(domain && { domain }),
  });
}

interface Session {
  accessToken: string | undefined;
  refreshToken: string | undefined;
}

export async function getSession(): Promise<Session> {
  const cookieStore = await cookies();
  return {
    accessToken: cookieStore.get(ACCESS_TOKEN_NAME)?.value,
    refreshToken: cookieStore.get(REFRESH_TOKEN_NAME)?.value,
  };
}

export async function getAccessToken(): Promise<string | undefined> {
  const cookieStore = await cookies();
  return cookieStore.get(ACCESS_TOKEN_NAME)?.value;
}

export async function getRefreshToken(): Promise<string | undefined> {
  const cookieStore = await cookies();
  return cookieStore.get(REFRESH_TOKEN_NAME)?.value;
}

export async function deleteSession(): Promise<void> {
  const cookieStore = await cookies();
  const domain = getCookieDomain();
  const isProduction = process.env.NODE_ENV === 'production';

  cookieStore.set(ACCESS_TOKEN_NAME, '', {
    httpOnly: true,
    secure: isProduction,
    sameSite: 'lax',
    maxAge: 0,
    path: '/',
    ...(domain && { domain }),
  });

  cookieStore.set(REFRESH_TOKEN_NAME, '', {
    httpOnly: true,
    secure: isProduction,
    sameSite: 'lax',
    maxAge: 0,
    path: '/',
    ...(domain && { domain }),
  });
}

export function getAdminUrl(): string {
  const url = process.env.ADMIN_URL || (
    process.env.NODE_ENV === 'production'
      ? 'https://admin.storeos.com'
      : 'http://admin.storeos.local:3000'
  );
  console.log('[getAdminUrl] NODE_ENV:', process.env.NODE_ENV, 'ADMIN_URL:', process.env.ADMIN_URL, '-> returning:', url);
  return url;
}

export function getAuthCallbackUrl(code: string): string {
  const adminUrl = getAdminUrl();
  return `${adminUrl}/auth/callback?code=${encodeURIComponent(code)}`;
}
