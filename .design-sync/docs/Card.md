---
category: Primitives
---
Card — the surface every panel in Harmost sits on (shadcn/ui, new-york).

Composed from parts, not props: `Card` > `CardHeader` (`CardTitle`, `CardDescription`, `CardAction`) > `CardContent` > `CardFooter`. All are plain `div`s taking `className`, so layout is done with Tailwind utilities.

```tsx
<Card>
  <CardHeader>
    <CardTitle className="text-sm font-medium text-muted-foreground uppercase tracking-widest">
      System Metrics
    </CardTitle>
  </CardHeader>
  <CardContent className="space-y-3">…</CardContent>
  <CardFooter className="flex justify-center">…</CardFooter>
</Card>
```

`Card` already supplies `rounded-xl border bg-card py-6 text-card-foreground shadow-sm` — do not re-declare background or radius on it. `CardHeader`/`CardContent`/`CardFooter` bring their own `px-6`; add only vertical rhythm.
