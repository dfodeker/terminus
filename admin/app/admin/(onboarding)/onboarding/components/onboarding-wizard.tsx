'use client';

import { useState, useActionState } from 'react';
import { StoreNameStep } from './store-name-step';
import { ProductTypeStep } from './product-type-step';
import { StepIndicator } from './step-indicator';
import { createShop } from '@/app/actions/createShop';
import type { ProductTypeId } from '@/lib/shops';

interface OnboardingData {
  storeName: string;
  storeHandle: string;
  productTypes: ProductTypeId[];
}

interface OnboardingWizardProps {
  userEmail: string;
}

const TOTAL_STEPS = 2;

export function OnboardingWizard({ userEmail }: OnboardingWizardProps) {
  const [step, setStep] = useState(0);
  const [data, setData] = useState<OnboardingData>({
    storeName: '',
    storeHandle: '',
    productTypes: [],
  });

  const [state, formAction, pending] = useActionState(createShop, {
    message: '',
    success: false,
    errors: {},
  });

  function handleStoreNameNext(name: string, handle: string) {
    setData((prev) => ({ ...prev, storeName: name, storeHandle: handle }));
    setStep(1);
  }

  function handleBack() {
    setStep((s) => Math.max(0, s - 1));
  }

  return (
    <div className="w-full max-w-2xl">
      <StepIndicator current={step} total={TOTAL_STEPS} />

      {step === 0 && (
        <StoreNameStep
          defaultName={data.storeName}
          userEmail={userEmail}
          onNext={handleStoreNameNext}
        />
      )}

      {step === 1 && (
        <ProductTypeStep
          selectedTypes={data.productTypes}
          onSelectionChange={(types) =>
            setData((prev) => ({ ...prev, productTypes: types }))
          }
          onBack={handleBack}
          storeData={data}
          userEmail={userEmail}
          formAction={formAction}
          pending={pending}
          serverState={state}
        />
      )}
    </div>
  );
}
