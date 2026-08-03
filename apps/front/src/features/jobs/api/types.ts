export type JobState =
  | 'accepted'
  | 'pulling_image'
  | 'creating_container'
  | 'starting_container'
  | 'running'
  | 'stopping'
  | 'cancelled'
  | 'succeeded'
  | 'failed'
  | 'timed_out';

export const TERMINAL_JOB_STATES: JobState[] = ['cancelled', 'succeeded', 'failed', 'timed_out'];

export function isTerminal(state: JobState): boolean {
  return TERMINAL_JOB_STATES.includes(state);
}

export interface JobSpec {
  image: string;
  command?: string[];
  args?: string[];
  env?: Record<string, string>;
  timeout_seconds?: number;
}

export interface Job {
  id: string;
  agent_id: string;
  state: JobState;
  spec: JobSpec;
  message: string;
  exit_code?: number;
  started_at?: string;
  finished_at?: string;
  created_at: string;
}

export interface JobLog {
  id: number;
  line: string;
  stream: 'stdout' | 'stderr';
  sequence: number;
  timestamp: string;
}
