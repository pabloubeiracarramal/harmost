import { useEffect, useRef, useState } from 'react';
import type { Agent } from '../api/types';

export interface MetricSample {
  /** Epoch ms of the heartbeat this sample came from. */
  t: number;
  cpu: number;
  memory: number;
  disk: number;
}

/** Rolling window: samples older than this are dropped (~20 points at the 30s heartbeat). */
export const METRICS_WINDOW_MS = 10 * 60 * 1000;

function percent(used: number | undefined, total: number | undefined): number {
  return total ? ((used ?? 0) / total) * 100 : 0;
}

/**
 * Buffers agent metric snapshots into a time series for the area charts.
 *
 * The hub persists only the *latest* snapshot per agent (see
 * `AgentRepo.UpdateMetrics`), so there is no history endpoint to query. We
 * accumulate in local component state instead — the same pattern job logs use:
 * an append-only live stream stays out of the query cache. History is therefore
 * per-mount and starts empty; a new sample lands whenever `last_seen_at` moves
 * (heartbeat WS invalidation or the 35s detail refetch).
 *
 * The series is a rolling `windowMs` window. Pruning is driven by appends, not a
 * timer, so an idle agent keeps its last points on screen rather than watching
 * the chart empty itself out — the trailing edge can lag the window by up to one
 * heartbeat.
 */
export function useMetricsHistory(
  agent: Agent | undefined,
  windowMs: number = METRICS_WINDOW_MS
): MetricSample[] {
  const [samples, setSamples] = useState<MetricSample[]>([]);
  const lastStamp = useRef<string | null>(null);
  const agentId = agent?.id;

  // Switching agents must not carry the previous agent's series over.
  useEffect(() => {
    lastStamp.current = null;
    setSamples([]);
  }, [agentId]);

  useEffect(() => {
    if (!agent || agent.status !== 'online' || !agent.last_seen_at) return;
    // The agent object gets a new identity on every refetch; only a moved
    // last_seen_at means the agent actually reported again.
    if (agent.last_seen_at === lastStamp.current) return;
    lastStamp.current = agent.last_seen_at;

    const sample: MetricSample = {
      t: new Date(agent.last_seen_at).getTime(),
      cpu: agent.cpu_usage_percent ?? 0,
      memory: percent(agent.memory_used_bytes, agent.memory_total_bytes),
      disk: percent(agent.disk_used_bytes, agent.disk_total_bytes),
    };

    // Cut off relative to the newest sample, not Date.now(), so a clock skew
    // between hub and browser can't discard everything the agent just reported.
    const cutoff = sample.t - windowMs;
    setSamples((prev) => [...prev, sample].filter((s) => s.t >= cutoff));
  }, [agent, windowMs]);

  return samples;
}
