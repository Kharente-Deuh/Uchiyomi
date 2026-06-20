# 15. Typed Suwayomi access via graphql-codegen and a committed SDL snapshot

- Status: Accepted
- Date: 2026-06-20

## Context

Uchiyomi talks to Suwayomi only through its GraphQL API (ADR-0001). We want
compile-time-typed operations without a running Suwayomi in CI, and a setup that
matches the existing "generate, don't commit, build in postinstall" pattern used
for the Prisma client (ADR-0008).

## Decision

- The Suwayomi GraphQL schema is captured as a **committed SDL snapshot**
  (`server/utils/suwayomi/schema.graphql`), refreshed with `pnpm suwayomi:schema`
  (introspects the live engine) — re-run on a Suwayomi upgrade.
- `graphql-codegen` (client-preset) generates typed document nodes from that
  snapshot plus our operation documents. Output
  (`server/utils/suwayomi/generated/`) is **gitignored** and built in `postinstall`
  alongside `prisma generate`.
- The shared client (`server/utils/suwayomi/`) executes `TypedDocumentNode`s over
  `graphql-request`, wrapped by a timeout+retry layer and a unified `SuwayomiError`.
- Operations, generated types, and mappers stay in each domain's `infrastructure/`
  layer (ADR-0013); application use cases depend on repository interfaces.

## Consequences

- `pnpm typecheck` catches operation/schema drift in CI with no running Suwayomi.
- The SDL snapshot can go stale vs a running Suwayomi; a future scheduled CI job
  may refresh-and-diff it.
- True request cancellation is not implemented (timeouts race, the socket lingers);
  acceptable for the BFF, revisit if needed.
