'use client';

import { useMemo } from 'react';

type PasswordStrength = 'weak' | 'fair' | 'good' | 'strong';

interface StrengthResult {
  strength: PasswordStrength;
  score: number;
}

const getPasswordStrength = (password: string): StrengthResult => {
  let score = 0;

  if (password.length >= 8) score++;
  if (password.length >= 12) score++;
  if (/[a-z]/.test(password) && /[A-Z]/.test(password)) score++;
  if (/\d/.test(password)) score++;
  if (/[^a-zA-Z0-9]/.test(password)) score++;

  if (score <= 1) return { strength: 'weak', score: 1 };
  if (score === 2) return { strength: 'fair', score: 2 };
  if (score === 3 || score === 4) return { strength: 'good', score: 3 };
  return { strength: 'strong', score: 4 };
};

const strengthConfig: Record<PasswordStrength, { color: string; textColor: string; text: string }> = {
  weak: { color: 'bg-red-500', textColor: 'text-red-600', text: 'Weak' },
  fair: { color: 'bg-yellow-500', textColor: 'text-yellow-600', text: 'Fair' },
  good: { color: 'bg-blue-500', textColor: 'text-blue-600', text: 'Good' },
  strong: { color: 'bg-green-500', textColor: 'text-green-600', text: 'Strong' },
};

interface PasswordStrengthIndicatorProps {
  password: string;
}

export function PasswordStrengthIndicator({ password }: PasswordStrengthIndicatorProps) {
  const strengthResult = useMemo(() => {
    if (!password) return null;
    return getPasswordStrength(password);
  }, [password]);

  if (!password || !strengthResult) return null;

  const config = strengthConfig[strengthResult.strength];

  return (
    <div id="password-strength" className="mt-2 space-y-1">
      <div className="flex gap-1">
        {[1, 2, 3, 4].map((level) => (
          <div
            key={level}
            className={`h-1.5 flex-1 rounded-full transition-colors ${
              level <= strengthResult.score ? config.color : 'bg-gray-200'
            }`}
          />
        ))}
      </div>
      <p className={`text-xs ${config.textColor}`}>
        Password strength: {config.text}
      </p>
    </div>
  );
}
