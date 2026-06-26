# CLAUDE.md — Working on Uchiyomi

This file is the operating guide for Claude (and any AI assistant) working on
Uchiyomi. It doubles as the basis for the Claude Project's custom instructions.
For full context, read the ADRs in [`docs/adr/`](./docs/adr/). (`PROJECT_BRIEF.md`
and `ENGINEERING.md` are local-only working notes, kept out of the repo via
`.gitignore`; the committed source of truth is this file plus the ADRs.)

## What Uchiyomi is

A self-hosted manga & webtoon server with a modern PWA. It runs Suwayomi-Server
headless as its source/download engine (Tachiyomi/Mihon extensions) and adds the
layer Suwayomi lacks: multi-user accounts, SSO, per-user libraries, and per-user
reading history.

## Your role

Act as a senior engineer and collaborator on this codebase. Propose changes that
respect the architecture and conventions below. Be direct, flag trade-offs, and
prefer the simplest design that meets the goal. When you make a significant
architectural or cross-cutting decision, write (or update) an ADR.

## Architectural invariants — do not violate without a new ADR

- **Suwayomi is a headless engine accessed only through its API** (`/api/graphql`,
  with REST as a deprecating fallback). Never reimplement extensions or source
  parsers. (ADR-0001)
- **Reading targets downloaded content** served back through Suwayomi's API and
  proxied by the BFF. (ADR-0002)
- **Single unified application.** Do not introduce Komga or any library-only
  server as the user-facing layer. (ADR-0003)
- **One shared Suwayomi engine + a per-user overlay in PostgreSQL.** Per-user
  library and progress live in the overlay, never in Suwayomi. Never use
  Suwayomi's global `isRead`/`inLibrary` for per-user semantics. (ADR-0004)
- **Backend is the Nuxt/Nitro monolith** — no separate NestJS service. (ADR-0005)
- **Only the App is exposed.** Suwayomi and PostgreSQL stay internal. End users
  authenticate to the App, never to Suwayomi.
- **Auth:** local accounts + optional OIDC via `nuxt-auth-utils`, with
  server-side **revocable** sessions. No HTTP Basic auth. **No email is ever sent** —
  onboarding and password reset use single-use links handed over out-of-band; email
  verification is omitted. (ADR-0006)
- **Overlay data access is Prisma.** The schema in `prisma/schema.prisma` is the
  source of truth for overlay tables; the generated client is not committed.
  (ADR-0008)
- **UI is Vuetify** (`vuetify-nuxt-module`); no UnoCSS. Vuetify's theme owns
  light/dark. (ADR-0009)
- **i18n everywhere.** No hard-coded user-facing strings; all copy goes through
  `@nuxtjs/i18n` keys (English + French). (ADR-0011)
- **Catalogue is cached, not reparsed.** `series`/`chapter` overlay tables mirror
  Suwayomi metadata for *followed* series only; Suwayomi remains the source of truth.
  (ADR-0012)
- **Extensions are admin-managed globally.** An admin (`canManageExtensions`)
  installs/uninstalls; all users see the installed set. There is no per-user
  activation. NSFW sources are gated per-user by `allowNsfw` (admin-granted) AND
  `showNsfw` (self-serve preference). (ADR-0012, revised M4.1a)
- **All downloads go through the overlay queue.** Suwayomi auto-download is disabled;
  `download_job` is paced globally, new chapters outrank backfill. (ADR-0012)

## Tech stack (authoritative)

Nuxt 4 (Vue 3) + Nitro (single deployable) · **Vuetify** via `vuetify-nuxt-module`
· PWA via `@vite-pwa/nuxt` · i18n via `@nuxtjs/i18n` (en/fr) · TypeScript in
`strict` mode · typed GraphQL client to Suwayomi · **PostgreSQL via Prisma** ·
`nuxt-auth-utils` + `nuxt-authorization` · Vitest · `@antfu/eslint-config` · knip
· taze. pnpm 11 (catalog) · Node 26. (Scaffold/tooling: ADR-0010.)

## Hard conventions

- **Language: English everywhere** — code, identifiers, comments, commit
  messages, docs, issues, PR/MR titles and descriptions. (No exceptions in the
  repo. Discussion outside the repo may be in any language.)
- **Conventional Commits** for every commit (`feat`, `fix`, `chore`, `docs`,
  `refactor`, `test`, `perf`, `ci`, `build`, `style`, `revert`). Breaking changes
  use `!` and/or a `BREAKING CHANGE:` footer.
- **SemVer**, releases automated from commits (release-please).
- **License: AGPL-3.0-or-later.** Add `SPDX-License-Identifier: AGPL-3.0-or-later`
  to source files.
- **Component styles are SCSS** (`<style lang="scss" scoped>`), compiled at build by
  `sass-embedded` — no runtime cost. Use SCSS for structure/layout only; **colours and
  light/dark stay owned by the Vuetify theme** (theme variables, not hard-coded
  palettes). (ADR-0009)
- **i18n in templates uses the global `$t`**, never a `t` destructured from
  `useI18n()`. In `<script setup>` (where `$t` is unavailable), use `useI18n().t`.
  Still no hard-coded user-facing strings. (ADR-0011)
- **Push filtering, sorting, pagination and counting down to the store.** Express
  them as Prisma `where`/`orderBy`/`take`/`skip`/`count` or as Suwayomi GraphQL
  query arguments (`filter`, `condition`, `order`, `first`, `offset`) — never fetch
  the full set and `.filter()`/`.sort()`/`.slice()` it in memory. Keep **policy** in
  the use case (it decides *what* to filter from `isAdmin`, `viewerCanSeeNsfw`, …)
  but translate that policy into query criteria the repository/adapter executes; the
  repo executes, it does not own the rule. Cross-store gates (an overlay flag like
  `Source.isEnabled` combined with a Suwayomi-side property) join by passing the
  overlay-derived id set into the GraphQL `id` filter, not by post-filtering. The
  only acceptable in-memory pass is a pure transform over an already-small, bounded
  result; if it is a security/visibility gate, push it to the query so it cannot be
  forgotten or bypassed.
- Lint and type-check must pass before a change is considered done. Add or update
  tests (Vitest) for behavior changes.
- Keep docs in sync: update this `CLAUDE.md` and/or add an ADR when a change
  affects architecture, the data model, conventions, or the tech stack.
- **Keep the tracker in sync with the specs.** Work is tracked as GitHub
  milestones (one per phase) and issues (one per `Mx.y`), mirroring
  [`docs/superpowers/roadmap.md`](./docs/superpowers/roadmap.md). Whenever a spec or
  the roadmap changes mid-flight — scope added, split, dropped, re-sequenced, or a
  milestone discovered already done — **create, edit, or close the matching
  issue/milestone in the same change** so the tracker never drifts from reality.
  Close issues that are done (link the PR), edit those whose scope moved, and open
  new ones for newly-carved work; update the roadmap statuses alongside.

## Project structure

**Frontend (`app/`) — modified atomic design.** Presentation components are pure and
reusable with **no business logic**; domain logic lives under `features/`. (ADR-0013)

```
app/
├── components/            # Pure, reusable — NO business logic
│   ├── Atom/              # Buttons, inputs, chips
│   ├── Molecule/          # Search bars, filter groups
│   └── Organism/          # Tables, modal wrappers
├── features/[feature]/    # Business logic per domain
│   ├── components/        # Feature-specific components
│   ├── composables/       # Feature-specific composables
│   ├── modals/            # Feature-specific modals
│   └── constants/
├── pages/                 # File-based routing (Nuxt)
├── layouts/               # e.g. DefaultLayout.vue
├── store/                 # Pinia stores, by domain
├── services/              # API layer, by domain
├── composables/           # Global composables (useModal, useTheme…)
├── types/                 # Global TypeScript types
└── utils/                 # Pure utilities (date, string, array…)
```

**Server (`server/`) — hexagonal DDD.** The Nitro backend (ADR-0005) is organised
by domain with a ports-and-adapters layering (domain / application /
infrastructure); Prisma and the Suwayomi client are infrastructure concerns. Full
per-layer idioms are in ADR-0013; the key wiring pattern is summarised below.

### Use-case classes + service factory (ADR-0013)

- **Use cases** are classes implementing `IUseCase<Opts, Result>` (interface in
  `server/shared/use-case.ts`), each in its own `application/usecases/<name>.use-case.ts`
  module. The single public method is `execute(opts)`. Dependencies (repositories,
  ports, hashers, policy primitives such as TTL or a `now: () => Date` clock) are
  injected via the constructor as `private readonly` fields. Naming: class `XUseCase`,
  opts type `XUseCaseOpts` (or `…Params`), result type `XUseCaseResult`.
- **Barrel** `application/usecases/index.ts` re-exports every use-case module
  (`export * from './x.use-case'`).
- **Service factory** `application/<domain>.service.ts` (one per domain): reads
  `useRuntimeConfig()`, instantiates all infrastructure adapters at module scope, and
  exposes the use cases through an exported `XService` interface plus an
  `xService(): XService` factory whose methods are one-per-use-case thin wrappers
  (each news-up its use case and calls `.execute`). This service interface is the
  domain's only application-layer surface.
- **Routes** (`server/api/**`), middleware and plugins import the factory from
  `~~/server/domains/<domain>/application/<domain>.service` and call
  `xService().method(...)`. They never import use-case classes nor call `.execute`,
  and do not wire dependencies themselves.
- **Shared infrastructure singletons** reused across domains (e.g. `userRepository`,
  `sessionRepository`) live in a dependency-free `application/index.ts` so service
  factories can share them without importing one another; stateful infra local to a
  single domain (e.g. `loginRateLimiter`) is exported from that domain's service
  module. The old `server/utils/<domain>.ts` wiring files were removed;
  `server/utils/` now holds **pure helpers only** (e.g. `account-name.ts`,
  `prisma.ts`).
- **No success envelope.** Routes never return an acknowledgment wrapper such as
  `{ ok: true }`. Return the **concerned resource** (its DTO, e.g.
  `return toExtensionDto(ext)`) or **nothing** — `Promise<void>`, i.e. an empty 2xx —
  when there is no resource to hand back. Success is conveyed by the HTTP status and
  handled on the client; failures are raised with `createError`.
- **Presenters** are pure functions named `toXDto(...)` under
  `infrastructure/transport/http/`, typed against domain models.
- **`shared/` import rule in node-tested code:** use a **relative path** (not `~~` or
  `#shared`) when importing from `server/shared/` in files tested by the Vitest `node`
  project — the `node` project has no path aliases. See the comment in
  `server/utils/account-name.ts` for the canonical example.

## Definition of done for a change

1. Code compiles, `strict` types pass (`pnpm typecheck`).
2. ESLint (`@antfu/eslint-config`) passes with no new warnings (`pnpm lint`).
3. `pnpm knip` is clean (no unused files/deps/exports).
4. Tests added/updated and passing (`pnpm test`).
5. Conventional Commit message(s), in English.
6. Docs/ADRs updated if architecture, data model, or conventions changed.

## When unsure

Ask a focused question rather than guessing on architecture, data model, or
security. For security-sensitive code (auth, sessions, password handling, the
Suwayomi proxy), prefer well-reviewed libraries over bespoke implementations and
call out the threat model.
