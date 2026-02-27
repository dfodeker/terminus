import { Search } from './search';
import { ProfileMenu } from './profile-menu';

export function TopBar() {
  return (
    <header className="flex items-center justify-between h-14 px-6 border-b border-gray-200 shrink-0 bg-gray-900 text-white">
      <div className="flex items-center gap-4">
        <span className="text-lg font-semibold">StoreOS</span>
      </div>
      <div className="flex-1 max-w-xl mx-8">
        <Search />
      </div>
      <div className="flex items-center">
        <ProfileMenu />
      </div>
    </header>
  );
}
