import { getToken, clearToken } from './auth';

export class ApiError extends Error {
  constructor(
    message: string,
    public status: number
  ) {
    super(message);
  }
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const token = getToken();
  const res = await fetch(path, {
    ...init,
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...init?.headers,
    },
  });
  if (res.status === 401) {
    // Token expired or revoked — drop it and send the user back to login.
    clearToken();
    window.location.href = '/login';
    throw new ApiError('session expired', 401);
  }
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new ApiError(body.error ?? `HTTP ${res.status}`, res.status);
  }
  if (res.status === 204) return undefined as T;
  return res.json();
}

export const api = {
  get: <T>(path: string) => request<T>(path),
  post: <T>(path: string, body?: unknown) =>
    request<T>(path, { method: 'POST', body: body ? JSON.stringify(body) : undefined }),
};

export interface AgentToken {
  id: string;
  name: string;
  agent_id?: string;
  created_at: string;
  last_used_at?: string;
}

export interface Agent {
  id: string;
  name: string;
  description: string;
  version: string;
  hostname: string;
  status: 'online' | 'offline';
  last_seen_at?: string;
  created_at: string;
  cpu_usage_percent?: number;
  memory_used_bytes?: number;
  memory_total_bytes?: number;
  disk_used_bytes?: number;
  disk_total_bytes?: number;
  running_containers?: number;
}

export interface User {
  id: string;
  email: string;
  name: string;
  avatar_url: string;
  org_id: string;
}

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
