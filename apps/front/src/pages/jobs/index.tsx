import { Link } from '@tanstack/react-router';
import { useAgents } from '@/features/agents';
import { useJobs, useJobsListSocket, JobStateBadge } from '@/features/jobs';
import { PageContainer } from '@/shared/components/layout/page-container/PageContainer';

export function JobsPage() {
  const { data: jobs = [], isLoading } = useJobs();
  const { data: agents = [] } = useAgents();
  useJobsListSocket();

  const agentName = (id: string) => {
    const a = agents.find((a) => a.id === id);
    return a ? (a.name !== 'pending' ? a.name : a.hostname) : id.slice(0, 8);
  };

  return (
    <PageContainer
      title="Jobs"
      actions={
        <Link
          to="/jobs/new"
          className="rounded-lg bg-indigo-600 px-4 py-2 text-sm font-medium hover:bg-indigo-500 transition"
        >
          New job
        </Link>
      }
    >
      {isLoading ? (
        <p className="text-neutral-500">Loading…</p>
      ) : jobs.length === 0 ? (
        <div className="rounded-xl border border-dashed border-neutral-700 bg-neutral-900/50 p-10 text-center">
          <p className="text-neutral-400">No jobs yet.</p>
          <p className="mt-1 text-sm text-neutral-600">
            Dispatch a container job to one of your agents to get started.
          </p>
        </div>
      ) : (
        <div className="overflow-hidden rounded-xl border border-neutral-800">
          <table className="w-full text-left text-sm">
            <thead className="bg-neutral-900 text-xs uppercase tracking-wider text-neutral-500">
              <tr>
                <th className="px-4 py-3 font-medium">Image</th>
                <th className="px-4 py-3 font-medium">Agent</th>
                <th className="px-4 py-3 font-medium">State</th>
                <th className="px-4 py-3 font-medium">Exit</th>
                <th className="px-4 py-3 font-medium">Created</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-neutral-800 bg-neutral-900/50">
              {jobs.map((job) => (
                <tr key={job.id} className="group relative hover:bg-neutral-800/50 transition">
                  <td className="px-4 py-3">
                    <Link
                      to="/jobs/$id"
                      params={{ id: job.id }}
                      className="font-mono text-white after:absolute after:inset-0"
                    >
                      {job.spec.image}
                    </Link>
                  </td>
                  <td className="px-4 py-3 text-neutral-400">{agentName(job.agent_id)}</td>
                  <td className="px-4 py-3">
                    <JobStateBadge state={job.state} />
                  </td>
                  <td className="px-4 py-3 font-mono text-neutral-400">
                    {job.exit_code ?? '—'}
                  </td>
                  <td className="px-4 py-3 text-neutral-500">
                    {new Date(job.created_at).toLocaleString()}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </PageContainer>
  );
}
