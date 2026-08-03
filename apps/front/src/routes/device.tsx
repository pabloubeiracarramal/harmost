import { createFileRoute, redirect } from '@tanstack/react-router';
import { isAuthenticated } from '@/shared/api/auth';
import { DevicePage } from '@/pages/device';

export const Route = createFileRoute('/device')({
  beforeLoad: () => {
    if (!isAuthenticated()) throw redirect({ to: '/login' });
  },
  component: DevicePage,
});
