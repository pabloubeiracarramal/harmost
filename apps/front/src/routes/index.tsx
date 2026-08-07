import { createFileRoute, redirect } from '@tanstack/react-router';
import { isAuthenticated } from '@/shared/api/auth';

export const Route = createFileRoute('/')({
  beforeLoad: () => {
    if (isAuthenticated()) throw redirect({ to: '/dashboard' });
    throw redirect({ to: '/login' });
  },
});
