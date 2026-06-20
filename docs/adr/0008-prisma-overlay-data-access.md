# 8. Prisma as the data-access layer for the PostgreSQL overlay

- Status: Accepted
- Date: 2026-06-20

## Context

The per-user overlay (ADR-0004) lives in PostgreSQL and owns accounts, auth
identities, revocable sessions (ADR-0006), subscriptions, categories, and reading
progress. The Nitro monolith (ADR-0005) needs a typed, migration-driven way to
talk to it that fits TypeScript `strict` and an ESM/Nitro bundle.

## Decision

Use **Prisma (v7)** as the ORM and migration tool for the overlay, with the new
**`prisma-client` generator** (ESM output, custom `output` path —
`prisma/generated`, gitignored). The schema lives in `prisma/schema.prisma`; a
singleton client is exposed from `server/utils/prisma.ts`.

Prisma 7 changes how the connection is configured:

- The datasource `url` is **no longer allowed in the schema**. The connection
  string lives in **`prisma.config.ts`** (`datasource.url = env('DATABASE_URL')`)
  for Migrate/CLI. Prisma 7 also stopped auto-loading `.env`, so the config loads
  it explicitly (`process.loadEnvFile()`, wrapped for CI/`generate`).
- The runtime client **requires a driver adapter**. We use **`@prisma/adapter-pg`**
  (with the `pg` driver), constructed from `DATABASE_URL` and passed to
  `new PrismaClient({ adapter })`.

## Alternatives considered

- **Drizzle ORM** — lighter and SQL-first, but Prisma's schema-driven migrations,
  mature tooling (Studio, migrate), and type ergonomics win for this scope.
- **Kysely / raw SQL** — maximum control, but more boilerplate and manual
  migration management than warranted here.
- **`prisma-client-js` (legacy generator)** — deprecated in Prisma 7 and not the
  recommended path for ESM bundlers; rejected.
- **`@prisma/adapter-pg` vs other adapters** — `pg` is the standard
  node-postgres driver and matches a self-hosted PostgreSQL; Accelerate/`prisma://`
  URLs are explicitly not used (PrismaPg expects a direct connection string).

## Consequences

- `DATABASE_URL` configures the connection (see `.env.example`), consumed by
  `prisma.config.ts` (CLI) and the `pg` adapter (runtime). Suwayomi may use the
  same PostgreSQL server in a separate database (PROJECT context / ADR-0004).
- The client is generated, not committed: `prisma generate` runs on `postinstall`
  and in CI before typecheck/build. `prisma/generated` is gitignored and excluded
  from ESLint and knip.
- Scripts: `db:generate`, `db:migrate` (dev), `db:deploy` (prod), `db:studio`.
- Migrations require a running PostgreSQL; the docker-compose dev environment is a
  prerequisite for the first migration.
- `tsconfig` must keep `moduleResolution: bundler` / ESM (already the Nuxt 4
  default) for the generated client to type-check.
