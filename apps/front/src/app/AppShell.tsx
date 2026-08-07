import type { ReactNode } from 'react';
import { Link, useNavigate } from '@tanstack/react-router';
import { useMe } from '@/features/auth';
import { clearToken } from '@/shared/api/auth';

export function AppShell({ children }: { children: ReactNode }) {
  const navigate = useNavigate();

  const { data: me } = useMe();

  const handleLogout = () => {
    clearToken();
    navigate({ to: '/login' });
  };

  return (
    <div className="min-h-screen bg-neutral-950 text-white">
      <header className="border-b border-neutral-800 px-6 py-4">
        <div className="mx-auto flex max-w-5xl items-center justify-between">
          <div className="flex items-center gap-8">
            <Link to="/dashboard" className="text-lg font-semibold">
              Harmost
            </Link>
            <nav className="flex items-center gap-1">
              <NavLink to="/dashboard" label="Agents" />
              <NavLink to="/jobs" label="Jobs" />
              <NavLink to="/tokens" label="Tokens" />
            </nav>
          </div>
          <div className="flex items-center gap-4">
            {me && (
              <span className="flex items-center gap-2">
                {me.avatar_url && (
                  <img
                    src={me.avatar_url}
                    alt=""
                    className="h-6 w-6 rounded-full border border-neutral-700"
                  />
                )}
                <span className="hidden text-sm text-neutral-300 sm:inline">
                  {me.name || me.email}
                </span>
              </span>
            )}
            <button
              onClick={handleLogout}
              className="text-sm text-neutral-400 hover:text-white transition"
            >
              Sign out
            </button>
          </div>
        </div>
      </header>
      <main className="mx-auto max-w-5xl px-6 py-10">{children}</main>
    </div>
  );
}

function NavLink({ to, label }: { to: string; label: string }) {
  return (
    <Link
      to={to}
      className="rounded-md px-3 py-1.5 text-sm text-neutral-400 transition hover:text-white"
      activeProps={{ className: 'rounded-md px-3 py-1.5 text-sm bg-neutral-800 text-white' }}
    >
      {label}
    </Link>
  );
}
