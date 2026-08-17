import { useMutation, useQueryClient } from '@tanstack/react-query';
import { api } from '@/shared/api/httpClient';
import { agentKeys } from './keys';

export function useApproveDevice() {
  return useMutation({
    mutationFn: (userCode: string) =>
      api.post('/api/v1/device/approve', { user_code: userCode }),
  });
}

export function useDeleteAgent() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (agentId: string) => api.delete(`/api/v1/agents/${agentId}`),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: agentKeys.lists() }),
  });
}

export function watchAgentContainers(agentId: string) {
  return api.post(`/api/v1/agents/${agentId}/containers/watch`);
}

export function unwatchAgentContainers(agentId: string) {
  return api.post(`/api/v1/agents/${agentId}/containers/unwatch`);
}

export function startContainer(agentId: string, containerId: string) {
  return api.post(`/api/v1/agents/${agentId}/containers/${containerId}/start`);
}

export function stopContainer(agentId: string, containerId: string) {
  return api.post(`/api/v1/agents/${agentId}/containers/${containerId}/stop`);
}

export function restartContainer(agentId: string, containerId: string) {
  return api.post(
    `/api/v1/agents/${agentId}/containers/${containerId}/restart`
  );
}

export function removeContainer(agentId: string, containerId: string) {
  return api.post(`/api/v1/agents/${agentId}/containers/${containerId}/remove`);
}
