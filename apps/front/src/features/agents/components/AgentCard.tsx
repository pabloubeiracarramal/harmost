import { Link } from '@tanstack/react-router';
import type { Agent } from '../api/types';

export function AgentCard({ agent }: { agent: Agent }) {
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
