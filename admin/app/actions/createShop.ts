'use server';

import { z } from 'zod';
import { organizationApi, shopApi } from '@/lib/api';
import { getAccessToken } from '@/lib/session';
import { redirect } from 'next/navigation';
import { slugify } from '@/lib/shops';

const schema = z.object({
  name: z.string().min(1, { message: 'Store name is required' }),
  handle: z.string().min(1),
  email: z.string().email(),
  product_types: z.array(z.string()).min(1, { message: 'Select at least one option' }),
});

export interface CreateShopState {
  message: string;
  success: boolean;
  errors: { [key: string]: string[] };
}

export async function createShop(
  _prevState: CreateShopState,
  formData: FormData
): Promise<CreateShopState> {
  const accessToken = await getAccessToken();

  if (!accessToken) {
    return { message: 'Not authenticated', success: false, errors: {} };
  }

  const name = formData.get('name') as string;
  const handle = (formData.get('handle') as string) || slugify(name);
  const email = formData.get('email') as string;
  const productTypes = formData.getAll('product_types') as string[];

  const validatedFields = schema.safeParse({
    name,
    handle,
    email,
    product_types: productTypes,
  });
  console.log('Validated fields:', validatedFields);
  if (!validatedFields.success) {
    return {
      message: 'Please fix the errors below.',
      success: false,
      errors: validatedFields.error.flatten().fieldErrors,
    };
  }

  // 1. Create organization
  const { data: org, error: orgError } = await organizationApi.create(
    {
      name: validatedFields.data.name,
      handle: validatedFields.data.handle,
      billing_email: validatedFields.data.email,
    },
    accessToken
  );
  console.log('Organization creation response:', { org, orgError });

  if (orgError || !org) {
    return { message: orgError || 'Failed to create organization', success: false, errors: {} };
  }

  // 2. Create shop under the organization
  const { data: shop, error: shopError } = await shopApi.create(
    {
      organization_id: org.id,
      name: validatedFields.data.name,
      handle: validatedFields.data.handle,
      subdomain: `${validatedFields.data.handle}.storeos.com`,
      currency: 'USD',
      email: validatedFields.data.email,
    },
    accessToken
  );

  console.log('Shop creation response:', { shop, shopError });

  if (shopError || !shop) {
    return { message: shopError || 'Failed to create store', success: false, errors: {} };
  }

  const storeHandle = shop.handle || handle;
  redirect(`/stores/${storeHandle}`);
}
