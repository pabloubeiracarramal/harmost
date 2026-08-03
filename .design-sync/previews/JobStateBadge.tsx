import { JobStateBadge } from 'front';

// Preview scaffolding uses inline styles on purpose: the shipped CSS is
// Tailwind compiled against apps/front's own content scan, so a utility class
// that appears ONLY in a preview file has no rule in _ds_bundle.css.
const row: React.CSSProperties = {
  display: 'flex',
  flexWrap: 'wrap',
  gap: 10,
  alignItems: 'center',
};

/** Every state the hub can report, in lifecycle order. */
export function AllStates() {
  return (
    <div style={row}>
      <JobStateBadge state="accepted" />
      <JobStateBadge state="pulling_image" />
      <JobStateBadge state="creating_container" />
      <JobStateBadge state="starting_container" />
      <JobStateBadge state="running" />
      <JobStateBadge state="stopping" />
      <JobStateBadge state="succeeded" />
      <JobStateBadge state="failed" />
      <JobStateBadge state="timed_out" />
      <JobStateBadge state="cancelled" />
    </div>
  );
}

/** In-flight states carry a pulsing dot — the job is still moving. */
export function Active() {
  return (
    <div style={row}>
      <JobStateBadge state="pulling_image" />
      <JobStateBadge state="creating_container" />
      <JobStateBadge state="starting_container" />
      <JobStateBadge state="running" />
      <JobStateBadge state="stopping" />
    </div>
  );
}

/** Terminal states are dot-free and colour-coded by outcome. */
export function Terminal() {
  return (
    <div style={row}>
      <JobStateBadge state="succeeded" />
      <JobStateBadge state="failed" />
      <JobStateBadge state="timed_out" />
      <JobStateBadge state="cancelled" />
    </div>
  );
}

/** How it reads in a jobs-list row, beside the image it ran. */
export function InListRow() {
  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 12, maxWidth: 460 }}>
      {[
        { image: 'ghcr.io/acme/api-tests:1.4.2', state: 'running' as const },
        { image: 'ghcr.io/acme/web-build:2026.7', state: 'succeeded' as const },
        { image: 'docker.io/library/postgres:16', state: 'failed' as const },
      ].map((j) => (
        <div
          key={j.image}
          style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: 16 }}
        >
          <span style={{ fontFamily: 'ui-monospace, monospace', fontSize: 13, color: '#d4d4d4' }}>
            {j.image}
          </span>
          <JobStateBadge state={j.state} />
        </div>
      ))}
    </div>
  );
}
