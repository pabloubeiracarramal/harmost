import { Link } from '@tanstack/react-router';

export function SidebarHeader() {

  return (
    <div className="flex h-14 shrink-0 items-center px-4 border-b">
      <Link to="/dashboard" className="text-lg font-semibold">
        Harmost
      </Link>
    </div>
  )

}