import { createFileRoute, redirect, useNavigate } from '@tanstack/react-router';
import { useState } from 'react';
import { isAuthenticated } from '@/lib/auth';
import { api } from '@/lib/api';

export const Route = createFileRoute('/device')({
  beforeLoad: () => {
    if (!isAuthenticated()) throw redirect({ to: '/login' });
  },
  component: DevicePage,
});

function DevicePage() {
  const navigate = useNavigate();
  const params = new URLSearchParams(
    typeof window !== 'undefined' ? window.location.search : ''
  );
  const userCode = params.get('code') ?? '';

  const [status, setStatus] = useState<'idle' | 'loading' | 'done' | 'error'>('idle');
  const [error, setError] = useState('');

  const handleApprove = async () => {
    setStatus('loading');
    try {
      await api.post('/api/v1/device/approve', { user_code: userCode });
      setStatus('done');
    } catch (e) {
      setError((e as Error).message);
      setStatus('error');
    }
  };

  return (
    <div className="flex min-h-screen items-center justify-center bg-neutral-950">
      <div className="w-full max-w-sm space-y-6 rounded-xl border border-neutral-800 bg-neutral-900 p-8">
        <div className="space-y-1">
          <h1 className="text-xl font-bold text-white">Approve Agent</h1>
          <p className="text-sm text-neutral-400">
            An agent is requesting access to your account.
          </p>
        </div>

        {userCode && (
          <div className="rounded-lg bg-neutral-800 p-4 text-center">
            <p className="text-xs uppercase tracking-widest text-neutral-500 mb-1">Pairing code</p>
            <p className="font-mono text-2xl font-bold tracking-widest text-white">{userCode}</p>
          </div>
        )}

        {status === 'done' ? (
          <div className="space-y-3">
            <p className="text-sm text-emerald-400">Agent approved! You can close this tab.</p>
            <button
              onClick={() => navigate({ to: '/dashboard' })}
              className="w-full rounded-lg bg-neutral-800 px-4 py-2 text-sm text-white hover:bg-neutral-700 transition"
            >
              Back to dashboard
            </button>
          </div>
        ) : (
          <div className="space-y-3">
            {status === 'error' && (
              <p className="text-sm text-red-400">{error}</p>
            )}
            <button
              onClick={handleApprove}
              disabled={status === 'loading' || !userCode}
              className="w-full rounded-lg bg-indigo-600 px-4 py-2.5 text-sm font-medium text-white transition hover:bg-indigo-500 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {status === 'loading' ? 'Approving…' : 'Approve Agent'}
            </button>
            <button
              onClick={() => navigate({ to: '/dashboard' })}
              className="w-full rounded-lg bg-transparent px-4 py-2 text-sm text-neutral-400 hover:text-white transition"
            >
              Cancel
            </button>
          </div>
        )}
      </div>
    </div>
  );
}
