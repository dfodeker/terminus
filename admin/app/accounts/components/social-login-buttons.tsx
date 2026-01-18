'use client';

import { AppleIcon, FacebookIcon, GoogleIcon } from './icons';

type AuthMode = 'login' | 'signup';

interface SocialLoginButtonsProps {
  mode: AuthMode;
}

const modeLabels: Record<AuthMode, string> = {
  login: 'Log in',
  signup: 'Sign up',
};

export function SocialLoginButtons({ mode }: SocialLoginButtonsProps) {
  const label = modeLabels[mode];

  return (
    <div className="space-y-3">
      <button
        type="button"
        className="w-full flex items-center justify-center gap-3 px-4 py-3 border border-gray-300 rounded-md hover:bg-gray-50 transition-colors"
      >
        <GoogleIcon />
        <span className="text-sm font-medium text-gray-700">{label} with Google</span>
      </button>

      <button
        type="button"
        className="w-full flex items-center justify-center gap-3 px-4 py-3 border border-gray-300 rounded-md hover:bg-gray-50 transition-colors"
      >
        <AppleIcon />
        <span className="text-sm font-medium text-gray-700">{label} with Apple</span>
      </button>

      <button
        type="button"
        className="w-full flex items-center justify-center gap-3 px-4 py-3 border border-gray-300 rounded-md hover:bg-gray-50 transition-colors"
      >
        <FacebookIcon />
        <span className="text-sm font-medium text-gray-700">{label} with Facebook</span>
      </button>
    </div>
  );
}
