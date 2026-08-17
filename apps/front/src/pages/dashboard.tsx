import { useEffect, useState } from 'react';
import { useNavigate } from '@tanstack/react-router';
import {
  useAgents,
  useAgentsListSocket,
  AgentCard,
  EmptyState,
  PairAgentDialog,
} from '@/features/agents';
import { Button } from '@/shared/components/ui/button';

export function DashboardPage({ pairCode }: { pairCode?: string }) {
  const { data: agents = [], isLoading } = useAgents();
  useAgentsListSocket();
  const [pairOpen, setPairOpen] = useState(!!pairCode);
  const navigate = useNavigate();

  // A code in the URL (from the link `agent pair` prints) opens the dialog
  // pre-filled once, then the URL is cleaned so a refresh doesn't reopen it.
  useEffect(() => {
    if (pairCode) {
      navigate({ to: '/dashboard', search: {}, replace: true });
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  return (
    <>
      <div className="mb-6 flex items-center justify-between gap-4">
        <p className="text-sm text-muted-foreground">
          Machines running the Harmost agent.
        </p>
        <Button onClick={() => setPairOpen(true)}>Pair new agent</Button>
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
      <PairAgentDialog
        open={pairOpen}
        onOpenChange={setPairOpen}
        initialCode={pairCode}
      />
    </>
  );
}
