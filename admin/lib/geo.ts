// lib/geo.ts
import { headers } from 'next/headers';

export interface GeoData {
  country: string | null;
  region: string | null;
  city: string | null;
}

export async function getGeoData(): Promise<GeoData> {
  const headersList = await headers();
  
  return {
    country: headersList.get('x-geo-country') || null,
    region: headersList.get('x-geo-region') || null,
    city: headersList.get('x-geo-city') || null,
  };
}

// Fallback for local dev
export function getDefaultCountry(country: string | null): string {
  return country || 'US';
}