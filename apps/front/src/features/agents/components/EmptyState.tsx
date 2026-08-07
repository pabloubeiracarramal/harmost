export function EmptyState() {
  return (
    <div className="rounded-xl border border-dashed border-neutral-700 bg-neutral-900/50 p-10 text-center">
      <p className="text-neutral-400">No agents yet.</p>
      <p className="mt-1 text-sm text-neutral-600">
        Run <code className="rounded bg-neutral-800 px-1 py-0.5 font-mono text-xs">harmost pair &lt;hub-url&gt;</code> on a machine to get started.
      </p>
    </div>
  );
}
