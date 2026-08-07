// Design-system entry for design-sync (see /docs or .design-sync/NOTES.md).
//
// apps/front is an application, not a published component library, so it has
// no dist/ entry the converter can bundle. This barrel is that entry: it
// re-exports the components the app already ships, nothing more. It is NOT
// part of the app build — Vite never sees it.
//
// Generated/maintained by design-sync. Regenerate by adding new component
// files here and to `componentSrcMap` in .design-sync/config.json.

// shared primitives (shadcn/ui, new-york)
export * from './src/shared/components/ui/card';
export * from './src/shared/components/ui/chart';
export * from './src/shared/components/ui/tooltip';

// The recharts primitives ChartContainer is designed to wrap. Re-exported
// deliberately: ChartContainer takes a recharts element as its child, so
// without these the chart components are unusable by anything consuming this
// bundle — and a second copy of recharts does not render inside the
// ResponsiveContainer this one provides. `Tooltip` is intentionally NOT
// re-exported here: that name belongs to the shadcn tooltip above. Use
// `ChartTooltip` for charts.
export {
  Area,
  AreaChart,
  Bar,
  BarChart,
  CartesianGrid,
  Cell,
  Label,
  Line,
  LineChart,
  Pie,
  PieChart,
  PolarAngleAxis,
  PolarRadiusAxis,
  RadialBar,
  RadialBarChart,
  ReferenceLine,
  ResponsiveContainer,
  XAxis,
  YAxis,
} from 'recharts';

// agents feature
export * from './src/features/agents/components/AgentCard';
export * from './src/features/agents/components/EmptyState';
export * from './src/features/agents/components/RadialGauge';
export * from './src/features/agents/components/StatRow';
export * from './src/features/agents/components/metrics-card/MetricAreaChart';
export * from './src/features/agents/components/metrics-card/MetricsCard';

// jobs feature
export * from './src/features/jobs/components/DetailRow';
export * from './src/features/jobs/components/Field';
export * from './src/features/jobs/components/Input';
export * from './src/features/jobs/components/JobStateBadge';
export * from './src/features/jobs/components/LogViewer';
