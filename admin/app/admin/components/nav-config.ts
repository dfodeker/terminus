export interface NavItem {
  label: string;
  href: string;
  icon: string;
  position?: 'bottom';
}

export const navItems: NavItem[] = [
  { label: 'Home',      href: '/',          icon: 'home' },
  { label: 'Orders',    href: '/orders',    icon: 'orders' },
  { label: 'Products',  href: '/products',  icon: 'products' },
  { label: 'Customers', href: '/customers', icon: 'customers' },
  { label: 'Discounts', href: '/discounts', icon: 'discounts' },
  { label: 'Analytics', href: '/analytics', icon: 'analytics' },
  { label: 'Settings',  href: '/settings',  icon: 'settings', position: 'bottom' },
];
