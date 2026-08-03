import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from 'front';

// Tooltips only appear on hover, which a static screenshot can't produce — so
// these force `open` to capture the resting open state. In real usage you omit
// `open` entirely and let hover/focus drive it.

/** The open state: trigger plus portalled content and arrow. */
export function Open() {
  return (
    <div style={{ paddingTop: 56, paddingBottom: 8, display: 'flex', justifyContent: 'center' }}>
      <TooltipProvider>
        <Tooltip open>
          <TooltipTrigger asChild>
            <span className="cursor-help text-xs text-muted-foreground">Updated 42 seconds ago</span>
          </TooltipTrigger>
          <TooltipContent side="bottom">
            <p>28 Jul 2026, 09:14:03</p>
          </TooltipContent>
        </Tooltip>
      </TooltipProvider>
    </div>
  );
}

/** Closed — only the trigger shows, which is the default resting state. */
export function Closed() {
  return (
    <div style={{ display: 'flex', justifyContent: 'center', padding: 8 }}>
      <TooltipProvider>
        <Tooltip>
          <TooltipTrigger asChild>
            <span className="cursor-help text-xs text-muted-foreground">Updated 42 seconds ago</span>
          </TooltipTrigger>
          <TooltipContent>
            <p>28 Jul 2026, 09:14:03</p>
          </TooltipContent>
        </Tooltip>
      </TooltipProvider>
    </div>
  );
}

/** How MetricsCard uses it: an exact timestamp behind a relative one. */
export function OnTimestamp() {
  return (
    <div style={{ paddingTop: 56, paddingBottom: 8, maxWidth: 420 }}>
      <div className="flex items-center justify-between">
        <span className="text-sm font-medium text-muted-foreground uppercase tracking-widest">
          System Metrics
        </span>
        <TooltipProvider>
          <Tooltip open>
            <TooltipTrigger asChild>
              <span className="cursor-help text-xs text-muted-foreground">Updated 42 seconds ago</span>
            </TooltipTrigger>
            <TooltipContent side="bottom">
              <p>28 Jul 2026, 09:14:03</p>
            </TooltipContent>
          </Tooltip>
        </TooltipProvider>
      </div>
    </div>
  );
}
