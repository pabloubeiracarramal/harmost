---
category: Jobs
---
JobStateBadge — a job's lifecycle state as a colour-coded pill.

```tsx
<JobStateBadge state={job.state} />
```

`state` is the `JobState` union: `accepted`, `pulling_image`, `creating_container`, `starting_container`, `running`, `stopping`, `cancelled`, `succeeded`, `failed`, `timed_out`. The underscores are rendered as spaces automatically — pass the raw state, never a prettified string.

In-flight states (`pulling_image` → `stopping`) additionally render a pulsing dot, so "still moving" is readable without parsing the word. Terminal states are colour-only: emerald for `succeeded`, red for `failed`/`timed_out`, neutral for `cancelled`.

The badge is `inline-flex` and `shrink-0` — safe to drop into a flex row beside truncating text without it being squeezed.

Also exports `formatJobState(state)` if you need the same label text outside the badge.
