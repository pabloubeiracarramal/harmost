import { useAgents } from '@/features/agents';
import { useAgentTokens, useRevokeToken } from '@/features/agent-tokens';
import { PageContainer } from '@/shared/components/layout/page-container/PageContainer';

export function TokensPage() {
  const { data: tokens = [], isLoading } = useAgentTokens();
  const { data: agents = [] } = useAgents();

  const agentName = (id?: string) => {
    if (!id) return '—';
    const a = agents.find((a) => a.id === id);
    return a ? (a.name !== 'pending' ? a.name : a.hostname) : id.slice(0, 8);
  };

  const revoke = useRevokeToken();

  return (
    <PageContainer title="Agent Tokens">
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
    </PageContainer>
  );
}
