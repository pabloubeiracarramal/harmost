import { useEffect, useRef } from 'react';
import { getToken } from '@/lib/auth';
import type { JobState } from '@/lib/api';

export interface AgentEvent {
  type: 'agent.connected' | 'agent.disconnected' | 'agent.heartbeat';
  agent_id: string;
  at: string;
}

export interface JobStatusEvent {
  type: 'job.status';
  agent_id: string;
  job_id: string;
  at: string;
  payload: {
    state: JobState;
    message?: string;
    exit_code?: number;
  };
}

export interface LogLine {
  line: string;
  stream: 'stdout' | 'stderr';
  sequence: number;
  timestamp: string;
}

export interface JobLogEvent {
  type: 'job.log';
  agent_id: string;
  job_id: string;
  at: string;
  payload: {
    lines: LogLine[];
  };
}

export type HubEvent = AgentEvent | JobStatusEvent | JobLogEvent;

const MAX_BACKOFF_MS = 30_000;

/**
 * Subscribes to the hub's /ws event stream. Reconnects automatically with
 * exponential backoff; the latest onEvent is always used, so callers don't
 * need to memoize it.
 */
export function useHubEvents(onEvent: (e: HubEvent) => void) {
  const handlerRef = useRef(onEvent);
  handlerRef.current = onEvent;

  useEffect(() => {
    let ws: WebSocket | null = null;
    let timer: number | undefined;
    let attempt = 0;
    let disposed = false;

    const connect = () => {
      const token = getToken();
      if (!token) return;

      const proto = window.location.protocol === 'https:' ? 'wss' : 'ws';
      ws = new WebSocket(
        `${proto}://${window.location.host}/ws?token=${encodeURIComponent(token)}`
      );

      ws.onopen = () => {
        attempt = 0;
      };
      ws.onmessage = (e) => {
        try {
          handlerRef.current(JSON.parse(e.data) as HubEvent);
        } catch {
          // ignore malformed frames
        }
      };
      ws.onerror = () => ws?.close();
      ws.onclose = () => {
        if (disposed) return;
        const delay = Math.min(1000 * 2 ** attempt, MAX_BACKOFF_MS);
        attempt += 1;
        timer = window.setTimeout(connect, delay);
      };
    };

    connect();

    return () => {
      disposed = true;
      window.clearTimeout(timer);
      ws?.close();
    };
  }, []);
}
