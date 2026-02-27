import { AuthLayout } from '@/app/accounts/components/auth-layout';

export default function OnboardingLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return <AuthLayout>{children}</AuthLayout>;
}
