import { getToken } from '@/shared/api/auth';

export interface AgentEvent {
  type: 'agent.connected' | 'agent.disconnected' | 'agent.heartbeat';
  agent_id: string;
  at: string;
}

// state is intentionally loose (not the feature-owned JobState union) — the
// WS wire layer must not depend on a feature's domain types.
export interface JobStatusEvent {
  type: 'job.status';
  agent_id: string;
  job_id: string;
  at: string;
  payload: {
    state: string;
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

type Listener = (e: HubEvent) => void;

/**
 * Owns the single WebSocket connection to the hub's /ws event stream.
 * Connects lazily on first subscriber, reconnects with exponential backoff
 * while at least one subscriber remains, and closes once the last one
 * unsubscribes.
 */
class WsClient {
  private listeners = new Set<Listener>();
  private ws: WebSocket | null = null;
  private timer: number | undefined;
  private attempt = 0;

  private connect() {
    if (this.ws) return;
    const token = getToken();
    if (!token) return;

    const proto = window.location.protocol === 'https:' ? 'wss' : 'ws';
    const ws = new WebSocket(
      `${proto}://${window.location.host}/ws?token=${encodeURIComponent(token)}`
    );
    this.ws = ws;

    ws.onopen = () => {
      this.attempt = 0;
    };
    ws.onmessage = (e) => {
      let event: HubEvent;
      try {
        event = JSON.parse(e.data) as HubEvent;
      } catch {
        return; // ignore malformed frames
      }
      for (const listener of this.listeners) listener(event);
    };
    ws.onerror = () => ws.close();
    ws.onclose = () => {
      this.ws = null;
      if (this.listeners.size === 0) return;
      const delay = Math.min(1000 * 2 ** this.attempt, MAX_BACKOFF_MS);
      this.attempt += 1;
      this.timer = window.setTimeout(() => this.connect(), delay);
    };
  }

  private disconnect() {
    window.clearTimeout(this.timer);
    this.attempt = 0;
    this.ws?.close();
    this.ws = null;
  }

  subscribe(listener: Listener): () => void {
    this.listeners.add(listener);
    this.connect();
    return () => {
      this.listeners.delete(listener);
      if (this.listeners.size === 0) this.disconnect();
    };
  }
}

export const wsClient = new WsClient();
