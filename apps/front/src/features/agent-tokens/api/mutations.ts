import { useMutation, useQueryClient } from '@tanstack/react-query';
import { api } from '@/shared/api/httpClient';
import { agentTokenKeys } from './keys';

export function useRevokeToken() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.post(`/api/v1/agent-tokens/${id}/revoke`),
    onSettled: () => queryClient.invalidateQueries({ queryKey: agentTokenKeys.lists() }),
  });
}
