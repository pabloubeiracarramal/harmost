import { useQuery } from '@tanstack/react-query';
import { api } from '@/shared/api/httpClient';
import { authKeys } from './keys';
import type { User } from './types';

export function useMe() {
  return useQuery<User>({
    queryKey: authKeys.me,
    queryFn: () => api.get<User>('/api/v1/me'),
    staleTime: 5 * 60 * 1000,
  });
}
