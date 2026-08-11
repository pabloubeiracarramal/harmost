import { useQueryClient } from '@tanstack/react-query';
import { useWsSubscribe } from '@/shared/ws/useWsSubscribe';
import type { HubEvent, LogLine } from '@/shared/ws/wsClient';
import { jobKeys } from './keys';
import type { Job } from './types';

/** Applies job.status events to the jobs list cache; refetches on jobs the list hasn't seen yet. */
export function useJobsListSocket() {
  const queryClient = useQueryClient();

  useWsSubscribe((e: HubEvent) => {
    if (e.type !== 'job.status') return;
    queryClient.setQueryData<Job[]>(jobKeys.list(), (prev) => {
      if (!prev) return prev;
      if (!prev.some((j) => j.id === e.job_id)) {
        // A job we've never seen (dispatched elsewhere) — refetch the list.
        queryClient.invalidateQueries({ queryKey: jobKeys.lists() });
        return prev;
      }
      return prev.map((j) =>
        j.id === e.job_id
          ? {
              ...j,
              state: e.payload.state,
              message: e.payload.message ?? j.message,
              exit_code: e.payload.exit_code ?? j.exit_code,
            }
          : j
      );
    });
  });
}

/** Applies job.status events for a single job to its detail cache. */
export function useJobDetailSocket(id: string) {
  const queryClient = useQueryClient();

  useWsSubscribe((e: HubEvent) => {
    if (e.type === 'job.status' && e.job_id === id) {
      queryClient.setQueryData<Job>(jobKeys.detail(id), (prev) =>
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
      queryClient.invalidateQueries({ queryKey: jobKeys.detail(id) });
    }
  });
}

/**
 * Forwards live job.log lines for a single job. Append-only streams stay out
 * of the query cache — the caller owns its own buffer (seeded from a REST
 * backfill), this just feeds it.
 */
export function useJobLogSocket(id: string, onLines: (lines: LogLine[]) => void) {
  useWsSubscribe((e: HubEvent) => {
    if (e.type === 'job.log' && e.job_id === id) onLines(e.payload.lines);
  });
}
