import { LayoutGrid } from 'lucide-react';

export function ProjectsPage() {
  return (
    <>
      <p className="mb-6 text-sm text-muted-foreground">Group agents and jobs by project.</p>
      <div className="flex flex-col items-center justify-center rounded-xl border border-dashed border-neutral-700 bg-neutral-900/50 p-10 text-center">
        <LayoutGrid className="size-8 text-neutral-600" />
        <p className="mt-3 text-neutral-400">Projects are coming soon.</p>
        <p className="mt-1 text-sm text-neutral-600">
          This is where you'll organize agents and jobs by project.
        </p>
      </div>
    </>
  );
}
