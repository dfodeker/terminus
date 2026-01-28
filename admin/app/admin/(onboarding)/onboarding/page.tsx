import { getAuthenticatedUser } from '@/lib/auth';
import { OnboardingWizard } from './components/onboarding-wizard';

export default async function OnboardingPage() {
  const user = await getAuthenticatedUser();
  const email = user?.email ?? '';

  return <OnboardingWizard userEmail={email} />;
}
