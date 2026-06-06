---
id: 0001
title: Use Nx as Monorepo Build and Task Orchestration Tool
date: 2026-06-06
status: Accepted
tags: [architecture, tooling, monorepo]
---

# ADR 0001: Use Nx as Monorepo Build and Task Orchestration Tool

## 1. Context and Problem Statement

The repository hosts multiple apps (`apps/front`, `apps/hub`, `apps/agent`) and shared libraries under `libs/`. Without a task orchestration layer, each app would require its own ad-hoc build scripts and there is no mechanism for caching, dependency-aware execution order, or a consistent developer interface across apps that may use different runtimes (TypeScript, Go, etc.).

As the number of apps and libs grows, running full builds and tests on every change becomes slow and error-prone. We need a single tool that understands the dependency graph of the monorepo and can run only what is affected by a given change.

## 2. Considered Options

* **Option A: Nx** — task orchestration with caching, affected-graph execution, plugin ecosystem for TypeScript, Go, and custom executors.
* **Option B: Turborepo** — similar caching and pipeline features, but primarily focused on JavaScript/TypeScript; first-class Go support is limited.
* **Option C: Raw Makefiles + shell scripts** — no dependencies, but no caching, no dependency graph, and high maintenance burden as the repo grows.

## 3. Decision

We will use **Nx** as the single task orchestration and build tool for the monorepo.

**Justification:** Nx provides local and remote caching out of the box, understands the project dependency graph to run only affected tasks, and supports non-JS runtimes via custom executors and community plugins. The `apps/` and `libs/` directory structure we already adopted maps directly to Nx conventions, minimizing retrofit cost. Turborepo was ruled out due to weaker Go support; raw scripts were ruled out due to long-term maintenance burden.

## 4. Consequences

**Positive:**
* Consistent `nx run <project>:<target>` interface for all apps regardless of runtime.
* Task caching avoids redundant rebuilds locally and in CI.
* `nx affected` limits CI work to only changed projects.
* Shared library dependency graph is explicit and enforced.

**Negative / Trade-offs:**
* Requires Node.js and `node_modules` at the repo root even for pure Go apps.
* New contributors must learn basic Nx concepts (`project.json`, targets, executors).
* Go apps need custom executors or `nx:run-commands` wrappers — no official first-party Go plugin.

---

## 5. AI Directives (System Rules)

* **MUST:** Always invoke builds, tests, and type-checks via `nx run <project>:<target>` (e.g., `nx run agent:build`, `nx run front:typecheck`). Never call `go build`, `tsc`, or `npm run` directly.
* **MUST NOT:** Add per-app shell scripts that duplicate what an Nx target already covers.
* **MUST NOT:** Bypass Nx caching by adding `--skip-nx-cache` unless explicitly instructed by the user for a specific debugging session.
* **REFERENCE:** See `nx.json` and each app's `project.json` for the canonical list of available targets.
