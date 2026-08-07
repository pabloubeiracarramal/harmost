---
category: Jobs
---
LogViewer — the scrolling container-log panel on a job detail page.

```tsx
<LogViewer lines={lines} live={!isTerminal(job.state)} />
```

`lines` is `LogLine[]` (`{ line, stream, sequence, timestamp }`); `live` drives the pulsing "live" indicator in the header. Sequence numbers render in a dim gutter and `stream === 'stderr'` lines render red.

Follow-on-append is built in: the panel stays pinned to the bottom as lines arrive, scrolling up pauses it, and a "↓ Jump to bottom" button appears until you scroll back down. You do not need to manage scroll position.

Capped at `max-h-[28rem]` and scrolls internally, so it will not push page layout around as output grows. Empty `lines` renders "No output yet." rather than collapsing.
