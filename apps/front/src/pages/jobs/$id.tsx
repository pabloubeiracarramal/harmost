import { Link } from '@tanstack/react-router';
import { useEffect, useRef, useState } from 'react';
import { AppShell } from '@/app/AppShell';
import type { LogLine } from '@/shared/ws/wsClient';
import {
  useJob,
  useJobLogs,
  useCancelJob,
  useJobDetailSocket,
  useJobLogSocket,
  isTerminal,
  JobStateBadge,
  DetailRow,
  LogViewer,
} from '@/features/jobs';

function formatDuration(ms: number): string {
  const s = Math.round(ms / 1000);
  if (s < 60) return `${s}s`;
  const m = Math.floor(s / 60);
  return `${m}m ${s % 60}s`;
}

export function JobDetailPage({ id }: { id: string }) {
  const { data: job, isLoading } = useJob(id);

  // ── logs: REST backfill merged with live WS chunks, deduped by sequence ──
  const [lines, setLines] = useState<LogLine[]>([]);
  const seenSeqs = useRef(new Set<number>());

  const appendLines = (incoming: LogLine[]) => {
    const fresh = incoming.filter((l) => !seenSeqs.current.has(l.sequence));
    if (!fresh.length) return;
    for (const l of fresh) seenSeqs.current.add(l.sequence);
    setLines((prev) => [...prev, ...fresh].sort((a, b) => a.sequence - b.sequence));
  };

  const { data: backfill } = useJobLogs(id);

  useEffect(() => {
    if (backfill) appendLines(backfill);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [backfill]);

  useJobDetailSocket(id);
  useJobLogSocket(id, appendLines);

  const cancel = useCancelJob(id);

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
