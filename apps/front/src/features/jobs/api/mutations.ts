import { useMutation, useQueryClient } from '@tanstack/react-query';
import { api } from '@/shared/api/httpClient';
import { jobKeys } from './keys';
import type { Job, JobSpec } from './types';

export function useDispatchJob() {
  return useMutation({
    mutationFn: (body: { agent_id: string; spec: JobSpec }) =>
      api.post<Job>('/api/v1/jobs', body),
  });
}

export function useCancelJob(id: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: () => api.post(`/api/v1/jobs/${id}/cancel`),
    onSettled: () => queryClient.invalidateQueries({ queryKey: jobKeys.detail(id) }),
  });
}
