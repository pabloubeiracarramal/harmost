import { Link } from '@tanstack/react-router';
import { AppShell } from '@/app/AppShell';
import {
  useAgent,
  useAgentDetailSocket,
  useAgentContainers,
  MetricsCard,
  ContainersCard,
  StatRow,
} from '@/features/agents';

export function AgentDetailPage({ id }: { id: string }) {
  const { data: agent, isLoading } = useAgent(id);
  useAgentDetailSocket(id);
  const containers = useAgentContainers(id);

  if (isLoading) {
    return (
      <AppShell>
        <p className="text-neutral-500">Loading…</p>
      </AppShell>
    );
  }

  if (!agent) {
    return (
      <AppShell>
        <p className="text-neutral-500">Agent not found.</p>
      </AppShell>
    );
  }

  const isOnline = agent.status === 'online';

  return (
    <AppShell>
      <div className="mx-auto max-w-3xl space-y-8">
        <div className="flex items-center gap-4">
          <Link to="/dashboard" className="text-neutral-400 hover:text-white transition text-sm">
            ← Agents
          </Link>
          <span className="text-neutral-700">/</span>
          <span className="text-sm font-medium truncate">{agent.name !== 'pending' ? agent.name : agent.hostname || 'Unnamed'}</span>
        </div>
        {/* Header */}
        <div className="flex items-start justify-between gap-4">
          <div>
            <h1 className="text-2xl font-bold">
              {agent.name !== 'pending' ? agent.name : agent.hostname || 'Unnamed agent'}
            </h1>
            <p className="text-neutral-400 mt-1">{agent.hostname}</p>
          </div>
          <span
            className={`mt-1 flex shrink-0 items-center gap-1.5 rounded-full px-3 py-1 text-sm font-medium ${
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
        {isOnline && <ContainersCard containers={containers} />}

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
    </AppShell>
  );
}
