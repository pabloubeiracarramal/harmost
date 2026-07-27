import { createFileRoute, redirect } from '@tanstack/react-router';
import { useQuery, useQueryClient, useMutation } from '@tanstack/react-query';
import { isAuthenticated } from '@/lib/auth';
import { api, type AgentToken, type Agent } from '@/lib/api';
import { AppShell } from '@/components/AppShell';

export const Route = createFileRoute('/tokens')({
  beforeLoad: () => {
    if (!isAuthenticated()) throw redirect({ to: '/login' });
  },
  component: TokensPage,
});

function TokensPage() {
  const queryClient = useQueryClient();

  const { data: tokens = [], isLoading } = useQuery<AgentToken[]>({
    queryKey: ['agent-tokens'],
    queryFn: () => api.get<AgentToken[]>('/api/v1/agent-tokens'),
  });

  const { data: agents = [] } = useQuery<Agent[]>({
    queryKey: ['agents'],
    queryFn: () => api.get<Agent[]>('/api/v1/agents'),
  });

  const agentName = (id?: string) => {
    if (!id) return '—';
    const a = agents.find((a) => a.id === id);
    return a ? (a.name !== 'pending' ? a.name : a.hostname) : id.slice(0, 8);
  };

  const revoke = useMutation({
    mutationFn: (id: string) => api.post(`/api/v1/agent-tokens/${id}/revoke`),
    onSettled: () => queryClient.invalidateQueries({ queryKey: ['agent-tokens'] }),
  });

  return (
    <AppShell>
      <div className="mb-6 flex items-center justify-between">
        <h2 className="text-xl font-semibold">Agent Tokens</h2>
      </div>

      {isLoading ? (
        <p className="text-neutral-500">Loading…</p>
      ) : tokens.length === 0 ? (
        <div className="rounded-xl border border-dashed border-neutral-700 bg-neutral-900/50 p-10 text-center">
          <p className="text-neutral-400">No agent tokens yet.</p>
          <p className="mt-1 text-sm text-neutral-600">
            Run <code className="text-neutral-400">agent pair &lt;hub-url&gt;</code> to pair a new
            agent.
          </p>
        </div>
      ) : (
        <div className="overflow-hidden rounded-xl border border-neutral-800">
          <table className="w-full text-left text-sm">
            <thead className="bg-neutral-900 text-xs uppercase tracking-wider text-neutral-500">
              <tr>
                <th className="px-4 py-3 font-medium">Name</th>
                <th className="px-4 py-3 font-medium">Agent</th>
                <th className="px-4 py-3 font-medium">Created</th>
                <th className="px-4 py-3 font-medium">Last used</th>
                <th className="px-4 py-3 font-medium" />
              </tr>
            </thead>
            <tbody className="divide-y divide-neutral-800 bg-neutral-900/50">
              {tokens.map((token) => (
                <tr key={token.id} className="hover:bg-neutral-800/50 transition">
                  <td className="px-4 py-3 font-mono text-white">{token.name}</td>
                  <td className="px-4 py-3 text-neutral-400">{agentName(token.agent_id)}</td>
                  <td className="px-4 py-3 text-neutral-500">
                    {new Date(token.created_at).toLocaleString()}
                  </td>
                  <td className="px-4 py-3 text-neutral-500">
                    {token.last_used_at ? new Date(token.last_used_at).toLocaleString() : 'Never'}
                  </td>
                  <td className="px-4 py-3 text-right">
                    <button
                      onClick={() => {
                        if (confirm(`Revoke token "${token.name}"? This cannot be undone.`)) {
                          revoke.mutate(token.id);
                        }
                      }}
                      disabled={revoke.isPending}
                      className="text-sm text-red-400 hover:text-red-300 transition disabled:opacity-50"
                    >
                      Revoke
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </AppShell>
  );
}
