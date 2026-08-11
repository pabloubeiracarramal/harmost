import type { components } from '@/shared/api/schema';

export type JobState = components['schemas']['JobState'];
export type JobSpec = components['schemas']['JobSpec'];
export type Job = components['schemas']['Job'];
export type JobLog = components['schemas']['JobLog'];

export const TERMINAL_JOB_STATES: JobState[] = ['cancelled', 'succeeded', 'failed', 'timed_out'];

export function isTerminal(state: JobState): boolean {
  return TERMINAL_JOB_STATES.includes(state);
}
