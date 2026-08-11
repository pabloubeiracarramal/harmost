import type { ContainerInfo } from '@/features/agents/api/types';
import { formatRelativeTime } from '@/features/agents/lib/formatRelativeTime';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/shared/components/ui/card';

interface ContainersCardProps {
  containers: ContainerInfo[];
}

export function ContainersCard({ containers }: ContainersCardProps) {
  return (
    <Card className="gap-4 pb-4">
      <CardHeader>
        <CardTitle className="text-base">Containers</CardTitle>
        <CardDescription className="text-xs">Running only, updated every ~5s</CardDescription>
      </CardHeader>

      <CardContent>
        {containers.length === 0 ? (
          <p className="text-sm text-neutral-500">No running containers.</p>
        ) : (
          <div className="space-y-1">
            {containers.map((c) => (
              <div
                key={c.id}
                className="flex items-center gap-3 rounded-lg border border-neutral-800 px-3 py-2"
              >
                <span className="h-2 w-2 shrink-0 animate-pulse rounded-full bg-emerald-400" />
                <div className="min-w-0 flex-1">
                  <p className="truncate text-sm font-medium text-white">{c.name}</p>
                  <p className="truncate font-mono text-xs text-neutral-500">{c.image}</p>
                </div>
                <span className="shrink-0 text-xs text-neutral-400">
                  {formatRelativeTime(c.started_at)}
                </span>
                <span className="shrink-0 font-mono text-xs text-neutral-500">
                  {c.id.slice(0, 12)}
                </span>
              </div>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  );
}
