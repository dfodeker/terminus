export interface CreateOrganizationPayload {
  name: string;
  handle: string;
  billing_email: string;
  plan?: string;
}

export interface Organization {
  id: string;
  name: string;
  handle: string;
  billing_email: string;
  plan: string;
}

export interface CreateShopPayload {
  organization_id: string;
  name: string;
  handle: string;
  subdomain: string;
  currency: string;
  email: string;
  locale?: string;
  timezone?: string;
  shop_owner?: string;
  phone?: string;
  custom_domain?: string;
}

export interface Shop {
  id: string;
  name: string;
  handle: string;
  subdomain: string;
  currency: string;
  email: string;
  custom_domain?: string;
  timezone?: string;
  shop_owner?: string;
  phone?: string;
  created_at?: string;
  updated_at?: string;
}

export interface ShopsListResponse {
  shops: Shop[];
  next_cursor?: string;
}

export const PRODUCT_TYPE_OPTIONS = [
  {
    id: 'physical',
    label: 'Products I buy or make myself',
    description: 'Shipped by me',
  },
  {
    id: 'digital',
    label: 'Digital products',
    description: 'Music, digital art, NFTs',
  },
  {
    id: 'dropshipping',
    label: 'Dropshipping products',
    description: 'Sourced and shipped by a third party',
  },
  {
    id: 'services',
    label: 'Services',
    description: 'Coaching, housekeeping, consulting',
  },
  {
    id: 'print_on_demand',
    label: 'Print-on-demand products',
    description: 'My designs, printed and shipped by a third party',
  },
  {
    id: 'undecided',
    label: "I'll decide later",
    description: '',
  },
] as const;

export type ProductTypeId = (typeof PRODUCT_TYPE_OPTIONS)[number]['id'];

export function slugify(text: string): string {
  return text
    .toLowerCase()
    .trim()
    .replace(/[^\w\s-]/g, '')
    .replace(/[\s_]+/g, '-')
    .replace(/^-+|-+$/g, '');
}

export function generateDefaultHandle(email: string): string {
  const local = email.split('@')[0] || 'my';
  return slugify(local) + '-store';
}
