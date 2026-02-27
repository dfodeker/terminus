'use server';

import { z } from 'zod';
import { authApi } from '@/lib/api';
import { getAuthCallbackUrl } from '@/lib/session';
import { redirect } from 'next/navigation';

const schema = z.object({
  email: z.string().email({
    message: 'Please enter a valid email address',
  }),
  password: z.string().min(1, { message: 'Password is required' }),
});

export interface LoginState {
  message: string;
  success: boolean;
  errors: { [key: string]: string[] };
}

export async function loginUser(
  _prevState: LoginState,
  formData: FormData
): Promise<LoginState> {
  const validatedFields = schema.safeParse({
    email: formData.get('email'),
    password: formData.get('password'),
  });

  if (!validatedFields.success) {
    return {
      message: 'Please check your credentials.',
      success: false,
      errors: validatedFields.error.flatten().fieldErrors,
    };
  }

  const { email, password } = validatedFields.data;
  const { data, error } = await authApi.login(email, password);

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
