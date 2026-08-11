import type { Agent } from '@/features/agents/api/types';
import { useMetricsHistory } from '@/features/agents/hooks/useMetricsHistory';
import { formatRelativeTime } from '@/features/agents/lib/formatRelativeTime';
import { MetricAreaChart } from './MetricAreaChart';
import {
  Card,
  CardAction,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from '@/shared/components/ui/card';
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/shared/components/ui/tooltip';

interface MetricsCardProps {
  agent: Agent;
}

function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${(bytes / Math.pow(k, i)).toFixed(1)} ${sizes[i]}`;
}

function formatCapacity(used: number | undefined, total: number | undefined): string | undefined {
  return total ? `${formatBytes(used ?? 0)} of ${formatBytes(total)}` : undefined;
}

function percent(used: number | undefined, total: number | undefined): number {
  return total ? ((used ?? 0) / total) * 100 : 0;
}

export function MetricsCard({ agent }: MetricsCardProps) {
  const samples = useMetricsHistory(agent);
  const containers = agent.running_containers ?? 0;

  // `pb-4` matches the footer's `pt-4`: Card's own `py-6` left 24px under the
  // footer text against 16px above it, so the line sat high in its band.
  return (
    <Card className="gap-4 pb-4">
      <CardHeader>
        <CardTitle className="text-base">System metrics</CardTitle>
        {/* The window itself is labelled on every plot, so this only carries
            what the span labels can't: how coarse the series is. */}
        <CardDescription className="text-xs">One point per heartbeat (~30s)</CardDescription>

        {agent.last_seen_at && (
          <CardAction>
            <TooltipProvider>
              <Tooltip>
                <TooltipTrigger asChild>
                  <span className="cursor-help text-xs text-muted-foreground transition-colors hover:text-foreground">
                    Updated {formatRelativeTime(agent.last_seen_at)}
                  </span>
                </TooltipTrigger>
                <TooltipContent>
                  <p>{new Date(agent.last_seen_at).toLocaleString()}</p>
                </TooltipContent>
              </Tooltip>
            </TooltipProvider>
          </CardAction>
        )}
      </CardHeader>

      <CardContent>
        <div className="grid grid-cols-1 gap-3 sm:grid-cols-3">
          <MetricAreaChart
            dataKey="cpu"
            label="CPU"
            value={agent.cpu_usage_percent ?? 0}
            samples={samples}
          />
          <MetricAreaChart
            dataKey="memory"
            label="Memory"
            value={percent(agent.memory_used_bytes, agent.memory_total_bytes)}
            sublabel={formatCapacity(agent.memory_used_bytes, agent.memory_total_bytes)}
            samples={samples}
          />
          <MetricAreaChart
            dataKey="disk"
            label="Disk"
            value={percent(agent.disk_used_bytes, agent.disk_total_bytes)}
            sublabel={formatCapacity(agent.disk_used_bytes, agent.disk_total_bytes)}
            samples={samples}
          />
        </div>
      </CardContent>

      {/* Always rendered: the container count dropping to zero must not resize
          the card, and the sampling hint needs somewhere to live that doesn't
          push the tiles around when it clears.

          `border-t-[1px]` rather than `border-t` on purpose — CardFooter carries
          a `[.border-t]:pt-6` variant that would beat this `pt-4`, and 24px on
          top of the card's own gap leaves the rule floating. */}
      <CardFooter className="justify-between gap-4 border-t-[1px] pt-4 text-xs text-muted-foreground">
        <span>
          <span className="font-medium text-foreground">{containers}</span> running container
          {containers !== 1 ? 's' : ''}
        </span>
        {samples.length < 2 && <span className="truncate">Collecting samples…</span>}
      </CardFooter>
    </Card>
  );
}
