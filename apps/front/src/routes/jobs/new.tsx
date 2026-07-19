import { createFileRoute, redirect, Link, useNavigate } from '@tanstack/react-router';
import { useQuery, useMutation } from '@tanstack/react-query';
import { useState } from 'react';
import { isAuthenticated } from '@/lib/auth';
import { api, type Agent, type Job, type JobSpec } from '@/lib/api';
import { AppShell } from '@/components/AppShell';

export const Route = createFileRoute('/jobs/new')({
  beforeLoad: () => {
    if (!isAuthenticated()) throw redirect({ to: '/login' });
  },
  component: NewJobPage,
});

/** Splits a shell-ish input on whitespace; empty input → undefined. */
function splitWords(s: string): string[] | undefined {
  const words = s.trim().split(/\s+/).filter(Boolean);
  return words.length ? words : undefined;
}

/** Parses KEY=VALUE lines; ignores blank lines. Returns an error message for malformed lines. */
function parseEnv(s: string): { env?: Record<string, string>; error?: string } {
  const env: Record<string, string> = {};
  for (const rawLine of s.split('\n')) {
    const line = rawLine.trim();
    if (!line) continue;
    const eq = line.indexOf('=');
    if (eq <= 0) return { error: `Invalid env line: "${line}" (expected KEY=VALUE)` };
    env[line.slice(0, eq)] = line.slice(eq + 1);
  }
  return { env: Object.keys(env).length ? env : undefined };
}

function NewJobPage() {
  const navigate = useNavigate();

  const { data: agents = [] } = useQuery<Agent[]>({
    queryKey: ['agents'],
    queryFn: () => api.get<Agent[]>('/api/v1/agents'),
  });
  const onlineAgents = agents.filter((a) => a.status === 'online');

  const [agentId, setAgentId] = useState('');
  const [image, setImage] = useState('');
  const [command, setCommand] = useState('');
  const [args, setArgs] = useState('');
  const [envText, setEnvText] = useState('');
  const [timeout, setTimeoutSecs] = useState('');
  const [formError, setFormError] = useState('');

  const dispatch = useMutation({
    mutationFn: (body: { agent_id: string; spec: JobSpec }) =>
      api.post<Job>('/api/v1/jobs', body),
    onSuccess: (job) => navigate({ to: '/jobs/$id', params: { id: job.id } }),
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    setFormError('');

    const { env, error } = parseEnv(envText);
    if (error) {
      setFormError(error);
      return;
    }
    const timeoutSeconds = timeout ? Number(timeout) : undefined;
    if (timeoutSeconds !== undefined && (!Number.isInteger(timeoutSeconds) || timeoutSeconds <= 0)) {
      setFormError('Timeout must be a positive whole number of seconds');
      return;
    }

    dispatch.mutate({
      agent_id: agentId,
      spec: {
        image: image.trim(),
        command: splitWords(command),
        args: splitWords(args),
        env,
        timeout_seconds: timeoutSeconds,
      },
    });
  };

  const error = formError || (dispatch.error as Error | null)?.message;

  return (
    <AppShell>
      <div className="mx-auto max-w-xl">
        <div className="mb-6 flex items-center gap-4">
          <Link to="/jobs" className="text-sm text-neutral-400 hover:text-white transition">
            ← Jobs
          </Link>
          <span className="text-neutral-700">/</span>
          <h2 className="text-sm font-medium">New job</h2>
        </div>

        <form onSubmit={handleSubmit} className="space-y-5 rounded-xl border border-neutral-800 bg-neutral-900 p-6">
          <Field label="Agent" required>
            <select
              value={agentId}
              onChange={(e) => setAgentId(e.target.value)}
              required
              className="w-full rounded-lg border border-neutral-700 bg-neutral-800 px-3 py-2 text-sm text-white focus:border-indigo-500 focus:outline-none"
            >
              <option value="" disabled>
                {onlineAgents.length ? 'Select an online agent…' : 'No agents online'}
              </option>
              {onlineAgents.map((a) => (
                <option key={a.id} value={a.id}>
                  {a.name !== 'pending' ? a.name : a.hostname}
                </option>
              ))}
            </select>
          </Field>

          <Field label="Image" required hint="e.g. alpine:3.20">
            <Input value={image} onChange={setImage} placeholder="alpine:3.20" required />
          </Field>

          <Field label="Command" hint="Overrides the image entrypoint; split on spaces">
            <Input value={command} onChange={setCommand} placeholder="sh -c" mono />
          </Field>

          <Field label="Args" hint="Split on spaces">
            <Input value={args} onChange={setArgs} placeholder='"echo hello"' mono />
          </Field>

          <Field label="Environment" hint="One KEY=VALUE per line">
            <textarea
              value={envText}
              onChange={(e) => setEnvText(e.target.value)}
              rows={3}
              placeholder={'FOO=bar\nDEBUG=1'}
              className="w-full rounded-lg border border-neutral-700 bg-neutral-800 px-3 py-2 font-mono text-sm text-white placeholder:text-neutral-600 focus:border-indigo-500 focus:outline-none"
            />
          </Field>

          <Field label="Timeout (seconds)" hint="Job is killed and marked timed_out after this">
            <Input value={timeout} onChange={setTimeoutSecs} placeholder="300" type="number" />
          </Field>

          {error && <p className="text-sm text-red-400">{error}</p>}

          <button
            type="submit"
            disabled={dispatch.isPending || !agentId || !image.trim()}
            className="w-full rounded-lg bg-indigo-600 px-4 py-2.5 text-sm font-medium text-white transition hover:bg-indigo-500 disabled:cursor-not-allowed disabled:opacity-50"
          >
            {dispatch.isPending ? 'Dispatching…' : 'Dispatch job'}
          </button>
        </form>
      </div>
    </AppShell>
  );
}

function Field({
  label,
  hint,
  required,
  children,
}: {
  label: string;
  hint?: string;
  required?: boolean;
  children: React.ReactNode;
}) {
  return (
    <label className="block">
      <span className="mb-1.5 block text-sm font-medium text-neutral-300">
        {label}
        {required && <span className="text-red-400"> *</span>}
      </span>
      {children}
      {hint && <span className="mt-1 block text-xs text-neutral-600">{hint}</span>}
    </label>
  );
}

function Input({
  value,
  onChange,
  placeholder,
  required,
  mono,
  type = 'text',
}: {
  value: string;
  onChange: (v: string) => void;
  placeholder?: string;
  required?: boolean;
  mono?: boolean;
  type?: string;
}) {
  return (
    <input
      type={type}
      value={value}
      onChange={(e) => onChange(e.target.value)}
      placeholder={placeholder}
      required={required}
      className={`w-full rounded-lg border border-neutral-700 bg-neutral-800 px-3 py-2 text-sm text-white placeholder:text-neutral-600 focus:border-indigo-500 focus:outline-none ${mono ? 'font-mono' : ''}`}
    />
  );
}
