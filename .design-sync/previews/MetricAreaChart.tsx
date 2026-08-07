import { MetricAreaChart } from 'front';

// A populated 10-minute window at the 30s heartbeat (~20 points) — what the
// chart looks like once an agent has been up a while. MetricsCard's own card
// can only ever show one sample (the history hook buffers per mount), so this
// is the component's real shape.
const START = Date.now() - 10 * 60 * 1000;
const STEP = 30_000;

const series = (shape: (i: number) => [number, number, number]) =>
  Array.from({ length: 20 }, (_, i) => {
    const [cpu, memory, disk] = shape(i);
    return { t: START + i * STEP, cpu, memory, disk };
  });

const steady = series((i) => [
  28 + Math.sin(i / 2.2) * 9 + (i % 3) * 1.6,
  46 + Math.cos(i / 3.1) * 4,
  23.8 + i * 0.04,
]);

const spiking = series((i) => [
  i < 11 ? 22 + Math.sin(i / 2) * 6 : Math.min(97, 34 + (i - 10) * 9),
  i < 11 ? 44 + Math.cos(i / 3) * 3 : Math.min(94, 48 + (i - 10) * 6),
  61 + i * 0.05,
]);

const grid: React.CSSProperties = {
  display: 'grid',
  gridTemplateColumns: 'repeat(3, minmax(0, 1fr))',
  gap: 16,
};

/** The three series side by side, exactly as MetricsCard lays them out. */
export function AllThree() {
  return (
    <div style={grid}>
      <MetricAreaChart dataKey="cpu" label="CPU" value={34.6} samples={steady} />
      <MetricAreaChart
        dataKey="memory"
        label="Memory"
        value={46.2}
        sublabel="4.4 GB / 15.5 GB"
        samples={steady}
      />
      <MetricAreaChart
        dataKey="disk"
        label="Disk"
        value={23.9}
        sublabel="110.0 GB / 460.4 GB"
        samples={steady}
      />
    </div>
  );
}

/** One series on its own, full width. */
export function Single() {
  return <MetricAreaChart dataKey="cpu" label="CPU" value={34.6} samples={steady} />;
}

/** A load spike climbing into the top of the fixed 0-100 scale. */
export function Spiking() {
  return (
    <div style={grid}>
      <MetricAreaChart dataKey="cpu" label="CPU" value={92.4} samples={spiking} />
      <MetricAreaChart
        dataKey="memory"
        label="Memory"
        value={92.9}
        sublabel="14.4 GB / 15.5 GB"
        samples={spiking}
      />
      <MetricAreaChart
        dataKey="disk"
        label="Disk"
        value={62.0}
        sublabel="285.4 GB / 460.4 GB"
        samples={spiking}
      />
    </div>
  );
}

/** Freshly mounted: one sample, so the series has nothing to draw yet. */
export function CollectingSamples() {
  return (
    <MetricAreaChart
      dataKey="cpu"
      label="CPU"
      value={12.4}
      samples={[{ t: Date.now(), cpu: 12.4, memory: 40, disk: 24 }]}
    />
  );
}
