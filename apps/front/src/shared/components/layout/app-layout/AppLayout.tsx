import type { ReactNode } from 'react';
import { TopBar } from '@/shared/components/layout/top-bar/TopBar';
import { PageContainer } from '@/shared/components/layout/page-container/PageContainer';
import { SidebarContainer } from '../sidebar/SidebarContainer';

interface AppLayoutProps {
  children: ReactNode;
}

export function AppLayout({  children }: AppLayoutProps) {
  return (
    <div className="flex h-screen w-full overflow-hidden bg-background text-foreground">
      <SidebarContainer/>
      <div className="flex flex-1 flex-col overflow-hidden">
        <TopBar />
        <main className="flex-1 overflow-y-auto">
          <PageContainer>{children}</PageContainer>
        </main>
      </div>
    </div>
  );
}
