import { SidebarHeader } from './SidebarHeader';
import { SidebarFooter } from './SidebarFooter';
import { SidebarNav } from './SidebarNav';

export function SidebarContainer() {
  return (
    <aside className="flex h-full w-64 shrink-0 flex-col border-r bg-background">
      
      <SidebarHeader/>

      <SidebarNav/>

      <SidebarFooter/>

    </aside>
  );
}
