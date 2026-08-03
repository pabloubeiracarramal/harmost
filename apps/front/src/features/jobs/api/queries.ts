import { useQuery } from '@tanstack/react-query';
import { api } from '@/shared/api/httpClient';
import { jobKeys } from './keys';
import type { Job, JobLog } from './types';

export function useJobs() {
  return useQuery<Job[]>({
    queryKey: jobKeys.list(),
    queryFn: () => api.get<Job[]>('/api/v1/jobs'),
  });
}

export function useJob(id: string) {
  return useQuery<Job>({
    queryKey: jobKeys.detail(id),
    queryFn: () => api.get<Job>(`/api/v1/jobs/${id}`),
  });
}

export function useJobLogs(id: string) {
  return useQuery<JobLog[]>({
    queryKey: jobKeys.logs(id),
    queryFn: () => api.get<JobLog[]>(`/api/v1/jobs/${id}/logs`),
  });
}
