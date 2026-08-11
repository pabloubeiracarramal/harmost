import { getToken, clearToken } from './auth';
import { ApiError } from '@/shared/types/api-error';

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const token = getToken();
  const res = await fetch(path, {
    ...init,
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...init?.headers,
    },
  });
  if (res.status === 401) {
    // Token expired or revoked — drop it and send the user back to login.
    clearToken();
    window.location.href = '/login';
    throw new ApiError('session expired', 401);
  }
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new ApiError(body.error ?? `HTTP ${res.status}`, res.status);
  }
  if (res.status === 204) return undefined as T;
  const text = await res.text();
  return text ? JSON.parse(text) : (undefined as T);
}

export const api = {
  get: <T>(path: string) => request<T>(path),
  post: <T>(path: string, body?: unknown) =>
    request<T>(path, { method: 'POST', body: body ? JSON.stringify(body) : undefined }),
};
