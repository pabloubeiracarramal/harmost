import {
  Card,
  CardAction,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
  StatRow,
} from 'front';

/** The full anatomy: header, title, description, action, content, footer. */
export function Anatomy() {
  return (
    <div style={{ maxWidth: 520 }}>
      <Card>
        <CardHeader>
          <CardTitle>build-runner-01</CardTitle>
          <CardDescription>ip-10-0-3-42.eu-west-1.compute.internal</CardDescription>
          <CardAction>
            <span className="text-xs text-muted-foreground">v0.4.1</span>
          </CardAction>
        </CardHeader>
        <CardContent>
          <StatRow label="Status" value="online" />
          <StatRow label="Running containers" value="3" />
        </CardContent>
        <CardFooter>
          <span className="text-sm text-muted-foreground">Last seen 42 seconds ago</span>
        </CardFooter>
      </Card>
    </div>
  );
}

/** The eyebrow-style header Harmost uses for its metric panels. */
export function PanelHeading() {
  return (
    <div style={{ maxWidth: 520 }}>
      <Card>
        <CardHeader>
          <CardTitle className="text-sm font-medium text-muted-foreground uppercase tracking-widest">
            System Metrics
          </CardTitle>
        </CardHeader>
        <CardContent>
          <StatRow label="CPU" value="34.6%" />
          <StatRow label="Memory" value="4.4 GB / 15.5 GB" />
          <StatRow label="Disk" value="110.0 GB / 460.4 GB" />
        </CardContent>
      </Card>
    </div>
  );
}

/** Header and content only — the footer is optional. */
export function HeaderAndContent() {
  return (
    <div style={{ maxWidth: 520 }}>
      <Card>
        <CardHeader>
          <CardTitle>Recent jobs</CardTitle>
          <CardDescription>Last 24 hours</CardDescription>
        </CardHeader>
        <CardContent>
          <StatRow label="Succeeded" value="128" />
          <StatRow label="Failed" value="3" />
        </CardContent>
      </Card>
    </div>
  );
}

/** Bare Card — the surface on its own, no parts. */
export function Surface() {
  return (
    <div style={{ maxWidth: 520 }}>
      <Card>
        <CardContent>
          <span className="text-sm text-card-foreground">
            Any content can sit directly in CardContent.
          </span>
        </CardContent>
      </Card>
    </div>
  );
}
