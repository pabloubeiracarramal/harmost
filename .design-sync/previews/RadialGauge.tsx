import { RadialGauge } from 'front';

const row: React.CSSProperties = {
  display: 'flex',
  flexWrap: 'wrap',
  gap: 28,
  alignItems: 'flex-start',
};

/** The three resource gauges as an agent detail page shows them. */
export function ResourceTrio() {
  return (
    <div style={row}>
      <RadialGauge percent={34.6} label="CPU" sublabel="8 cores" />
      <RadialGauge percent={46.2} label="Memory" sublabel="4.4 GB / 15.5 GB" />
      <RadialGauge percent={23.9} label="Disk" sublabel="110.0 GB / 460.4 GB" />
    </div>
  );
}

/**
 * The colour scale, swept. Hue slides green → amber → red as usage climbs,
 * so LOW is good — this gauge is for resource pressure, not health.
 */
export function ColourScale() {
  return (
    <div style={row}>
      <RadialGauge percent={8} label="Idle" />
      <RadialGauge percent={35} label="Light" />
      <RadialGauge percent={62} label="Busy" />
      <RadialGauge percent={87} label="Heavy" />
      <RadialGauge percent={99.2} label="Saturated" />
    </div>
  );
}

/** Without a sublabel the caption sits directly under the ring. */
export function LabelOnly() {
  return (
    <div style={row}>
      <RadialGauge percent={54.8} label="CPU" />
    </div>
  );
}

/** Values outside 0-100 are clamped rather than overflowing the ring. */
export function Clamped() {
  return (
    <div style={row}>
      <RadialGauge percent={0} label="Floor" sublabel="0%" />
      <RadialGauge percent={100} label="Ceiling" sublabel="100%" />
    </div>
  );
}
