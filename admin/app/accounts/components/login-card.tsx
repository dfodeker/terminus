'use client';

import Link from 'next/link';
import { useActionState, useState } from 'react';
import { loginUser } from '@/app/actions/loginUser';
import {
  AuthDivider,
  AuthFooter,
  FormInput,
  FormMessage,
  PasswordInput,
  SocialLoginButtons,
  SubmitButton,
} from './index';

export function LoginForm() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');

  const [state, formAction, pending] = useActionState(loginUser, {
    message: '',
    success: false,
    errors: {},
  });

  return (
    <div className="w-full max-w-md bg-white rounded-lg shadow-xl p-8 space-y-6">
      {/* Header */}
      <div className="space-y-2">
        <h1 className="text-2xl font-semibold text-gray-900">Log in to StoreOS</h1>
        <p className="text-sm text-gray-600">Welcome back! Please enter your details.</p>
      </div>

      <FormMessage message={state.message} success={state.success} />

      {/* Login Form */}
      <form className="space-y-4" action={formAction}>
        <FormInput
          id="email"
          name="email"
          label="Email"
          type="email"
          value={email}
          onChange={setEmail}
          error={state.errors?.email}
        />

        <PasswordInput
          id="password"
          name="password"
          label="Password"
          value={password}
          onChange={setPassword}
          error={state.errors?.password}
        />

        <div className="flex justify-end">
          <Link
            href="/accounts/forgot-password"
            className="text-sm text-blue-600 hover:text-blue-700"
          >
            Forgot password?
          </Link>
        </div>

        <SubmitButton pending={pending} pendingText="Logging in...">
          Log in
        </SubmitButton>
      </form>

      <AuthDivider />

      <SocialLoginButtons mode="login" />

      {/* Sign Up Link */}
      <div className="text-center">
        <p className="text-sm text-gray-600">
          Don&apos;t have an account?{' '}
          <Link href="/signup" className="text-blue-600 hover:text-blue-700 font-medium">
            Sign up →
          </Link>
        </p>
      </div>

      <AuthFooter />
    </div>
  );
}
