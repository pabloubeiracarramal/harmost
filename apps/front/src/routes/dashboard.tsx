import { createFileRoute, redirect, Link } from '@tanstack/react-router';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { isAuthenticated } from '@/lib/auth';
import { api, type Agent } from '@/lib/api';
import { useHubEvents, type HubEvent } from '@/hooks/useHubEvents';
import { AppShell } from '@/components/AppShell';

export const Route = createFileRoute('/dashboard')({
  beforeLoad: () => {
    if (!isAuthenticated()) throw redirect({ to: '/login' });
  },
  component: Dashboard,
});

function Dashboard() {
  const queryClient = useQueryClient();

  const { data: agents = [], isLoading } = useQuery<Agent[]>({
    queryKey: ['agents'],
    queryFn: () => api.get<Agent[]>('/api/v1/agents'),
  });

  useHubEvents((e: HubEvent) => {
    queryClient.setQueryData<Agent[]>(['agents'], (prev = []) => {
      if (e.type === 'agent.connected') {
        return prev.map((a) =>
          a.id === e.agent_id ? { ...a, status: 'online', last_seen_at: e.at } : a
        );
      }
      if (e.type === 'agent.disconnected') {
        return prev.map((a) =>
          a.id === e.agent_id ? { ...a, status: 'offline' } : a
        );
      }
      if (e.type === 'agent.heartbeat') {
        return prev.map((a) =>
          a.id === e.agent_id ? { ...a, last_seen_at: e.at } : a
        );
      }
      return prev;
    });
  });

  return (
    <AppShell>
      <div className="mb-6 flex items-center justify-between">
        <h2 className="text-xl font-semibold">Agents</h2>
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
    </AppShell>
  );
}

function AgentCard({ agent }: { agent: Agent }) {
  const isOnline = agent.status === 'online';

  return (
    <Link
      to="/agents/$id"
      params={{ id: agent.id }}
      className="block rounded-xl border border-neutral-800 bg-neutral-900 p-5 hover:border-neutral-700 transition"
    >
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0">
          <p className="truncate font-medium">{agent.name}</p>
          <p className="truncate text-sm text-neutral-400">{agent.hostname}</p>
        </div>
        <span
          className={`mt-0.5 flex shrink-0 items-center gap-1.5 rounded-full px-2.5 py-0.5 text-xs font-medium ${
            isOnline
              ? 'bg-emerald-500/10 text-emerald-400'
              : 'bg-neutral-700/60 text-neutral-400'
          }`}
        >
          <span
            className={`h-1.5 w-1.5 rounded-full ${isOnline ? 'bg-emerald-400' : 'bg-neutral-500'}`}
          />
          {isOnline ? 'Online' : 'Offline'}
        </span>
      </div>
      {agent.version && (
        <p className="mt-3 text-xs text-neutral-500">v{agent.version}</p>
      )}
      {agent.last_seen_at && (
        <p className="mt-1 text-xs text-neutral-600">
          Last seen {new Date(agent.last_seen_at).toLocaleString()}
        </p>
      )}
    </Link>
  );
}

function EmptyState() {
  return (
    <div className="rounded-xl border border-dashed border-neutral-700 bg-neutral-900/50 p-10 text-center">
      <p className="text-neutral-400">No agents yet.</p>
      <p className="mt-1 text-sm text-neutral-600">
        Run <code className="rounded bg-neutral-800 px-1 py-0.5 font-mono text-xs">harmost pair &lt;hub-url&gt;</code> on a machine to get started.
      </p>
    </div>
  );
}
