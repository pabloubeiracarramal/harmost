import { LogViewer } from 'front';

const line = (sequence: number, text: string, stream: 'stdout' | 'stderr' = 'stdout') => ({
  sequence,
  line: text,
  stream,
  timestamp: new Date(Date.UTC(2026, 6, 28, 9, 14, sequence)).toISOString(),
});

const buildLog = [
  line(1, '+ docker run --rm ghcr.io/acme/api-tests:1.4.2 pytest -q'),
  line(2, 'Unable to find image locally, pulling…'),
  line(3, '1.4.2: Pulling from acme/api-tests'),
  line(4, 'a1b2c3d4e5f6: Pull complete'),
  line(5, 'Digest: sha256:9f2c1e7b4a8d3c5e6f10b2a4d8e9c7f3b1a5d6e8c9f0a2b4d6e8f1c3a5b7d9e0'),
  line(6, ''),
  line(7, '==================== test session starts ===================='),
  line(8, 'platform linux — Python 3.12.4, pytest-8.2.1'),
  line(9, 'collected 128 items'),
  line(10, ''),
  line(11, 'tests/test_auth.py ..........................          [ 20%]'),
  line(12, 'tests/test_jobs.py ..................................  [ 47%]'),
  line(13, 'tests/test_agents.py ..............................    [ 71%]'),
  line(14, 'tests/test_hub.py ....................................  [ 98%]'),
  line(15, 'tests/test_ws.py .                                     [100%]'),
  line(16, ''),
  line(17, '===================== 128 passed in 24.31s ====================='),
];

const failingLog = [
  line(1, '+ docker run --rm ghcr.io/acme/web-build:2026.7 npm run build'),
  line(2, '> web@2026.7 build'),
  line(3, '> vite build'),
  line(4, ''),
  line(5, 'vite v7.0.2 building for production...'),
  line(6, 'transforming...'),
  line(7, 'src/routes/dashboard.tsx:14:22 — error TS2551', 'stderr'),
  line(8, "  Property 'agentz' does not exist on type 'DashboardData'.", 'stderr'),
  line(9, '  Did you mean ‘agents’?', 'stderr'),
  line(10, '', 'stderr'),
  line(11, 'error during build:', 'stderr'),
  line(12, 'Error: Build failed with 1 error.', 'stderr'),
  line(13, 'npm ERR! Lifecycle script `build` failed with error code 1', 'stderr'),
];

/** A job still streaming — the header shows the live indicator. */
export function Streaming() {
  return <LogViewer lines={buildLog} live />;
}

/** A finished job: same output, no live dot. */
export function Completed() {
  return <LogViewer lines={buildLog} live={false} />;
}

/** stderr lines render red so a failure is scannable without reading it. */
export function WithStderr() {
  return <LogViewer lines={failingLog} live={false} />;
}

/** Before the first line arrives. */
export function Empty() {
  return <LogViewer lines={[]} live />;
}
