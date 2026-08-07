import { useQuery } from '@tanstack/react-query';
import { api } from '@/shared/api/httpClient';
import { agentKeys } from './keys';
import type { Agent } from './types';

export function useAgents() {
  return useQuery<Agent[]>({
    queryKey: agentKeys.list(),
    queryFn: () => api.get<Agent[]>('/api/v1/agents'),
  });
}

export function useAgent(id: string) {
  return useQuery<Agent>({
    queryKey: agentKeys.detail(id),
    queryFn: () => api.get<Agent>(`/api/v1/agents/${id}`),
    refetchInterval: 35_000,
  });
}
