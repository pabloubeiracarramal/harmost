import { StatRow } from 'front';

const panel: React.CSSProperties = { maxWidth: 420 };

/** An agent's detail panel — rows stack directly, no divider of your own. */
export function AgentDetails() {
  return (
    <div style={panel}>
      <StatRow label="Hostname" value="ip-10-0-3-42.eu-west-1.compute.internal" />
      <StatRow label="Version" value="0.4.1" />
      <StatRow label="Status" value="online" />
      <StatRow label="Running containers" value="3" />
      <StatRow label="Last seen" value="42 seconds ago" />
    </div>
  );
}

/** The last row drops its border automatically. */
export function TwoRows() {
  return (
    <div style={panel}>
      <StatRow label="CPU" value="34.6%" />
      <StatRow label="Memory" value="4.4 GB / 15.5 GB" />
    </div>
  );
}

/** Values render monospace, so identifiers align down the column. */
export function MonospaceValues() {
  return (
    <div style={panel}>
      <StatRow label="Agent ID" value="9c1f2a7e-4b3d-4a91-8e52-1d7c6b0a3f84" />
      <StatRow label="Fingerprint" value="SHA256:9f2c1e7b4a8d3c5e" />
      <StatRow label="Token" value="hmt_a1b2c3d4e5f6" />
    </div>
  );
}
