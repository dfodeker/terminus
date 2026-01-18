'use client';

import Link from 'next/link';
import { useActionState, useState } from 'react';
import { CountrySelect } from '@/app/components/ui/country-selector';
import createUser from '@/app/actions/createUser';
import {
  AuthDivider,
  AuthFooter,
  FormInput,
  FormMessage,
  PasswordInput,
  PasswordStrengthIndicator,
  SocialLoginButtons,
  SubmitButton,
} from './index';

interface SignUpFormProps {
  defaultCountry?: string;
}

export function SignUpForm({ defaultCountry = 'CA' }: SignUpFormProps) {
  const [firstName, setFirstName] = useState('');
  const [lastName, setLastName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');

  const [state, formAction, pending] = useActionState(createUser, {
    message: '',
    success: false,
    errors: {},
  });

  return (
    <div className="w-full max-w-md bg-white rounded-lg shadow-xl p-8 space-y-6">
      {/* Region Selector */}
      <div className="flex items-center gap-2">
        <CountrySelect defaultCountry={defaultCountry} />
      </div>

      {/* Header */}
      <div className="space-y-2">
        <h1 className="text-2xl font-semibold text-gray-900">Create a StoreOS account</h1>
        <p className="text-sm text-gray-600">Get 3 days free, then 3 months for $1/month</p>
      </div>

      <FormMessage message={state.message} success={state.success} />

      {/* Sign Up Form */}
      <form className="space-y-4" action={formAction}>
        <div className="grid grid-cols-2 gap-3">
          <FormInput
            id="first_name"
            name="first_name"
            label="First name"
            type="text"
            value={firstName}
            onChange={setFirstName}
            error={state.errors?.first_name}
          />
          <FormInput
            id="last_name"
            name="last_name"
            label="Last name"
            type="text"
            value={lastName}
            onChange={setLastName}
            error={state.errors?.last_name}
          />
        </div>

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
        >
          <PasswordStrengthIndicator password={password} />
        </PasswordInput>

        <SubmitButton pending={pending} pendingText="Creating account...">
          Create account
        </SubmitButton>
      </form>

      <AuthDivider />

      <SocialLoginButtons mode="signup" />

      {/* Login Link */}
      <div className="text-center">
        <p className="text-sm text-gray-600">
          Already have an account?{' '}
          <Link href="/login" className="text-blue-600 hover:text-blue-700 font-medium">
            Log in →
          </Link>
        </p>
      </div>

      <AuthFooter />
    </div>
  );
}
