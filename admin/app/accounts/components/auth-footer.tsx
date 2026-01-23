import Link from 'next/link';

export function AuthFooter() {
  return (
    <>
      {/* Terms */}
      <div className="text-xs text-gray-500">
        <p>
          By proceeding, you agree to the{' '}
          <Link href="/terms" className="text-blue-600 hover:underline">
            Terms and Conditions
          </Link>
          {' '}and{' '}
          <Link href="/privacy" className="text-blue-600 hover:underline">
            Privacy Policy
          </Link>
        </p>
      </div>

      {/* Footer Links */}
      <div className="flex gap-4 text-xs text-gray-500 pt-2">
        <Link href="/help" className="hover:text-gray-700">
          Help
        </Link>
        <Link href="/privacy" className="hover:text-gray-700">
          Privacy
        </Link>
        <Link href="/terms" className="hover:text-gray-700">
          Terms
        </Link>
      </div>
    </>
  );
}
