import { useState } from 'react';
import type { ContainerActionKind, ContainerInfo } from '@/features/agents/api/types';
import type { PendingAction } from '@/features/agents/hooks/useAgentContainers';
import { formatRelativeTime } from '@/features/agents/lib/formatRelativeTime';
import { formatBytes } from '@/features/agents/lib/formatBytes';
import { Button } from '@/shared/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/shared/components/ui/card';

interface ContainersCardProps {
  containers: ContainerInfo[];
  pending: Record<string, PendingAction>;
  onAction: (containerId: string, action: ContainerActionKind) => void;
}

// paused/restarting still count as "running" for action-availability purposes
// — there's no separate pause/unpause action, Stop/Restart work on either.
const RUNNING_LIKE = new Set(['running', 'paused', 'restarting']);

const STATE_DOT: Record<string, string> = {
  running: 'bg-emerald-400 animate-pulse',
  paused: 'bg-amber-400 animate-pulse',
  restarting: 'bg-amber-400 animate-pulse',
};
const STATE_DOT_DEFAULT = 'bg-neutral-500';

function ActionButton({
  label,
  pendingLabel,
  disabled,
  pending,
  variant = 'outline',
  onClick,
}: {
  label: string;
  pendingLabel: string;
  disabled: boolean;
  pending: boolean;
  variant?: 'outline' | 'destructive';
  onClick: () => void;
}) {
  return (
    <Button variant={variant} size="xs" onClick={onClick} disabled={disabled || pending}>
      {pending ? pendingLabel : label}
    </Button>
  );
}

export function ContainersCard({ containers, pending, onAction }: ContainersCardProps) {
  const [expandedId, setExpandedId] = useState<string | null>(null);

  return (
    <Card className="gap-4 pb-4">
      <CardHeader>
        <CardTitle className="text-base">Containers</CardTitle>
        <CardDescription className="text-xs">Updated every ~5s</CardDescription>
      </CardHeader>

      <CardContent>
        {containers.length === 0 ? (
          <p className="text-sm text-neutral-500">No containers.</p>
        ) : (
          <div className="space-y-1">
            {containers.map((c) => {
              const running = RUNNING_LIKE.has(c.state);
              const rowPending = pending[c.id];
              const hasDetail = c.ports.length > 0 || c.volumes.length > 0;
              const expanded = expandedId === c.id;

              return (
                <div key={c.id} className="rounded-lg border border-neutral-800 px-3 py-2">
                  <div className="flex flex-wrap items-center gap-3">
                    <span
                      className={`h-2 w-2 shrink-0 rounded-full ${STATE_DOT[c.state] ?? STATE_DOT_DEFAULT}`}
                    />
                    <div className="min-w-0 flex-1">
                      <button
                        onClick={() => hasDetail && setExpandedId(expanded ? null : c.id)}
                        disabled={!hasDetail}
                        className={`truncate text-left text-sm font-medium text-white ${
                          hasDetail ? 'cursor-pointer hover:underline' : ''
                        }`}
                      >
                        {c.name}
                      </button>
                      <p className="truncate font-mono text-xs text-neutral-500">{c.image}</p>
                    </div>

                    {c.stats && (
                      <span className="shrink-0 text-xs text-neutral-400">
                        {c.stats.cpu_usage_percent.toFixed(0)}% · {formatBytes(c.stats.memory_usage_bytes)}
                      </span>
                    )}
                    <span className="shrink-0 text-xs text-neutral-400">
                      {formatRelativeTime(c.started_at)}
                    </span>
                    <span className="shrink-0 font-mono text-xs text-neutral-500">
                      {c.id.slice(0, 12)}
                    </span>

                    <div className="flex shrink-0 items-center gap-1.5">
                      <ActionButton
                        label="Start"
                        pendingLabel="Starting…"
                        disabled={running}
                        pending={rowPending?.action === 'start'}
                        onClick={() => onAction(c.id, 'start')}
                      />
                      <ActionButton
                        label="Stop"
                        pendingLabel="Stopping…"
                        disabled={!running}
                        pending={rowPending?.action === 'stop'}
                        onClick={() => onAction(c.id, 'stop')}
                      />
                      <ActionButton
                        label="Restart"
                        pendingLabel="Restarting…"
                        disabled={!running}
                        pending={rowPending?.action === 'restart'}
                        onClick={() => onAction(c.id, 'restart')}
                      />
                      <ActionButton
                        label="Remove"
                        pendingLabel="Removing…"
                        disabled={running}
                        pending={rowPending?.action === 'remove'}
                        variant="destructive"
                        onClick={() => {
                          if (window.confirm(`Remove container "${c.name}"?`)) onAction(c.id, 'remove');
                        }}
                      />
                    </div>
                  </div>

                  {rowPending?.error && (
                    <p className="mt-1.5 pl-5 text-xs text-red-400">{rowPending.error}</p>
                  )}

                  {expanded && hasDetail && (
                    <div className="mt-2 space-y-1 border-t border-neutral-800 pl-5 pt-2 text-xs text-neutral-400">
                      {c.ports.map((p, i) => (
                        <p key={`port-${i}`} className="font-mono">
                          {p.public_port
                            ? `${p.host_ip || '0.0.0.0'}:${p.public_port} → ${p.private_port}/${p.type}`
                            : `${p.private_port}/${p.type} (not published)`}
                        </p>
                      ))}
                      {c.volumes.map((v, i) => (
                        <p key={`mount-${i}`} className="truncate font-mono">
                          {v.name || v.source} → {v.destination}
                          {v.read_only ? ' (ro)' : ''}
                        </p>
                      ))}
                    </div>
                  )}
                </div>
              );
            })}
          </div>
        )}
      </CardContent>
    </Card>
  );
}
