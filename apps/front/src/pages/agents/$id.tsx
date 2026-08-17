import { useNavigate } from '@tanstack/react-router';
import {
  useAgent,
  useAgentDetailSocket,
  useAgentContainers,
  useDeleteAgent,
  MetricsCard,
  ContainersCard,
  StatRow,
} from '@/features/agents';
import { Button } from '@/shared/components/ui/button';
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from '@/shared/components/ui/alert-dialog';

export function AgentDetailPage({ id }: { id: string }) {
  const { data: agent, isLoading } = useAgent(id);
  useAgentDetailSocket(id);
  const { containers, pending, performAction } = useAgentContainers(id);
  const deleteAgent = useDeleteAgent();
  const navigate = useNavigate();

  if (isLoading) {
    return <p className="text-neutral-500">Loading…</p>;
  }

  if (!agent) {
    return <p className="text-neutral-500">Agent not found.</p>;
  }

  const isOnline = agent.status === 'online';
  const displayName =
    agent.name !== 'pending' ? agent.name : agent.hostname || 'Unnamed agent';

  return (
    <div className="space-y-8">
      <div className="flex items-start justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">{displayName}</h1>
          {agent.hostname && (
            <p className="mt-1 text-sm text-muted-foreground">
              {agent.hostname}
            </p>
          )}
        </div>
        <div className="flex shrink-0 items-center gap-3">
          <span
            className={`flex items-center gap-1.5 rounded-full px-3 py-1 text-sm font-medium ${
              isOnline
                ? 'bg-emerald-500/10 text-emerald-400'
                : 'bg-neutral-700/60 text-neutral-400'
            }`}
          >
            <span
              className={`h-2 w-2 rounded-full ${
                isOnline ? 'bg-emerald-400 animate-pulse' : 'bg-neutral-500'
              }`}
            />
            {isOnline ? 'Online' : 'Offline'}
          </span>
          <AlertDialog>
            <AlertDialogTrigger asChild>
              <Button variant="destructive" size="sm">
                Unpair agent
              </Button>
            </AlertDialogTrigger>
            <AlertDialogContent>
              <AlertDialogHeader>
                <AlertDialogTitle>Unpair {displayName}?</AlertDialogTitle>
                <AlertDialogDescription>
                  This removes the agent from your dashboard and revokes its
                  pairing token, so it can no longer connect. Its job history is
                  kept. This can't be undone from the UI — pair it again to
                  reconnect.
                </AlertDialogDescription>
              </AlertDialogHeader>
              <AlertDialogFooter>
                <AlertDialogCancel>Cancel</AlertDialogCancel>
                <AlertDialogAction
                  variant="destructive"
                  disabled={deleteAgent.isPending}
                  onClick={() =>
                    deleteAgent.mutate(id, {
                      onSuccess: () => navigate({ to: '/dashboard' }),
                    })
                  }
                >
                  {deleteAgent.isPending ? 'Unpairing…' : 'Unpair agent'}
                </AlertDialogAction>
              </AlertDialogFooter>
            </AlertDialogContent>
          </AlertDialog>
        </div>
      </div>

      {/* Metrics */}
      {isOnline && <MetricsCard agent={agent} />}

      {/* Containers */}
      {isOnline && (
        <ContainersCard
          containers={containers}
          pending={pending}
          onAction={performAction}
        />
      )}

      {/* Info */}
      <div className="rounded-xl border border-neutral-800 bg-neutral-900 px-5">
        <StatRow label="ID" value={agent.id} />
        {agent.description && <StatRow label="OS" value={agent.description} />}
        {agent.version && (
          <StatRow label="Version" value={`v${agent.version}`} />
        )}
        <StatRow
          label="Paired"
          value={new Date(agent.created_at).toLocaleString()}
        />
        {agent.last_seen_at && (
          <StatRow
            label="Last seen"
            value={new Date(agent.last_seen_at).toLocaleString()}
          />
        )}
      </div>
    </div>
  );
}
