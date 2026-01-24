'use server';

import { z } from 'zod';
import { authApi } from '@/lib/api';
import { getAuthCallbackUrl } from '@/lib/session';
import { redirect } from 'next/navigation';

const schema = z.object({
  first_name: z.string().min(1, { message: 'First name is required' }),
  last_name: z.string().min(1, { message: 'Last name is required' }),
  email: z.string().email({
    message: 'Please enter a valid email address',
  }),
  password: z
    .string()
    .min(8, { message: 'Password must be at least 8 characters' })
    .regex(/[a-z]/, { message: 'Password must contain at least one lowercase letter' })
    .regex(/[A-Z]/, { message: 'Password must contain at least one uppercase letter' })
    .regex(/\d/, { message: 'Password must contain at least one number' }),
});

export interface CreateUserState {
  message: string;
  success: boolean;
  errors: { [key: string]: string[] };
}

export default async function createUser(
  _prevState: CreateUserState,
  formData: FormData
): Promise<CreateUserState> {
  const validatedFields = schema.safeParse({
    first_name: formData.get('first_name'),
    last_name: formData.get('last_name'),
    email: formData.get('email'),
    password: formData.get('password'),
  });

  if (!validatedFields.success) {
    return {
      message: 'There were errors with your submission.',
      success: false,
      errors: validatedFields.error.flatten().fieldErrors,
    };
  }

  const { data, error } = await authApi.register(validatedFields.data);

  if (error) {
    return {
      message: error,
      success: false,
      errors: {},
    };
  }

  if (!data?.code) {
    return {
      message: 'Authentication failed - no code received',
      success: false,
      errors: {},
    };
  }

  // Redirect to admin subdomain with auth code
  redirect(getAuthCallbackUrl(data.code));
}
