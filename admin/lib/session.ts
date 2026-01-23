import { cookies } from 'next/headers';

const TOKEN_NAME = 'auth_token';
const TOKEN_MAX_AGE = 60 * 60 * 24 * 7; // 7 days

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

export async function createSession(token: string): Promise<void> {
  const cookieStore = await cookies();
  const domain = getCookieDomain();
  
  cookieStore.set(TOKEN_NAME, token, {
    httpOnly: true,
    secure: process.env.NODE_ENV === 'production',
    sameSite: 'lax',
    maxAge: TOKEN_MAX_AGE,
    path: '/',
    ...(domain && { domain }),
  });
  
}

export async function getSession(): Promise<string | undefined> {
  const cookieStore = await cookies();
  return cookieStore.get(TOKEN_NAME)?.value;
}

export async function deleteSession(): Promise<void> {
  const cookieStore = await cookies();
  const domain = getCookieDomain();
  
  cookieStore.set(TOKEN_NAME, '', {
    httpOnly: true,
    secure: process.env.NODE_ENV === 'production',
    sameSite: 'lax',
    maxAge: 0,
    path: '/',
    ...(domain && { domain }),
  });
}

export function getAdminUrl(): string {
  const baseUrl = process.env.ADMIN_URL || (
    process.env.NODE_ENV === 'production'
      ? 'https://admin.storeos.com'
      : 'http://admin.storeos.local:3000'
  );
  return baseUrl;
}
