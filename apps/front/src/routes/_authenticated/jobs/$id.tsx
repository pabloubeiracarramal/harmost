import { createFileRoute } from '@tanstack/react-router';
import { JobDetailPage } from '@/pages/jobs/$id';

export const Route = createFileRoute('/_authenticated/jobs/$id')({
  component: RouteComponent,
});

function RouteComponent() {
  const { id } = Route.useParams();
  return <JobDetailPage id={id} />;
}
