---
id: "0002"
title: Inter-Service Communication Protocols
date: 2026-06-06
status: Accepted
tags: [architecture, golang, frontend, grpc, websocket]
---

# ADR 0002: Inter-Service Communication Protocols

## 1. Context and Problem Statement

Harmost has three services that must exchange data in real time: the React frontend (`front`), the Go hub (`hub`), and the Go agent (`agent`). The hub is cloud-hosted SaaS; agents run as system services on user machines and connect outbound to the hub.

Two distinct communication boundaries need a protocol decision:

1. **front ↔ hub** — The UI needs live updates: agent connection status, job progress, and streaming container logs. A pure request/response protocol (REST) would require polling and cannot push log chunks efficiently.

2. **hub ↔ agent** — The hub must dispatch job specs to agents and receive a continuous stream of log chunks and status events back. The connection must be long-lived, survive network hiccups via reconnect logic, and be initiated by the agent (since the agent is behind a user's firewall).

## 2. Considered Options

**front ↔ hub:**
* Option A: REST + polling — simple, but wastes bandwidth and adds latency to live log streaming
* Option B: Server-Sent Events (SSE) — efficient for one-way push; cannot send commands from frontend to hub without a separate REST call
* Option C: WebSocket — bidirectional, single persistent connection per client; handles both inbound commands (trigger job) and outbound events (log chunks, status) over one channel

**hub ↔ agent:**
* Option A: REST (hub polls agent) — requires agent to expose an inbound port, which breaks behind firewalls/NAT
* Option B: gRPC unary — agent calls hub for each event; no persistent stream; no server-push from hub to agent
* Option C: gRPC bidirectional streaming — agent opens one long-lived stream to hub; hub pushes job specs down; agent streams log chunks and status events up; agent reconnects with exponential back-off on disconnect

## 3. Decision

- **front ↔ hub**: WebSocket
- **hub ↔ agent**: gRPC bidirectional streaming

**Justification:** WebSocket gives the frontend a single persistent channel that can receive pushed log chunks and send commands (job trigger, cancel) without separate REST endpoints. gRPC bidirectional streaming lets agents connect outbound through firewalls, gives strongly-typed proto contracts, and supports efficient binary framing for log chunk payloads. The two protocols are consistent in spirit: both are long-lived, bidirectional, and initiated by the downstream party.

## 4. Consequences

**Positive:**
* Real-time log streaming end-to-end with no polling overhead
* Agent can sit behind NAT/firewalls — it always initiates the outbound gRPC connection
* gRPC proto files act as the canonical API contract between hub and agent
* WebSocket simplifies the frontend: one connection handles commands and events

**Negative / Trade-offs:**
* gRPC requires a `.proto` file and code generation step for both hub and agent
* WebSocket connections require careful lifecycle management on the hub (registry of active clients per job)
* Load balancers must support WebSocket and gRPC (HTTP/2) — standard on modern cloud providers but worth verifying

---

## 5. AI Directives (System Rules)

* **MUST:** All hub ↔ agent interactions MUST be defined in a `.proto` file and communicated over the existing gRPC bidirectional stream. Never add a REST or WebSocket endpoint between hub and agent.
* **MUST:** All front ↔ hub real-time interactions MUST use the WebSocket channel. Do not add a polling endpoint as a fallback.
* **MUST NOT:** Never have hub initiate an outbound TCP connection to an agent. The agent always connects to hub.
* **REFERENCE:** See `docs/architecture.md` for the full data flow diagram.
