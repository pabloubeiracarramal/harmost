import {
  useAgent,
  useAgentDetailSocket,
  useAgentContainers,
  MetricsCard,
  ContainersCard,
  StatRow,
} from '@/features/agents';
import { PageContainer } from '@/shared/components/layout/page-container/PageContainer';

export function AgentDetailPage({ id }: { id: string }) {
  const { data: agent, isLoading } = useAgent(id);
  useAgentDetailSocket(id);
  const { containers, pending, performAction } = useAgentContainers(id);

  if (isLoading) {
    return (
      <PageContainer>
        <p className="text-neutral-500">Loading…</p>
      </PageContainer>
    );
  }

  if (!agent) {
    return (
      <PageContainer>
        <p className="text-neutral-500">Agent not found.</p>
      </PageContainer>
    );
  }

  const isOnline = agent.status === 'online';
  const displayName = agent.name !== 'pending' ? agent.name : agent.hostname || 'Unnamed agent';

  return (
    <PageContainer>
      <div className="mx-auto max-w-3xl space-y-8">
        <div className="flex items-start justify-between gap-4">
          <div>
            <h1 className="text-2xl font-bold tracking-tight">{displayName}</h1>
            {agent.hostname && (
              <p className="mt-1 text-sm text-muted-foreground">{agent.hostname}</p>
            )}
          </div>
          <span
            className={`flex shrink-0 items-center gap-1.5 rounded-full px-3 py-1 text-sm font-medium ${
              isOnline ? 'bg-emerald-500/10 text-emerald-400' : 'bg-neutral-700/60 text-neutral-400'
            }`}
          >
            <span className={`h-2 w-2 rounded-full ${isOnline ? 'bg-emerald-400 animate-pulse' : 'bg-neutral-500'}`} />
            {isOnline ? 'Online' : 'Offline'}
          </span>
        </div>

        {/* Metrics */}
        {isOnline && <MetricsCard agent={agent} />}

        {/* Containers */}
        {isOnline && (
          <ContainersCard containers={containers} pending={pending} onAction={performAction} />
        )}

        {/* Info */}
        <div className="rounded-xl border border-neutral-800 bg-neutral-900 px-5">
          <StatRow label="ID" value={agent.id} />
          {agent.description && <StatRow label="OS" value={agent.description} />}
          {agent.version && <StatRow label="Version" value={`v${agent.version}`} />}
          <StatRow label="Paired" value={new Date(agent.created_at).toLocaleString()} />
          {agent.last_seen_at && (
            <StatRow label="Last seen" value={new Date(agent.last_seen_at).toLocaleString()} />
          )}
        </div>
      </div>
    </PageContainer>
  );
}
