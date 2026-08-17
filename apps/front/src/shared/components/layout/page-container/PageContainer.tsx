import type { ReactNode } from 'react';

interface PageContainerProps {
  children: ReactNode;
}

export function PageContainer({ children }: PageContainerProps) {
  return (
    <div className="mx-auto flex h-full max-w-7xl flex-col p-6 lg:p-8">
      {children}
    </div>
  );
}
