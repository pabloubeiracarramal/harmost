import { DetailRow } from 'front';

const panel: React.CSSProperties = { maxWidth: 520 };

/** A job's detail panel, with the agent row linking through. */
export function JobDetails() {
  return (
    <div style={panel}>
      <DetailRow label="Job ID" value="7f3a9c21-8e4b-4d6f-a1c3-5b7d9e0f2a48" />
      <DetailRow label="Image" value="ghcr.io/acme/api-tests:1.4.2" />
      <DetailRow label="Agent" value="build-runner-01" link="/agents/9c1f2a7e" />
      <DetailRow label="Exit code" value="0" />
      <DetailRow label="Duration" value="24.31s" />
    </div>
  );
}

/** Plain values render as white monospace text. */
export function PlainValues() {
  return (
    <div style={panel}>
      <DetailRow label="Started" value="28 Jul 2026, 09:14:03" />
      <DetailRow label="Finished" value="28 Jul 2026, 09:14:27" />
    </div>
  );
}

/** With `link`, the value becomes an indigo anchor. */
export function Linked() {
  return (
    <div style={panel}>
      <DetailRow label="Agent" value="build-runner-01" link="/agents/9c1f2a7e" />
      <DetailRow label="Registry" value="ghcr.io/acme/api-tests" link="https://ghcr.io/acme/api-tests" />
    </div>
  );
}

/** Long values truncate so the row stays on one line. */
export function Truncation() {
  return (
    <div style={panel}>
      <DetailRow
        label="Digest"
        value="sha256:9f2c1e7b4a8d3c5e6f10b2a4d8e9c7f3b1a5d6e8c9f0a2b4d6e8f1c3a5b7d9e0"
      />
    </div>
  );
}
