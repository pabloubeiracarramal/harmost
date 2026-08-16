import { useEffect, useRef, useState } from 'react';
import { getToken } from '@/shared/api/auth';
import {
  watchAgentContainers,
  unwatchAgentContainers,
  startContainer,
  stopContainer,
  restartContainer,
  removeContainer,
} from '../api/mutations';
import { useAgentContainersSocket, useAgentContainerActionsSocket } from '../api/socket';
import type { ContainerInfo, ContainerActionKind } from '../api/types';

const ACTION_CALLS: Record<ContainerActionKind, (agentId: string, containerId: string) => Promise<unknown>> = {
  start: startContainer,
  stop: stopContainer,
  restart: restartContainer,
  remove: removeContainer,
};

// The agent should always answer well before this — it's a fallback for a
// disconnect mid-action, not the expected path.
const ACTION_TIMEOUT_MS = 15_000;

export interface PendingAction {
  action: ContainerActionKind;
  error?: string;
}

export interface UseAgentContainers {
  containers: ContainerInfo[];
  /** In-flight (or failed) actions, keyed by container id. */
  pending: Record<string, PendingAction>;
  performAction: (containerId: string, action: ContainerActionKind) => void;
}

/**
 * Watches an agent's containers (as before) and additionally exposes
 * start/stop/restart/remove. `pending[containerId]` is set optimistically by
 * `performAction` and cleared once the agent's `agent.container_action`
 * result arrives — or, if it never does (e.g. the agent disconnects
 * mid-action), replaced with a timeout error after ACTION_TIMEOUT_MS. A
 * failed action's error sticks around until the next action attempt on that
 * container supersedes it.
 */
export function useAgentContainers(id: string): UseAgentContainers {
  const [containers, setContainers] = useState<ContainerInfo[]>([]);
  const [pending, setPending] = useState<Record<string, PendingAction>>({});
  const timers = useRef<Record<string, number>>({});

  const clearTimer = (containerId: string) => {
    const timer = timers.current[containerId];
    if (timer !== undefined) {
      window.clearTimeout(timer);
      delete timers.current[containerId];
    }
  };

  useAgentContainersSocket(id, setContainers);

  useAgentContainerActionsSocket(id, (result) => {
    clearTimer(result.container_id);
    setPending((prev) => {
      if (result.success) {
        const next = { ...prev };
        delete next[result.container_id];
        return next;
      }
      return {
        ...prev,
        [result.container_id]: { action: result.action, error: result.error || 'Action failed' },
      };
    });
  });

  const performAction = (containerId: string, action: ContainerActionKind) => {
    clearTimer(containerId);
    setPending((prev) => ({ ...prev, [containerId]: { action } }));

    timers.current[containerId] = window.setTimeout(() => {
      setPending((prev) =>
        prev[containerId]?.action === action && !prev[containerId].error
          ? { ...prev, [containerId]: { action, error: 'Timed out waiting for the agent' } }
          : prev
      );
    }, ACTION_TIMEOUT_MS);

    ACTION_CALLS[action](id, containerId).catch((err: unknown) => {
      clearTimer(containerId);
      setPending((prev) => ({
        ...prev,
        [containerId]: { action, error: err instanceof Error ? err.message : 'Action failed' },
      }));
    });
  };

  useEffect(() => {
    setContainers([]);
    setPending({});
    watchAgentContainers(id).catch(() => {});

    // Unmount cleanup below isn't guaranteed to run on a tab close, which
    // would otherwise leave the agent polling Docker forever. sendBeacon
    // can't carry the Authorization header this endpoint requires, so a
    // keepalive fetch is the closest best-effort equivalent.
    const token = getToken();
    const releaseOnUnload = () => {
      if (!token) return;
      fetch(`/api/v1/agents/${id}/containers/unwatch`, {
        method: 'POST',
        keepalive: true,
        headers: { Authorization: `Bearer ${token}` },
      }).catch(() => {});
    };
    window.addEventListener('pagehide', releaseOnUnload);

    return () => {
      window.removeEventListener('pagehide', releaseOnUnload);
      unwatchAgentContainers(id).catch(() => {});
      Object.values(timers.current).forEach(window.clearTimeout);
      timers.current = {};
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [id]);

  return { containers, pending, performAction };
}
