import { createFileRoute, redirect } from '@tanstack/react-router';
import { isAuthenticated } from '@/shared/api/auth';
import { DashboardPage } from '@/pages/dashboard';

export const Route = createFileRoute('/dashboard')({
  beforeLoad: () => {
    if (!isAuthenticated()) throw redirect({ to: '/login' });
  },
  component: DashboardPage,
});
