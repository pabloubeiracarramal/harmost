import { useQueryClient } from '@tanstack/react-query';
import { useWsSubscribe } from '@/shared/ws/useWsSubscribe';
import type { HubEvent } from '@/shared/ws/wsClient';
import { agentKeys } from './keys';
import type { Agent } from './types';

/** Applies agent.connected/disconnected/heartbeat events to the agents list cache. */
export function useAgentsListSocket() {
  const queryClient = useQueryClient();

  useWsSubscribe((e: HubEvent) => {
    if (e.type === 'agent.connected') {
      queryClient.setQueryData<Agent[]>(agentKeys.list(), (prev = []) =>
        prev.map((a) =>
          a.id === e.agent_id ? { ...a, status: 'online', last_seen_at: e.at } : a
        )
      );
    }
    if (e.type === 'agent.disconnected') {
      queryClient.setQueryData<Agent[]>(agentKeys.list(), (prev = []) =>
        prev.map((a) => (a.id === e.agent_id ? { ...a, status: 'offline' } : a))
      );
    }
    if (e.type === 'agent.heartbeat') {
      queryClient.setQueryData<Agent[]>(agentKeys.list(), (prev = []) =>
        prev.map((a) => (a.id === e.agent_id ? { ...a, last_seen_at: e.at } : a))
      );
    }
  });
}

/** Invalidates a single agent's detail cache on any agent.* event concerning it. */
export function useAgentDetailSocket(id: string) {
  const queryClient = useQueryClient();

  useWsSubscribe((e: HubEvent) => {
    if (e.type.startsWith('agent.') && e.agent_id === id) {
      queryClient.invalidateQueries({ queryKey: agentKeys.detail(id) });
    }
  });
}
