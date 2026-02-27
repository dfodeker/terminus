'use client';

import { useState, useRef, useEffect } from 'react';
import { useAuth } from '@/app/admin/providers/auth-provider';
import { logout } from '@/app/actions/logout';

interface Store {
  id: string;
  name: string;
  initials: string;
}

// Placeholder until store data comes from the backend
const currentStore: Store = { id: '1', name: 'My Store', initials: 'MS' };
const stores: Store[] = [currentStore];

export function ProfileMenu() {
  const { user } = useAuth();
  const [isOpen, setIsOpen] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);

  const userInitials = user.email
    ? user.email.substring(0, 2).toUpperCase()
    : '??';

  useEffect(() => {
    function handleClick(e: MouseEvent) {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        setIsOpen(false);
      }
    }
    document.addEventListener('mousedown', handleClick);
    return () => document.removeEventListener('mousedown', handleClick);
  }, []);

  return (
    <div ref={containerRef} className="relative">
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="flex items-center gap-2 px-2 py-1 rounded-lg hover:bg-gray-800"
      >
        <span className="w-8 h-8 rounded-md bg-green-500 text-white flex items-center justify-center text-xs font-bold">
          {currentStore.initials}
        </span>
        <span className="text-sm font-medium">{currentStore.name}</span>
      </button>

      {isOpen && (
        <div className="absolute right-0 top-full mt-2 w-72 bg-white text-gray-900 rounded-xl shadow-lg border border-gray-200 z-50 overflow-hidden">
          {/* Store picker */}
          <div className="p-2">
            {stores.map((store) => (
              <div
                key={store.id}
                className="flex items-center gap-3 px-3 py-2 rounded-lg hover:bg-gray-100"
              >
                <span className="w-8 h-8 rounded-md bg-green-500 text-white flex items-center justify-center text-xs font-bold shrink-0">
                  {store.initials}
                </span>
                <span className="text-sm font-medium">{store.name}</span>
                {store.id === currentStore.id && (
                  <svg className="w-5 h-5 ml-auto text-gray-900" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2} strokeLinecap="round" strokeLinejoin="round">
                    <polyline points="20 6 9 17 4 12" />
                  </svg>
                )}
              </div>
            ))}

            <button className="flex items-center gap-3 px-3 py-2 rounded-lg hover:bg-gray-100 w-full text-left">
              <span className="w-8 h-8 rounded-md bg-gray-200 flex items-center justify-center shrink-0">
                <svg className="w-4 h-4 text-gray-600" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2} strokeLinecap="round" strokeLinejoin="round">
                  <path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z" />
                  <polyline points="9 22 9 12 15 12 15 22" />
                </svg>
              </span>
              <span className="text-sm font-medium">All stores</span>
            </button>
          </div>

          <div className="border-t border-gray-200" />

          {/* User section */}
          <div className="p-2">
            <div className="flex items-center gap-3 px-3 py-2">
              <span className="w-8 h-8 rounded-md bg-green-500 text-white flex items-center justify-center text-xs font-bold shrink-0">
                {userInitials}
              </span>
              <span className="text-sm">{user.email || 'No email'}</span>
            </div>

            <form action={logout}>
              <button
                type="submit"
                className="flex items-center gap-3 px-3 py-2 rounded-lg hover:bg-gray-100 w-full text-left"
              >
                <span className="w-8 h-8 flex items-center justify-center shrink-0">
                  <svg className="w-5 h-5 text-gray-600" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2} strokeLinecap="round" strokeLinejoin="round">
                    <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4" />
                    <polyline points="16 17 21 12 16 7" />
                    <line x1="21" y1="12" x2="9" y2="12" />
                  </svg>
                </span>
                <span className="text-sm font-medium">Log out</span>
              </button>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
