import { createFileRoute, redirect } from '@tanstack/react-router';
import { isAuthenticated } from '@/shared/api/auth';
import { AgentDetailPage } from '@/pages/agents/$id';

export const Route = createFileRoute('/agents/$id')({
  beforeLoad: () => {
    if (!isAuthenticated()) throw redirect({ to: '/login' });
  },
  component: RouteComponent,
});

function RouteComponent() {
  const { id } = Route.useParams();
  return <AgentDetailPage id={id} />;
}
