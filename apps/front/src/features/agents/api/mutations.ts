import { useMutation } from '@tanstack/react-query';
import { api } from '@/shared/api/httpClient';

export function useApproveDevice() {
  return useMutation({
    mutationFn: (userCode: string) =>
      api.post('/api/v1/device/approve', { user_code: userCode }),
  });
}
