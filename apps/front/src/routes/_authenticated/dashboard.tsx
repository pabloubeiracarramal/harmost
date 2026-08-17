import { createFileRoute } from '@tanstack/react-router';
import { DashboardPage } from '@/pages/dashboard';

export const Route = createFileRoute('/_authenticated/dashboard')({
  validateSearch: (search: Record<string, unknown>): { code?: string } => ({
    code: typeof search.code === 'string' ? search.code : undefined,
  }),
  component: RouteComponent,
});

function RouteComponent() {
  const { code } = Route.useSearch();
  return <DashboardPage pairCode={code} />;
}
