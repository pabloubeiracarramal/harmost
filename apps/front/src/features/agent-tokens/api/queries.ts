import { useQuery } from '@tanstack/react-query';
import { api } from '@/shared/api/httpClient';
import { agentTokenKeys } from './keys';
import type { AgentToken } from './types';

export function useAgentTokens() {
  return useQuery<AgentToken[]>({
    queryKey: agentTokenKeys.list(),
    queryFn: () => api.get<AgentToken[]>('/api/v1/agent-tokens'),
  });
}
