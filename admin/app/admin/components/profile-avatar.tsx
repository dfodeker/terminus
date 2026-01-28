'use client';

import { useAuth } from '@/app/admin/providers/auth-provider';

export function ProfileAvatar() {
  const { user } = useAuth();
  const initial = user.email ? user.email.charAt(0).toUpperCase() : '?';

  return (
    <div className="w-8 h-8 rounded-full bg-gray-600 text-white flex items-center justify-center text-sm font-medium">
      {initial}
    </div>
  );
}
