import { useNavigate } from '@tanstack/react-router';
import { useState } from 'react';
import { useAgents } from '@/features/agents';
import { useDispatchJob, Field, Input, type JobSpec } from '@/features/jobs';
import { PageContainer } from '@/shared/components/layout/page-container/PageContainer';

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

export function NewJobPage() {
  const navigate = useNavigate();

  const { data: agents = [] } = useAgents();
  const onlineAgents = agents.filter((a) => a.status === 'online');

  const [agentId, setAgentId] = useState('');
  const [image, setImage] = useState('');
  const [command, setCommand] = useState('');
  const [args, setArgs] = useState('');
  const [envText, setEnvText] = useState('');
  const [timeout, setTimeoutSecs] = useState('');
  const [formError, setFormError] = useState('');

  const dispatch = useDispatchJob();

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

    const spec: JobSpec = {
      image: image.trim(),
      command: splitWords(command),
      args: splitWords(args),
      env,
      timeout_seconds: timeoutSeconds,
    };

    dispatch.mutate(
      { agent_id: agentId, spec },
      { onSuccess: (job) => navigate({ to: '/jobs/$id', params: { id: job.id } }) }
    );
  };

  const error = formError || (dispatch.error as Error | null)?.message;

  return (
    <PageContainer>
      <div className="mx-auto max-w-xl">
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
    </PageContainer>
  );
}
