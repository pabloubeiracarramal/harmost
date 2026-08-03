import { MetricsCard } from 'front';

// MetricsCard derives its series from useMetricsHistory(agent), which buffers
// samples per mount from moving `last_seen_at` values — there is no history
// endpoint on the hub. A static preview therefore holds exactly one sample, so
// the sparklines are intentionally sparse here and the footer's "Collecting
// samples…" hint shows. The headline percentages, capacity sublabels, tooltip
// trigger and container count are all live. See MetricAreaChart for the
// populated series. (Recorded in .design-sync/NOTES.md.)
const agent = (over: Record<string, unknown> = {}) => ({
  id: '9c1f2a7e-4b3d-4a91-8e52-1d7c6b0a3f84',
  name: 'build-runner-01',
  description: 'Primary CI runner',
  version: '0.4.1',
  hostname: 'ip-10-0-3-42.eu-west-1.compute.internal',
  status: 'online',
  // Anchored to render time, not a literal date: the header renders a relative
  // time ("42 seconds ago"), so a hardcoded ISO date reads as "in 2 years"
  // whenever the rendering clock disagrees with the authoring one.
  last_seen_at: new Date(Date.now() - 42_000).toISOString(),
  created_at: new Date(Date.now() - 90 * 24 * 60 * 60 * 1000).toISOString(),
  cpu_usage_percent: 34.6,
  memory_used_bytes: 4_724_464_025,
  memory_total_bytes: 16_642_998_272,
  disk_used_bytes: 118_111_600_640,
  disk_total_bytes: 494_384_795_648,
  running_containers: 3,
  ...over,
});

/** A healthy runner mid-build. */
export function Online() {
  return <MetricsCard agent={agent() as never} />;
}

/** Memory and disk near capacity — the figures are the warning, not a colour. */
export function UnderPressure() {
  return (
    <MetricsCard
      agent={
        agent({
          name: 'build-runner-04',
          cpu_usage_percent: 92.4,
          memory_used_bytes: 15_461_882_265,
          disk_used_bytes: 464_722_101_862,
          running_containers: 11,
        }) as never
      }
    />
  );
}

/** Idle agent: the footer stays, reading "0 running containers". */
export function Idle() {
  return (
    <MetricsCard
      agent={
        agent({
          name: 'laptop-pablo',
          hostname: 'pablo-thinkpad',
          cpu_usage_percent: 2.1,
          memory_used_bytes: 6_012_954_214,
          disk_used_bytes: 210_453_397_504,
          running_containers: 0,
        }) as never
      }
    />
  );
}
