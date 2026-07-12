import { createFileRoute, useNavigate } from '@tanstack/react-router';
import { useEffect } from 'react';
import { setToken } from '@/lib/auth';

export const Route = createFileRoute('/auth/callback')({
  component: AuthCallback,
});

function AuthCallback() {
  const navigate = useNavigate();

  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    const token = params.get('token');
    if (token) {
      setToken(token);
      navigate({ to: '/dashboard', replace: true });
    } else {
      navigate({ to: '/login', replace: true });
    }
  }, [navigate]);

  return (
    <div className="flex min-h-screen items-center justify-center bg-neutral-950">
      <p className="text-neutral-400">Signing in…</p>
    </div>
  );
}
