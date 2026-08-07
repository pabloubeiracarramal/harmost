import { AgentCard } from 'front';

const agent = (over: Record<string, unknown> = {}) => ({
  id: '9c1f2a7e-4b3d-4a91-8e52-1d7c6b0a3f84',
  name: 'build-runner-01',
  description: 'Primary CI runner',
  version: '0.4.1',
  hostname: 'ip-10-0-3-42.eu-west-1.compute.internal',
  status: 'online',
  // Fixed literals on purpose: AgentCard prints last_seen_at ABSOLUTELY via
  // toLocaleString(), so a Date.now()-relative value just renders whatever the
  // rendering machine's clock says. (MetricsCard is the opposite case — it
  // formats relatively, so it must be anchored to Date.now().)
  last_seen_at: '2026-07-28T09:14:03.000Z',
  created_at: '2026-05-02T11:00:00.000Z',
  ...over,
});

const grid: React.CSSProperties = {
  display: 'grid',
  gridTemplateColumns: 'repeat(auto-fill, minmax(260px, 1fr))',
  gap: 16,
  maxWidth: 860,
};

/** The dashboard grid: a live fleet, online and offline mixed. */
export function Fleet() {
  return (
    <div style={grid}>
      <AgentCard agent={agent() as never} />
      <AgentCard
        agent={
          agent({
            id: 'b2e4d6f8-1a3c-4e5f-9d7b-2c4a6e8f0b1d',
            name: 'build-runner-02',
            hostname: 'ip-10-0-3-77.eu-west-1.compute.internal',
          }) as never
        }
      />
      <AgentCard
        agent={
          agent({
            id: 'c3f5e7a9-2b4d-4f6a-8e0c-3d5b7f9a1c2e',
            name: 'laptop-pablo',
            hostname: 'pablo-thinkpad',
            status: 'offline',
            version: '0.3.9',
            last_seen_at: '2026-07-28T03:22:41.000Z',
          }) as never
        }
      />
    </div>
  );
}

/** Online: emerald pill with its status dot. */
export function Online() {
  return (
    <div style={{ maxWidth: 320 }}>
      <AgentCard agent={agent() as never} />
    </div>
  );
}

/** Offline: neutral pill, and the last-seen line carries the weight. */
export function Offline() {
  return (
    <div style={{ maxWidth: 320 }}>
      <AgentCard
        agent={
          agent({
            name: 'laptop-pablo',
            hostname: 'pablo-thinkpad',
            status: 'offline',
            last_seen_at: '2026-07-28T03:22:41.000Z',
          }) as never
        }
      />
    </div>
  );
}

/** A long hostname truncates rather than wrapping the tile. */
export function LongHostname() {
  return (
    <div style={{ maxWidth: 320 }}>
      <AgentCard
        agent={
          agent({
            name: 'eu-west-1-ephemeral-spot-runner-0042',
            hostname: 'ip-10-0-114-203.eu-west-1.compute.internal',
          }) as never
        }
      />
    </div>
  );
}
