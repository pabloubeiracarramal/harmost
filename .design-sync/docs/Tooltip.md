---
category: Primitives
---
Tooltip — radix-backed hover/focus tooltip (shadcn/ui, new-york).

Four parts: `TooltipProvider` (required ancestor) > `Tooltip` > `TooltipTrigger` + `TooltipContent`. Use `asChild` on the trigger to attach to your own element rather than injecting a button.

```tsx
<TooltipProvider>
  <Tooltip>
    <TooltipTrigger asChild>
      <span className="cursor-help text-xs text-muted-foreground">Updated 42 seconds ago</span>
    </TooltipTrigger>
    <TooltipContent><p>28 Jul 2026, 09:14:03</p></TooltipContent>
  </Tooltip>
</TooltipProvider>
```

`delayDuration` defaults to 0 (instant). Content renders in a portal with an arrow, styled `bg-foreground text-background` — inverted against the surface on purpose.
