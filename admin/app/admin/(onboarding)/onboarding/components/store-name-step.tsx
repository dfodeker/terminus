'use client';

import { useState, useMemo } from 'react';
import { slugify, generateDefaultHandle } from '@/lib/shops';

interface StoreNameStepProps {
  defaultName: string;
  userEmail: string;
  onNext: (name: string, handle: string) => void;
}

export function StoreNameStep({ defaultName, userEmail, onNext }: StoreNameStepProps) {
  const [name, setName] = useState(defaultName);

  const handle = useMemo(() => {
    if (name.trim()) return slugify(name);
    return generateDefaultHandle(userEmail);
  }, [name, userEmail]);

  function handleSkip() {
    const fallbackHandle = generateDefaultHandle(userEmail);
    onNext('', fallbackHandle);
  }

  return (
    <div className="bg-white rounded-lg shadow-xl p-8 space-y-6">
      <div className="space-y-2">
        <h1 className="text-2xl font-semibold text-gray-900">Name your store</h1>
        <p className="text-sm text-gray-600">
          You can always change this later in your settings.
        </p>
      </div>

      <div>
        <label htmlFor="store-name" className="block text-sm font-medium text-gray-700 mb-1">
          Store name
        </label>
        <input
          id="store-name"
          type="text"
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="My awesome store"
          className="w-full px-3 py-2.5 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-gray-900 focus:border-transparent"
        />
      </div>

      <div className="text-sm text-gray-500">
        Your store will be at{' '}
        <span className="font-mono text-gray-800">{handle}.storeos.com</span>
      </div>

      <div className="flex gap-3">
        <button
          type="button"
          onClick={handleSkip}
          className="flex-1 py-3 px-4 border border-gray-300 rounded-md text-gray-700 font-medium hover:bg-gray-50"
        >
          Skip
        </button>
        <button
          type="button"
          onClick={() => onNext(name, handle)}
          className="flex-1 bg-gray-900 text-white py-3 px-4 rounded-md font-medium hover:bg-gray-800"
        >
          Next
        </button>
      </div>
    </div>
  );
}
