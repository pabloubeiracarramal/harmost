import { createFileRoute, redirect, useNavigate, Outlet } from '@tanstack/react-router';
import { isAuthenticated, clearToken } from '@/shared/api/auth';
import { useMe } from '@/features/auth';
import { AppLayout } from '@/shared/components/layout/app-layout/AppLayout';
import { SidebarContainer } from '@/shared/components/layout/sidebar/SidebarContainer';

export const Route = createFileRoute('/_authenticated')({
  beforeLoad: () => {
    if (!isAuthenticated()) throw redirect({ to: '/login' });
  },
  component: RouteComponent,
});

function RouteComponent() {
  const navigate = useNavigate();
  const { data: me } = useMe();

  const handleLogout = () => {
    clearToken();
    navigate({ to: '/login' });
  };

  return (
    <AppLayout sidebar={<SidebarContainer user={me} onLogout={handleLogout} />}>
      <Outlet />
    </AppLayout>
  );
}
