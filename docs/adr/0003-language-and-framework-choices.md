---
id: "0003"
title: Language and Framework Choices
date: 2026-06-06
status: Accepted
tags: [architecture, golang, frontend, typescript, react]
---

# ADR 0003: Language and Framework Choices

## 1. Context and Problem Statement

Harmost is a polyglot monorepo. Before the codebase grows, we need an explicit record of which language/framework is canonical for each service tier and why — so that future contributors and AI agents do not introduce additional runtimes ad hoc.

There are three distinct service tiers:
1. **Frontend web UI** — needs a rich component model, type safety, and fast iteration
2. **Hub** — cloud-hosted SaaS backend; needs gRPC, PostgreSQL, and high concurrency
3. **Agent** — long-running system service on user machines; needs cross-platform packaging, minimal runtime footprint, and gRPC support

## 2. Considered Options

**Frontend:**
* Option A: React + TypeScript — large ecosystem, strong typing, widely understood, good gRPC-web/WebSocket library support
* Option B: Vue or Svelte — smaller learning surface; less ecosystem breadth for our UI component needs (shadcn/ui targets React)

**Hub and Agent:**
* Option A: Go — single static binary, excellent gRPC support (`google.golang.org/grpc`), low memory footprint, strong standard library for HTTP/WebSocket
* Option B: Node.js / TypeScript — shares language with frontend; weaker gRPC story; heavier runtime for a system-service agent
* Option C: Rust — ideal footprint for agent, but steep learning curve and slower iteration for hub business logic

## 3. Decision

- **Frontend**: React 19 + TypeScript 5, bundled with Vite
- **Hub**: Go
- **Agent**: Go

**Justification:** Go gives both hub and agent a single static binary, first-class gRPC support, and a small runtime — critical for an agent that users install on their own machines. Sharing Go between hub and agent means proto-generated code and utility packages can live in shared Go modules inside `libs/`. React + TypeScript is the natural choice given the selected UI component library (shadcn/ui) and routing/data-fetching libraries (TanStack Router, TanStack Query).

## 4. Consequences

**Positive:**
* Hub and agent share the same generated gRPC stubs — one proto compile step
* Single static binary for agent simplifies cross-platform packaging (Linux, macOS, Windows)
* TypeScript on the frontend catches API shape mismatches at compile time
* No runtime version management for end users installing the agent

**Negative / Trade-offs:**
* Node.js is required at the repo root even for pure Go development (Nx dependency)
* Go and TypeScript cannot natively share types; an API schema layer (e.g., OpenAPI or generated TS types from proto) is needed if strict front-hub type sharing is required
* Adding a new service in a different language requires a new ADR

---

## 5. AI Directives (System Rules)

* **MUST:** All backend services (hub, agent, and any future services) MUST be written in Go unless a new ADR explicitly supersedes this decision.
* **MUST:** The frontend MUST be written in TypeScript. Do not introduce plain JavaScript files in `apps/front/src/`.
* **MUST NOT:** Do not add a Python, Ruby, Node.js, or Rust service without first writing and accepting a new ADR.
* **MUST NOT:** Do not add a second frontend framework (Vue, Svelte, Angular) alongside React.
* **REFERENCE:** See `docs/adr/0002-inter-service-communication.md` for the protocol choices that depend on these language choices.
