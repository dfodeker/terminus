import { navItems } from './nav-config';
import { iconMap } from './icons';
import { SidebarNavItem } from './sidebar-nav-item';

export function Sidebar() {
  const topItems = navItems.filter((item) => item.position !== 'bottom');
  const bottomItems = navItems.filter((item) => item.position === 'bottom');

  return (
    <aside className="flex flex-col w-60 shrink-0 border-r border-gray-200 bg-gray-50 pt-4 pb-4">
      <nav className="flex flex-col gap-1 px-3">
        {topItems.map((item) => {
          const Icon = iconMap[item.icon];
          return (
            <SidebarNavItem key={item.href} href={item.href} label={item.label}>
              {Icon && <Icon className="w-5 h-5" />}
            </SidebarNavItem>
          );
        })}
      </nav>
      <nav className="mt-auto flex flex-col gap-1 px-3">
        {bottomItems.map((item) => {
          const Icon = iconMap[item.icon];
          return (
            <SidebarNavItem key={item.href} href={item.href} label={item.label}>
              {Icon && <Icon className="w-5 h-5" />}
            </SidebarNavItem>
          );
        })}
      </nav>
    </aside>
  );
}
