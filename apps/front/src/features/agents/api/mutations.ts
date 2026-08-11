import { useMutation } from '@tanstack/react-query';
import { api } from '@/shared/api/httpClient';

export function useApproveDevice() {
  return useMutation({
    mutationFn: (userCode: string) =>
      api.post('/api/v1/device/approve', { user_code: userCode }),
  });
}

export function watchAgentContainers(agentId: string) {
  return api.post(`/api/v1/agents/${agentId}/containers/watch`);
}

export function unwatchAgentContainers(agentId: string) {
  return api.post(`/api/v1/agents/${agentId}/containers/unwatch`);
}
