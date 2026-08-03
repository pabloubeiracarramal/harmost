import { createFileRoute, redirect } from '@tanstack/react-router';
import { isAuthenticated } from '@/shared/api/auth';
import { JobDetailPage } from '@/pages/jobs/$id';

export const Route = createFileRoute('/jobs/$id')({
  beforeLoad: () => {
    if (!isAuthenticated()) throw redirect({ to: '/login' });
  },
  component: RouteComponent,
});

function RouteComponent() {
  const { id } = Route.useParams();
  return <JobDetailPage id={id} />;
}
