import { NextRequest, NextResponse } from "next/server";
import {geolocation} from "@vercel/functions"

// Map ISO country codes to display names



export default function proxy(request: NextRequest) {
    const host = request.headers.get("host") || "";
    const pathname = request.nextUrl.pathname;
    
    const subdomain = getSubdomain(host);
    
    // If someone tries to access /admin from any subdomain (other than admin),
    // redirect them to admin.storeos.local (dev) or admin.storeos.com (prod)
    if (pathname.startsWith('/admin') && subdomain !== 'admin') {
        const url = request.nextUrl.clone();
        const port = host.includes(':') ? ':' + host.split(':')[1] : '';
        
        if (host.includes('.storeos.local') || host.includes('.localhost')) {
            url.host = `admin.storeos.local${port}`;
        } else {
            // For production: redirect to admin.yourdomain.com
            const baseDomain = getBaseDomain(host);
            url.host = `admin.${baseDomain}${port}`;
        }
        
        // Remove /admin prefix since admin subdomain will add it back
        url.pathname = pathname.replace(/^\/admin/, '') || '/';
        
        return NextResponse.redirect(url);
    }
    
    if (!subdomain) {
        return NextResponse.next();
    }
    
    const geo = geolocation(request);

    const requestHeaders = new Headers(request.headers)
    requestHeaders.set("x-subdomain", subdomain);
    requestHeaders.set('x-geo-country', geo.country || '');
    requestHeaders.set('x-geo-region', geo.countryRegion || '');
    requestHeaders.set('x-geo-city', geo.city || '');



    if (subdomain==="admin") {
        const url = request.nextUrl.clone();
        url.pathname = `/admin${request.nextUrl.pathname}`;
        return NextResponse.rewrite(url, {
            request: {
                headers: requestHeaders
            }
        });
    }
    if (subdomain=== "accounts") {
        const url = request.nextUrl.clone();
        url.pathname = `/accounts${request.nextUrl.pathname}`;
        return NextResponse.rewrite(url, {
            request: {
                headers: requestHeaders
            }
        });// this can be rewritten to other subdomains like user, dashboard, etc.
    }
    return NextResponse.next({
        request: {
            headers: requestHeaders
        }
    });

}

function getSubdomain(host: string): string | null {
  // Remove port for local dev
  const hostname = host.split(':')[0];

  // Handle local dev subdomains: admin.storeos.local
  if (hostname.endsWith('.storeos.local')) {
    return hostname.replace('.storeos.local', '');
  }

  // Handle localhost subdomains: admin.localhost (fallback)
  if (hostname.endsWith('.localhost')) {
    return hostname.replace('.localhost', '');
  }

  // Handle production: admin.ourapp.com
  const parts = hostname.split('.');
  if (parts.length > 2) {
    return parts[0];
  }

  return null;
}

function getBaseDomain(host: string): string {
  // Remove port
  const hostname = host.split(':')[0];
  
  // For local dev with storeos.local
  if (hostname.endsWith('.storeos.local') || hostname === 'storeos.local') {
    return 'storeos.local';
  }
  
  // For localhost (fallback)
  if (hostname.endsWith('.localhost') || hostname === 'localhost') {
    return 'localhost';
  }
  
  // For production: get the last two parts (e.g., ourapp.com)
  const parts = hostname.split('.');
  if (parts.length >= 2) {
    return parts.slice(-2).join('.');
  }
  
  return hostname;
}

export const config = {
  matcher: [
    // Match all paths except static files and api
    '/((?!_next/static|_next/image|favicon.ico).*)',
  ],
};