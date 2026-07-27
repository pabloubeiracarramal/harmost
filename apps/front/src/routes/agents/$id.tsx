import { createFileRoute, redirect, Link } from '@tanstack/react-router';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { isAuthenticated } from '@/lib/auth';
import { api, type Agent } from '@/lib/api';
import { useHubEvents } from '@/hooks/useHubEvents';
import { AppShell } from '@/components/AppShell';

export const Route = createFileRoute('/agents/$id')({
  beforeLoad: () => {
    if (!isAuthenticated()) throw redirect({ to: '/login' });
  },
  component: AgentDetail,
});

function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${(bytes / Math.pow(k, i)).toFixed(1)} ${sizes[i]}`;
}

function ArcGauge({ percent, label, sublabel }: { percent: number; label: string; sublabel?: string }) {
  const r = 42;
  const cx = 56;
  const cy = 56;
  const startAngle = -210;
  const endAngle = 30;
  const totalArc = endAngle - startAngle;
  const filledArc = (Math.min(Math.max(percent, 0), 100) / 100) * totalArc;

  function polarToCart(angleDeg: number, radius: number) {
    const a = ((angleDeg - 90) * Math.PI) / 180;
    return { x: cx + radius * Math.cos(a), y: cy + radius * Math.sin(a) };
  }

  function describeArc(start: number, end: number, radius: number) {
    const s = polarToCart(start, radius);
    const e = polarToCart(end, radius);
    const large = end - start > 180 ? 1 : 0;
    return `M ${s.x} ${s.y} A ${radius} ${radius} 0 ${large} 1 ${e.x} ${e.y}`;
  }

  const hue = 142 - (percent / 100) * 80;
  const color = `hsl(${hue}, 70%, 55%)`;

  return (
    <div className="flex flex-col items-center gap-2">
      <svg width="112" height="80" viewBox="0 0 112 88">
        <path
          d={describeArc(startAngle, endAngle, r)}
          fill="none"
          stroke="#262626"
          strokeWidth="10"
          strokeLinecap="round"
        />
        {percent > 0 && (
          <path
            d={describeArc(startAngle, startAngle + filledArc, r)}
            fill="none"
            stroke={color}
            strokeWidth="10"
            strokeLinecap="round"
          />
        )}
        <text x={cx} y={cy + 4} textAnchor="middle" fill="white" fontSize="16" fontWeight="600">
          {percent > 0 ? `${percent.toFixed(1)}%` : '—'}
        </text>
      </svg>
      <p className="text-sm font-medium text-white">{label}</p>
      {sublabel && <p className="text-xs text-neutral-500">{sublabel}</p>}
    </div>
  );
}

function StatRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-center justify-between py-2.5 border-b border-neutral-800 last:border-0">
      <span className="text-sm text-neutral-400">{label}</span>
      <span className="text-sm font-mono text-white">{value}</span>
    </div>
  );
}

function AgentDetail() {
  const { id } = Route.useParams();
  const queryClient = useQueryClient();

  const { data: agent, isLoading } = useQuery<Agent>({
    queryKey: ['agents', id],
    queryFn: () => api.get<Agent>(`/api/v1/agents/${id}`),
    refetchInterval: 35_000,
  });

  useHubEvents((e) => {
    if (e.type.startsWith('agent.') && e.agent_id === id) {
      queryClient.invalidateQueries({ queryKey: ['agents', id] });
    }
  });

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
  const memPercent = agent.memory_total_bytes
    ? ((agent.memory_used_bytes ?? 0) / agent.memory_total_bytes) * 100
    : 0;
  const diskPercent = agent.disk_total_bytes
    ? ((agent.disk_used_bytes ?? 0) / agent.disk_total_bytes) * 100
    : 0;

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

        {/* Gauges */}
        {isOnline && (
          <div className="rounded-xl border border-neutral-800 bg-neutral-900 p-6">
            <h2 className="text-sm font-medium text-neutral-400 uppercase tracking-widest mb-6">
              System Metrics
            </h2>
            <div className="grid grid-cols-3 gap-4">
              <ArcGauge
                percent={agent.cpu_usage_percent ?? 0}
                label="CPU"
              />
              <ArcGauge
                percent={memPercent}
                label="Memory"
                sublabel={
                  agent.memory_total_bytes
                    ? `${formatBytes(agent.memory_used_bytes ?? 0)} / ${formatBytes(agent.memory_total_bytes)}`
                    : undefined
                }
              />
              <ArcGauge
                percent={diskPercent}
                label="Disk"
                sublabel={
                  agent.disk_total_bytes
                    ? `${formatBytes(agent.disk_used_bytes ?? 0)} / ${formatBytes(agent.disk_total_bytes)}`
                    : undefined
                }
              />
            </div>
            {(agent.running_containers ?? 0) > 0 && (
              <p className="mt-6 text-center text-sm text-neutral-400">
                <span className="text-white font-medium">{agent.running_containers}</span> running container{agent.running_containers !== 1 ? 's' : ''}
              </p>
            )}
          </div>
        )}

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
