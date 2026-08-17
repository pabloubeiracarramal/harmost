import { useAgents, useAgentsListSocket, AgentCard, EmptyState } from '@/features/agents';
import { PageContainer } from '@/shared/components/layout/page-container/PageContainer';

export function DashboardPage() {
  const { data: agents = [], isLoading } = useAgents();
  useAgentsListSocket();

  return (
    <PageContainer>
      <div className="mb-6 flex items-center justify-between gap-4">
        <p className="text-sm text-muted-foreground">Machines running the Harmost agent.</p>
        <a
          href="/device"
          className="rounded-lg bg-indigo-600 px-4 py-2 text-sm font-medium hover:bg-indigo-500 transition"
        >
          Pair new agent
        </a>
      </div>
      {isLoading ? (
        <p className="text-neutral-500">Loading…</p>
      ) : agents.length === 0 ? (
        <EmptyState />
      ) : (
        <div className="grid gap-3 sm:grid-cols-2">
          {agents.map((agent) => (
            <AgentCard key={agent.id} agent={agent} />
          ))}
        </div>
      )}
    </PageContainer>
  );
}
