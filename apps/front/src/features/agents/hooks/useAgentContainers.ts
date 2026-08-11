import { useEffect, useState } from 'react';
import { getToken } from '@/shared/api/auth';
import { watchAgentContainers, unwatchAgentContainers } from '../api/mutations';
import { useAgentContainersSocket } from '../api/socket';
import type { ContainerInfo } from '../api/types';

/**
 * Watches an agent's running containers for as long as this hook stays
 * mounted: POSTs watch on mount / unwatch on unmount (or when `id` changes),
 * buffering the live `agent.containers` pushes in local state. Push-only, so
 * — like job logs and metrics history — it never touches the query cache.
 */
export function useAgentContainers(id: string): ContainerInfo[] {
  const [containers, setContainers] = useState<ContainerInfo[]>([]);

  useAgentContainersSocket(id, setContainers);

  useEffect(() => {
    setContainers([]);
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
    };
  }, [id]);

  return containers;
}
