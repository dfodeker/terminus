'use client';

import { useTransition } from 'react';
import { PRODUCT_TYPE_OPTIONS, type ProductTypeId } from '@/lib/shops';
import type { CreateShopState } from '@/app/actions/createShop';

interface ProductTypeStepProps {
  selectedTypes: ProductTypeId[];
  onSelectionChange: (types: ProductTypeId[]) => void;
  onBack: () => void;
  storeData: { storeName: string; storeHandle: string };
  userEmail: string;
  formAction: (formData: FormData) => void;
  pending: boolean;
  serverState: CreateShopState;
}

export function ProductTypeStep({
  selectedTypes,
  onSelectionChange,
  onBack,
  storeData,
  userEmail,
  formAction,
  pending,
  serverState,
}: ProductTypeStepProps) {
  function toggleType(id: ProductTypeId) {
    if (id === 'undecided') {
      onSelectionChange(['undecided']);
      return;
    }
    const without = selectedTypes.filter((t) => t !== 'undecided');
    if (without.includes(id)) {
      onSelectionChange(without.filter((t) => t !== id));
    } else {
      onSelectionChange([...without, id]);
    }
  }

  const [submitting, startTransition] = useTransition();

  function handleSubmit() {
    const fd = new FormData();
    fd.set('name', storeData.storeName || storeData.storeHandle);
    fd.set('handle', storeData.storeHandle);
    fd.set('email', userEmail);
    const types = selectedTypes.length > 0 ? selectedTypes : ['undecided'];
    types.forEach((t) => fd.append('product_types', t));
    startTransition(() => {
      formAction(fd);
    });
  }

  return (
    <div className="bg-white rounded-lg shadow-xl p-8 space-y-6">
      <div className="space-y-2">
        <h1 className="text-2xl font-semibold text-gray-900">What do you plan to sell?</h1>
        <p className="text-sm text-gray-600">
          We&apos;ll get you the right features and tools
        </p>
      </div>

      {serverState.message && !serverState.success && (
        <p className="text-sm text-red-600">{serverState.message}</p>
      )}

      <div className="grid grid-cols-2 gap-3">
        {PRODUCT_TYPE_OPTIONS.map((option) => {
          const isSelected = selectedTypes.includes(option.id);
          return (
            <button
              key={option.id}
              type="button"
              onClick={() => toggleType(option.id)}
              className={`relative p-4 rounded-lg border-2 text-left transition-colors ${
                isSelected
                  ? 'border-gray-900 bg-gray-50'
                  : 'border-gray-200 hover:border-gray-400'
              }`}
            >
              <div className="absolute top-3 right-3">
                {isSelected ? (
                  <div className="w-5 h-5 bg-gray-900 rounded flex items-center justify-center">
                    <svg className="w-3 h-3 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={3} d="M5 13l4 4L19 7" />
                    </svg>
                  </div>
                ) : (
                  <div className="w-5 h-5 border-2 border-gray-300 rounded" />
                )}
              </div>
              <p className="text-sm font-medium text-gray-900 pr-6">{option.label}</p>
              {option.description && (
                <p className="text-xs text-gray-500 mt-1">{option.description}</p>
              )}
            </button>
          );
        })}
      </div>

      <div className="flex justify-between items-center">
        <button
          type="button"
          onClick={onBack}
          className="flex items-center gap-1 text-sm text-gray-600 hover:text-gray-900"
        >
          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
          </svg>
          Back
        </button>
        <button
          type="button"
          onClick={handleSubmit}
          disabled={pending || submitting}
          className="flex items-center gap-1 bg-gray-900 text-white py-2.5 px-6 rounded-md font-medium hover:bg-gray-800 disabled:opacity-70 disabled:cursor-not-allowed"
        >
          {pending || submitting ? 'Creating store...' : 'Next'}
          {!pending && (
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
            </svg>
          )}
        </button>
      </div>
    </div>
  );
}
