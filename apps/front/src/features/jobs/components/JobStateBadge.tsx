import type { JobState } from '../api/types';

const STYLES: Record<JobState, string> = {
  // Reachable when the hub cannot map an agent-reported status; rendered
  // neutrally rather than hidden, so an unmappable job is still visible.
  unspecified: 'bg-neutral-700/60 text-neutral-400',
  accepted: 'bg-neutral-700/60 text-neutral-300',
  pulling_image: 'bg-sky-500/10 text-sky-400',
  creating_container: 'bg-sky-500/10 text-sky-400',
  starting_container: 'bg-sky-500/10 text-sky-400',
  running: 'bg-indigo-500/10 text-indigo-400',
  stopping: 'bg-amber-500/10 text-amber-400',
  cancelled: 'bg-neutral-700/60 text-neutral-400',
  succeeded: 'bg-emerald-500/10 text-emerald-400',
  failed: 'bg-red-500/10 text-red-400',
  timed_out: 'bg-red-500/10 text-red-400',
};

const ACTIVE_STATES: JobState[] = [
  'pulling_image',
  'creating_container',
  'starting_container',
  'running',
  'stopping',
];

export function formatJobState(state: JobState): string {
  return state.replace(/_/g, ' ');
}

export function JobStateBadge({ state }: { state: JobState }) {
  const style = STYLES[state] ?? 'bg-neutral-700/60 text-neutral-300';
  const active = ACTIVE_STATES.includes(state);

  return (
    <span
      className={`inline-flex shrink-0 items-center gap-1.5 rounded-full px-2.5 py-0.5 text-xs font-medium ${style}`}
    >
      {active && <span className="h-1.5 w-1.5 animate-pulse rounded-full bg-current" />}
      {formatJobState(state)}
    </span>
  );
}
