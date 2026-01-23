import { AuthLayout } from '../components/auth-layout';
import { SignUpForm } from '../components/signup-card';

export default function SignUpPage() {
  return (
    <AuthLayout>
      <SignUpForm />
    </AuthLayout>
  );
}
