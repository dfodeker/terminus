import { redirect } from 'next/navigation';
import { Shell } from '../components/shell';
import { getAccessToken } from '@/lib/session';
import { shopApi } from '@/lib/api';

export default async function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const accessToken = await getAccessToken();

  if (accessToken) {
    const { data } = await shopApi.list(accessToken);
    if (!data?.shops || data.shops.length === 0) {
      redirect('/onboarding');
    }
  }

  return <Shell>{children}</Shell>;
}
