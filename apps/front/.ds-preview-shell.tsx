// Preview-only surface for design-sync cards — wired via `provider` in
// .design-sync/config.json. Not part of the app build.
//
// Two things every Harmost card needs that the preview harness doesn't give it:
//
// 1. The dark surface. index.html hardcodes <html class="dark"> and AppShell
//    establishes `bg-neutral-950 text-white`. Every component below features/
//    is authored against that surface (bg-neutral-900, text-white,
//    border-neutral-800) and reads as broken on the harness's white body.
//    The negative margin cancels the harness's 24px body padding so the dark
//    surface goes full-bleed instead of floating as a panel on white.
//
// 2. A router. AgentCard renders a TanStack Router <Link>, which throws
//    ("Cannot read properties of null (reading 'stores')") outside a
//    RouterProvider. RouterProvider renders the matched route rather than its
//    own children, so the card's tree is passed down through SlotContext and
//    re-emitted by the root route's component.

import { createContext, useContext, type ReactNode } from 'react';
import {
  createMemoryHistory,
  createRootRoute,
  createRoute,
  createRouter,
  RouterProvider,
} from '@tanstack/react-router';

// `dark` goes on <html>, not just the wrapper below: Radix renders tooltip and
// popover content into a portal on document.body, which sits OUTSIDE any
// wrapper div. Scoping the class to the wrapper left portalled content
// resolving :root (light) tokens — a `bg-foreground` pill came out near-black
// on a near-black page, i.e. invisible.
if (typeof document !== 'undefined') {
  document.documentElement.classList.add('dark');
}

const SlotContext = createContext<ReactNode>(null);

function RootSlot() {
  return <>{useContext(SlotContext)}</>;
}

const rootRoute = createRootRoute({ component: RootSlot });

// Registered so the destinations components actually link to resolve. The
// components render nothing — only the root slot is ever displayed.
const childRoutes = ['/', '/dashboard', '/jobs', '/jobs/$id', '/agents/$id', '/tokens'].map((path) =>
  createRoute({ getParentRoute: () => rootRoute, path, component: () => null }),
);

const router = createRouter({
  routeTree: rootRoute.addChildren(childRoutes),
  history: createMemoryHistory({ initialEntries: ['/'] }),
});

export function HarmostPreviewShell({ children }: { children: ReactNode }) {
  return (
    <div
      className="dark bg-neutral-950 text-white"
      style={{ margin: -24, padding: 24, minHeight: 'calc(100vh - 48px)' }}
    >
      <SlotContext.Provider value={children}>
        <RouterProvider router={router as never} />
      </SlotContext.Provider>
    </div>
  );
}
