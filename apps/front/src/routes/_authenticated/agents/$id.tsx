import { createFileRoute } from '@tanstack/react-router';
import { AgentDetailPage } from '@/pages/agents/$id';

export const Route = createFileRoute('/_authenticated/agents/$id')({
  component: RouteComponent,
});

function RouteComponent() {
  const { id } = Route.useParams();
  return <AgentDetailPage id={id} />;
}
