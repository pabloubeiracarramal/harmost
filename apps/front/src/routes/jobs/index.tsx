import { createFileRoute, redirect } from '@tanstack/react-router';
import { isAuthenticated } from '@/shared/api/auth';
import { JobsPage } from '@/pages/jobs';

export const Route = createFileRoute('/jobs/')({
  beforeLoad: () => {
    if (!isAuthenticated()) throw redirect({ to: '/login' });
  },
  component: JobsPage,
});
