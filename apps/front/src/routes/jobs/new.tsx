import { createFileRoute, redirect } from '@tanstack/react-router';
import { isAuthenticated } from '@/shared/api/auth';
import { NewJobPage } from '@/pages/jobs/new';

export const Route = createFileRoute('/jobs/new')({
  beforeLoad: () => {
    if (!isAuthenticated()) throw redirect({ to: '/login' });
  },
  component: NewJobPage,
});
