import { Link } from '@tanstack/react-router';
import { Bot, ListChecks, LayoutGrid } from 'lucide-react';

const NAV_ITEMS = [
    { to: '/projects', label: 'Projects', icon: LayoutGrid },
    { to: '/dashboard', label: 'Agents', icon: Bot },
    { to: '/jobs', label: 'Jobs', icon: ListChecks },
] as const;

export function SidebarNav() {

    return (
        <nav className="flex flex-1 flex-col gap-1 py-2 px-2">
            {NAV_ITEMS.map((item) => (
                <Link
                    key={item.to}
                    to={item.to}
                    className="flex items-center gap-2 rounded-md px-3 py-2 text-sm text-muted-foreground transition hover:bg-muted hover:text-foreground"
                    activeProps={{
                        className: 'flex items-center gap-2 rounded-md px-3 py-2 text-sm bg-muted text-foreground font-medium',
                    }}
                >
                    <item.icon className="size-4" />
                    {item.label}
                </Link>
            ))}
        </nav>
    )

}