import { ReactNode } from 'react';

interface AuthLayoutProps {
  children: ReactNode;
}

export function AuthLayout({ children }: AuthLayoutProps) {
  return (
    <section className="relative flex items-center justify-center min-h-screen overflow-hidden">
      {/* Dark background with gradient overlay */}
      <div className="absolute inset-0 bg-gradient-to-br from-gray-900 via-gray-800 to-black">
        {/* Decorative product image grid effect */}
        <div className="absolute inset-0 opacity-30">
          <div className="grid grid-cols-4 gap-4 p-4 transform -rotate-12 scale-125 origin-center">
            {Array.from({ length: 16 }).map((_, i) => (
              <div
                key={i}
                className="aspect-square rounded-lg bg-gradient-to-br from-gray-700 to-gray-800"
              />
            ))}
          </div>
        </div>
        {/* Additional gradient overlay for depth */}
        <div className="absolute inset-0 bg-gradient-to-t from-black/60 via-transparent to-black/40" />
      </div>

      {/* Logo */}
      <div className="absolute top-6 left-6 z-10">
        <div className="w-8 h-8 bg-white rounded flex items-center justify-center">
          <span className="text-black font-bold text-lg">S</span>
        </div>
      </div>

      {/* Content */}
      <div className="relative z-10">{children}</div>
    </section>
  );
}
