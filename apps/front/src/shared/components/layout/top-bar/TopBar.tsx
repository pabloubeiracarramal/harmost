import { Link, useRouterState } from '@tanstack/react-router';
import { ChevronRight } from 'lucide-react';

const SECTIONS: Record<string, { label: string; to: string }> = {
  dashboard: { label: 'Agents', to: '/dashboard' },
  agents: { label: 'Agents', to: '/dashboard' },
  jobs: { label: 'Jobs', to: '/jobs' },
  tokens: { label: 'Tokens', to: '/tokens' },
  projects: { label: 'Projects', to: '/projects' },
};

function getBreadcrumbs(pathname: string): { label: string; to?: string }[] {
  const [first, second] = pathname.split('/').filter(Boolean);
  const section = SECTIONS[first];
  if (!section) return [];

  const crumbs: { label: string; to?: string }[] = [section];
  if (second === 'new') {
    crumbs.push({ label: 'New job' });
  } else if (second) {
    crumbs.push({ label: second.slice(0, 8) });
  }
  return crumbs;
}

export function TopBar() {
  const pathname = useRouterState({ select: (s) => s.location.pathname });
  const crumbs = getBreadcrumbs(pathname);

  return (
    <header className="flex h-14 shrink-0 items-center gap-1.5 border-b px-6 text-sm">
      {crumbs.map((crumb, i) => {
        const isLast = i === crumbs.length - 1;
        return (
          <span key={i} className="flex items-center gap-1.5">
            {i > 0 && <ChevronRight className="size-3.5 text-muted-foreground" />}
            {crumb.to && !isLast ? (
              <Link to={crumb.to} className="text-muted-foreground transition hover:text-foreground">
                {crumb.label}
              </Link>
            ) : (
              <span className={isLast ? 'font-semibold text-foreground' : 'text-muted-foreground'}>
                {crumb.label}
              </span>
            )}
          </span>
        );
      })}
    </header>
  );
}
