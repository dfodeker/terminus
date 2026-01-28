'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';

interface SidebarNavItemProps {
  href: string;
  label: string;
  children?: React.ReactNode;
}

export function SidebarNavItem({ href, label, children }: SidebarNavItemProps) {
  const pathname = usePathname();
  const isActive = href === '/' ? pathname === '/' : pathname.startsWith(href);

  return (
    <Link
      href={href}
      className={`flex items-center gap-3 px-3 py-2 rounded-md text-sm font-medium ${
        isActive
          ? 'bg-gray-200 text-gray-900'
          : 'text-gray-600 hover:bg-gray-100 hover:text-gray-900'
      }`}
    >
      {children}
      {label}
    </Link>
  );
}
