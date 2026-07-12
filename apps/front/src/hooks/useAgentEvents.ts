import { useEffect, useCallback } from 'react';
import { getToken } from '@/lib/auth';

export type AgentEventType = 'agent.connected' | 'agent.disconnected' | 'agent.heartbeat';

export interface AgentEvent {
  type: AgentEventType;
  agent_id: string;
  at: string;
}

export function useAgentEvents(onEvent: (e: AgentEvent) => void) {
  const stableOnEvent = useCallback(onEvent, []);

  useEffect(() => {
    const token = getToken();
    if (!token) return;

    const ws = new WebSocket(`ws://${window.location.host}/ws?token=${encodeURIComponent(token)}`);

    ws.onmessage = (e) => {
      try {
        const event: AgentEvent = JSON.parse(e.data);
        stableOnEvent(event);
      } catch {
        // ignore malformed frames
      }
    };

    ws.onerror = () => ws.close();

    return () => {
      ws.close();
    };
  }, [stableOnEvent]);
}
