import { EmptyState } from 'front';

/** The dashboard before any agent has paired. */
export function NoAgents() {
  return (
    <div style={{ maxWidth: 640 }}>
      <EmptyState />
    </div>
  );
}

/** At full page width, as the dashboard actually renders it. */
export function FullWidth() {
  return <EmptyState />;
}
