import { useQueryClient } from '@tanstack/react-query';
import { useWsSubscribe } from '@/shared/ws/useWsSubscribe';
import type { HubEvent } from '@/shared/ws/wsClient';
import { agentKeys } from './keys';
import type { Agent, ContainerInfo, ContainerActionPayload } from './types';

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

/**
 * Forwards live agent.containers snapshots for a single agent. Push-only,
 * never fetched via REST, so — like job logs — it stays out of the query
 * cache; the caller owns its own buffer, this just feeds it.
 */
export function useAgentContainersSocket(id: string, onContainers: (containers: ContainerInfo[]) => void) {
  useWsSubscribe((e: HubEvent) => {
    if (e.type === 'agent.containers' && e.agent_id === id) onContainers(e.payload.containers);
  });
}

/** Forwards the outcome of a start/stop/restart/remove request for a single agent's containers. */
export function useAgentContainerActionsSocket(id: string, onResult: (result: ContainerActionPayload) => void) {
  useWsSubscribe((e: HubEvent) => {
    if (e.type === 'agent.container_action' && e.agent_id === id) onResult(e.payload);
  });
}
