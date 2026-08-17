import { LogOut } from 'lucide-react';
import { useNavigate } from '@tanstack/react-router';
import { useMe } from '@/features/auth';
import { clearToken } from '@/shared/api/auth';
import { Avatar, AvatarFallback, AvatarImage } from '@/shared/components/ui/avatar';
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuLabel,
    DropdownMenuSeparator,
    DropdownMenuTrigger,
} from '@/shared/components/ui/dropdown-menu';

export function SidebarFooter() {
    const navigate = useNavigate();
    const { data: user } = useMe();

    const handleLogout = () => {
        clearToken();
        navigate({ to: '/login' });
    };

    if (!user) return null;

    const displayName = user.name || user.email;
    const initial = displayName.charAt(0).toUpperCase();

    return (
        <div className="border-t px-4 py-4">
            <DropdownMenu>
                <DropdownMenuTrigger className="flex w-full items-center gap-2 rounded-md text-left focus:outline-hidden focus-visible:ring-2 focus-visible:ring-primary focus-visible:ring-offset-2">
                    <Avatar size="sm">
                        {user.avatar_url && <AvatarImage src={user.avatar_url} alt="" />}
                        <AvatarFallback>{initial}</AvatarFallback>
                    </Avatar>
                    <span className="truncate text-sm text-muted-foreground">{displayName}</span>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="start" className="w-56">
                    <DropdownMenuLabel className="truncate">{displayName}</DropdownMenuLabel>
                    <DropdownMenuSeparator />
                    <DropdownMenuItem variant="destructive" onClick={handleLogout}>
                        <LogOut className="h-4 w-4" /> Sign out
                    </DropdownMenuItem>
                </DropdownMenuContent>
            </DropdownMenu>
        </div>
    )

}
