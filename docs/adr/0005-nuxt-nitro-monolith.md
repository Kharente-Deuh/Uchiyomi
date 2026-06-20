# 5. Backend as a Nuxt/Nitro monolith, not a separate NestJS service

- Status: Accepted
- Date: 2026-06-20

## Context

The app needs a backend for the BFF/proxy, authentication, and the per-user
overlay. The question was whether to put it in Nuxt's Nitro server or a dedicated
service (e.g. NestJS), with an initial worry about performance.

The workload is I/O-bound (proxying Suwayomi, serving/caching images, DB reads and
writes), not CPU-bound. Heavy background work (downloads, source updates) is
handled by Suwayomi, so the app has minimal background processing.

## Decision

Implement the backend as a **Nuxt/Nitro monolith** — a single deployable that also
serves the PWA. Do not add a separate NestJS service.

## Alternatives considered

- **Separate NestJS backend** — rejected: no performance benefit for an I/O-bound
  app (framework choice is not the bottleneck), and it adds a second deployable
  with no meaningful structural gain at this scope.

## Consequences

- One container for the app (plus Suwayomi and PostgreSQL) → simplest ops on a NAS.
- If the overlay's domain logic later grows substantially, a structured backend
  could be reconsidered — but that would be a maintainability decision, never a
  performance one (and would warrant a new ADR).
- A dedicated worker process (e.g. BullMQ or a Nitro task) can be added later for
  any bespoke background jobs without adopting NestJS.
