import { Link } from '@tanstack/react-router';

export interface SidebarUser {
  name?: string | null;
  email?: string | null;
  avatar_url?: string | null;
}

interface SidebarContainerProps {
  user?: SidebarUser | null;
  onLogout: () => void;
}

const NAV_ITEMS = [
  { to: '/dashboard', label: 'Agents' },
  { to: '/jobs', label: 'Jobs' },
  { to: '/tokens', label: 'Tokens' },
] as const;

export function SidebarContainer({ user, onLogout }: SidebarContainerProps) {
  return (
    <aside className="flex h-full w-64 shrink-0 flex-col border-r bg-background">
      <div className="px-4 py-4">
        <Link to="/dashboard" className="text-lg font-semibold">
          Harmost
        </Link>
      </div>

      <nav className="flex flex-1 flex-col gap-1 px-2">
        {NAV_ITEMS.map((item) => (
          <Link
            key={item.to}
            to={item.to}
            className="rounded-md px-3 py-2 text-sm text-muted-foreground transition hover:bg-muted hover:text-foreground"
            activeProps={{
              className: 'rounded-md px-3 py-2 text-sm bg-muted text-foreground font-medium',
            }}
          >
            {item.label}
          </Link>
        ))}
      </nav>

      <div className="border-t px-4 py-4">
        {user && (
          <div className="mb-3 flex items-center gap-2">
            {user.avatar_url && (
              <img
                src={user.avatar_url}
                alt=""
                className="h-6 w-6 rounded-full border border-border"
              />
            )}
            <span className="truncate text-sm text-muted-foreground">
              {user.name || user.email}
            </span>
          </div>
        )}
        <button
          onClick={onLogout}
          className="text-sm text-muted-foreground transition hover:text-foreground"
        >
          Sign out
        </button>
      </div>
    </aside>
  );
}
