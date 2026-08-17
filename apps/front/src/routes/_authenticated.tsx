import { createFileRoute, redirect, Outlet } from '@tanstack/react-router';
import { isAuthenticated } from '@/shared/api/auth';
import { AppLayout } from '@/shared/components/layout/app-layout/AppLayout';

export const Route = createFileRoute('/_authenticated')({
  beforeLoad: () => {
    if (!isAuthenticated()) throw redirect({ to: '/login' });
  },
  component: RouteComponent,
});

function RouteComponent() {
  return (
    <AppLayout>
      <Outlet />
    </AppLayout>
  );
}
