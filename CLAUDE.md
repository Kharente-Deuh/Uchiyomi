# CLAUDE.md — Working on Uchiyomi

This file is the operating guide for Claude (and any AI assistant) working on
Uchiyomi. It doubles as the basis for the Claude Project's custom instructions.
For full context, read [`PROJECT_BRIEF.md`](./PROJECT_BRIEF.md),
[`ENGINEERING.md`](./ENGINEERING.md), and the ADRs in
[`docs/adr/`](./docs/adr/).

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
  server-side **revocable** sessions. No HTTP Basic auth. (ADR-0006)

## Tech stack (authoritative)

Nuxt (Vue 3) + Nitro (single deployable) · PWA via `@vite-pwa/nuxt` · TypeScript
in `strict` mode · typed GraphQL client to Suwayomi · PostgreSQL · `nuxt-auth-utils`
+ `nuxt-authorization` · Vitest · `@nuxt/eslint`.

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
- Lint and type-check must pass before a change is considered done. Add or update
  tests (Vitest) for behavior changes.
- Keep docs in sync: update `PROJECT_BRIEF.md` / `ENGINEERING.md` / ADRs when a
  change affects them.

## Definition of done for a change

1. Code compiles, `strict` types pass.
2. `@nuxt/eslint` passes (no new warnings).
3. Tests added/updated and passing.
4. Conventional Commit message(s), in English.
5. Docs/ADRs updated if architecture, data model, or conventions changed.

## When unsure

Ask a focused question rather than guessing on architecture, data model, or
security. For security-sensitive code (auth, sessions, password handling, the
Suwayomi proxy), prefer well-reviewed libraries over bespoke implementations and
call out the threat model.
