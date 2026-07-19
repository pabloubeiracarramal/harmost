import { createFileRoute, redirect, Link } from '@tanstack/react-router';
import { useQuery, useQueryClient, useMutation } from '@tanstack/react-query';
import { useEffect, useRef, useState } from 'react';
import { isAuthenticated } from '@/lib/auth';
import { api, isTerminal, type Job, type JobLog } from '@/lib/api';
import { useHubEvents, type LogLine } from '@/hooks/useHubEvents';
import { AppShell } from '@/components/AppShell';
import { JobStateBadge } from '@/components/JobStateBadge';

export const Route = createFileRoute('/jobs/$id')({
  beforeLoad: () => {
    if (!isAuthenticated()) throw redirect({ to: '/login' });
  },
  component: JobDetail,
});

function JobDetail() {
  const { id } = Route.useParams();
  const queryClient = useQueryClient();

  const { data: job, isLoading } = useQuery<Job>({
    queryKey: ['jobs', id],
    queryFn: () => api.get<Job>(`/api/v1/jobs/${id}`),
  });

  // ── logs: REST backfill merged with live WS chunks, deduped by sequence ──
  const [lines, setLines] = useState<LogLine[]>([]);
  const seenSeqs = useRef(new Set<number>());

  const appendLines = (incoming: LogLine[]) => {
    const fresh = incoming.filter((l) => !seenSeqs.current.has(l.sequence));
    if (!fresh.length) return;
    for (const l of fresh) seenSeqs.current.add(l.sequence);
    setLines((prev) => [...prev, ...fresh].sort((a, b) => a.sequence - b.sequence));
  };

  const { data: backfill } = useQuery<JobLog[]>({
    queryKey: ['jobs', id, 'logs'],
    queryFn: () => api.get<JobLog[]>(`/api/v1/jobs/${id}/logs`),
  });

  useEffect(() => {
    if (backfill) appendLines(backfill);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [backfill]);

  useHubEvents((e) => {
    if (e.type === 'job.status' && e.job_id === id) {
      queryClient.setQueryData<Job>(['jobs', id], (prev) =>
        prev
          ? {
              ...prev,
              state: e.payload.state,
              message: e.payload.message ?? prev.message,
              exit_code: e.payload.exit_code ?? prev.exit_code,
            }
          : prev
      );
      // started_at/finished_at aren't in the event — refetch on transitions.
      queryClient.invalidateQueries({ queryKey: ['jobs', id] });
    }
    if (e.type === 'job.log' && e.job_id === id) {
      appendLines(e.payload.lines);
    }
  });

  const cancel = useMutation({
    mutationFn: () => api.post(`/api/v1/jobs/${id}/cancel`),
    onSettled: () => queryClient.invalidateQueries({ queryKey: ['jobs', id] }),
  });

  if (isLoading) {
    return (
      <AppShell>
        <p className="text-neutral-500">Loading…</p>
      </AppShell>
    );
  }

  if (!job) {
    return (
      <AppShell>
        <p className="text-neutral-500">Job not found.</p>
      </AppShell>
    );
  }

  return (
    <AppShell>
      <div className="space-y-6">
        <div className="flex items-center gap-4">
          <Link to="/jobs" className="text-sm text-neutral-400 hover:text-white transition">
            ← Jobs
          </Link>
          <span className="text-neutral-700">/</span>
          <span className="truncate font-mono text-sm">{job.id.slice(0, 8)}</span>
        </div>

        <div className="flex items-start justify-between gap-4">
          <div className="min-w-0">
            <h1 className="truncate font-mono text-2xl font-bold">{job.spec.image}</h1>
            {job.message && <p className="mt-1 text-sm text-neutral-400">{job.message}</p>}
          </div>
          <div className="flex shrink-0 items-center gap-3">
            <JobStateBadge state={job.state} />
            {!isTerminal(job.state) && (
              <button
                onClick={() => cancel.mutate()}
                disabled={cancel.isPending}
                className="rounded-lg border border-red-500/40 px-3 py-1.5 text-sm font-medium text-red-400 transition hover:bg-red-500/10 disabled:opacity-50"
              >
                {cancel.isPending ? 'Cancelling…' : 'Cancel'}
              </button>
            )}
          </div>
        </div>

        {cancel.error && (
          <p className="text-sm text-red-400">{(cancel.error as Error).message}</p>
        )}

        {/* Spec + timing */}
        <div className="grid gap-4 sm:grid-cols-2">
          <div className="rounded-xl border border-neutral-800 bg-neutral-900 px-5">
            <DetailRow label="Job ID" value={job.id} />
            <DetailRow label="Agent" value={job.agent_id} link={`/agents/${job.agent_id}`} />
            {job.spec.command && <DetailRow label="Command" value={job.spec.command.join(' ')} />}
            {job.spec.args && <DetailRow label="Args" value={job.spec.args.join(' ')} />}
            {job.spec.timeout_seconds ? (
              <DetailRow label="Timeout" value={`${job.spec.timeout_seconds}s`} />
            ) : null}
            {job.spec.env && (
              <DetailRow label="Env" value={Object.keys(job.spec.env).join(', ')} />
            )}
          </div>
          <div className="rounded-xl border border-neutral-800 bg-neutral-900 px-5">
            <DetailRow label="Created" value={new Date(job.created_at).toLocaleString()} />
            {job.started_at && (
              <DetailRow label="Started" value={new Date(job.started_at).toLocaleString()} />
            )}
            {job.finished_at && (
              <DetailRow label="Finished" value={new Date(job.finished_at).toLocaleString()} />
            )}
            {job.started_at && job.finished_at && (
              <DetailRow
                label="Duration"
                value={formatDuration(
                  new Date(job.finished_at).getTime() - new Date(job.started_at).getTime()
                )}
              />
            )}
            {job.exit_code != null && (
              <DetailRow label="Exit code" value={String(job.exit_code)} />
            )}
          </div>
        </div>

        <LogViewer lines={lines} live={!isTerminal(job.state)} />
      </div>
    </AppShell>
  );
}

function DetailRow({ label, value, link }: { label: string; value: string; link?: string }) {
  return (
    <div className="flex items-center justify-between gap-4 border-b border-neutral-800 py-2.5 last:border-0">
      <span className="shrink-0 text-sm text-neutral-400">{label}</span>
      {link ? (
        <a href={link} className="truncate font-mono text-sm text-indigo-400 hover:text-indigo-300">
          {value}
        </a>
      ) : (
        <span className="truncate font-mono text-sm text-white">{value}</span>
      )}
    </div>
  );
}

function formatDuration(ms: number): string {
  const s = Math.round(ms / 1000);
  if (s < 60) return `${s}s`;
  const m = Math.floor(s / 60);
  return `${m}m ${s % 60}s`;
}

const AT_BOTTOM_THRESHOLD_PX = 40;

function LogViewer({ lines, live }: { lines: LogLine[]; live: boolean }) {
  const containerRef = useRef<HTMLDivElement>(null);
  // Pinned = follow new output. Scrolling up pauses; scrolling back down resumes.
  const [pinned, setPinned] = useState(true);

  useEffect(() => {
    const el = containerRef.current;
    if (el && pinned) el.scrollTop = el.scrollHeight;
  }, [lines, pinned]);

  const handleScroll = () => {
    const el = containerRef.current;
    if (!el) return;
    setPinned(el.scrollHeight - el.scrollTop - el.clientHeight < AT_BOTTOM_THRESHOLD_PX);
  };

  const jumpToBottom = () => {
    const el = containerRef.current;
    if (el) el.scrollTop = el.scrollHeight;
    setPinned(true);
  };

  return (
    <div className="overflow-hidden rounded-xl border border-neutral-800">
      <div className="flex items-center justify-between border-b border-neutral-800 bg-neutral-900 px-4 py-2.5">
        <span className="text-xs font-medium uppercase tracking-widest text-neutral-500">
          Logs
        </span>
        <span className="flex items-center gap-3">
          {live && (
            <span className="flex items-center gap-1.5 text-xs text-neutral-500">
              <span className="h-1.5 w-1.5 animate-pulse rounded-full bg-emerald-400" />
              live
            </span>
          )}
          {!pinned && (
            <button
              onClick={jumpToBottom}
              className="rounded-md bg-neutral-800 px-2 py-1 text-xs text-neutral-300 transition hover:bg-neutral-700"
            >
              ↓ Jump to bottom
            </button>
          )}
        </span>
      </div>
      <div
        ref={containerRef}
        onScroll={handleScroll}
        className="max-h-[28rem] overflow-y-auto bg-neutral-950 p-4 font-mono text-xs leading-5"
      >
        {lines.length === 0 ? (
          <p className="text-neutral-600">No output yet.</p>
        ) : (
          lines.map((l) => (
            <div key={l.sequence} className="flex gap-3 whitespace-pre-wrap break-all">
              <span className="shrink-0 select-none text-neutral-700">
                {String(l.sequence).padStart(4, ' ')}
              </span>
              <span className={l.stream === 'stderr' ? 'text-red-400' : 'text-neutral-200'}>
                {l.line}
              </span>
            </div>
          ))
        )}
      </div>
    </div>
  );
}
