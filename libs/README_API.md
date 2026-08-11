# harmost-api

OpenAPI 3.1 contract for **front ↔ hub**, mirroring the role `libs/harmost-proto` plays for
hub ↔ agent. `openapi.yaml` is the single source of truth for the REST surface (`/api/v1/*`,
`/auth/github*`) *and* the `/ws` event payloads — OpenAPI has no channel concept, so WS frames are
documented as `components/schemas` with no `paths` entry rather than as an operation.

See ADR 0010 (`docs/adr/0010-openapi-contract-for-front-hub.md`) for the full rationale and the
AI directives that govern this boundary.

## Layout

```
openapi.yaml         # the contract — the only file you hand-edit
oapi-codegen.yaml     # Go generation config (models only, see below)
redocly.yaml          # lint config for `harmost-api:lint`
gen/api.gen.go        # generated Go models — committed, never hand-edited
                       # (apps/front/src/shared/api/schema.d.ts is the TS half, generated
                       # into the front app rather than in this lib)
```

Generation is deliberately scoped to **models/types only** — no server or client stubs. The hub's
chi handlers and the front's `httpClient.ts`/TanStack Query hooks keep their current shape; only the
payload *shapes* are generated.

## Editing the contract

1. Edit `openapi.yaml`:
   - **New REST endpoint:** add a `paths` entry (request/response schemas, JWT bearer security like
     the existing routes) plus whatever new `components/schemas` it needs.
   - **New WS event:** don't add a `paths` entry. Instead add a payload schema (e.g.
     `JobStatusPayload`), an envelope schema wrapping it (`type` as a literal string enum,
     `agent_id`/`job_id`/`at`, `payload: {$ref: ...}`), and list that envelope in the `HubEvent`
     `oneOf` (bottom of the schemas section).
2. Regenerate:
   ```
   nx run harmost-api:generate
   ```
   Runs `oapi-codegen` (Go models → `gen/api.gen.go`) and `openapi-typescript` (TS types →
   `apps/front/src/shared/api/schema.d.ts`). Commit both alongside the spec change.
3. Lint and verify generation is in sync:
   ```
   nx run harmost-api:lint    # redocly lint on openapi.yaml
   nx run harmost-api:check   # regenerate + git diff --exit-code — catches "forgot to regenerate"
   ```
4. Wire it up by hand — generation doesn't touch routing or handlers:
   - Hub: add the handler/route in `apps/hub/internal/transport/httpapi/`, mapping through
     `convert.go` if the domain type differs from the generated one (the pattern used for
     `JobSpec`, which has three Go representations — proto, `domain.JobSpec`, `api.JobSpec` — because
     letting `domain` import the generated package would invert the transport → service → repository
     layering). For a new WS event, `events.Event` stays hand-written (`HubEvent`'s `oneOf` is
     excluded from Go generation — see `oapi-codegen.yaml`'s `exclude-schemas`) — construct/publish it
     manually, but the payload struct comes from `gen/api.gen.go`.
   - Front: each feature's `api/types.ts` re-exports the generated schema type — never a new
     hand-written interface. For a WS event, extend `apps/front/src/shared/ws/wsClient.ts`'s handling.
5. Final check:
   ```
   nx run harmost-api:check
   nx run hub:build
   nx run front:typecheck
   ```
   The last one is the only thing that catches TS-side drift.

## Rules

- Never hand-edit `gen/api.gen.go` or `apps/front/src/shared/api/schema.d.ts` — regenerate instead.
- Never reintroduce a hand-written wire type in `apps/hub/internal/transport/httpapi/` or
  `apps/front/src/features/*/api/types.ts` — those `types.ts` files exist only to re-export the
  generated schema.
- Never describe a hub ↔ agent message here — that stays in `libs/harmost-proto` (ADR 0002).
- `oapi-codegen` installs via a `go.mod` `tool` directive (see `oapi-codegen.yaml`'s header) — no
  global binary needed, unlike `harmost-proto`'s `buf`.
